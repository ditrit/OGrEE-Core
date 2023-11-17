package controllers

var C = Controller{
	API:     API,
	Ogree3D: Ogree3D,
}

type Controller struct {
	API     APIPort
	Ogree3D Ogree3DPort
}
