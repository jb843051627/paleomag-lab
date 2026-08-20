package model

import "fmt"

type Orientation struct {
	Declination float64 `json:"declination"`
	Inclination float64 `json:"inclination"`
	Coordinate  string  `json:"coordinate"`
	Source      string  `json:"source"`
}

func (o Orientation) Validate() error {
	if o.Declination < 0 || o.Declination >= 360 {
		return fmt.Errorf("%w: declination must be in [0,360)", ErrInvalid)
	}
	if o.Inclination < -90 || o.Inclination > 90 {
		return fmt.Errorf("%w: inclination must be in [-90,90]", ErrInvalid)
	}
	if o.Coordinate == "" || o.Source == "" {
		return fmt.Errorf("%w: orientation coordinate and source are required", ErrInvalid)
	}
	return nil
}

func (o Orientation) IsGeographicallyPlausible() bool {
	return o.Declination >= 0 && o.Declination < 360 && o.Inclination >= -90 && o.Inclination <= 90
}
