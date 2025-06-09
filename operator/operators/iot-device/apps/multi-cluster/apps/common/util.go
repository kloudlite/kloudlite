package common

import (
	"time"

	"github.com/kloudlite/operator/operators/wireguard/apps/multi-cluster/constants"
)

func ReconWait() {
	time.Sleep(constants.ReconDuration * time.Second)
}
