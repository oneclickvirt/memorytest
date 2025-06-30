package memory

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	. "github.com/oneclickvirt/defaultset"
	"github.com/oneclickvirt/mbw"
)

// hasRootPermission 检测是否有root权限
func hasRootPermission() bool {
	if runtime.GOOS == "windows" {
		return true // Windows环境直接返回true
	}
	return os.Getuid() == 0
}

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

// simpleMemoryTest 使用 mbw 库进行内存测试
func simpleMemoryTest(language string) string {
	if EnableLoger {
		Logger.Info("Running mbw memory test")
	}
	// 获取 mbw 命令
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
	// 如果有临时文件，确保在函数结束时清理
	if tempFile != "" {
		defer func() {
			if removeErr := os.Remove(tempFile); removeErr != nil && EnableLoger {
				Logger.Error("Failed to remove temp file: " + removeErr.Error())
			}
			// 尝试删除临时目录
			if tempDir := strings.TrimSuffix(tempFile, "/mbw-linux-amd64"); tempDir != tempFile {
				os.Remove(tempDir)
			}
		}()
	}
	// 执行 mbw 测试
	var cmd *exec.Cmd
	testSize := "256" // 默认测试 256MB
	if strings.Contains(mbwCmd, "sudo") {
		// 包含 sudo 的情况
		parts := strings.Fields(mbwCmd)
		args := append(parts[1:], "-n", "5", testSize)
		cmd = exec.Command("sudo", args...)
	} else {
		// 不包含 sudo 的情况
		cmd = exec.Command(mbwCmd, "-n", "5", testSize)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		if EnableLoger {
			Logger.Error("Failed to execute mbw: " + err.Error())
		}
		if language == "en" {
			return "Memory test execution failed\n"
		} else {
			return "内存测试执行失败\n"
		}
	}
	results := parseMBWOutput(string(output))
	var result strings.Builder
	if language == "en" {
		if speed, ok := results["MEMCPY"]; ok {
			result.WriteString(fmt.Sprintf("Memory Copy Speed (MEMCPY)   : %.2f MB/s (256 MiB)\n", speed))
		}
		if speed, ok := results["DUMB"]; ok {
			result.WriteString(fmt.Sprintf("Memory Copy Speed (DUMB)     : %.2f MB/s (256 MiB)\n", speed))
		}
		if speed, ok := results["MCBLOCK"]; ok {
			result.WriteString(fmt.Sprintf("Memory Copy Speed (MCBLOCK)  : %.2f MB/s (256 MiB)\n", speed))
		}
	} else {
		if speed, ok := results["MEMCPY"]; ok {
			result.WriteString(fmt.Sprintf("内存复制速度(读+写) (MEMCPY)   : %.2f MB/s (256 MiB)\n", speed))
		}
		if speed, ok := results["DUMB"]; ok {
			result.WriteString(fmt.Sprintf("内存复制速度(读+写) (DUMB)     : %.2f MB/s (256 MiB)\n", speed))
		}
		if speed, ok := results["MCBLOCK"]; ok {
			result.WriteString(fmt.Sprintf("内存复制速度(读+写) (MCBLOCK)  : %.2f MB/s (256 MiB)\n", speed))
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
