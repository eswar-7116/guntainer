package main

import (
	"fmt"
	"os"
	"os/exec"
)

const PIPE_NAME = "guntainer_child_pipe"
const PIPE_MSG = "__GUNTAINER_CHILD__"
const CHILD_ENV_KEY = "_GUNTAINER_CHILD"

func main() {
	if len(os.Args) < 2 {
		throwError("no sub-command provided", 1)
	}

	switch os.Args[1] {
	case "run":
		parent()
	default:
		throwError("invalid sub-command.\nTry running guntainer help", 1)
	}
}

func parent() {
	if len(os.Args) < 3 {
		os.Args = append(os.Args, "bash")
	}

	if val, found := os.LookupEnv(CHILD_ENV_KEY); found && val == "1" && os.Args[0] == "/proc/self/exe" {
		child(os.NewFile(3, PIPE_NAME))
		return
	}

	r, w, err := os.Pipe()
	if err != nil {
		throwError(err.Error(), 1)
	}
	w.Write([]byte(PIPE_MSG))
	w.Close()

	cmd := exec.Command("/proc/self/exe", append([]string{"run"}, os.Args[2:]...)...)
	cmd.Env = append(os.Environ(), CHILD_ENV_KEY+"=1")
	cmd.ExtraFiles = append(cmd.ExtraFiles, r)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
}

func child(pipe *os.File) {
	buf := make([]byte, 32)
	n, err := pipe.Read(buf)
	pipe.Close()
	if err != nil || pipe.Name() != PIPE_NAME || string(buf[:n]) != PIPE_MSG {
		throwError("unauthorized access to child process", 1)
	}

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		throwError(err.Error(), 1)
	}
}

func throwError(msg string, status int) {
	fmt.Fprintln(os.Stderr, "Error: "+msg)
	os.Exit(status)
}
