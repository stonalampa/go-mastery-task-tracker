// Command task-cli tracks tasks in a JSON file in the current directory.
package main

import (
	"fmt"
	"io"
	"os"
)

const usage = `Usage:
  task-cli add <description>
  task-cli update <id> <description>
  task-cli delete <id>
  task-cli mark-in-progress <id>
  task-cli mark-done <id>
  task-cli list [todo|in-progress|done]
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes one task-cli command and returns the process exit code.
// args excludes the program name.
func run(args []string, _, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 1
	}
	fmt.Fprintf(stderr, "error: unknown command %q\n%s", args[0], usage)
	return 1
}
