package models

type Netbox struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Port     string `json:"port"`
}

type OpenDCIM struct {
	DcimPort    string `json:"dcimPort" binding:"required"`
	AdminerPort string `json:"adminerPort" binding:"required"`
}

type Nautobot struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Port     string `json:"port"`
}

type BackendServer struct {
	Host      string `json:"host" binding:"required"`
	User      string `json:"user" binding:"required"`
	Password  string `json:"password"`
	Pkey      string `json:"pkey"`
	PkeyPass  string `json:"pkeypass"`
	DstPath   string `json:"dstpath" binding:"required"`
	RunPort   string `json:"runport" binding:"required"`
	AtStartup bool   `json:"startup"`
}
