package ogreetypes

import "encoding/json"

type Domain struct {
	Header
	Color Color `json:"color,omitempty"`
}

type DomainAlias Domain

type DomainJson struct {
	Category string `json:"category"`
	Header
	Attributes DomainAlias `json:"attributes"`
}

func (s Domain) MarshalJSON() ([]byte, error) {
	return json.Marshal(DomainJson{
		Category:   "domain",
		Header:     s.Header,
		Attributes: DomainAlias(s),
	})
}

func (s *Domain) UnmarshalJSON(data []byte) error {
	var sjson DomainJson
	if err := json.Unmarshal(data, &sjson); err != nil {
		return err
	}
	*s = Domain(sjson.Attributes)
	s.Header = sjson.Header
	return nil
}
