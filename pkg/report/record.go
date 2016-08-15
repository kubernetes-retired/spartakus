package report

type Record struct {
	Version       string  `json:"version"`   // required
	Timestamp     string  `json:"timestamp"` // provided by server, client values are ignored
	ClusterID     string  `json:"clusterID"` // required
	MasterVersion *string `json:"masterVersion,omitempty"`
	Nodes         []Node  `json:"nodes,omitempty"`
}

type Node struct {
	//FIXME: decide if ID is MachineID or SystemUUID
	ID                      string     `json:"id"` // required
	OperatingSystem         *string    `json:"operatingSystem,omitempty"`
	OSImage                 *string    `json:"osImage,omitempty"`
	KernelVersion           *string    `json:"kernelVersion,omitempty"`
	Architecture            *string    `json:"architecture,omitempty"`
	ContainerRuntimeVersion *string    `json:"containerRuntimeVersion,omitempty"`
	KubeletVersion          *string    `json:"kubeletVersion,omitempty"`
	Capacity                []Resource `json:"capacity,omitempty"`
}

type Resource struct {
	Resource string `json:"resource"` // required
	Value    string `json:"value"`    // required
}
