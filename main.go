package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 {
		throwError("no sub-command provided", 1)
	}

	switch os.Args[1] {
	case "run":
		parent()
	default:
		throwError("invalid sub-command.\nTry running ./container help", 1)
	}
}

func parent() {
	if len(os.Args) < 3 {
		os.Args = append(os.Args, "bash")
	}

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err == nil {
		throwError("failed to run command: "+os.Args[2], 1)
	}
}

func throwError(msg string, status int) {
	fmt.Fprintln(os.Stderr, "Error: "+msg)
	os.Exit(status)
}
