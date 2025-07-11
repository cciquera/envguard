package main

import (
	"github.com/cciquera/envguard/cmd"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
        fmt.Println("envguard version:", version)
        return
    }
	cmd.Execute()
}
