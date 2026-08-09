package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	backendA := exec.Command(
		"./backend",
		"--port=9001",
		"--name=Backend A",
	)

	backendB := exec.Command(
		"./backend",
		"--port=9002",
		"--name=Backend B",
	)

	backendA.Stdout = os.Stdout
	backendA.Stderr = os.Stderr
	backendB.Stdout = os.Stdout
	backendB.Stderr = os.Stderr

	if err := backendA.Start(); err != nil {
		log.Fatalf("failed to start Backend A: %v", err)
	}

	if err := backendB.Start(); err != nil {
		_ = backendA.Process.Kill()
		log.Fatalf("failed to start Backend B: %v", err)
	}

	server := exec.Command("./nimbuslb")
	server.Stdout = os.Stdout
	server.Stderr = os.Stderr
	server.Stdin = os.Stdin
	server.SysProcAttr = &syscall.SysProcAttr{}

	if err := server.Run(); err != nil {
		log.Printf("NimbusLB stopped: %v", err)
	}

	_ = backendA.Process.Kill()
	_ = backendB.Process.Kill()
}
