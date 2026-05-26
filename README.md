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

## Usage

```bash
sysinfo
```

## What It Shows

| Field | Source |
|---|---|
| Host | os.Hostname() |
| IP | Primary non-loopback IPv4 address |
| OS | /etc/os-release |
| Kernel | /proc/version |
| Arch | Go runtime |
| CPU | /proc/cpuinfo |
| Memory | /proc/meminfo |
| Disk | syscall.Statfs on / |
| Uptime | /proc/uptime |
| Load | /proc/loadavg |

