package ogreetypes

import "encoding/json"

type Group struct {
	Header  `json:"-"`
	Content string `json:"content"`
}

type GroupAlias Group

type GroupJson struct {
	Category string `json:"category"`
	Header
	Attributes GroupAlias `json:"attributes"`
}

func (g Group) MarshalJSON() ([]byte, error) {
	return json.Marshal(GroupJson{
		Category:   "group",
		Header:     g.Header,
		Attributes: GroupAlias(g),
	})
}

func (g *Group) UnmarshalJSON(data []byte) error {
	var gjson GroupJson
	if err := json.Unmarshal(data, &gjson); err != nil {
		return err
	}
	*g = Group(gjson.Attributes)
	g.Header = gjson.Header
	return nil
}
