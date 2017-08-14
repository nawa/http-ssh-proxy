## http-ssh-proxy

Tool that allows to proxy your web environment through ssh connections. Could be helpful for cluster of nodes when http ports are hidden from external access by firewall but you have access to these nodes using ssh. 

### Use case
Personally I'm using `http-ssh-proxy` to access to Spark cluster deployed on AWS and its ports are hidden behind firewall. Two nodes in cluster exist
	
- master - Master UI have address `10.1.1.1:8080` and available directly from my machine without any forwardings. Master node also servers as worker and its UI address is `10.1.1.1:8081` - this port is not available. I have ssh access only to this node and it sees all nodes and ports in whole cluster
- worker - separate machine with worker's responsibility only. UI on `10.1.1.2:8080`

To summarize - only master node is visible but using ssh access through it I can access to each other node

### Configuration
Configuration for use case described above

```yaml
app-port: 8080
start-page: master
hosts:
    master:
        address: 10.1.1.1:8080
    worker-1-8081:
        address: 10.1.1.1:8081
        forwarding:
            server: 10.1.1.1:22
            user: ssh-username
            private-key: ~/.ssh/id_rsa
            password: #in case of password
    worker-2-8081:
        address: 10.1.1.2:8081
        forwarding:
            server: 10.1.1.1:22
            user: ssh-username
            private-key: ~/.ssh/id_rsa
            password: #in case of password
``` 

- `app-port` main port of the tool
- `start-page` main page showing one of defined hosts below
- `hosts` the list of hosts to be proxied. `master`, `worker-1-8081`, `worker-2-8081` will be used for rewrite links on proxied pages
	- `host.address` address you want to proxy
	- `host.forwarding` host can be hidden or not. If it isn't visible you have to open it using ssh port forwarding with settings under this section
	- `host.forwarding.server` address of ssh server for which `host.address` is visible
	- `host.forwarding.user`, `private-key`, `password` ssh connection paramaters. Private key or password could be used

### Run

- `go get github.com/nawa/http-ssh-proxy`
- create `config.yml` in your working folder
- `go run http-ssh-proxy.go` or `go build http-ssh-proxy.go && http-ssh-proxy`