# operation

This is a simple project for running a command on some nodes to compare speed of running on serial and parallel. This script with some updates can be used for collecting telemetry parameters on some nodes. It is returning standard output and standard error from every node which passing command is running on it.

We can run this command in two ways:
  operation [OPTIONS]
  operation -help

[OPTIONS]:
  -help         Shows this text.

  -nodes		It is a string which includes Comma separated list of Nodes. 
                default value is: ""

  -username  	It is SSH username for nodes.
                default value is: ""
  
  -password     It is SSH password for nodes. 
                default value is: ""
  
  -command      It is the command to be executed on the nodes.
                default value is: ""

How to build the binary on Linux and Mac:
```bash
$ go build -o ssh_run main.go
```

How to use it:
```bash
$ ./ssh_run -nodes "node1.net,node2.net,node3.net" \
-username "user1" \
-password "user1_password" \
-command "hostname"
```

