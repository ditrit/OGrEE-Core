package models

// swagger:model Devices
type Devices struct {

	// Primary key of device
	// in: _key
	// read only: true
	Key string `json:"_key"`
	// name of Devices
	// in: _name
	// example: storage_bay
	Name string `json:"_name"`
	// group_name of Devices
	// in: group_name
	// example: GS00OPSAN06
	GroupName string `json:"group_name"`
	// category of Devices
	// in: category
	// example: port
	Category string `json:"category"`
	// sp_name of Devices
	// in: sp_name
	// example: sp_b
	SpName string `json:"sp_name"`

	// sp_port_id of Devices
	// in: sp_port_id
	// example: 0
	SpPortId string `json:"sp_port_id"`

	// hba_device_name of Devices
	// in: hba_device_name
	// example: nsa.*
	HbaDeviceName string `json:"hba_device_name"`

	// storage_group_name of Devices
	// in: storage_group_name
	// example: storage
	StorageGroupName string `json:"storage_group_name"`

	// Date of device's creation
	// in: created
	// example: 2016-04-22
	Created string `json:"created"`

	// Date of device's expiration
	// in: expired
	// example: 3000-01-01
	Expired string `json:"expired"`

}

// swagger:model SuccessResponse
type SuccessResponse struct {
	// Success
	// in : array
	Devices []Devices
}


