package env

import (
	"os"
	"testing"
)

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue int
		expected     int
		shouldSetEnv bool
	}{
		{
			name:         "returns environment variable when set",
			key:          "TEST_PORT",
			envValue:     "8080",
			defaultValue: 3000,
			expected:     8080,
			shouldSetEnv: true,
		},
		{
			name:         "returns default when env not set",
			key:          "MISSING_PORT",
			envValue:     "",
			defaultValue: 3000,
			expected:     3000,
			shouldSetEnv: false,
		},
		{
			name:         "returns default when env value is invalid",
			key:          "INVALID_PORT",
			envValue:     "not-a-number",
			defaultValue: 3000,
			expected:     3000,
			shouldSetEnv: true,
		},
		{
			name:         "handles zero value",
			key:          "ZERO_PORT",
			envValue:     "0",
			defaultValue: 3000,
			expected:     0,
			shouldSetEnv: true,
		},
		{
			name:         "handles negative value",
			key:          "NEGATIVE_PORT",
			envValue:     "-1",
			defaultValue: 3000,
			expected:     -1,
			shouldSetEnv: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: set or unset environment variable
			if tt.shouldSetEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			// Execute
			result := GetEnvInt(tt.key, tt.defaultValue)

			// Assert
			if result != tt.expected {
				t.Errorf("GetEnvInt(%q, %d) = %d; want %d",
					tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvString(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue string
		expected     string
		shouldSetEnv bool
	}{
		{
			name:         "returns environment variable when set",
			key:          "TEST_SECRET",
			envValue:     "my-secret-key",
			defaultValue: "default-secret",
			expected:     "my-secret-key",
			shouldSetEnv: true,
		},
		{
			name:         "returns default when env not set",
			key:          "MISSING_SECRET",
			envValue:     "",
			defaultValue: "default-secret",
			expected:     "default-secret",
			shouldSetEnv: false,
		},
		{
			name:         "handles empty string value",
			key:          "EMPTY_SECRET",
			envValue:     "",
			defaultValue: "default-secret",
			expected:     "",
			shouldSetEnv: true,
		},
		{
			name:         "handles whitespace value",
			key:          "WHITESPACE_SECRET",
			envValue:     "   ",
			defaultValue: "default-secret",
			expected:     "   ",
			shouldSetEnv: true,
		},
		{
			name:         "handles special characters",
			key:          "SPECIAL_SECRET",
			envValue:     "secret!@#$%^&*()",
			defaultValue: "default-secret",
			expected:     "secret!@#$%^&*()",
			shouldSetEnv: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: set or unset environment variable
			if tt.shouldSetEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			// Execute
			result := GetEnvString(tt.key, tt.defaultValue)

			// Assert
			if result != tt.expected {
				t.Errorf("GetEnvString(%q, %q) = %q; want %q",
					tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

// TestGetEnvIntConcurrent tests thread safety
func TestGetEnvIntConcurrent(t *testing.T) {
	key := "CONCURRENT_TEST_PORT"
	os.Setenv(key, "5000")
	defer os.Unsetenv(key)

	done := make(chan bool)
	for range 10 {
		go func() {
			result := GetEnvInt(key, 3000)
			if result != 5000 {
				t.Errorf("Expected 5000, got %d", result)
			}
			done <- true
		}()
	}

	for range 10 {
		<-done
	}
}

// TestGetEnvStringConcurrent tests thread safety
func TestGetEnvStringConcurrent(t *testing.T) {
	key := "CONCURRENT_TEST_SECRET"
	expected := "concurrent-secret"
	os.Setenv(key, expected)
	defer os.Unsetenv(key)

	done := make(chan bool)
	for range 10 {
		go func() {
			result := GetEnvString(key, "default")
			if result != expected {
				t.Errorf("Expected %q, got %q", expected, result)
			}
			done <- true
		}()
	}

	for range 10 {
		<-done
	}
}

// Benchmark tests
func BenchmarkGetEnvInt(b *testing.B) {
	os.Setenv("BENCH_PORT", "8080")
	defer os.Unsetenv("BENCH_PORT")

	for b.Loop() {
		GetEnvInt("BENCH_PORT", 3000)
	}
}

func BenchmarkGetEnvString(b *testing.B) {
	os.Setenv("BENCH_SECRET", "my-secret")
	defer os.Unsetenv("BENCH_SECRET")

	for b.Loop() {
		GetEnvString("BENCH_SECRET", "default")
	}
}
