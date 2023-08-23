package models

// swagger:model ErrorResponse
type ErrorResponse struct {
	// Error Response Message
	// in: message
	Message string `json:"message"`
}

// swagger:parameters CreateDevices
type ReqDevicesBody struct {
	//  in: body
	Body Devices `json:"body"`
}

// swagger:parameters CreateConnection
type ReqConnBody struct {
	//  in: body
	Body Connection `json:"body"`
}

type ErrorMessage struct {

	StatusCode int `json:"statuscode"`
	Message string `json:"message"`
}
// swagger:model DatabaseInfo
type DatabaseInfo struct {
	// Host url of database
	// in: host
	// example: http://localhost:8529
	Host string `json:"host"`

	// Database name
	// in: database
	// example: _system
	Database string `json:"database"`
	
	// User of database
	// in: user
	// example: root
	User string `json:"user"`
	
	// Password of the user
	// in: password
	// example: password
	Password string `json:"password"`
}

// swagger:parameters ConnectBDD
type ReqBDDBody struct {
	//  in: body
	Body DatabaseInfo `json:"body"`
}