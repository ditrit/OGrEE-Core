package models

type User struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Token    string `json:"token"`
}