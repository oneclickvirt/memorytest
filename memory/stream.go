package memory

import (
	"fmt"
	"os/exec"
	"strings"

	. "github.com/oneclickvirt/defaultset"
	"github.com/oneclickvirt/stream"
)

// StreamTest 使用 stream 进行内存测试
func StreamTest(language string) string {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
		Logger.Info("Running STREAM memory test")
	}
	streamCmd, tempFile, err := stream.GetStream()
	if err != nil {
		if EnableLoger {
			Logger.Error(fmt.Sprintf("Failed to get stream binary: %v", err))
		}
		return ""
	}
	if tempFile != "" {
		defer func() {
			if cleanErr := stream.CleanStream(tempFile); cleanErr != nil && EnableLoger {
				Logger.Warn(fmt.Sprintf("Failed to clean stream temp file: %v", cleanErr))
			}
		}()
	}
	cmd := exec.Command(streamCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if EnableLoger {
			Logger.Error(fmt.Sprintf("STREAM test failed: %v", err))
		}
		return ""
	}
	return parseStreamOutput(string(output), language)
}

// parseStreamOutput 解析 STREAM 输出
func parseStreamOutput(output, language string) string {
	lines := strings.Split(output, "\n")
	var result strings.Builder
	inFunctionSection := false
	functionHeaderFound := false
	for _, line := range lines {
		if strings.Contains(line, "Function") && strings.Contains(line, "Best Rate MB/s") {
			functionHeaderFound = true
			inFunctionSection = true
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}
		if functionHeaderFound && inFunctionSection {
			trimmedLine := strings.TrimSpace(line)
			if strings.HasPrefix(trimmedLine, "---") || strings.Contains(trimmedLine, "Solution Validates") {
				break
			}
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
		return ""
	}
	return result.String()
}
