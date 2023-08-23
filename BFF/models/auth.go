package models


// swagger:model LoginInput
type LoginInput struct {
	// email
	// in: email
	Email string `json:"email" binding:"required"`
	// username
	// in: password
	Password string `json:"password" binding:"required"`
}

type User struct {
	ID uint `json:"id"`
	Email string `json:"email" binding:"required"`
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
