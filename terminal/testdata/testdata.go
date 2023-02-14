package main

import (
	"fmt"

	"github.com/bluekaki/pkg/terminal"
)

func main() {
	fmt.Println(terminal.ReadPwd(true, "dummy pwd"))

	fmt.Println(terminal.ReadSomething("dummy xxx"))
}
