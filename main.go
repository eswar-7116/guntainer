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
		throwError("Error: invalid sub-command.\nTry running: guntainer run /bin/bash")
	}
}

func printHelp() {
	fmt.Println(`Usage: guntainer <command> [args...]

Available commands:
  run <program>     Run a program inside a container (default: sh)
  help              Show this help message

Example:
  guntainer run /bin/bash`)
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
	must(err, "creating pipe")

	_, err = w.Write([]byte(PIPE_MSG))
	must(err, "writing to pipe")
	w.Close()

	// Child process command
	cmd := exec.Command("/proc/self/exe", append([]string{"run"}, os.Args[2:]...)...)

	cmd.Env = append(os.Environ(), CHILD_ENV_KEY+"=1")
	cmd.ExtraFiles = append(cmd.ExtraFiles, r)

	// Assign input, output and error streams
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Isolate namespaces and users
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
	must(cmd.Run(), "running the container process")
}

func child(pipe *os.File) {
	buf := make([]byte, 32)
	n, err := pipe.Read(buf)
	pipe.Close()
	if err != nil || pipe.Name() != PIPE_NAME || string(buf[:n]) != PIPE_MSG {
		throwError("Error: unauthorized child entry")
	}

	// Set new hostname
	syscall.Sethostname([]byte("guntainer"))
	rootfsPath := filepath.Join(os.TempDir(), RootfsName)
	fmt.Println(">> init: chroot to", rootfsPath)
	must(syscall.Chroot(rootfsPath), "changing root in container")  // Change root filesystem
	must(os.Chdir("/"), "changing directory to container root")  // Change current directory to root
	must(syscall.Mount("proc", "proc", "proc", 0, ""), "mount proc directory")  // Mount /proc

	fmt.Printf(">> running %s\n", os.Args[2])
	cmd := exec.Command(os.Args[2], os.Args[3:]...)  // Given command
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(cmd.Run(), "running "+os.Args[2]+" in the container")

	must(syscall.Unmount("proc", 0), "unmounting proc directory")
}

func must(err error, when string) {
	if err != nil {
		throwError("Error while " + when + ": " + err.Error())
	}
}

func throwError(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
