# ELP Packet Simulation Tool

This tool simulates devices sending packets to a target UDP server. It allows multiple devices to run concurrently, each sending data at configurable intervals.

## Features

- Simulates multiple devices with unique MAC addresses.
- Sends ELP packets with device, RSSI, and IR data.
- Configurable via a simple configuration file.
- Logging to both console and file.
- Command-line configuration.

## Usage

Run the program with command-line flags to configure the simulation:

### Available Flags

| Flag | Description | Default Value |
|------|-------------|---------------|
| `-devices` | Number of devices to simulate | 1 |
| `-interval` | Interval between packets (seconds) | 2 |
| `-ip` | Target IP address | 127.0.0.1 |
| `-port` | Target port number | 8552 |
| `-mac` | Base MAC address | C4:CB:6B:23:00:01 |

### Example

```
go run main.go -ip 127.0.0.1 -port 8552 -devices 1 -mac C4:CB:6B:23:00:01 -interval 5
```

This will start the simulation with 2 devices, sending packets every second to 127.0.0.1:8552, using C4:CB:6B:23:00:01 as the base MAC address.
