package system

type DiskInfo struct {
	Mount   string `json:"mount"`
	Total   string `json:"total"`
	Used    string `json:"used"`
	Free    string `json:"free"`
	Percent string `json:"percent"`
}

type NetworkInterface struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	RX   string `json:"rx"`
	TX   string `json:"tx"`
}

type PortInfo struct {
	Port    string `json:"port"`
	Proto   string `json:"proto"`
	State   string `json:"state"`
	PID     string `json:"pid"`
	Command string `json:"command"`
}

type Process struct {
	PID     string `json:"pid"`
	Command string `json:"command"`
	CPU     string `json:"cpu"`
	Memory  string `json:"memory"`
}

type GPUInfo struct {
	Name          string `json:"name"`
	DriverVersion string `json:"driver_version"`
	MemoryTotal   string `json:"memory_total"`
	MemoryUsed    string `json:"memory_used"`
	Utilization   string `json:"utilization"`
}
