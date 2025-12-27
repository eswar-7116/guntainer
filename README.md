# Guntainer

Guntainer is a minimal container runtime implemented in Go. It demonstrates how Linux containers work internally by directly using Linux kernel primitives such as namespaces and filesystem isolation, without relying on Docker or other container engines.

## Features

* ### Linux Namespace Isolation:

  Uses Linux namespaces (PID, Mount, UTS, IPC, User) to isolate processes, providing the core mechanism that makes a container a container.

* ### Unprivileged Containers:

  Leverages user namespaces to map container root (UID 0) to the invoking host user, allowing containers to run without real root privileges.

* ### Filesystem Isolation:

  Sets up an isolated root filesystem using a minimal Ubuntu rootfs combined with `chroot` and mount namespaces.

* ### Self Re-exec Model:

  The runtime re-executes itself to transition from the host context into the container context, mirroring how real container runtimes bootstrap isolated processes.

* ### Simple CLI:

  Provides a minimal command-line interface to execute arbitrary commands inside the isolated environment.

## Usage

### Run a shell inside the container:

```bash
guntainer run /bin/bash
```

### Run a command:

```bash
guntainer run id
```

The command will run with UID 0 inside the container while remaining unprivileged on the host.

## Installation

### Using Go Install (Recommended):

```bash
go install github.com/eswar-7116/guntainer@latest
```

Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is in your `PATH`.

### From Source:

```bash
git clone https://github.com/eswar-7116/guntainer.git
cd guntainer
go build -o guntainer
```

## Platform Support

* ### Linux-only:

  Guntainer depends on Linux-specific kernel features such as namespaces, `clone`, UID/GID mapping, mount namespaces, etc.

  * Linux: ✅ supported
  * macOS: ❌ not supported
  * Windows: ❌ not supported

macOS is Unix-based but does not implement Linux kernel primitives required by this project.

## Limitations

* ### No Networking Isolation:

  Network namespaces are not implemented. Containers share the host network stack.

* ### No Resource Limits:

  cgroups are not used because they need elevated privileges. CPU and memory usage are unrestricted.

* ### No Image Management:

  There is no layered image format, registry support, or caching mechanism.

* ### Not Production-Ready:

  This runtime is intended strictly for learning and experimentation.

## Conclusion

Guntainer is a minimal, educational container runtime designed to explain containers from first principles. It focuses on clarity and correctness over features, making it suitable for developers interested in Linux internals, systems programming, and understanding how tools like Docker work beneath the surface.

---

#### <div align="center">If you like this project, please give this repo a star 🌟</div>

---
