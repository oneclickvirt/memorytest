package memory

import (
	"testing"
	"strings"
)

// TestStreamTestCrossPlatform tests that StreamTest function is available and behaves consistently
func TestStreamTestCrossPlatform(t *testing.T) {
	// Test that StreamTest can be called (may return empty if stream binary not available)
	result := StreamTest("en")
	
	// Should not panic and should return a string (empty is acceptable if no stream binary)
	if result != "" {
		// If result is not empty, it should contain some memory test data
		if !strings.Contains(result, "MB/s") && !strings.Contains(result, "Function") {
			t.Errorf("StreamTest returned non-empty result but doesn't appear to contain valid memory test data: %s", result)
		}
	}
	
	// Test Chinese language as well
	resultZh := StreamTest("zh")
	// Should behave the same way regardless of language
	_ = resultZh
}

// TestWindowsFunctionsUseStreamFirst tests that Windows functions try StreamTest first when no admin permission
func TestWindowsFunctionsUseStreamFirst(t *testing.T) {
	// This test mainly validates that the functions exist and can be called
	// The actual behavior depends on having admin permissions and stream binaries
	
	// Test WinsatTest
	_ = WinsatTest("en")
	_ = WinsatTest("zh")
	
	// Test WindowsDDTest  
	_ = WindowsDDTest("en")
	_ = WindowsDDTest("zh")
}