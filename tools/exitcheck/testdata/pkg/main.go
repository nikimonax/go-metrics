package main

import "os"

func main() {
	foo()
	os.Exit(67) // want "os.Exit\" usage"
}

func foo() {
	os.Exit(42)
}
