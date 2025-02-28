package golang

import (
	"fmt"
	"tee/process/golang/vm"
)

func Deploy(userCode []byte) ([]byte, error) {
	// dynamic load user code
	err := vm.InitializeInterpreter(userCode)
	if err != nil {
		fmt.Println("Error initializing interpreter:", err)
		return nil, err
	}

	// save the initial state
	state, err := vm.GetStates()
	if err != nil {
		fmt.Println("Error saving state:", err)
		return nil, err
	}
	return state, nil
}

func Execute(userCode []byte, state []byte, funcName string, args string) ([]byte, interface{}, error) {
	// dynamic load user code
	err := vm.InitializeInterpreter(userCode)
	if err != nil {
		fmt.Println("Error initializing interpreter:", err)
		return nil, nil, err
	}

	// load the state
	err = vm.SetStates(state)
	if err != nil {
		fmt.Println("Error loading state:", err)
		return nil, nil, err
	}

	// Call
	result, err := vm.CallMethod(funcName, args)
	if err != nil {
		fmt.Println("Error calling method:", err)
		return nil, nil, err
	}

	// Save the updated state
	state, err = vm.GetStates()
	if err != nil {
		fmt.Println("Error saving state:", err)
		return nil, nil, err
	}

	return state, result, nil
}
