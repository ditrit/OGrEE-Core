package models

type Metadata struct{
	Name string `json:"name"`
}

type Namespace struct {
	ApiVersion string `json:"apiVersion"`
	Kind string `json:"kind"`
	Metadata Metadata`json:"metadata,omitempty"`
}
type ContainerStatuses struct{
	Name string `json:"name"`
	Ready bool `json:"ready"`
	RestartCount int `json:"restartCount"`
	Started bool `json:"started"`
}
type Status struct {
	ContainerStatuses []ContainerStatuses`json:"containerStatuses"`
	Phase string `json:"phase"`
	HostIp string `json:"hostIp"`
	StartTime string `json:"startTime"`
}

type Port struct {
	ContainerPort int `json:"containerPort"`
	Name string `json:"name"`
	Protocol string `json:"protocol"`
}
type Container struct{
	Image string `json:"image"`
	Name string `json:"name"`
	ImagePullPolicy string `json:"imagePullPolicy"`
	Ports []Port `json:"ports"`
}
type SpecPod struct{
	Containers []Container `json:"containers"`
}
type Pod struct {
	ApiVersion string `json:"apiVersion"`
	Kind string `json:"kind"`
	Status Status `json:"status"`
	Spec SpecPod `json:"spec"`
}
