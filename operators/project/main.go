package main

import (
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/project/controller"
)

func main() {
	// profiler := profile.Start(profile.MemProfile)
	// time.AfterFunc(1*time.Minute, func() {
	// 	profiler.Stop()
	// })
	mgr := operator.New("projects")
	controller.RegisterInto(mgr)
	mgr.Start()
}
