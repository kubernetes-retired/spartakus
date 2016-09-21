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

	"github.com/kubernetes-incubator/spartakus/pkg/report"
	"github.com/thockin/logr"
)

type Database interface {
	Store(report.Record) error
}

func NewDatabase(log logr.Logger, dbspec string) (Database, error) {
	for name, plug := range plugins {
		is, db, err := plug.Attempt(log, dbspec)
		if !is {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to initialize %q: %v", name, err)
		}
		if db == nil {
			return nil, fmt.Errorf("no error, but result was nil: %q", name)
		}
		log.V(0).Infof("using %q database", name)
		return db, nil
	}

	return nil, fmt.Errorf("unknown spec: %q", dbspec)
}
