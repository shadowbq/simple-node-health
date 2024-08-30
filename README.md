# Simple Node Health

[![Go Report Card](https://goreportcard.com/badge/github.com/shadowbq/simple-node-health)](https://goreportcard.com/report/github.com/shadowbq/simple-node-health)
[![GoDoc](https://godoc.org/github.com/shadowbq/simple-node-health?status.svg)](https://godoc.org/github.com/shadowbq/simple-node-health)

## Dev Prerequisites

* Golang

## Usage

```shell
./simple-node-health --help
A simple tool to check hardware EXT4 devices and run DNS queries

Usage:
  simple-node-health [flags]
  simple-node-health [command]

Available Commands:
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

```
./simple-node-health settings
	domain: cloudflare.com
	verbose: false
	port: 8080
```

## Build

```shell
Make
Product Version 1.0.0

Checking Build Dependencies ---->

Cleaning Build ---->
rm -f -rf pkg/*
rm -f -rf build/*
rm -f -rf tmp/*

Building ---->
env GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/shadowbq/simple-node-health/cmd.Version=1.0.0 " -o build/simple-node-health_linux_amd64 main.go
env GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/shadowbq/simple-node-health/cmd.Version=1.0.0 " -o build/simple-node-health_darwin_amd64 main.go
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