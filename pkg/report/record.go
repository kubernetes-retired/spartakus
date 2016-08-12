package report

type Record struct {
	ID            string  `json:"id"` // required
	MasterVersion *string `json:"masterVersion,omitempty"`
	Nodes         []Node  `json:"nodes,omitempty"`
}

type Node struct {
	ID                      string     `json:"id"` // required
	OSImage                 *string    `json:"osImage,omitempty"`
	KernelVersion           *string    `json:"kernelVersion,omitempty"`
	ContainerRuntimeVersion *string    `json:"containerRuntimeVersion,omitempty"`
	KubeletVersion          *string    `json:"kubeletVersion,omitempty"`
	Capacity                []Resource `json:"capacity,omitempty"`
}

type Resource struct {
	Resource string `json:"resource"` // required
	Value    string `json:"value"`    // required
}
