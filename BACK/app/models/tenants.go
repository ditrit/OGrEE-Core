package models

type Tenant struct {
	Name             string `json:"name" binding:"required"`
	CustomerPassword string `json:"customerPassword"`
	ApiUrl           string `json:"apiUrl"`
	WebUrl           string `json:"webUrl"`
	ApiPort          string `json:"apiPort"`
	WebPort          string `json:"webPort"`
	DocUrl           string `json:"docUrl"`
	DocPort          string `json:"docPort"`
	HasWeb           bool   `json:"hasWeb"`
	HasBFF           bool   `json:"hasBFF,omitempty"`
	HasDoc           bool   `json:"hasDoc"`
	AssetsDir        string `json:"assetsDir"`
	ImageTag         string `json:"imageTag"`
}

type ContainerInfo struct {
	Name        string `json:"Names" binding:"required"`
	LastStarted string `json:"RunningFor" binding:"required"`
	Status      string `json:"State" binding:"required"`
	Image       string `json:"Image" binding:"required"`
	Size        string `json:"Size"`
	Ports       string `json:"Ports" binding:"required"`
}

type DbFile struct {
	Name             string
	CustomerPassword string
	Env              []Env
	ConfigMap        []ConfigMap
	Image            string
	Tag              string
	Port             int
	Volume           []PersistentVolumeClaim
}

type APIFile struct {
	Ingress          bool
	Namespace        string
	FullnameOverride string
	BDDPass          string
	Port             int
	Env              []Env
	ConfigMap        []ConfigMap
	Image            string
	Tag              string
}

type APPFile struct {
	Namespace        string
	FullnameOverride string
	Port             int
	Tag              string
}

type Backup struct {
	DBPassword string `json:"password" binding:"required"`
	IsDownload bool   `json:"shouldDownload"`
}
