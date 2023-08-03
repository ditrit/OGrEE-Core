package models

// swagger:model Connection
type Connection struct {

	// Primary key of device
	// in: _key
	// read only: true
	Key string `json:"_key"`

	// from Device
	// in: _from
	// example: devices/*
	From string `json:"_from"`

	// To device
	// in: _to
	// example: devices/*
	To string `json:"_to"`

	// Type of connection
	// in: type
	// example: parent of (between partens)
	Type string `json:"type"`

	// Date of connection's creation
	// in: created
	// example: 2016-04-22
	Created string `json:"created"`

	// Date of connection's expiration
	// in: expired
	// example: 3000-01-01
	Expired string `json:"expired"`

}

// swagger:model SuccessConResponse
type SuccessConResponse struct {
	// Success
	// in : array
	Connections []Connection
}