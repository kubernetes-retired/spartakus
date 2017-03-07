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

package database

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/kubernetes-incubator/spartakus/pkg/report"
	"github.com/thockin/logr"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	bigquery "google.golang.org/api/bigquery/v2"
)

func init() {
	registerPlugin("bigquery", bigqueryPlugin{})
}

type bigqueryPlugin struct{}

func (plug bigqueryPlugin) Attempt(log logr.Logger, dbspec string) (bool, Database, error) {
	isMine, project, dataset, table, err := parseBigquerySpec(dbspec)
	if !isMine {
		return false, nil, nil
	}
	if err != nil {
		return true, nil, err
	}
	db, err := newBigqueryDatabase(log, project, dataset, table)
	return true, db, err
}

func (plug bigqueryPlugin) ExampleSpec() string {
	return "bigquery://project.dataset.table"
}

var bqre = regexp.MustCompile(`^bigquery://([^.]+)\.([^.]+)\.([^.]+)$`)

// Parse a bigquery spec, formatted as `bigquery://project.dataset.table`.  If
// the argument appears to be a bigquery spec, the first return (bool) will be
// true. The 3 string returns are project, dataset, and table respectively.
// This will only return an error if it believes the argument is a biquery spec,
// but it can't parse properly.
func parseBigquerySpec(dbspec string) (bool, string, string, string, error) {
	if !strings.HasPrefix(dbspec, "bigquery://") {
		return false, "", "", "", nil
	}
	subs := bqre.FindStringSubmatch(dbspec)
	if len(subs) != 4 {
		return true, "", "", "", fmt.Errorf("invalid bigquery spec: %q", dbspec)
	}
	return true, subs[1], subs[2], subs[3], nil
}

func newBigqueryDatabase(log logr.Logger, project, dataset, table string) (Database, error) {
	ctx := context.Background()

	// This assumes that:
	//  a) a GCE ServiceAccount has been created for this app
	//  b) the ServiceAccount is an owner of the dataset for this app
	//  c) the credentials for the ServiceAccount are in a file
	//  d) env GOOGLE_APPLICATION_CREDENTIALS=<path to credentials file>
	ts, err := google.DefaultTokenSource(ctx, bigquery.BigqueryScope)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth2 token: %v", err)
	}
	oauthClient := oauth2.NewClient(ctx, ts)

	bq, err := bigquery.New(oauthClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create Bigquery client: %v", err)
	}
	return bqDatabase{
		log:     log,
		bq:      bq,
		project: project,
		dataset: dataset,
		table:   table,
	}, nil
}

type bqDatabase struct {
	log     logr.Logger
	bq      *bigquery.Service
	project string
	dataset string
	table   string
}

func (bqdb bqDatabase) Store(rec report.Record) error {
	tds := bqdb.bq.Tabledata
	req := &bigquery.TableDataInsertAllRequest{
		Rows: []*bigquery.TableDataInsertAllRequestRows{
			&bigquery.TableDataInsertAllRequestRows{
				Json: makeRow(rec),
			},
		},
	}
	call := tds.InsertAll(bqdb.project, bqdb.dataset, bqdb.table, req)
	resp, err := call.Do()
	if err != nil {
		return err
	}
	bqdb.log.V(9).Infof("bigquery response: %s", spew.Sprintf("%#v", resp))

	return nil
}

func makeRow(rec report.Record) map[string]bigquery.JsonValue {
	row := map[string]bigquery.JsonValue{
		"version":       rec.Version,
		"timestamp":     rec.Timestamp,
		"clusterID":     rec.ClusterID,
		"masterVersion": rec.MasterVersion,
	}
	nodes := []map[string]bigquery.JsonValue{}
	for _, n := range rec.Nodes {
		nodes = append(nodes, makeNode(n))
	}
	row["nodes"] = nodes
	extensions := []map[string]bigquery.JsonValue{}
	for _, e := range rec.Extensions {
		extensions = append(extensions, makeExtension(e))
	}
	row["extensions"] = extensions
	return row
}

func makeNode(node report.Node) map[string]bigquery.JsonValue {
	n := map[string]bigquery.JsonValue{
		"id":                      node.ID,
		"operatingSystem":         node.OperatingSystem,
		"osImage":                 node.OSImage,
		"kernelVersion":           node.KernelVersion,
		"architecture":            node.Architecture,
		"containerRuntimeVersion": node.ContainerRuntimeVersion,
		"kubeletVersion":          node.KubeletVersion,
		"cloudProvider":           node.CloudProvider,
	}
	capacity := []map[string]bigquery.JsonValue{}
	for _, c := range node.Capacity {
		capacity = append(capacity, makeResource(c))
	}
	n["capacity"] = capacity
	return n
}

func makeResource(res report.Resource) map[string]bigquery.JsonValue {
	r := map[string]bigquery.JsonValue{
		"resource": res.Resource,
		"value":    res.Value,
	}
	return r
}

func makeExtension(ext report.Extension) map[string]bigquery.JsonValue {
	e := map[string]bigquery.JsonValue{
		"name":  ext.Name,
		"value": ext.Value,
	}
	return e
}
