package controllers

import "time"

var Clock ClockPort = &clockPortImpl{}

type ClockPort interface {
	Now() time.Time
}

type clockPortImpl struct{}

func (clock clockPortImpl) Now() time.Time {
	return time.Now()
}
