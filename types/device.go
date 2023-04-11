package ogreetypes

import "encoding/json"

type DeviceOrientation string

type DeviceTemplate = RackTemplate

type Device struct {
	Header
	FbxModel    string             `json:"fbxModel,omitempty"`
	Height      float64            `json:"height"`
	HeightUnit  RackUnit           `json:"heightUnit"`
	Orientation DeviceOrientation  `json:"orientation"`
	Size        Vector2            `json:"-"`
	SizeUnit    MetricImperialUnit `json:"sizeUnit"`
	Slot        string             `json:"slot,omitempty"`
	Template    string             `json:"template"`
	Type        string             `json:"type,omitempty"`
	PosU        *int               `json:"posU,omitempty"`
	SizeU       *int               `json:"sizeU,omitempty"`
}

type DeviceAlias Device

type DeviceJsonAttributes struct {
	DeviceAlias
	SizeAux Vector2Wrapper `json:"size"`
}

type DeviceJson struct {
	Category string `json:"category"`
	Header
	Attributes DeviceJsonAttributes `json:"attributes"`
}

func (d Device) MarshalJSON() ([]byte, error) {
	return json.Marshal(DeviceJson{
		Category: "device",
		Header:   d.Header,
		Attributes: DeviceJsonAttributes{
			DeviceAlias: DeviceAlias(d),
			SizeAux:     Vector2Wrapper{d.Size},
		},
	})
}

func (d *Device) UnmarshalJSON(data []byte) error {
	var djson DeviceJson
	if err := json.Unmarshal(data, &djson); err != nil {
		return err
	}
	*d = Device(djson.Attributes.DeviceAlias)
	d.Header = djson.Header
	d.Size = djson.Attributes.SizeAux.v
	return nil
}
