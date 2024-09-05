# Simple Node Health

[![Go Report Card](https://goreportcard.com/badge/github.com/shadowbq/simple-node-health)](https://goreportcard.com/report/github.com/shadowbq/simple-node-health)
[![GoDoc](https://godoc.org/github.com/shadowbq/simple-node-health?status.svg)](https://godoc.org/github.com/shadowbq/simple-node-health)

The `simple-node-health` tool is a versatile Go-based utility designed to monitor the health of Linux systems by providing an easy-to-use HTTP server and command-line interface. It performs essential checks, such as monitoring EXT4 file systems for read-only status, conducting DNS queries, and reporting the overall status of the node. This tool is best utilized when coupled with monitoring tools like Uptime Kuma or other systems that can monitor states via web calls, allowing users to integrate these health checks into their broader monitoring setups and ensure real-time visibility into server performance and stability in production environments.

The web routes are protected with an OAUTH `bearer` issued via `form:client_id` `form:client_secret` `token` method.

## Dev Prerequisites

* Golang 1.22

## Usage

```shell
$> ./simple-node-health --help
A simple tool to check hardware EXT4 devices and run DNS queries

Usage:
  simple-node-health [flags]
  simple-node-health [command]

Available Commands:
  check         Run various checks
  completion    Generate the autocompletion script for the specified shell
  create-client Create a new client_id and client_secret and append them to the config file
  help          Help about any command
  settings      Print the current configuration settings
  show-routes   Show all registered HTTP routes
  version

Flags:
  -d, --domain string   Domain to query with dig (default "cloudflare.com")
  -h, --help            help for simple-node-health
  -p, --port int        Port for the web server (default 8080)
      --verbose         verbose output

Use "simple-node-health [command] --help" for more information about a command.
```

Inspect its default settings

```shell
./simple-node-health settings
	domain: cloudflare.com
	verbose: false
	port: 8080
```

Run the DNS example

```shell
./simple-node-health check dns
{
  "response": [
    "104.16.133.229",
    "104.16.132.229"
  ]
}
```

## Liveliness Check

The `/ready` endpoint can be used to check for liveliness of the web application. (It does not need authentication)

```shell
$> curl http://localhost:8080/ready
{"status":"ok"}
```

## Web OAUTH2 Tokens

This web application is designed to be used with OAUTH2 tokens. It has a built-in token request endpoint and a token check endpoint. The `/token` request endpoint can be used to request a token with a `client_id` and `client_secret`. 

Use the CLI to create new clients 

```shell
$> simple-node-health create-client
New client_id and client_secret added:
client_id: 0ea7386e827d0a33
client_secret: f72029251b56cf0b730e989f1af77c03
```

The token `/check` endpoint can be used to check the validity of the token. Use the header `Authorization: Bearer <access_token>` when make calls to any of the secured endpoints. 


```shell
$> curl -X POST -d "client_id=81573e4c363622a6&client_secret=85e26051190016e5f3edb9d15c9803a9" http://localhost:8080/token
```

```json
{"access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfaWQiOiI4MTU3M2U0YzM2MzYyMmE2IiwiZXhwIjoxNzI1Mzg3MTQ1fQ.f0OVB8ehLvDgak4-JcrHI-vblrkV9YLJpnmnJKFQSJY", "token_type": "Bearer", "expires_in": 3600} 
```

```shell
$> curl -H "Authorization: Bearer not-a-real-token" http://localhost:8080/check
Unauthorized: Invalid token
```

```shell
$> curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfaWQiOiI4MTU3M2U0YzM2MzYyMmE2IiwiZXhwIjoxNzI1Mzg3MTQ1fQ.f0OVB8ehLvDgak4-JcrHI-vblrkV9YLJpnmnJKFQSJY" http://localhost:8080/check
```

```json
{"status":"ok"}
```

## Web 

Self document the URL routes that are available 

```json
$> ./simple-node-health show-routes
./build/simple-node-health_darwin_amd64 show-routes
2024/09/04 16:32:26 Clients size from Config: 5
{
  "routes": [
    "/",
    "/check",
    "/check/disks",
    "/check/dns",
    "/ready",
    "/token"
  ]
}
```



## Build

```shell
Product Version 2.0.0

Checking Build Dependencies ---->

Cleaning Build ---->
rm -f -rf pkg/*
rm -f -rf build/*
rm -f -rf tmp/*
rm -f -rf support/usr/local/bin/*

Building ---->
env GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/shadowbq/simple-node-health/cmd.Version=2.0.0 " -o build/simple-node-health_linux_amd64 main.go
env GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/shadowbq/simple-node-health/cmd.Version=2.0.0 " -o build/simple-node-health_darwin_amd64 main.go

Packaging ---->
cp build/simple-node-health_linux_amd64 support/usr/local/bin/simple-node-health
# Replace {{VERSION}} in the control template with the actual version
sed 's/{{VERSION}}/2.0.0/g' support/DEBIAN/control.tpl > support/DEBIAN/control
chmod 0644 support/DEBIAN/control
# Build the .deb package
dpkg-deb --build support
dpkg-deb: building package 'simple-node-health' in 'support.deb'.
mv support.deb build/simple-node-health_2.0.0_amd64.deb
```

## Support - Running as a Service

### Step 1: Copy binary and service files

Copy the `build/simple-node-health_linux_amd64` to `/usr/local/bin/` 

Change the file perms `chmod +x /usr/local/bin/simple-node-health` 

### Step 2: Create a Systemd Service File

Next, create a systemd service unit file for your Go application. Systemd service files are typically located in the /etc/systemd/system/ directory. 

Copy the `support/snh.service` to `/etc/systemd/system` 

### Step 3: Validate the Configuration of the Service File

Using a file viewer: `less /etc/systemd/system/snh.service`

### Step 4: Reload Systemd and Enable the Service

Reload the systemd manager configuration to read the new service file:

`sudo systemctl daemon-reload`

Enable the service to start on boot:

`sudo systemctl enable snh.service`

### Step 5: Start and Check the Status of the Service

Start your new service:

`sudo systemctl start snh.service`

Check the status of the service to ensure it is running correctly:

`sudo systemctl status snh.service`

### Step 6: Adjust Firewall Settings (if necessary)

If you have a firewall enabled (such as `ufw` on Ubuntu), allow traffic on the port your service is using (in this example, port `8080`):

`sudo ufw allow 8080`