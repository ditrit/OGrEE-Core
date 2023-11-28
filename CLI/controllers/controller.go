package controllers

var C = Controller{
	API:     API,
	Ogree3D: Ogree3D,
	Clock:   Clock,
}

type Controller struct {
	API     APIPort
	Ogree3D Ogree3DPort
	Clock   ClockPort
}
