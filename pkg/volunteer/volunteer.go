/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package volunteer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kubernetes-incubator/spartakus/pkg/database"
	"github.com/kubernetes-incubator/spartakus/pkg/report"
	"github.com/kubernetes-incubator/spartakus/pkg/version"
	"github.com/thockin/logr"
)

func New(log logr.Logger, clusterID string, period time.Duration, db database.Database, extensionsPath string) (*volunteer, error) {
	kcw, err := newKubeClientWrapper()
	if err != nil {
		return nil, err
	}
	pel := pathExtensionsLister(extensionsPath)
	return newVolunteer(log, clusterID, period, db, kcw, kcw, pel), nil
}

func newVolunteer(
	log logr.Logger,
	clusterID string,
	period time.Duration,
	db database.Database,
	nodeLister nodeLister,
	serverVersioner serverVersioner,
	extensionsLister extensionsLister) *volunteer {

	return &volunteer{
		log:              log,
		clusterID:        clusterID,
		period:           period,
		database:         db,
		nodeLister:       nodeLister,
		serverVersioner:  serverVersioner,
		extensionsLister: extensionsLister,
	}
}

type volunteer struct {
	clusterID        string
	period           time.Duration
	database         database.Database
	log              logr.Logger
	nodeLister       nodeLister
	serverVersioner  serverVersioner
	extensionsLister extensionsLister
}

func (v *volunteer) Run() error {
	v.log.V(0).Infof("started volunteer")
	for {
		if err := v.runOnce(); err != nil {
			v.log.Errorf("%v", err)
		} else {
			v.log.V(0).Infof("report successfully sent")
		}
		if v.period == 0 {
			return nil
		}
		v.log.V(0).Infof("next attempt in %v", v.period)
		<-time.After(v.period)
	}
	// This can never be reached, and `go vet` complains if code is here.
}

func (v *volunteer) runOnce() error {
	rec, err := v.generateRecord()
	if err != nil {
		return fmt.Errorf("failed generating report: %v", err)
	}

	if err = v.send(rec); err != nil {
		return fmt.Errorf("failed sending report: %v", err)
	}

	return nil
}

func (v *volunteer) generateRecord() (report.Record, error) {
	svrVer, err := v.serverVersioner.ServerVersion()
	if err != nil {
		return report.Record{}, err
	}

	nodes, err := v.nodeLister.ListNodes()
	if err != nil {
		return report.Record{}, err
	}

	extensions, err := v.extensionsLister.ListExtensions()
	if err != nil {
		v.log.Errorf("failed to list extensions: %v", err)
		extensions = []report.Extension{}
	}

	rec := report.Record{
		Version:       version.VERSION,
		Timestamp:     strconv.FormatInt(time.Now().Unix(), 10),
		ClusterID:     v.clusterID,
		MasterVersion: &svrVer,
		Nodes:         nodes,
		Extensions:    extensions,
	}

	return rec, nil
}

func (v *volunteer) send(rec report.Record) error {
	return v.database.Store(rec)
}

type extensionsLister interface {
	// ListExtensions returns a slice of report.Extensions, since that is the
	// schema of the extensions in the database. Returning the array is nicer
	// than returning a map since it means we don't have to encode any logic
	// about transforming a map of extensions into an array of extensions in
	// other parts of the the package. The format of the extensions file
	// different from the database schema to be less verbose: writing
	// {"k1": "v1", "k2": "v2"} is easier than [{"name": "k1", "value": "v1"}...
	ListExtensions() ([]report.Extension, error)
}

// pathExtensionsLister is higher level implementation of the extensionsLister
// interface that reads extensions from a path, be it a directory or a file.
type pathExtensionsLister string

// ListExtensions returns a slice of report.Extensions containing the
// custom extensions that the user may want to report.
func (p pathExtensionsLister) ListExtensions() ([]report.Extension, error) {
	var extensions []report.Extension

	if p == "" {
		return extensions, nil
	}

	f, err := os.Stat(string(p))
	if err != nil {
		return nil, fmt.Errorf("failed to stat extensions path: %v", err)
	}

	var paths []string
	if f.IsDir() {
		fis, err := ioutil.ReadDir(string(p))
		if err != nil {
			return nil, fmt.Errorf("failed to open extensions directory: %v", err)
		}

		for _, fi := range fis {
			// Ignore directories and files with a leading `.`.
			if !fi.IsDir() && !strings.HasPrefix(fi.Name(), ".") {
				paths = append(paths, filepath.Join(string(p), fi.Name()))
			}
		}
	} else {
		paths = append(paths, string(p))
	}

	for _, path := range paths {
		extensionsBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read extensions file: %v", err)
		}

		es, err := byteExtensionsLister(extensionsBytes).ListExtensions()
		if err == nil {
			extensions = append(extensions, es...)
		}
	}

	return extensions, nil
}

// byteExtensionsLister is a basic implementation of the extensionsLister
// interface that reads extensions from a byte array.
type byteExtensionsLister []byte

// ListExtensions returns a slice of report.Extensions containing the
// custom extensions that the user may want to report.
func (b byteExtensionsLister) ListExtensions() ([]report.Extension, error) {
	var extensions []report.Extension

	if len(b) == 0 {
		return extensions, nil
	}

	extensionsMap := make(map[string]string)
	err := json.Unmarshal(b, &extensionsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extensions data: %v", err)
	}

	for k, v := range extensionsMap {
		extensions = append(extensions, report.Extension{Name: k, Value: v})
	}

	return extensions, nil
}
