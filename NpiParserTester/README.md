# NPI Test Tool

A testing utility designed to simulate NPI (Network Processor Interface) packet transmission. This tool generates and sends iBeacon advertisement packets over TCP connections, simulating multiple tags and infrastructure devices.

## Purpose

This application helps test and validate NPI parser systems by:
- Simulating multiple infrastructure devices and tags
- Generating valid NPI packets with proper checksums
- Sending packets at specified intervals
- Incrementing sequence numbers automatically
- Simulating real device behavior

## Features

- Configure number of tags and infrastructure devices
- Set custom transmission intervals
- Automatic sequence number management (1-65000)
- Proper MAC address handling
- Configurable host and port
- Simulates device data including:
  - Temperature
  - Battery level
  - Humidity
  - Motion status
  - Safety switch status
  - Charger status

## Usage

### Running the Application

```bash
go run main.go [flags]
```

### Available Flags

- `-host`: Target host address (default: "127.0.0.1")
- `-port`: Target port number (default: "8888")
- `-tags`: Number of tags to simulate (default: 1)
- `-infra`: Number of infrastructure devices (default: 1)
- `-interval`: Interval between packets in seconds (default: 1)

### Example

```bash
go run main.go -host 127.0.0.1 -port 8888 -tags 5 -infra 2 -interval 1
```

This command will:
- Connect to localhost on port 8888
- Simulate 5 different tags
- Use 2 infrastructure devices
- Send packets every 1 second

## Packet Flow

1. Each infrastructure device creates its own TCP connection
2. For each tag, at every interval:
   - Creates NPI packet with proper format
   - Includes iBeacon advertisement data
   - Increments sequence numbers
   - Sends packet through TCP connection

## MAC Address Format

- Infrastructure MACs: `C4:CB:6B:50:00:XX` (XX increments for each device)
- Tag MACs: `C4:CB:6B:23:00:XX` (XX increments for each tag)

## Building

```bash
go build -o npi-test
```
