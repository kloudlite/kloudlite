package common

import (
	"time"

	"github.com/kloudlite/operator/apps/multi-cluster/constants"
)

func ReconWait() {
	time.Sleep(constants.ReconDuration * time.Second)
}
