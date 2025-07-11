package main

import (
	"fmt"
	"os"

	"github.com/cciquera/envguard/cmd"
)

var version = "0.1.1"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("envguard version:", version)
		return
	}
	cmd.Execute()
}
