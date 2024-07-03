package views

import "fmt"

func PrintDryRunErrors(dryRunErrors []error) {
	// print error quantity
	fmt.Println("####################")
	errCountMsg := fmt.Sprint("Errors found: ", len(dryRunErrors))
	if len(dryRunErrors) > 0 {
		fmt.Println("\033[31m" + errCountMsg + "\033[0m")
	} else {
		fmt.Println("\u001b[32m" + errCountMsg + "\u001b[0m")
	}

	// print error recap
	for idx, err := range dryRunErrors {
		fmt.Println("\033[31m# Error", idx, "\033[0m")
		fmt.Println(err)
	}
}
