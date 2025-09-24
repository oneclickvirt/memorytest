package memory

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	. "github.com/oneclickvirt/defaultset"
	"github.com/oneclickvirt/stream"
)

// StreamTest 使用 stream 进行内存测试 (cross-platform)
func StreamTest(language string) string {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
		Logger.Info("Running STREAM memory test")
	}

	var streamCmd string
	var tempFile string
	var err error

	// First try to get stream binary from the embedded library
	streamCmd, tempFile, err = stream.GetStream()
	if err != nil {
		if EnableLoger {
			Logger.Warn(fmt.Sprintf("Failed to get stream from library: %v, trying system binaries", err))
		}
		
		// Fallback to original logic - try different stream binary names based on architecture and OS
		var streamBinaries []string
		if runtime.GOOS == "windows" {
			streamBinaries = []string{
				"./stream-windows-amd64.exe",
				"./stream.exe",
				"stream-windows-amd64.exe",
				"stream.exe",
				"./stream-windows-amd64",
				"./stream",
				"stream-windows-amd64",
				"stream",
			}
		} else {
			streamBinaries = []string{
				"./stream-linux-amd64",
				"./stream",
				"stream-linux-amd64",
				"stream",
			}
		}

		for _, binary := range streamBinaries {
			if _, err := os.Stat(binary); err == nil {
				streamCmd = binary
				break
			}
			// Also check if it's available in PATH
			if _, err := exec.LookPath(binary); err == nil {
				streamCmd = binary
				break
			}
		}

		if streamCmd == "" {
			if EnableLoger {
				Logger.Warn("STREAM binary not found, falling back to alternative test")
			}
			return "" // Return empty to indicate fallback needed
		}
	}

	// Clean up temporary file if it was created
	if tempFile != "" {
		defer func() {
			if cleanErr := stream.CleanStream(tempFile); cleanErr != nil && EnableLoger {
				Logger.Warn(fmt.Sprintf("Failed to clean stream temp file: %v", cleanErr))
			}
		}()
	}

	// Execute STREAM test
	cmd := exec.Command(streamCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if EnableLoger {
			Logger.Error(fmt.Sprintf("STREAM test failed: %v", err))
		}
		return "" // Return empty to indicate fallback needed
	}

	// Parse STREAM output to extract the Function section
	return parseStreamOutput(string(output), language)
}

// parseStreamOutput 解析 STREAM 输出，提取 Function 部分
func parseStreamOutput(output, language string) string {
	lines := strings.Split(output, "\n")
	var result strings.Builder
	
	// Find the start and end of the Function section
	inFunctionSection := false
	functionHeaderFound := false
	
	for _, line := range lines {
		// Look for the Function header line
		if strings.Contains(line, "Function") && strings.Contains(line, "Best Rate MB/s") {
			functionHeaderFound = true
			inFunctionSection = true
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}
		
		// If we found the header, keep collecting lines until we hit a line of dashes or empty line
		if functionHeaderFound && inFunctionSection {
			trimmedLine := strings.TrimSpace(line)
			// Stop when we encounter the ending dashes or validation line
			if strings.HasPrefix(trimmedLine, "---") || strings.Contains(trimmedLine, "Solution Validates") {
				break
			}
			// Skip empty lines at the beginning, but include data lines
			if trimmedLine != "" {
				result.WriteString(line)
				result.WriteString("\n")
			}
		}
	}
	
	if !functionHeaderFound {
		if EnableLoger {
			Logger.Error("Could not parse STREAM output - Function section not found")
		}
		return "" // Return empty to indicate parsing failed
	}
	
	return result.String()
}