package ogreetypes

import "encoding/json"

type SiteOrientation string

type Site struct {
	Header
	Orientation    *SiteOrientation `json:"orientation,omitempty"`
	ReservedColor  *Color           `json:"reservedColor,omitempty"`
	TechnicalColor *Color           `json:"technicalColor,omitempty"`
	UsableColor    *Color           `json:"usableColor,omitempty"`
}

type SiteAlias Site

type SiteJson struct {
	Category string `json:"category"`
	Header
	Attributes SiteAlias `json:"attributes"`
}

func (s Site) MarshalJSON() ([]byte, error) {
	return json.Marshal(SiteJson{
		Category:   "site",
		Header:     s.Header,
		Attributes: SiteAlias(s),
	})
}

func (s *Site) UnmarshalJSON(data []byte) error {
	var sjson SiteJson
	if err := json.Unmarshal(data, &sjson); err != nil {
		return err
	}
	*s = Site(sjson.Attributes)
	s.Header = sjson.Header
	return nil
}
