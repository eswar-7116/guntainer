package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	rootfsTarballURL = "https://cdimage.ubuntu.com/ubuntu-base/releases/24.04/release/ubuntu-base-24.04.3-base-amd64.tar.gz"
	RootfsName       = "guntainer-ubuntu-rootfs"
	cacheDir         = ".cache/guntainer"
	tarballRelPath   = cacheDir + "/ubuntu-base-24.04.3-amd64.tar.gz"
)

func SetupRoot() error {
	fmt.Println("Setting up root filesystem...")

	cacheDirPath := filepath.Join(os.Getenv("HOME"), cacheDir)
	fmt.Printf("Creating cache directory at %s\n", cacheDirPath)
	if err := os.MkdirAll(cacheDirPath, 0755); err != nil {
		return err
	}

	tarballPath := filepath.Join(os.Getenv("HOME"), tarballRelPath)
	if _, err := os.Stat(tarballPath); os.IsNotExist(err) {
		fmt.Printf("Tarball not found at %s. Downloading...\n", tarballPath)
		if err := downloadFile(rootfsTarballURL, tarballPath); err != nil {
			return err
		}
		fmt.Println("Download completed.")
	} else {
		fmt.Println("Tarball already exists. Skipping download.")
	}

	rootfsPath := filepath.Join(os.TempDir(), RootfsName)
	fmt.Printf("Cleaning up old rootfs at %s\n", rootfsPath)
	os.RemoveAll(rootfsPath)

	fmt.Printf("Extracting tarball to %s\n", rootfsPath)
	if err := extractTarGz(tarballPath, rootfsPath); err != nil {
		return err
	}

	fmt.Println("Root filesystem setup complete.")
	return nil
}

func downloadFile(url, dest string) error {
	outputFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	_, err = io.Copy(outputFile, res.Body)
	return err
}

func extractTarGz(tarGzPath, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create destination dir: %w", err)
	}

	cmd := exec.Command("tar", "-xzf", tarGzPath, "-C", dest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tar failed: %v\nOutput: %s", err, string(output))
	}
	return nil
}
