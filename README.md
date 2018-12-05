## github.com/jim-minter/proxy

Proof-of-concept codebase which allows proxying of TCP connections, through a
TLS and client-certificate protected listener, to selected other listeners
running on the same hostname as the proxy.

### Getting started

1. Generate sample self-signed client and proxy keys and certificates

```bash
$ go run ./cmd/generatecerts
$ ls *.crt *.key
client.crt  proxy.crt  sampleserver.crt
client.key  proxy.key  sampleserver.key
```

2. Run a sample https server (this runs on port 8444)

```bash
$ go run ./cmd/sampleserver &
$ curl --insecure --silent \
    https://$(hostname):8444/
Hello, world!
```

3. Run the proxy (by default it runs on port 8443 and allows forwarding to 22
   and 8444)

```bash
$ go run ./cmd/proxy -h
Usage of ./proxy:
  -allowed string
    	allowed proxy ports (default "22,8444")
  -listen string
    	listen port (default "8443")
$ go run ./cmd/proxy &
1970/01/01 00:00:00 proxy is running
```

4. Access the sample https server via the proxy using curl

```bash
$ curl --insecure --silent \
    --proxy https://$(hostname):8443/ --proxy-insecure \
    --proxy-cert client.crt --proxy-key client.key \
    https://$(hostname):8444/
Hello, world!
127.0.0.1:12345 - client [01/Jan/1970 00:00:00] "CONNECT localhost:8444 HTTP/1.1" 200 0
```

5. Access the sample https server via the proxy using Go

```bash
$ go run ./cmd/sampleclient https://$(hostname):8443/ https://$(hostname):8444/
Hello, world!
127.0.0.1:12345 - client [01/Jan/1970 00:00:00] "CONNECT localhost:8444 HTTP/1.1" 200 0
```

6. SSH via the proxy

```bash
$ ssh -o ProxyCommand='go run ./cmd/corkscrew https://$(hostname):8443/ $(hostname):22' \
    $(hostname)
user@localhost's password:
$ logout
127.0.0.1:12345 - client [01/Jan/1970 00:00:00] "CONNECT localhost:22 HTTP/1.1" 200 0
```

### Components

* `proxy` - the actual proxy itself
* `corkscrew` - ssh helper (ProxyCommand) to connect via the proxy
* `sampleclient` - sample Golang program which makes an HTTPS call via the proxy
* `generatecerts` - generates sample keys and certificates
* `sampleserver` - sample TLS server
