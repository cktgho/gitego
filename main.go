// main.go

package main

import (
	"github.com/bgreenwell/gitego/cmd" // IMPORTANT: Replace with your module path!
)

func main() {
	// All the application logic now lives in the cmd package.
	// This keeps main.go clean and simple.
	cmd.Execute()
}

