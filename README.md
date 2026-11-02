# sysinfo

A lightweight, cross-platform system information CLI tool written in Go. Minimal external dependencies — reads directly from Linux virtual filesystems, macOS system utilities/syscalls, and Windows Win32 APIs.

## Example Output

```
Host:    host.example.com
IP:      192.168.0.10
OS:      Rocky Linux 10.1 (Red Quartz)
Kernel:  6.12.0-124.56.1.el10_1.x86_64
Arch:    amd64
CPU:     AMD Ryzen 7 3700X 8-Core Processor (16 cores)
Memory:  5.4 GB / 31.0 GB
Disk:    41.0 GB / 69.9 GB
Uptime:  10 days, 20 hours, 28 minutes
Load:    0.08, 0.13, 0.12 (1, 5, 15 min)
GPU:     NVIDIA GeForce RTX 3080
```

## Requirements

- Go 1.22 or higher (when building from source)
- Linux, macOS, or Windows

## Download

Pre-built binaries are available on the [Releases](https://github.com/zemation/sysinfo/releases/latest) page.

| Binary | OS | Architecture |
|---|---|---|
| `sysinfo-linux-amd64` | Linux | x86_64 (64-bit) |
| `sysinfo-linux-arm64` | Linux | ARM64 (Raspberry Pi, etc) |
| `sysinfo-darwin-amd64` | macOS | Intel (x86_64) |
| `sysinfo-darwin-arm64` | macOS | Apple Silicon (M1/M2/M3/M4) |
| `sysinfo-windows-amd64.exe` | Windows | x86_64 (64-bit) |
| `sysinfo-windows-arm64.exe` | Windows | ARM64 |

### Install with Go (All Platforms)

```bash
go install github.com/zemation/sysinfo@latest
```

### Install from release (Linux / macOS)

```bash
# Download the binary for your OS and architecture (e.g., linux-amd64 or darwin-arm64)
curl -LO https://github.com/zemation/sysinfo/releases/latest/download/sysinfo-linux-amd64

# Make it executable
chmod +x sysinfo-linux-amd64

# Install system-wide
sudo mv sysinfo-linux-amd64 /usr/local/bin/sysinfo
```

### Install from release (Windows PowerShell)

```powershell
# Download the binary directly into WindowsApps (already on your user PATH)
Invoke-WebRequest -Uri https://github.com/zemation/sysinfo/releases/latest/download/sysinfo-windows-amd64.exe -OutFile "$env:LOCALAPPDATA\Microsoft\WindowsApps\sysinfo.exe"

# Run sysinfo from anywhere
sysinfo
```

## Build from Source

```bash
git clone https://github.com/zemation/sysinfo.git
cd sysinfo
go build -o sysinfo main.go
```

On Windows:
```powershell
git clone https://github.com/zemation/sysinfo.git
cd sysinfo
go build -o sysinfo.exe main.go
```

## Commands

| Command | Description |
|---|---|
| `sysinfo` | System overview |
| `sysinfo all` | All information in one output |
| `sysinfo gpu` | GPU model, driver version, VRAM usage, and utilization |
| `sysinfo disk` | Disk usage per mount point / drive |
| `sysinfo processes cpu` | Top processes by CPU usage |
| `sysinfo processes memory` | Top processes by memory usage |
| `sysinfo network interfaces` | Network interfaces, IPs, RX/TX stats |
| `sysinfo network ports` | Listening ports and owning processes |

## Flags

### Global Flags
Available on all commands:

```
--json, -j        Output as JSON
--watch, -w       Continuously refresh output in real time
--interval, -i    Refresh interval in seconds when using --watch (default 2)
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
| 🟡 Yellow | Warning — 70% to 89% |
| 🔴 Red | Critical — 90% and above |

Color applies to:
- `sysinfo disk` — `USE%` column
- `sysinfo gpu` — `Utilization` value
- `sysinfo processes cpu` — `CPU%` column
- `sysinfo processes memory` — `MEM%` column
- `sysinfo all` — `USE%`, `Utilization`, `CPU%`, and `MEM%` values

Color is automatically disabled when piping output or redirecting to a file so it does not interfere with tools like `jq`.

## Usage Examples

```bash
# System overview
sysinfo

# All info at once
sysinfo all

# GPU details, VRAM, and utilization
sysinfo gpu

# Live watch mode (refreshes every 2s by default, or custom interval with -i)
sysinfo --watch
sysinfo gpu -w -i 1
sysinfo processes cpu -w

# Disk usage per mount / drive
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

# Elevated privileges may be required for full port ownership output on Linux/macOS
sudo sysinfo network ports

# JSON output (color disabled automatically)
sysinfo --json
sysinfo all --json
sysinfo gpu --json
sysinfo disk --json
sysinfo processes cpu --count 10 --json
sysinfo network interfaces --json

# Pipe JSON to jq
sysinfo --json | jq '.cpu'
sysinfo gpu --json | jq '.utilization'
sysinfo processes memory --json | jq '.[0]'
```

## Project Structure

```
sysinfo/
├── main.go                      # Entry point
├── cmd/
│   ├── root.go                  # Default sysinfo command & --watch loop
│   ├── all.go                   # sysinfo all
│   ├── gpu.go                   # sysinfo gpu
│   ├── disk.go                  # sysinfo disk
│   ├── processes.go             # sysinfo processes
│   └── network.go               # sysinfo network
└── system/
    ├── common.go                # Shared cross-platform helpers
    ├── types.go                 # Shared data structures & JSON tags
    ├── color.go                 # ANSI color output & column padding helpers
    ├── color_windows.go         # Windows Virtual Terminal ANSI enablement
    ├── info_{linux,darwin,windows}.go
    ├── disk_{linux,darwin,windows}.go
    ├── gpu_{linux,darwin,windows}.go
    ├── processes_{linux,darwin,windows}.go
    └── network_{linux,darwin,windows}.go
```

## Data Sources

| Field | Linux | macOS (`darwin`) | Windows |
|---|---|---|---|
| Host | `os.Hostname()` | `os.Hostname()` | `os.Hostname()` |
| IP | `net.InterfaceAddrs()` | `net.InterfaceAddrs()` | `net.InterfaceAddrs()` |
| OS | `/etc/os-release` | `sw_vers` | Registry `CurrentVersion` |
| Kernel | `/proc/version` | `uname -r` | Registry `CurrentBuildNumber` / `UBR` |
| Arch | `runtime.GOARCH` | `runtime.GOARCH` | `runtime.GOARCH` |
| CPU | `/proc/cpuinfo` | `sysctl machdep.cpu` | Registry `CentralProcessor\0` + `runtime.NumCPU()` |
| Memory | `/proc/meminfo` | `sysctl hw.memsize` + `vm_stat` | `GlobalMemoryStatusEx` (`kernel32.dll`) |
| Disk | `syscall.Statfs` & `/proc/mounts` | `syscall.Getfsstat` | `GetLogicalDrives` + `GetDiskFreeSpaceExW` |
| Uptime | `/proc/uptime` | `sysctl kern.boottime` | `GetTickCount64` (`kernel32.dll`) |
| Load | `/proc/loadavg` | `sysctl vm.loadavg` | `n/a (Windows)` |
| GPU | `nvidia-smi`, `rocm-smi`, `/sys/class/drm` | `system_profiler SPDisplaysDataType` | `nvidia-smi` / Display Class Registry |
| Network stats | `/sys/class/net/[iface]/statistics` | `netstat -I [iface] -b` | `GetIfEntry2` (`iphlpapi.dll`) |
| Listening ports | `/proc/net/{tcp,tcp6,udp,udp6}` + `/proc/[pid]/fd` | `lsof -nP` | `GetExtendedTcpTable` & `GetExtendedUdpTable` (`iphlpapi.dll`) |
| Process info | `/proc/[pid]/stat`, `/proc/[pid]/comm` | `ps axo pid=,%cpu=,%mem=,comm=` | `CreateToolhelp32Snapshot`, `GetProcessMemoryInfo`, `GetProcessTimes` |

---

## Roadmap

### Cross-Platform Support
- [x] **Linux** (`amd64`, `arm64`) — Direct `/proc` and `/sys` virtual filesystem readers
- [x] **macOS** (`amd64`, `arm64`) — Native `syscall.Getfsstat` and system utilities (`sysctl`, `vm_stat`, `lsof`, `ps`)
- [x] **Windows** (`amd64`, `arm64`) — Native Win32 API syscalls (`kernel32.dll`, `iphlpapi.dll`, `psapi.dll`) and Windows Registry

### Features & Enhancements
- [x] Dedicated `sysinfo gpu` subcommand exposing VRAM usage, driver version, and GPU utilization
- [x] Live watch / refresh mode (`--watch` / `-w` and `--interval` / `-i`)
- [x] Native Win32 `GetExtendedTcpTable` / `GetExtendedUdpTable` for Windows listening ports
- [ ] Package manager distributions (Homebrew tap for macOS/Linux, Winget / Scoop manifests for Windows)
- [ ] Real-time delta CPU usage sampling for processes (`--sample` flag)
- [ ] Temperature & sensor telemetry (`sysinfo sensors`)
