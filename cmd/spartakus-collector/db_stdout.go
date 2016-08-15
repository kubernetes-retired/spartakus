package main

import (
	"encoding/json"
	"fmt"

	"k8s.io/spartakus/pkg/report"
)

type stdoutDatabase struct{}

func (db stdoutDatabase) Store(rec report.Record) error {
	j, err := json.MarshalIndent(rec, "", "    ")
	if err != nil {
		return err
	}
	fmt.Println(string(j))
	return nil
}
