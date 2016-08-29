package volunteer

import (
	"crypto/md5"
	"encoding/hex"
	"sort"

	kclient "k8s.io/client-go/1.4/kubernetes"
	kapi "k8s.io/client-go/1.4/pkg/api"
	kv1 "k8s.io/client-go/1.4/pkg/api/v1"
	kfields "k8s.io/client-go/1.4/pkg/fields"
	klabels "k8s.io/client-go/1.4/pkg/labels"
	krest "k8s.io/client-go/1.4/rest"

	"k8s.io/spartakus/pkg/report"
)

type nodeLister interface {
	ListNodes() ([]report.Node, error)
}

type serverVersioner interface {
	ServerVersion() (string, error)
}

func nodeFromKubeNode(kn *kv1.Node) report.Node {
	n := report.Node{
		ID:                      getID(kn),
		OperatingSystem:         strPtr(kn.Status.NodeInfo.OperatingSystem),
		OSImage:                 strPtr(kn.Status.NodeInfo.OSImage),
		KernelVersion:           strPtr(kn.Status.NodeInfo.KernelVersion),
		Architecture:            strPtr(kn.Status.NodeInfo.Architecture),
		ContainerRuntimeVersion: strPtr(kn.Status.NodeInfo.ContainerRuntimeVersion),
		KubeletVersion:          strPtr(kn.Status.NodeInfo.KubeletVersion),
	}
	// We want to iterate the resources in a deterministic order.
	keys := []string{}
	for k, _ := range kn.Status.Capacity {
		keys = append(keys, string(k))
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := kn.Status.Capacity[kv1.ResourceName(k)]
		n.Capacity = append(n.Capacity, report.Resource{
			Resource: string(k),
			Value:    v.String(),
		})
	}
	return n
}

func getID(kn *kv1.Node) string {
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
	if str == "" {
		return nil
	}
	p := new(string)
	*p = str
	return p
}

func newKubeClientWrapper() (*kubeClientWrapper, error) {
	kubeConfig, err := krest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	kubeClient, err := kclient.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	return &kubeClientWrapper{client: kubeClient}, nil
}

type kubeClientWrapper struct {
	client *kclient.Clientset
}

func (k *kubeClientWrapper) ListNodes() ([]report.Node, error) {
	knl, err := k.client.Core().Nodes().List(kapi.ListOptions{
		LabelSelector: klabels.Everything(),
		FieldSelector: kfields.Everything(),
	})
	if err != nil {
		return nil, err
	}
	nodes := make([]report.Node, len(knl.Items))
	for i := range knl.Items {
		kn := &knl.Items[i]
		nodes[i] = nodeFromKubeNode(kn)
	}
	return nodes, nil
}

func (k *kubeClientWrapper) ServerVersion() (string, error) {
	info, err := k.client.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return info.String(), nil
}
