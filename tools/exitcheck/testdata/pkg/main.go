package main

import ex "os"

type executor struct{}

func (executor) Exit(code int) {}

var os executor

func main() {
	foo()
	os.Exit(67) // method call on local var named "os", must not be reported
	ex.Exit(2)  // aliased real os.Exit, must be reported // want "os.Exit\" usage"
}

func foo() {
	os.Exit(42)
}
