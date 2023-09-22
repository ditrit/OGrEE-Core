package models

type Netbox struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Port     string `json:"port"`
}