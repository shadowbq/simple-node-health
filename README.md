# Simple Node Health

[![Go Report Card](https://goreportcard.com/badge/github.com/shadowbq/simple-node-health)](https://goreportcard.com/report/github.com/shadowbq/simple-node-health)
[![GoDoc](https://godoc.org/github.com/shadowbq/simple-node-health?status.svg)](https://godoc.org/github.com/shadowbq/simple-node-health)

The `simple-node-health` tool is a versatile Go-based utility designed to monitor the health of Linux systems by providing an easy-to-use HTTP server and command-line interface. It performs essential checks, such as monitoring EXT4 file systems for read-only status, conducting DNS queries, and reporting the overall status of the node. This tool is best utilized when coupled with monitoring tools like Uptime Kuma or other systems that can monitor states via web calls, allowing users to integrate these health checks into their broader monitoring setups and ensure real-time visibility into server performance and stability in production environments.

## Dev Prerequisites

* Golang

## Usage

```shell
$> ./simple-node-health --help
A simple tool to check hardware EXT4 devices and run DNS queries

Usage:
  simple-node-health [flags]
  simple-node-health [command]

Available Commands:
  check       Run various checks
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  settings    Print the current configuration settings

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
./simple-node-health check checkdns
{
  "response": [
    "104.16.133.229",
    "104.16.132.229"
  ]
}
```

## Build

```shell
make
Product Version 1.0.1

Checking Build Dependencies ---->

Cleaning Build ---->
rm -f -rf pkg/*
rm -f -rf build/*
rm -f -rf tmp/*
rm -f -rf support/usr/local/bin/*

Building ---->
env GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/shadowbq/simple-node-health/cmd.Version=1.0.1 " -o build/simple-node-health_linux_amd64 main.go
env GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/shadowbq/simple-node-health/cmd.Version=1.0.1 " -o build/simple-node-health_darwin_amd64 main.go

Packaging ---->
cp build/simple-node-health_linux_amd64 support/usr/local/bin/simple-node-health
# Replace {{VERSION}} in the control template with the actual version
sed 's/{{VERSION}}/1.0.1/g' support/DEBIAN/control.tpl > support/DEBIAN/control
chmod 0644 support/DEBIAN/control
# Build the .deb package
dpkg-deb --build support
dpkg-deb: building package 'simple-node-health' in 'support.deb'.
mv support.deb build/simple-node-health_1.0.1_amd64.deb
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