package vm

import (
	"fmt"
	"strings"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

var (
	interpreter *interp.Interpreter
)

// InitializeInterpreter initializes the yaegi interpreter
func InitializeInterpreter(userCode []byte) error {
	interpreter = interp.New(interp.Options{})
	// Import Go standard library
	interpreter.Use(stdlib.Symbols)

	// dynamic interpret user code
	_, err := interpreter.Eval(string(userCode))
	if err != nil {
		return fmt.Errorf("failed to interpret user code: %v", err)
	}
	return nil
}

func SetStates(states []byte) error {
	// set the state
	_, err := CallMethod("SetStates", fmt.Sprintf("[]byte(`%s`)", string(states)))
	if err != nil {
		return fmt.Errorf("failed to call SetStates: %v", err)
	}

	return nil
}

// SaveState saves the current state to a JSON string
func GetStates() ([]byte, error) {
	// get the current state
	result, err := CallMethod("GetStates", "")
	if err != nil {
		return nil, fmt.Errorf("failed to call GetStates: %v", err)
	}
	return result.([]byte), nil
}

// CallMethod calls a method in the interpreted code
func CallMethod(methodName string, arg string) (interface{}, error) {
	// construct the expression to call the method
	arg = strings.Trim(arg, "\"")
	expr := fmt.Sprintf("%s(%s)", methodName, arg)
	// print("expr: ", expr, "\n")

	// interpret the expression
	result, err := interpreter.Eval(expr)
	if err != nil {
		return nil, fmt.Errorf("failed to call method %s: %v", methodName, err)
	}

	return result.Interface(), nil
}
