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
- Linux (macOS and Windows support planned — see Roadmap)

## Download

Pre-built binaries are available on the [Releases](https://github.com/zemation/sysinfo/releases/latest) page.

| Binary | OS | Architecture |
|---|---|---|
| sysinfo-linux-amd64 | Linux | x86_64 (64-bit) |
| sysinfo-linux-arm64 | Linux | ARM64 (Raspberry Pi, etc) |

### Install from release

```bash
# Download the binary for your architecture
curl -LO https://github.com/zemation/sysinfo/releases/latest/download/sysinfo-linux-amd64

# Make it executable
chmod +x sysinfo-linux-amd64

# Install system-wide
sudo mv sysinfo-linux-amd64 /usr/local/bin/sysinfo
```

## Build from Source

```bash
git clone https://github.com/zemation/sysinfo.git
cd sysinfo
go build -o sysinfo main.go
sudo mv sysinfo /usr/local/bin/
```

## Commands

| Command | Description |
|---|---|
| `sysinfo` | System overview |
| `sysinfo all` | All information in one output |
| `sysinfo disk` | Disk usage per mount point |
| `sysinfo processes cpu` | Top processes by CPU usage |
| `sysinfo processes memory` | Top processes by memory usage |
| `sysinfo network interfaces` | Network interfaces, IPs, RX/TX stats |
| `sysinfo network ports` | Listening ports and owning processes |

## Flags

### Global Flags
Available on all commands:

```
--json, -j    Output as JSON
```

### Command Flags

```
sysinfo processes --count, -c    Number of processes to show (default 5)
```

## Color Output

When running in a terminal, sysinfo color codes usage values automatically:

| Color | Meaning |
|---|---|
| 🟢 Green | Normal — below 70% |
| 🟡 Yellow | Warning — above 70% |
| 🔴 Red | Critical — above 90% |

Color applies to:
- `sysinfo disk` — USE% column
- `sysinfo processes cpu` — CPU% column
- `sysinfo processes memory` — MEM% column

Color is automatically disabled when piping output or redirecting to a file so it does not interfere with tools like `jq`.

## Usage Examples

```bash
# System overview
sysinfo

# All info at once
sysinfo all

# Disk usage per mount
sysinfo disk

# Top processes (default 5)
sysinfo processes cpu
sysinfo processes memory

# Custom process count
sysinfo processes cpu --count 10
sysinfo processes memory -c 3

# Network
sysinfo network interfaces
sysinfo network ports

# Requires sudo for full port ownership output
sudo sysinfo network ports

# JSON output (color disabled automatically)
sysinfo --json
sysinfo disk --json
sysinfo processes cpu --count 10 --json
sysinfo network interfaces --json

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
    ├── color.go         # ANSI color output helpers
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

---

## Roadmap

### Cross-Platform Support
Currently sysinfo is Linux-only. The plan is to add macOS and Windows support using Go build tags — platform-specific files that the compiler picks automatically based on the target OS.

```
system/
├── info_linux.go       # current implementation
├── info_darwin.go      # macOS implementation (planned)
└── info_windows.go     # Windows implementation (planned)
```

#### macOS
macOS is the closer port. Most functionality can be implemented using `sysctl` system calls and `/usr/bin/sw_vers`. Disk usage already works since macOS supports `syscall.Statfs`.

| Feature | Approach |
|---|---|
| OS / Kernel | sw_vers + uname |
| CPU | sysctl hw.model, hw.logicalcpu |
| Memory | sysctl hw.memsize, vm.page_free_count |
| Disk | syscall.Statfs (already works) |
| Uptime | sysctl kern.boottime |
| Load | sysctl vm.loadavg |
| Processes | sysctl + kinfo_proc |
| Network ports | net.inet.tcp (sysctl) |

#### Windows
Windows requires a different approach — WMI (Windows Management Instrumentation) via the `github.com/StackExchange/wmi` package for most system data.

| Feature | Approach |
|---|---|
| OS / Kernel | WMI Win32_OperatingSystem |
| CPU | WMI Win32_Processor |
| Memory | WMI Win32_OperatingSystem |
| Disk | syscall.GetDiskFreeSpaceEx |
| Uptime | WMI Win32_OperatingSystem.LastBootUpTime |
| Load | Not natively available on Windows |
| Processes | WMI Win32_Process |
| Network ports | netstat via WMI or net.tcp registry |
