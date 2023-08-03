package functions

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strings"
)

var generator *regexp.Regexp = regexp.MustCompile(`^([[:lower:]]|[[:digit:]]|-)*`)

// GenReadableName generates a readable name from a given string (seed).
// It finds matches in the seed following regex `^([[:lower:]]|[[:digit:]]|-)*`.
// It also limits maximum match text length to 20 (by default)
func GenReadableName(seed string, maxMatchLen ...int) string {
	if seed == "" {
		return ""
	}
	maxLength := 20
	if len(maxMatchLen) > 0 {
		maxLength = maxMatchLen[0]
	}

	sp := generator.FindStringSubmatch(strings.ToLower(seed))
	if len(sp) > 0 {
		return fmt.Sprintf("%s-%d", sp[0][:int(math.Min(float64(maxLength), float64(len(sp[0]))))], rand.Intn(999999))
	}
	return fmt.Sprintf("%s-%d", seed[0:int(math.Min(float64(10), float64(len(seed))))], rand.Intn(999999))
}

// GenValidK8sResourceNames generates a list of valid k8s resource names from a given string (seed).
// It makes use of GenReadableName to generate a readable name from the seed.
// It also checks if the generated name is valid k8s resource name.

// TODO (nxtcoder17): need to benchmark this function
func GenValidK8sResourceNames(seed string, count int) []string {
	names := make([]string, count)
	for i := 0; i < count; i++ {
		gn := GenReadableName(seed)
		for !IsValidK8sResourceName(gn) {
			gn = GenReadableName(gn)
		}
		names[i] = gn
	}
	return names
}
