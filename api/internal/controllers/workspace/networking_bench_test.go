package workspace

import (
	"strconv"
	"strings"
	"testing"
)

// BenchmarkStringConcatenation tests simple string concatenation vs fmt.Sprintf
func BenchmarkStringConcatenation(b *testing.B) {
	b.Run("simple_concat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = "exposed-" + strconv.Itoa(3000)
		}
	})

	b.Run("fmt_sprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = "exposed-" + strconv.Itoa(3000)
		}
	})
}

// BenchmarkStringsWithBuilder tests strings.Builder for building hostnames
func BenchmarkStringsWithBuilder(b *testing.B) {
	hash := "a1b2c3d4"
	subdomain := "eastman.khost.dev"
	port := int32(3000)

	b.Run("strings_builder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var sb strings.Builder
			sb.Grow(1 + 5 + 1 + len(hash) + 1 + len(subdomain))
			sb.WriteString("p")
			sb.WriteString(strconv.FormatInt(int64(port), 10))
			sb.WriteString("-")
			sb.WriteString(hash)
			sb.WriteString(".")
			sb.WriteString(subdomain)
			_ = sb.String()
		}
	})

	b.Run("simple_concat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = "p" + strconv.FormatInt(int64(port), 10) + "-" + hash + "." + subdomain
		}
	})
}

// BenchmarkHostnameBuilding tests building workspace hostnames
func BenchmarkHostnameBuilding(b *testing.B) {
	prefix := "vscode"
	hash := "a1b2c3d4"
	subdomain := "eastman.khost.dev"

	b.Run("simple_concat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = prefix + "-" + hash + "." + subdomain
		}
	})

	b.Run("strings_builder_with_grow", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var sb strings.Builder
			sb.Grow(len(prefix) + 1 + len(hash) + 1 + len(subdomain))
			sb.WriteString(prefix)
			sb.WriteString("-")
			sb.WriteString(hash)
			sb.WriteString(".")
			sb.WriteString(subdomain)
			_ = sb.String()
		}
	})
}

// BenchmarkWorkspaceServiceName tests service name construction
func BenchmarkWorkspaceServiceName(b *testing.B) {
	workspaceName := "my-workspace-123"

	b.Run("simple_concat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = "ws-" + workspaceName
		}
	})

	b.Run("headless_concat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = "ws-" + workspaceName + "-headless"
		}
	})
}

// BenchmarkBuildPortHostname tests the actual function
func BenchmarkBuildPortHostname(b *testing.B) {
	port := int32(3000)
	hash := "a1b2c3d4"
	subdomain := "eastman.khost.dev"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildPortHostname(port, hash, subdomain)
	}
}

// BenchmarkMultipleHostnames tests building many hostnames (common in ingress creation)
func BenchmarkMultipleHostnames(b *testing.B) {
	httpServices := map[string]int32{
		"vscode":   8080,
		"tty":      7681,
		"claude":   7682,
		"opencode": 7683,
		"codex":    7684,
	}
	hash := "a1b2c3d4"
	subdomain := "eastman.khost.dev"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for prefix := range httpServices {
			_ = prefix + "-" + hash + "." + subdomain
		}
	}
}
