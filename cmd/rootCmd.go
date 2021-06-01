package cmd

import "runtime"

func Execute() {
	println("Congratulations, nobody cares")
	return
}

func Exit() {
	runtime.Goexit()
}
