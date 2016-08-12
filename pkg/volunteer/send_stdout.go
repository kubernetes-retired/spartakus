package volunteer

import (
	"encoding/json"
	"fmt"

	"k8s.io/spartakus/pkg/report"
)

type stdoutRecordSender struct{}

func (repo stdoutRecordSender) Send(rec report.Record) error {
	j, err := json.MarshalIndent(rec, "", "    ")
	if err != nil {
		return err
	}
	fmt.Println(string(j))
	return nil
}

func newStdoutRecordSender() (stdoutRecordSender, error) {
	return stdoutRecordSender{}, nil
}
