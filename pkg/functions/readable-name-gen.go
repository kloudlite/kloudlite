package functions

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
)

func GenReadableName(seed string) string {
	if seed == "" {
		return ""
	}

	re := regexp.MustCompile(`^(\w[a-z]+)([^a-z].*)`)
	sp := re.FindStringSubmatch(seed)
	if len(sp) > 0 {
		return fmt.Sprintf("%s-%d", sp[1], rand.Intn(999999))
	}
	return fmt.Sprintf("%s-%d", seed[0:int(math.Min(float64(10), float64(len(seed))))], rand.Intn(999999))
}
