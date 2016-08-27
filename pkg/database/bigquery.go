package database

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/prometheus/common/log"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	bigquery "google.golang.org/api/bigquery/v2"
	"k8s.io/spartakus/pkg/collector"
	"k8s.io/spartakus/pkg/logr"
	"k8s.io/spartakus/pkg/report"
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
	if len(subs) != 3 {
		return true, "", "", "", fmt.Errorf("invalid bigquery spec: %q", dbspec)
	}
	return true, subs[0], subs[1], subs[2], nil
}

func newBigqueryDatabase(log logr.Logger, project, dataset, table string) (collector.Database, error) {
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
		bq:      bq,
		project: project,
		dataset: dataset,
		table:   table,
	}, nil
}

type bqDatabase struct {
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
		log.Errorf("TIM: err %v", err)
	} else {
		spew.Dump(resp)
	}
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
