package database

import (
	"encoding/json"
	"fmt"

	"github.com/thockin/logr"
	"k8s.io/spartakus/pkg/report"
)

func init() {
	registerPlugin("stdout", stdoutPlugin{})
}

type stdoutPlugin struct{}

func (plug stdoutPlugin) Attempt(_ logr.Logger, dbspec string) (bool, Database, error) {
	if dbspec != "stdout" {
		return false, nil, nil
	}
	return true, stdoutDatabase{}, nil
}

func (plug stdoutPlugin) ExampleSpec() string {
	return "stdout"
}

type stdoutDatabase struct{}

func (db stdoutDatabase) Store(rec report.Record) error {
	j, err := json.MarshalIndent(rec, "", "    ")
	if err != nil {
		return err
	}
	fmt.Println(string(j))
	return nil
}
