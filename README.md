# sysinfo

A lightweight system information CLI tool written in Go. No external dependencies for data collection — reads directly from Linux virtual filesystems.

## Example Output

```
Host:    sam.example.com
IP:      192.168.0.10
OS:      Rocky Linux 10.1 (Red Quartz)
Kernel:  6.12.0-124.56.1.el10_1.x86_64
Arch:    amd64
CPU:     AMD Ryzen 7 3700X 8-Core Processor (16 cores)
Memory:  5.4 GB / 31.0 GB
Disk:    41.0 GB / 69.9 GB
Uptime:  10 days, 20 hours, 28 minutes
Load:    0.08, 0.13, 0.12 (1, 5, 15 min)
```

## Requirements

- Go 1.20 or higher
- Linux (reads from /proc and /sys)

## Build

```bash
git clone https://github.com/zemation/sysinfo.git
cd sysinfo
go build -o sysinfo main.go
```

## Install

```bash
sudo mv sysinfo /usr/local/bin/
```

## Commands

| Command | Description |
|---|---|
| `sysinfo` | System overview |
| `sysinfo all` | All information in one output |
| `sysinfo disk` | Disk usage per mount point |
| `sysinfo processes cpu` | Top 5 processes by CPU usage |
| `sysinfo processes memory` | Top 5 processes by memory usage |
| `sysinfo network interfaces` | Network interfaces, IPs, RX/TX stats |
| `sysinfo network ports` | Listening ports and owning processes |

## Global Flags

```
--json, -j    Output as JSON
```

Works on all commands:

```bash
sysinfo --json
sysinfo disk --json
sysinfo processes cpu --json
sysinfo network interfaces --json
```

## Usage Examples

```bash
# System overview
sysinfo

# All info at once
sysinfo all

# Disk usage per mount
sysinfo disk

# Top processes
sysinfo processes cpu
sysinfo processes memory

# Network
sysinfo network interfaces
sysinfo network ports

# Requires sudo for full output (port ownership)
sudo sysinfo network ports

# Pipe JSON to jq
sysinfo --json | jq '.cpu'
sysinfo processes memory --json | jq '.[0]'
```

## Project Structure

```
sysinfo/
├── main.go              # Entry point
├── cmd/
│   ├── root.go          # Default sysinfo command
│   ├── all.go           # sysinfo all
│   ├── disk.go          # sysinfo disk
│   ├── processes.go     # sysinfo processes
│   └── network.go       # sysinfo network
└── system/
    ├── info.go          # System info functions
    ├── disk.go          # Disk mount functions
    ├── processes.go     # Process listing and sorting
    └── network.go       # Network interface and port functions
```

## Data Sources

| Field | Source |
|---|---|
| Host | os.Hostname() |
| IP | net.InterfaceAddrs() |
| OS | /etc/os-release |
| Kernel | /proc/version |
| Arch | runtime.GOARCH |
| CPU | /proc/cpuinfo |
| Memory | /proc/meminfo |
| Disk | syscall.Statfs |
| Uptime | /proc/uptime |
| Load | /proc/loadavg |
| Network stats | /sys/class/net/[iface]/statistics |
| Listening ports | /proc/net/tcp, /proc/net/udp |
| Process info | /proc/[pid]/stat, /proc/[pid]/comm |
