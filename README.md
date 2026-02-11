# IOT MQTT Packet Sender

A collection of testing utilities for simulating IoT device communication over UDP, TCP, and MQTT-related workflows.  
This project is designed for packet generation, network interface testing, and local infrastructure validation.

---

## Overview

This repository contains multiple tools for:

- Simulating IoT device packet transmission
- Sending hardcoded TCP byte streams
- Testing UDP listener behavior
- Simulating NPI (Network Processor Interface) communication
- Generating and sending test packets for infrastructure validation

These tools are primarily intended for development, QA testing, and network debugging scenarios.

---

## Components

### 1. ELP Packet Simulation Tool

Simulates devices sending packets to a target UDP server.

**Features:**
- Multiple concurrent device simulation
- Configurable send intervals
- UDP packet transmission
- Local log support

Use case:
- Load testing UDP servers
- Simulating edge device behavior
- Infrastructure validation

---

### 2. NPI Test Tool

A testing utility to simulate NPI (Network Processor Interface) packet transmission.

**Features:**
- Generates iBeacon advertisement packets
- Sends packets over TCP connections
- Simulates multiple tags and infrastructure devices
- Designed for integration and system-level testing

Use case:
- NPI communication validation
- Beacon packet simulation
- Network processor testing

---

### 3. SendTcpBytes

Utility to send hardcoded TCP byte packets to a specified endpoint.

Use case:
- Raw TCP packet testing
- Debugging custom protocol implementations
- Integration testing

---

### 4. UDPListenerTester

Tool for testing UDP listener implementations.

Use case:
- Verifying UDP server behavior
- Packet reception validation
- Network debugging

---

## Requirements

- Go (recommended 1.18+)
- Network access to target test server
- Proper firewall configuration for UDP/TCP testing

---

## Installation

Clone the repository:

```bash
git clone https://github.com/<your-username>/IOT-mqtt-packet-sender.git
cd IOT-mqtt-packet-sender
```

---

## Running the Tools

Each tool is located in its respective folder. Navigate to the desired directory and run:

```bash
go run main.go
```

Or build:

```bash
go build
./binary-name
```

Refer to individual folder documentation (if available) for configuration parameters and usage details.

---

## Configuration

Most tools allow configuration of:

- Target host
- Target port
- Packet interval
- Device simulation count
- Logging options

Modify configuration within each tool’s source files as required.

---

## Logging

Some components support local logging of transmitted or received packets for debugging and verification.

---

## Use Cases

- IoT device simulation
- MQTT backend testing
- UDP/TCP packet debugging
- Beacon and NPI protocol testing
- Infrastructure load testing

---

## Disclaimer

This project is intended for testing and development purposes only.  
Ensure you have authorization before sending packets to any production systems.

---
