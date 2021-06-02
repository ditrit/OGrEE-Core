package cmd

import "runtime"

func Execute() {
	println("Congratulations, nobody cares")
	return
}

func Exit() {
	runtime.Goexit()
}

func Disp(x map[string]string) {
	for i, k := range x {
		println("We got: ", i, " and ", k)
	}
}
