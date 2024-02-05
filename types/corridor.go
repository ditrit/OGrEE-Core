package ogreetypes

type Temperature string

type CorridorAttributes struct {
	Content     string      `json:"content"`
	Temperature Temperature `json:"temperature"`
}
