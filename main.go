package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

var nodes string
var username string
var password string
var command string
var mode string
var help bool

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

func showHelp() {
	str := `
We can run this command in two ways:
  operation [OPTIONS]
  operation -help

[OPTIONS]:
  -help         Shows this text.

  -nodes		It is a string which includes Comma separated list of Nodes. 
                Default value is: ""

  -username  	It is SSH username for nodes.
                Default value is: ""
  
  -password     It is SSH password for nodes. 
                Default value is: ""
  
  -command      It is the command to be executed on the nodes.
                Default value is: ""

  -mode			It is definging if we want to run the command on nodes serial or parallel.
				Could be either "serial" or "parallel"
				Default value is: "parallel"

Example:
$ ./operation -nodes "node1.net,node2.net,node3.net" -username "user1" -password "user1_password" -command "hostname" -m "serial"
`
	fmt.Println(str)
}

func main() {

	flag.StringVar(&nodes, "n", "", "Comma separated list of Nodes")
	flag.StringVar(&username, "u", "", "SSH username")
	flag.StringVar(&password, "p", "", "SSH password")
	flag.StringVar(&command, "c", "", "Command to be executed")
	flag.StringVar(&mode, "m", "parallel", "Running \"serial\" or \"parallel\"")
	flag.BoolVar(&help, "h", false, "Showing help")
	flag.Parse()

	nodes_slice := strings.Split(nodes, ",")

	if help {
		showHelp()
		return
	}
	start := time.Now()
	switch mode {
	case "serial": // Doing Serial
		for _, node := range nodes_slice {
			stdout, stderr, err := sshRun(node, username, password, command)
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
	case "parallel": // Doing Parallel
		var wg sync.WaitGroup
		start = time.Now()
		for _, node := range nodes_slice {
			wg.Add(1)
			go func(node, username, password, cmd string) {
				defer wg.Done()
				stdout, stderr, err := sshRun(node, username, password, command)
				if err != nil {
					log.Fatalf("error running command: %v\nstderr: %s", err, stderr)
				}
				fmt.Println("STDOUT:")
				fmt.Println(stdout)
				if stderr != "" {
					fmt.Println("STDERR:")
					fmt.Println(stderr)
				}
			}(node, username, password, command)
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
		for _, node := range nodes_slice {
			wg1.Add(1)
			go func(node string, username, password, cmd string, stdOutChannel chan sshOut) {
				defer wg1.Done()
				stdout, stderr, err := sshRun(node, username, password, command)
				if err != nil {
					log.Fatalf("error running command: %v\nstderr: %s", err, stderr)
				}
				stdOutChannel <- sshOut{stdOut: stdout, stdErr: stderr}
			}(node, username, password, command, stdOutChannel)
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
	default:
		log.Fatalf("Error: mode(-m) can be either \"serial\" or \"parallel\"")
	}
}
