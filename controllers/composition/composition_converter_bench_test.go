package composition

import (
	"strconv"
	"strings"
	"testing"
)

// BenchmarkVolumeNameConstruction tests volume name building for config files
func BenchmarkVolumeNameConstruction(b *testing.B) {
	filename := "nginx.conf"

	b.Run("simple_concat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = "env-file-" + filename
		}
	})

	b.Run("strings_builder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var sb strings.Builder
			sb.Grow(9 + len(filename))
			sb.WriteString("env-file-")
			sb.WriteString(filename)
			_ = sb.String()
		}
	})
}

// BenchmarkPortNameConstruction tests service port name building
func BenchmarkPortNameConstruction(b *testing.B) {
	b.Run("simple_concat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = "port-" + strconv.Itoa(i%100)
		}
	})
}

// BenchmarkCPUConversion tests CPU format conversion
func BenchmarkCPUConversion(b *testing.B) {
	nanoCPUs := float64(0.5)

	b.Run("strings_builder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			millicores := int(nanoCPUs * 1000)
			var sb strings.Builder
			sb.Grow(5)
			sb.WriteString(strconv.Itoa(millicores))
			sb.WriteString("m")
			_ = sb.String()
		}
	})

	b.Run("fmt_sprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			millicores := int(nanoCPUs * 1000)
			_ = strconv.Itoa(millicores) + "m"
		}
	})
}

// BenchmarkMultipleVolumeNames tests building many volume names
func BenchmarkMultipleVolumeNames(b *testing.B) {
	filenames := []string{"nginx.conf", "app.yaml", "redis.conf", "postgres.conf", "mongo.conf"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, filename := range filenames {
			safeFilename := strings.ReplaceAll(filename, ".", "-")
			_ = "env-file-" + safeFilename
			_ = "env-file-" + filename
		}
	}
}

// BenchmarkConvertCPU tests the actual function
func BenchmarkConvertCPU(b *testing.B) {
	testCases := []float64{0.1, 0.5, 1.0, 2.0, 4.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, cpu := range testCases {
			_ = convertCPU(cpu)
		}
	}
}
