package ogreetypes

import "encoding/json"

type Temperature string

type Corridor struct {
	Header
	Content     string      `json:"content"`
	Temperature Temperature `json:"temperature"`
}

type CorridorAlias Corridor

type CorridorJson struct {
	Category string `json:"category"`
	Header
	Attributes CorridorAlias `json:"attributes"`
}

func (c Corridor) MarshalJSON() ([]byte, error) {
	return json.Marshal(CorridorJson{
		Category:   "corridor",
		Header:     c.Header,
		Attributes: CorridorAlias(c),
	})
}

func (c *Corridor) UnmarshalJSON(data []byte) error {
	var cjson CorridorJson
	if err := json.Unmarshal(data, &cjson); err != nil {
		return err
	}
	*c = Corridor(cjson.Attributes)
	c.Header = cjson.Header
	return nil
}
