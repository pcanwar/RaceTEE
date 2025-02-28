package main

import "encoding/json"

var sum = 0
var product = 1

// Add is a function that performs addition
func Add(a int) int {
	sum += a
	return sum
}

// Mul is a function that performs multiplication
func Mul(a int) int {
	product *= a
	return product
}

type States struct {
	Sum     int // global variable: addition result
	Product int // global variable: multiplication result
}

// IMPORTANT: GetStates and SetStates must be defined in the user code
// GetStates returns the current state witch will be saved and load lately by the VM using the SetStates function
func GetStates() []byte {
	res, err := json.Marshal(States{Sum: sum, Product: product})
	if err != nil {
		return nil
	}
	return res
}

func SetStates(states []byte) {
	var newStates States
	err := json.Unmarshal(states, &newStates)
	if err != nil {
		return
	}

	sum = newStates.Sum
	product = newStates.Product
}
