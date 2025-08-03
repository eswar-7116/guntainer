package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

const (
	PIPE_NAME     = "guntainer_child_pipe"
	PIPE_MSG      = "__GUNTAINER_CHILD__"
	CHILD_ENV_KEY = "_GUNTAINER_CHILD"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	switch os.Args[1] {
	case "run":
		parent()
	case "help", "--help", "-h":
		printHelp()
	default:
		throwError("invalid sub-command.\nTry running: guntainer run /bin/sh")
	}
}

func printHelp() {
	fmt.Println(`Usage: guntainer <command> [args...]

Available commands:
  run <program>     Run a program inside a container (default: sh)
  help              Show this help message

Example:
  guntainer run /bin/sh`)
}

func parent() {
	if len(os.Args) < 3 {
		os.Args = append(os.Args, "sh")
	}

	// Check if already in child
	if val, _ := os.LookupEnv(CHILD_ENV_KEY); val == "1" && os.Args[0] == "/proc/self/exe" {
		child(os.NewFile(3, PIPE_NAME))
		return
	}

	fmt.Println(">> Setting up root filesystem")
	SetupRoot()

	r, w, err := os.Pipe()
	must(err)

	_, err = w.Write([]byte(PIPE_MSG))
	must(err)
	w.Close()

	cmd := exec.Command("/proc/self/exe", append([]string{"run"}, os.Args[2:]...)...)

	cmd.Env = append(os.Environ(), CHILD_ENV_KEY+"=1")
	cmd.ExtraFiles = append(cmd.ExtraFiles, r)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER,
		Unshareflags: syscall.CLONE_NEWNS,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
		GidMappingsEnableSetgroups: false,
	}

	fmt.Println(">> Entering container")
	must(cmd.Run())
}

func child(pipe *os.File) {
	buf := make([]byte, 32)
	n, err := pipe.Read(buf)
	pipe.Close()
	if err != nil || pipe.Name() != PIPE_NAME || string(buf[:n]) != PIPE_MSG {
		throwError("unauthorized child entry")
	}

	syscall.Sethostname([]byte("guntainer"))
	rootfsPath := filepath.Join(os.TempDir(), RootfsName)
	fmt.Println(">> init: chroot to", rootfsPath)
	must(syscall.Chroot(rootfsPath))
	must(os.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	fmt.Printf(">> running %s\n", os.Args[2])
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(cmd.Run())

	must(syscall.Unmount("proc", 0))
}

func must(err error) {
	if err != nil {
		throwError(err.Error())
	}
}

func throwError(msg string) {
	fmt.Fprintln(os.Stderr, "Error:", msg)
	os.Exit(1)
}
