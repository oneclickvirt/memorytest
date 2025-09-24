package memory

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/oneclickvirt/dd"
	. "github.com/oneclickvirt/defaultset"
)

// WinsatTest 通过 winsat 测试内存读写
func WinsatTest(language string) string {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	if !hasRootPermission() {
		if language == "en" {
			fmt.Println("Current system detected no admin permission")
		} else {
			fmt.Println("当前检测到系统无admin权限")
		}
		// Try STREAM first when no admin permission, fallback to mbw if not available
		streamResult := StreamTest(language)
		if streamResult != "" && strings.TrimSpace(streamResult) != "" {
			return streamResult
		}
		return simpleMemoryTest(language)
	}
	var result string
	cmd := exec.Command("winsat", "mem")
	output, err := cmd.Output()
	if err != nil {
		if EnableLoger {
			Logger.Info(fmt.Sprintf("Error running winsat command: %v %s\n", strings.TrimSpace(string(output)), err.Error()))
		}
		return simpleMemoryTest(language)
	} else {
		tempList := strings.Split(string(output), "\n")
		for _, l := range tempList {
			if strings.Contains(l, "MB/s") {
				tempL := strings.Split(l, " ")
				tempText := strings.TrimSpace(tempL[len(tempL)-2])
				if language == "en" {
					result += "Memory Performance: "
				} else {
					result += "内存性能: "
				}
				result += tempText + "MB/s" + "\n"
			}
		}
	}
	return result
}

// WindowsDDTest 在Windows环境下使用dd命令测试内存IO
func WindowsDDTest(language string) string {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	if !hasRootPermission() {
		if language == "en" {
			fmt.Println("Current system detected no admin permission")
		} else {
			fmt.Println("当前检测到系统无admin权限")
		}
		// Try STREAM first when no admin permission, fallback to mbw if not available
		streamResult := StreamTest(language)
		if streamResult != "" && strings.TrimSpace(streamResult) != "" {
			return streamResult
		}
		return simpleMemoryTest(language)
	}
	var result string
	var err error
	var tempText string
	var records float64 = 1024.0
	// 创建临时文件路径
	tempDir := os.Getenv("TEMP")
	if tempDir == "" {
		tempDir = "C:\\Windows\\Temp"
	}
	testWriteFile := tempDir + "\\testfile.test"
	testReadFile := tempDir + "\\testfile_read.test"
	// 确保清理临时文件
	defer os.Remove(testWriteFile)
	defer os.Remove(testReadFile)
	// Write test - 在Windows上使用%TEMP%目录
	// dd if=/dev/zero of=%TEMP%\testfile.test bs=1M count=1024
	sizes := []string{"1024", "128", "64"}
	for _, size := range sizes {
		tempText, err = execWindowsDDTest("NUL", testWriteFile, "1M", size, true)
		if EnableLoger {
			Logger.Info("Write test:")
			Logger.Info(tempText)
		}
		if err == nil {
			break
		}
		os.Remove(testWriteFile)
	}
	if err == nil || strings.Contains(tempText, "No space left on device") {
		writeResult, err := parseWindowsOutput(tempText, language, records)
		if err == nil {
			if language == "en" {
				result += "Single Seq Write Speed: "
			} else {
				result += "单线程顺序写速度: "
			}
			result += writeResult
		} else {
			if EnableLoger {
				Logger.Info(fmt.Sprintf("Error parsing write test: %v\n", err.Error()))
			}
			return simpleMemoryTest(language)
		}
	} else {
		if EnableLoger {
			Logger.Info(fmt.Sprintf("Error running write test: %v %s\n", strings.TrimSpace(tempText), err.Error()))
		}
		return simpleMemoryTest(language)
	}
	// Read test - 在Windows上从临时文件读取到NUL
	for _, size := range sizes {
		tempText, err = execWindowsDDTest(testWriteFile, "NUL", "1M", size, false)
		if err != nil || strings.Contains(tempText, "Invalid argument") || strings.Contains(tempText, "Permission denied") {
			tempText, _ = execWindowsDDTest(testWriteFile, testReadFile, "1M", size, false)
		}
		if EnableLoger {
			Logger.Info("Read test:")
			Logger.Info(tempText)
		}
		if err == nil {
			break
		}
		os.Remove(testReadFile)
	}
	if err == nil || strings.Contains(tempText, "No space left on device") {
		readResult, err := parseWindowsOutput(tempText, language, records)
		if err == nil {
			if language == "en" {
				result += "Single Seq Read  Speed: "
			} else {
				result += "单线程顺序读速度: "
			}
			result += readResult
		} else {
			if EnableLoger {
				Logger.Info(fmt.Sprintf("Error parsing read test: %v\n", err.Error()))
			}
			return simpleMemoryTest(language)
		}
	} else {
		if EnableLoger {
			Logger.Info(fmt.Sprintf("Error running read test: %v %s\n", strings.TrimSpace(tempText), err.Error()))
		}
		return simpleMemoryTest(language)
	}
	return result
}

// execWindowsDDTest 在Windows环境执行dd命令测试内存IO
func execWindowsDDTest(ifKey, ofKey, bs, blockCount string, isWrite bool) (string, error) {
	_ = isWrite
	var tempText string
	var cmd2 *exec.Cmd
	// 获取dd命令路径，Windows环境可能需要额外检查
	ddCmd, ddPath, err := dd.GetDD()
	defer dd.CleanDD(ddPath)
	if err != nil {
		return "", err
	}
	if ddCmd == "" {
		return "", fmt.Errorf("execWindowsDDTest: ddCmd is NULL")
	}
	// Windows环境下调整路径格式和命令
	ifKey = strings.ReplaceAll(ifKey, "/", "\\")
	ofKey = strings.ReplaceAll(ofKey, "/", "\\")
	parts := strings.Split(ddCmd, " ")
	cmd2 = exec.Command(parts[0], append(parts[1:], "if="+ifKey, "of="+ofKey, "bs="+bs, "count="+blockCount)...)
	stderr2, err := cmd2.StderrPipe()
	if err == nil {
		if err := cmd2.Start(); err == nil {
			outputBytes, err := io.ReadAll(stderr2)
			if err == nil {
				tempText = string(outputBytes)
			} else {
				return "", err
			}
		} else {
			return "", err
		}
	} else {
		return "", err
	}
	return tempText, nil
}

// parseWindowsOutput 解析Windows环境下dd命令输出结果
func parseWindowsOutput(tempText, language string, records float64) (string, error) {
	_ = language
	var result string
	// Windows环境下dd命令输出格式可能与Linux不同，需要适当调整
	lines := strings.Split(tempText, "\n")
	for _, line := range lines {
		// 匹配Windows环境下dd输出的性能数据行
		if strings.Contains(line, "bytes") || strings.Contains(line, "字节") {
			var separator string
			if strings.Contains(line, "bytes") {
				separator = ","
			} else {
				separator = "，"
			}
			parts := strings.Split(line, separator)
			if len(parts) < 3 || len(parts) > 4 {
				continue // Windows下格式可能有差异，跳过不匹配的行
			}
			// 尝试提取时间和速度信息
			timeIndex := -1
			speedIndex := -1
			for i, part := range parts {
				if strings.Contains(part, "sec") || strings.Contains(part, "秒") {
					timeIndex = i
				}
				if strings.Contains(part, "bytes/sec") || strings.Contains(part, "字节/秒") {
					speedIndex = i
				}
			}
			if timeIndex == -1 || speedIndex == -1 {
				continue
			}
			// 解析使用时间
			usageTime, err := parseUsageTime(parts[timeIndex])
			if err != nil {
				continue
			}
			// 解析IO速度
			ioSpeed, ioSpeedFlat, err := parseIOSpeed(parts[speedIndex])
			if err != nil {
				continue
			}
			// 计算IOPS
			iops := records / usageTime
			iopsText := formatIOPS(iops, usageTime)
			// 格式化结果
			result += fmt.Sprintf("%-30s\n", strings.TrimSpace(ioSpeed)+" "+ioSpeedFlat+"("+iopsText+")")
			break
		}
	}
	if result == "" {
		return "", fmt.Errorf("无法解析dd命令输出")
	}
	return result, nil
}
