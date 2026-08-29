package flagextra

import (
	"flag"
	"fmt"
	"os"
)

func NewFlagSet() *flag.FlagSet {
	var program string

	if len(os.Args) > 0 {
		program = os.Args[0]
	}

	cmd := flag.NewFlagSet(program, flag.ExitOnError)
	cmd.Usage = func() {
		_, _ = fmt.Fprintf(cmd.Output(), "Usage of %s:\n", program)
		cmd.PrintDefaults()
	}

	return cmd
}
