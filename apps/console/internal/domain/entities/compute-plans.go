package entities

type ComputePlan struct {
	Name                  string  `yaml:"name"`
	Desc                  string  `yaml:"desc"`
	SharingEnabled        bool    `yaml:"sharingEnabled"`
	DedicatedEnabled      bool    `yaml:"dedicatedEnabled"`
	MemoryPerCPU          float64 `yaml:"memoryPerCPU"`
	MaxSharedCPUPerPod    float64 `yaml:"maxSharedCPUPerPod"`
	MaxDedicatedCPUPerPod float64 `yaml:"maxDedicatedCPUPerPod"`
}
