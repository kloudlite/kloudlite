package entities

type MemoryUsage struct {
	Quantity float64 `json:"quantity,omitempty"`
	Unit     string  `json:"unit,omitempty"`
}

type CPUUsage struct {
	Quantity float64 `json:"quantity,omitempty"`
	Unit     string  `json:"unit,omitempty"`
}

type ComputePlan struct {
	Provider string      `json:"provider,omitempty"`
	Name     string      `json:"name,omitempty"`
	Region   string      `json:"region,omitempty"`
	Memory   MemoryUsage `json:"memory"`
	Cpu      CPUUsage    `json:"cpu"`
}
