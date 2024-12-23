# ELP Packet Simulation Tool

This tool simulates devices sending packets to a target UDP server. It allows multiple devices to run concurrently, each sending data at configurable intervals.

## Features

- Simulates multiple devices with unique MAC addresses.
- Sends ELP packets with device, RSSI, and IR data.
- Configurable via a simple configuration file.

## Configuration

The tool uses a configuration file (`config.json`) to define runtime parameters. Below is an example of the `config.json` structure:

```
{
  "NumDevices": 5,
  "IntervalSec": 2,
  "TargetIP": "127.0.0.1",
  "TargetPort": 8080,
  "BaseMAC": "C4:CB:6B:50:00:00"
}
```

## Configuration Parameters:

- **NumberDevices**: Number of simulated devices.
- **IntervalSec**: Interval (in seconds) between packet transmissions.
- **TargetIP**: IP address of the UDP server.
- **TargetPort**: Port of the UDP server.
- **BaseMAC**: Base MAC address for devices. Each device increments the MAC.

## How to Run

1. **Setup Configuration**: Update a config.json file in the project directory following the format above.
2. **Run the Program**:
   ```
   go run main.go
   ```
3. **Output**: The program logs device packet transmission to the console.

## Flow Overview

1. **Initialization**:
   - Read configuration from `config.json`.
   - Parse the number of devices, interval, target address, and base MAC.
2. **Device Simulation**:
   - Each device is assigned a unique MAC address based on the base MAC.
   - A goroutine is created for each device to handle concurrent execution.
3. **Packet Generation**:
   - Each device generates:
     - Device data.
     - RSSI data from multiple APs.
     - IR data.
   - Packets are built with the data and a sequence number.
4. **Packet Transmission**:
   - Packets are sent to the target UDP server.
   - Logs are generated for each transmission, showing sequence number and MAC.
5. **Repeat**:
   - The process continues at the configured interval until stopped.
