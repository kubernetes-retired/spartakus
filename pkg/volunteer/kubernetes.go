package volunteer

import (
	"crypto/md5"
	"encoding/hex"

	kclient "k8s.io/client-go/1.4/kubernetes"
	kapi "k8s.io/client-go/1.4/pkg/api"
	kv1 "k8s.io/client-go/1.4/pkg/api/v1"
	kfields "k8s.io/client-go/1.4/pkg/fields"
	klabels "k8s.io/client-go/1.4/pkg/labels"
	"k8s.io/spartakus/pkg/report"
)

type nodeLister interface {
	ListNodes() ([]report.Node, error)
}

type serverVersioner interface {
	ServerVersion() (string, error)
}

func nodeFromKubernetesAPINode(kn kv1.Node) report.Node {
	n := report.Node{
		ID:                      getID(kn),
		OSImage:                 strPtr(kn.Status.NodeInfo.OSImage),
		KernelVersion:           strPtr(kn.Status.NodeInfo.KernelVersion),
		ContainerRuntimeVersion: strPtr(kn.Status.NodeInfo.ContainerRuntimeVersion),
		KubeletVersion:          strPtr(kn.Status.NodeInfo.KubeletVersion),
	}
	for k, v := range kn.Status.Capacity {
		n.Capacity = append(n.Capacity, report.Resource{
			Resource: string(k),
			Value:    v.String(),
		})
	}
	return n
}

func getID(kn kv1.Node) string {
	// We don't want to report the node's Name - that is PII.  The MachineID is
	// apparently not always populated and SystemUUID is ill-defined.  Let's
	// just hash them all together.  It should be stable, and this reduces risk
	// of PII leakage.
	return hashOf(kn.Name + kn.Status.NodeInfo.MachineID + kn.Status.NodeInfo.SystemUUID)
}

func hashOf(str string) string {
	hasher := md5.New()
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil)[0:])
}

func strPtr(str string) *string {
	p := new(string)
	*p = str
	return p
}

type kubernetesClientWrapper struct {
	client *kclient.Clientset
}

func (k *kubernetesClientWrapper) ListNodes() ([]report.Node, error) {
	knl, err := k.client.Core().Nodes().List(kapi.ListOptions{
		LabelSelector: klabels.Everything(),
		FieldSelector: kfields.Everything(),
	})
	if err != nil {
		return nil, err
	}
	nodes := make([]report.Node, len(knl.Items))
	for i, kn := range knl.Items {
		nodes[i] = nodeFromKubernetesAPINode(kn)
	}
	return nodes, nil
}

func (k *kubernetesClientWrapper) ServerVersion() (string, error) {
	i, err := k.client.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return i.String(), nil
}
