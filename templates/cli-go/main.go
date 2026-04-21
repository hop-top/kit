package main

import (
	"os"

	"{{module_prefix}}/{{app_name}}/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
