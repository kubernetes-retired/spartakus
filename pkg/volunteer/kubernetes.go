package volunteer

import (
	"crypto/md5"
	"encoding/hex"

	kapi "k8s.io/kubernetes/pkg/api"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kfields "k8s.io/kubernetes/pkg/fields"
	klabels "k8s.io/kubernetes/pkg/labels"
	"k8s.io/spartakus/pkg/report"
)

type nodeLister interface {
	List() ([]report.Node, error)
}

type serverVersioner interface {
	ServerVersion() (string, error)
}

func nodeFromKubernetesAPINode(kn kapi.Node) report.Node {
	n := report.Node{
		ID:                      hashOf(kn.Name),
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
	client *kclient.Client
}

func (k *kubernetesClientWrapper) List() ([]report.Node, error) {
	knl, err := k.client.Nodes().List(kapi.ListOptions{
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
