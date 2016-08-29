package volunteer

import (
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	kresource "k8s.io/client-go/1.4/pkg/api/resource"
	kv1 "k8s.io/client-go/1.4/pkg/api/v1"
	"k8s.io/spartakus/pkg/report"
)

func TestNodeFromKubeNode(t *testing.T) {
	testCases := []struct {
		input  kv1.Node
		expect report.Node
	}{
		{
			input: kv1.Node{
				ObjectMeta: kv1.ObjectMeta{
					Name: "kname",
				},
			},
			expect: report.Node{},
		},
		{
			input: kv1.Node{
				ObjectMeta: kv1.ObjectMeta{
					Name: "kname",
				},
				Status: kv1.NodeStatus{
					Capacity: kv1.ResourceList{
						// unsorted
						"r2": kresource.MustParse("200"),
						"r3": kresource.MustParse("300"),
						"r1": kresource.MustParse("100"),
					},
				},
			},
			expect: report.Node{
				Capacity: []report.Resource{
					{Resource: "r1", Value: "100"},
					{Resource: "r2", Value: "200"},
					{Resource: "r3", Value: "300"},
				},
			},
		},
		{
			input: kv1.Node{
				ObjectMeta: kv1.ObjectMeta{
					Name: "kname",
				},
				Status: kv1.NodeStatus{
					NodeInfo: kv1.NodeSystemInfo{
						OperatingSystem:         "os",
						OSImage:                 "image",
						KernelVersion:           "kernel",
						Architecture:            "architecture",
						ContainerRuntimeVersion: "runtime",
						KubeletVersion:          "kubelet",
					},
				},
			},
			expect: report.Node{
				OperatingSystem:         strPtr("os"),
				OSImage:                 strPtr("image"),
				KernelVersion:           strPtr("kernel"),
				Architecture:            strPtr("architecture"),
				ContainerRuntimeVersion: strPtr("runtime"),
				KubeletVersion:          strPtr("kubelet"),
			},
		},
	}

	for i, tc := range testCases {
		n := nodeFromKubeNode(&tc.input)
		if n.ID == "" || n.ID == tc.input.Name {
			t.Errorf("[%d] expected anonymized ID, got %q", i, n.ID)
		}
		n.ID = ""
		if !reflect.DeepEqual(n, tc.expect) {
			t.Errorf("[%d]: did not get expected result:\n%s", i, pretty.Compare(n, tc.expect))
		}
	}
}
