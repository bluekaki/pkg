package env

import (
	"fmt"
	"os"
	"strings"
)

var activated *environment

type environment struct {
	value string
}

func (e *environment) Value() string {
	return e.value
}

func (e *environment) IsDev() bool {
	return e.value == "dev"
}

func (e *environment) IsTest() bool {
	return e.value == "test"
}

func (e *environment) IsPre() bool {
	return e.value == "pre"
}

func (e *environment) IsProd() bool {
	return e.value == "prod"
}

func (e *environment) t() {}

func init() {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("env"))) {
	case "dev":
		activated = &environment{value: "dev"}
	case "test":
		activated = &environment{value: "test"}
	case "pre":
		activated = &environment{value: "pre"}
	case "prod":
		activated = &environment{value: "prod"}
	default:
		activated = &environment{value: "test"}
		fmt.Println("Warning: environment 'env' don't found or it's illegal, the default 'test' will be used.")
	}

	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	fmt.Println("'env' is " + activated.Value())
	fmt.Println("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
}

// Activated the activated env
func Activated() *environment {
	return activated
}
