//go:build !(linux && amd64) && !(linux && 386) && !(linux && arm64) && !(linux && riscv64) && !(linux && mips64) && !(linux && mips64le) && !(linux && ppc64le) && !(windows && amd64) && !(windows && 386)

package memory

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	. "github.com/oneclickvirt/defaultset"
	"github.com/oneclickvirt/mbw"
)

// parseMBWOutput 解析 mbw 输出
func parseMBWOutput(output string) map[string]float64 {
	results := make(map[string]float64)
	lines := strings.Split(output, "\n")
	avgRegex := regexp.MustCompile(`AVG\s+Method:\s+(\w+)\s+.*Copy:\s+([\d.]+)\s+MiB/s`)
	for _, line := range lines {
		if matches := avgRegex.FindStringSubmatch(line); matches != nil {
			method := matches[1]
			speed, err := strconv.ParseFloat(matches[2], 64)
			if err == nil {
				results[method] = speed
			}
		}
	}
	return results
}

func simpleMemoryTest(language string) string {
	if EnableLoger {
		Logger.Info("Running mbw memory test")
	}
	mbwCmd, tempFile, err := mbw.GetMBW()
	if err != nil {
		fmt.Println(mbwCmd, tempFile, err.Error())
		if EnableLoger {
			Logger.Error("Failed to get mbw command: " + err.Error())
		}
		if language == "en" {
			return "Memory test failed: mbw command not available\n"
		} else {
			return "内存测试失败: mbw 命令不可用\n"
		}
	}
	testSizes := []string{"1024", "512", "256", "128", "64", "32"}
	var output []byte
	var execErr error
	var successfulSize string
	for _, size := range testSizes {
		var cmd *exec.Cmd
		if strings.Contains(mbwCmd, "sudo") {
			parts := strings.Fields(mbwCmd)
			args := append(parts[1:], "-n", "5", size)
			cmd = exec.Command("sudo", args...)
		} else {
			cmd = exec.Command(mbwCmd, "-n", "5", size)
		}
		output, execErr = cmd.CombinedOutput()
		if execErr == nil && !bytes.Contains(output, []byte("Cannot allocate memory")) {
			successfulSize = size
			break
		}
		if EnableLoger {
			Logger.Warn(fmt.Sprintf("Test with size %sMB failed: %v", size, execErr))
		}
	}
	if successfulSize == "" {
		if EnableLoger {
			Logger.Error("All memory test sizes failed")
		}
		if language == "en" {
			return "Memory test execution failed: insufficient memory\n"
		} else {
			return "内存测试执行失败：可用内存不足\n"
		}
	}
	results := parseMBWOutput(string(output))
	var result strings.Builder
	if language == "en" {
		if speed, ok := results["MEMCPY"]; ok {
			result.WriteString(fmt.Sprintf("Memory Copy Speed (MEMCPY)   : %10.2f MB/s \n", speed))
		}
		if speed, ok := results["DUMB"]; ok {
			result.WriteString(fmt.Sprintf("Memory Copy Speed (DUMB)     : %10.2f MB/s \n", speed))
		}
		if speed, ok := results["MCBLOCK"]; ok {
			result.WriteString(fmt.Sprintf("Memory Copy Speed (MCBLOCK)  : %10.2f MB/s \n", speed))
		}
	} else {
		if speed, ok := results["MEMCPY"]; ok {
			result.WriteString(fmt.Sprintf("内存复制速度(读+写) (MEMCPY)   : %10.2f MB/s \n", speed))
		}
		if speed, ok := results["DUMB"]; ok {
			result.WriteString(fmt.Sprintf("内存复制速度(读+写) (DUMB)     : %10.2f MB/s \n", speed))
		}
		if speed, ok := results["MCBLOCK"]; ok {
			result.WriteString(fmt.Sprintf("内存复制速度(读+写) (MCBLOCK)  : %10.2f MB/s \n", speed))
		}
	}
	if len(results) == 0 {
		if EnableLoger {
			Logger.Info("No results parsed from mbw output, returning raw output")
		}
		result.WriteString("\nRaw output:\n")
		result.WriteString(string(output))
	}
	return result.String()
}
