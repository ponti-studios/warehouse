package random

import (
	"fmt"
	"os"
)

func variables() {
	var goos string = os.Getenv("GOOS")
	fmt.Printf("The operating system is: %s\n", goos)
	path := os.Getenv("PATH")
	fmt.Printf("The PATH is: %s\n", path)
	var a string = "meow"
	var b int = 5
	fmt.Println(a)
	fmt.Println(b)
}

func Run() error {
	fmt.Println("Random Go examples")
	fmt.Println("Available: variables, pointers, numbers, constants, if_else")
	fmt.Println("These are example files, run individually to test:")
	fmt.Println("  go run constants.go")
	fmt.Println("  go run hello_world.go")
	fmt.Println("  go run if_else.go")
	fmt.Println("  go run numbers.go")
	fmt.Println("  go run pointers.go")
	return nil
}
