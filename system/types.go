package system

type DiskInfo struct {
	Mount   string
	Total   string
	Used    string
	Free    string
	Percent string
}

type NetworkInterface struct {
	Name string
	IP   string
	RX   string
	TX   string
}

type PortInfo struct {
	Port    string
	Proto   string
	State   string
	PID     string
	Command string
}

type Process struct {
	PID     string
	Command string
	CPU     string
	Memory  string
}

type GPUInfo struct {
	Name          string
	DriverVersion string
	MemoryTotal   string
	MemoryUsed    string
	Utilization   string
}
