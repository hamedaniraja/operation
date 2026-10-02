package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sort"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

func sshRun(node, username, password, cmd string) (string, string, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password), // or use PublicKeys, see below
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // NOT for production, see notes
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(node, "22")

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return "", "", fmt.Errorf("failed to dial %s: %w", addr, err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	if err := session.Run(cmd); err != nil {
		return stdoutBuf.String(), stderrBuf.String(), fmt.Errorf("command failed: %w", err)
	}

	return stdoutBuf.String(), stderrBuf.String(), nil

}

func main() {
	nodes := []string{"master", "worker01", "worker02"}
	username := ""
	password := ""
	cmd := ""

	// Doing Serial
	start := time.Now()
	for _, node := range nodes {
		stdout, stderr, err := sshRun(node, username, password, cmd)
		if err != nil {
			log.Fatalf("error running command: %v\nstderr: %s", err, stderr)
		}

		fmt.Println("STDOUT:")
		fmt.Println(stdout)
		if stderr != "" {
			fmt.Println("STDERR:")
			fmt.Println(stderr)
		}
	}
	fmt.Println("Time serial: ", time.Since(start))

	// Doing Parallel
	var wg sync.WaitGroup
	start = time.Now()
	for _, node := range nodes {
		wg.Add(1)
		go func(node, username, password, cmd string) {
			defer wg.Done()
			stdout, stderr, err := sshRun(node, username, password, cmd)
			if err != nil {
				log.Fatalf("error running command: %v\nstderr: %s", err, stderr)
			}

			fmt.Println("STDOUT:")
			fmt.Println(stdout)
			if stderr != "" {
				fmt.Println("STDERR:")
				fmt.Println(stderr)
			}
		}(node, username, password, cmd)
	}
	wg.Wait()
	fmt.Println("Time waitGroup: ", time.Since(start))

	// Doing Parallel using channel
	type sshOut struct {
		stdOut string
		stdErr string
	}

	var wg1 sync.WaitGroup
	stdOutChannel := make(chan sshOut)
	start = time.Now()
	for _, node := range nodes {
		wg1.Add(1)
		go func(node string, username, password, cmd string, stdOutChannel chan sshOut) {
			defer wg1.Done()
			stdout, stderr, err := sshRun(node, username, password, cmd)
			if err != nil {
				log.Fatalf("error running command: %v\nstderr: %s", err, stderr)
			}
			stdOutChannel <- sshOut{stdOut: stdout, stdErr: stderr}
		}(node, username, password, cmd, stdOutChannel)
	}

	go func() {
		wg1.Wait()
		close(stdOutChannel)
	}()

	results := []sshOut{}
	for output := range stdOutChannel {
		results = append(results, output)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].stdOut < results[j].stdOut
	})
	for _, result := range results {
		fmt.Println("STDOUT:")
		fmt.Println(result.stdOut)
		if result.stdErr != "" {
			fmt.Println("STDERR:")
			fmt.Println(result.stdErr)
		}
	}
	fmt.Println("Time channel: ", time.Since(start))
}
