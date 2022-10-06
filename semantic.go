package main

import "fmt"

//This file contains code for semantic checking
//(ensure correct data types, syntax etc)

//The data type validator functions return empty
//errors for now since the error is dependent
//on the respective node in the AST (ast.go)
func AssertString(x *node, item string) (string, error) {
	valInf, err := (*x).execute()
	if err != nil {
		return "", err
	}
	val, ok := valInf.(string)
	if !ok {
		return "", fmt.Errorf(item + " should be a string value")
	}
	return val, nil
}

func AssertInt(x *node, item string) (int, error) {
	valInf, err := (*x).execute()
	if err != nil {
		return -1, err
	}
	val, ok := valInf.(int)
	if !ok {
		return -1, fmt.Errorf(item + " should be an int value")
	}
	return val, nil
}

func AssertFloat64(x *node, item string) (float64, error) {
	valInf, err := (*x).execute()
	if err != nil {
		return -1, err
	}
	val, ok := valInf.(float64)
	if !ok {
		return -1, fmt.Errorf(item + " should be a float value")
	}
	return val, nil
}

func AssertFloat64Arr(x *node, item string) ([]float64, error) {
	valInf, err := (*x).execute()
	if err != nil {
		return nil, err
	}
	val, ok := valInf.([]float64)
	if !ok {
		return nil, fmt.Errorf(item + " should be a float array")
	}
	return val, nil
}

func AssertBool(x *node, item string) (bool, error) {
	valInf, err := (*x).execute()
	if err != nil {
		return false, err
	}
	val, ok := valInf.(bool)
	if !ok {
		return false, fmt.Errorf(item + " should be a boolean value")
	}
	return val, nil
}

//Check if x is a certain value
func AssertInStringValues(x interface{}, potential []string) bool {
	for i := range potential {
		if x == potential[i] {
			return true
		}
	}
	return false
}
