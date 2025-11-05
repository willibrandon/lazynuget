package platform

import (
	"testing"
)

// BenchmarkNormalize tests path normalization performance
// Target: <1ms per FR-010
// See: T098, FR-010
func BenchmarkNormalize(b *testing.B) {
	platform, err := New()
	if err != nil {
		b.Fatalf("Failed to create platform: %v", err)
	}
	resolver, err := NewPathResolver(platform)
	if err != nil {
		b.Fatalf("Failed to create path resolver: %v", err)
	}

	tests := []struct {
		name string
		path string
	}{
		{"simple_relative", "configs/app.yml"},
		{"nested_relative", "../../configs/nested/deep/app.yml"},
		{"absolute_unix", "/usr/local/bin/myapp"},
		{"absolute_windows", "C:\\Users\\test\\AppData\\Local\\myapp"},
		{"mixed_separators", "configs\\subdir/file.txt"},
		{"dots_in_path", "./configs/../data/./file.txt"},
		{"long_path", "very/long/path/with/many/segments/to/test/performance/configs/nested/deep/structure/app.yml"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = resolver.Normalize(tt.path)
			}
		})
	}
}

// BenchmarkValidate tests path validation performance
// See: T098, FR-011
func BenchmarkValidate(b *testing.B) {
	platform, err := New()
	if err != nil {
		b.Fatalf("Failed to create platform: %v", err)
	}
	resolver, err := NewPathResolver(platform)
	if err != nil {
		b.Fatalf("Failed to create path resolver: %v", err)
	}

	tests := []struct {
		name string
		path string
	}{
		{"simple_valid", "configs/app.yml"},
		{"complex_valid", "../../configs/nested/deep/app.yml"},
		{"absolute_valid", "/usr/local/bin/myapp"},
		{"invalid_null_byte", "path\x00with\x00null"},
		{"invalid_control_chars", "path\nwith\rcontrol"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = resolver.Validate(tt.path)
			}
		})
	}
}

// BenchmarkIsAbsolute tests absolute path detection performance
// See: T098
func BenchmarkIsAbsolute(b *testing.B) {
	platform, err := New()
	if err != nil {
		b.Fatalf("Failed to create platform: %v", err)
	}
	resolver, err := NewPathResolver(platform)
	if err != nil {
		b.Fatalf("Failed to create path resolver: %v", err)
	}

	tests := []struct {
		name string
		path string
	}{
		{"relative", "configs/app.yml"},
		{"unix_absolute", "/usr/local/bin/myapp"},
		{"windows_absolute", "C:\\Users\\test\\AppData"},
		{"unc_path", "\\\\server\\share\\file"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = resolver.IsAbsolute(tt.path)
			}
		})
	}
}

// BenchmarkResolve tests relative path resolution performance
// See: T098
func BenchmarkResolve(b *testing.B) {
	platform, err := New()
	if err != nil {
		b.Fatalf("Failed to create platform: %v", err)
	}
	resolver, err := NewPathResolver(platform)
	if err != nil {
		b.Fatalf("Failed to create path resolver: %v", err)
	}

	tests := []struct {
		name string
		path string
	}{
		{"simple_relative", "app.yml"},
		{"nested_relative", "configs/nested/app.yml"},
		{"dot_path", "./configs/app.yml"},
		{"dotdot_path", "../configs/app.yml"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_, _ = resolver.Resolve(tt.path)
			}
		})
	}
}

// BenchmarkConfigDir tests config directory retrieval performance
// See: T098
func BenchmarkConfigDir(b *testing.B) {
	platform, err := New()
	if err != nil {
		b.Fatalf("Failed to create platform: %v", err)
	}
	resolver, err := NewPathResolver(platform)
	if err != nil {
		b.Fatalf("Failed to create path resolver: %v", err)
	}

	b.ReportAllocs()
	for b.Loop() {
		_, _ = resolver.ConfigDir()
	}
}

// BenchmarkCacheDir tests cache directory retrieval performance
// See: T098
func BenchmarkCacheDir(b *testing.B) {
	platform, err := New()
	if err != nil {
		b.Fatalf("Failed to create platform: %v", err)
	}
	resolver, err := NewPathResolver(platform)
	if err != nil {
		b.Fatalf("Failed to create path resolver: %v", err)
	}

	b.ReportAllocs()
	for b.Loop() {
		_, _ = resolver.CacheDir()
	}
}
