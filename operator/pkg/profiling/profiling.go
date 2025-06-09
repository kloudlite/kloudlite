package profiling

import (
	"time"

	"github.com/pkg/profile"
)

// ProfileCPU starts CPU profiling, and stops after the given duration.
func ProfileCPU(duration time.Duration) {
	profiler := profile.Start(profile.CPUProfile)
	time.AfterFunc(duration, func() {
		profiler.Stop()
	})
}

// ProfileCPU starts Memory profiling, and stops after the given duration.
func ProfileMem(duration time.Duration) {
	profiler := profile.Start(profile.MemProfile)
	time.AfterFunc(duration, func() {
		profiler.Stop()
	})
}
