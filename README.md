# operation

This is a simple project for running a command on some nodes to compare speed of running on serial and parallel. This script with some updates can be used for collecting telemetry parameters on some nodes. It is returning standard output and standard error from every node which passing command is running on it.

We can run this script by:  
```
$ ./ssh_run [OPTIONS]
```
[OPTIONS]:  
&emsp; -help&emsp; Shows this text.

&emsp; -nodes&emsp; It is a string which includes Comma separated list of Nodes. 
                default value is: ""

&emsp; -username&emsp; It is SSH username for nodes. default value is: ""
  
&emsp; -password&emsp; It is SSH password for nodes. default value is: ""
  
&emsp; -command&emsp; It is the command to be executed on the nodes. default value is: ""

How to build the binary on Linux and Mac:
```bash
$ cd /path/to/project/dir
$ go build -o ssh_run main.go
```

How to use it (example):
```bash
$ ssh_run -nodes "node1.net,node2.net,node3.net" \
-username "user1" \
-password "user1_password" \
-command "hostname"
```

