package models

import (
	"cli/utils"
	"fmt"
	"strconv"
)

// Rack inner attributes
type Breaker struct {
	Powerpanel string  `json:"powerpanel"`
	Type       string  `json:"type,omitempty"`
	Circuit    string  `json:"circuit,omitempty"`
	Intensity  float64 `json:"intensity,omitempty"`
	Tag        string  `json:"tag,omitempty"`
}

func ValuesToBreaker(values []any) (string, Breaker, error) {
	nMandatory := 2
	// mandatory name
	mandatoryErr := fmt.Errorf("at least %d values (name and powerpanel) expected to add a breaker", nMandatory)
	if len(values) < nMandatory {
		return "", Breaker{}, mandatoryErr
	}
	name, err := utils.ValToString(values[0], "name")
	if err != nil {
		return name, Breaker{}, err
	}
	powerpanel, err := utils.ValToString(values[1], "powerpanel")
	if err != nil {
		return name, Breaker{}, err
	}
	if len(name) <= 0 || len(powerpanel) <= 0 {
		return name, Breaker{}, mandatoryErr
	}
	var breakerType string
	var circuit string
	var intensityStr string
	var tag string
	for index, receiver := range []*string{&breakerType, &circuit, &intensityStr, &tag} {
		err = setOptionalParam(index+nMandatory, values, receiver)
		if err != nil {
			return name, Breaker{}, err
		}
	}
	var intensity float64
	if intensity, err = strconv.ParseFloat(intensityStr,
		64); intensityStr != "" && (err != nil || intensity <= 0) {
		return name, Breaker{}, fmt.Errorf("invalid value for intensity, it should be a positive number")
	}

	return name, Breaker{Powerpanel: powerpanel, Type: breakerType,
		Circuit: circuit, Intensity: intensity, Tag: tag}, nil
}

// Helpers
func setOptionalParam(index int, values []any, receiver *string) error {
	if len(values) > index {
		value, err := utils.ValToString(values[index], fmt.Sprintf("optional %d", index))
		if err != nil {
			return err
		}
		*receiver = value
	}
	return nil
}
