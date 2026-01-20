package ogreetypes

type SiteOrientation string

type SiteAttributes struct {
	Orientation    *SiteOrientation `json:"orientation,omitempty"`
	ReservedColor  *Color           `json:"reservedColor,omitempty"`
	TechnicalColor *Color           `json:"technicalColor,omitempty"`
	UsableColor    *Color           `json:"usableColor,omitempty"`
}
