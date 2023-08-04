package models


// swagger:model LoginInput
type LoginInput struct {
	// username
	// in: username
	Username string `json:"username" binding:"required"`
	// username
	// in: password
	Password string `json:"password" binding:"required"`
}

type User struct {
	ID uint `json:"id"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// swagger:model SuccessLogin
type SuccessLogin struct {
	Token string `json:"token" binding:"required"`
}

// swagger:parameters LoginToApi
type ReqLoginBody struct {
	//  in: body
	Body LoginInput `json:"body"`
}
