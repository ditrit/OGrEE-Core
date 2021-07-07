package main

import "strings"

func derivePath(x string) {
	res := "/"
	arr := strings.Split(x, "/")
	for i := range arr {
		println(arr[i])
		if arr[i] != ".." {
			//println("Adding: ", arr[i])
			res += arr[i] + "/"
		} else if arr[i] == ".." && res != "" {
			//res[0:strings.LastIndexByte(
			res = res[0:strings.LastIndexByte(res[0:strings.LastIndexByte(res, '/')], '/')]
			res += "/"
		}
	}
	println()
	println("FINAL RESULT...")
	println(res)
}
func main() {
	//PASSED!
	//println("TEST CASE 1: ")
	//derivePath("CED/BETA/../A")

	//PASSED!
	//println("TEST CASE 2: ")
	//derivePath("CED/..")

	//PASSED
	//println("TEST CASE 3: ")
	//derivePath("CED/BETA/../..")

	//PASSED
	//println("TEST CASE 4: ")
	//derivePath("CED/BETA/R1/50/..")

	derivePath("../CED")
}
