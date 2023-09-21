package models

type Service struct {
  Type string `yaml:"type,omitempty"`
  Port int `yaml:"port,omitempty"`
}

type Host struct {
	Host string `yaml:"host,omitempty"`
}

type Ingress struct{
  Enabled bool `yaml:"enabled"`
  ClassName string `yaml:"className,omitempty"`
  EntryPoints []string `yaml:"entryPoints,omitempty"`
  Hosts []Host `yaml:"hosts,omitempty"`
}

type Image struct{
  Repository string `yaml:"repository,omitempty"`
  PullPolicy string `yaml:"pullPolicy,omitempty"`
  Tag string `yaml:"tag,omitempty"`
}
type Env struct{
	Name string `yaml:"name,omitempty"`
	Value string `yaml:"value,omitempty"`
}
type ImagePullSecrets struct{
	Name string `yaml:"name,omitempty"`
}

type ConfigMapData struct {
	NountPath string `yaml:"mountPath,omitempty"`
	Name string `yaml:"name,omitempty"`
	Value string `yaml:"value,omitempty"`
}

type ConfigMap struct {
	Name string `yaml:"name,omitempty"`
	Data []ConfigMapData `yaml:"data,omitempty"`
}

type PersistentVolumeClaim struct {
	Name string `yaml:"name,omitempty"`
    Storage string `yaml:"storage,omitempty"`
    MountPath string `yaml:"mountPath,omitempty"`
}

type Deployement struct {

	ReplicaCount int `yaml:"replicaCount"`
	Name string `yaml:"fullnameOverride"`

	ImagePullSecrets []ImagePullSecrets `yaml:"imagePullSecrets,omitempty"`
	Image Image `yaml:"image"`
	Env []Env `yaml:"env,omitempty"`

	Service Service `yaml:"service"`
	Ingress Ingress `yaml:"ingress"`

	ConfigMap []ConfigMap `yaml:"configmap,omitempty"`
	PersistentVolumeClaim []PersistentVolumeClaim `yaml:"persistentVolumeClaim,omitempty"`
	SecurityContext interface{} `yaml:"securityContext"`
}