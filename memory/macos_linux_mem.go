package memory

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/oneclickvirt/dd"
	. "github.com/oneclickvirt/defaultset"
)

// runSysBenchCommand 执行 sysbench 命令进行测试
func runSysBenchCommand(numThreads, oper, maxTime, version string) (string, error) {
	// version <= 1.0.17
	// 读测试
	// sysbench --test=memory --num-threads=1 --memory-block-size=1M --memory-total-size=102400G --memory-oper=read --max-time=5 --memory-access-mode=seq run 2>&1
	// 写测试
	// sysbench --test=memory --num-threads=1 --memory-block-size=1M --memory-total-size=102400G --memory-oper=write --max-time=5 --memory-access-mode=seq run 2>&1
	// version >= 1.0.18
	// 读测试
	// sysbench memory --num-threads=1 --memory-block-size=1M --memory-total-size=102400G --memory-oper=read --time=5 --memory-access-mode=seq run 2>&1
	// 写测试
	// sysbench memory --num-threads=1 --memory-block-size=1M --memory-total-size=102400G --memory-oper=write --time=5 --memory-access-mode=seq run 2>&1
	// memory options:
	//  --memory-block-size=SIZE    size of memory block for test [1K]
	//  --memory-total-size=SIZE    total size of data to transfer [100G]
	//  --memory-scope=STRING       memory access scope {global,local} [global]
	//  --memory-hugetlb[=on|off]   allocate memory from HugeTLB pool [off]
	//  --memory-oper=STRING        type of memory operations {read, write, none} [write]
	//  --memory-access-mode=STRING memory access mode {seq,rnd} [seq]
	var command *exec.Cmd
	if strings.Contains(version, "1.0.18") || strings.Contains(version, "1.0.19") || strings.Contains(version, "1.0.20") {
		command = exec.Command("sysbench", "memory", "--num-threads="+numThreads, "--memory-block-size=1M", "--memory-total-size=102400G", "--memory-oper="+oper, "--time="+maxTime, "--memory-access-mode=seq", "run")
	} else {
		command = exec.Command("sysbench", "--test=memory", "--num-threads="+numThreads, "--memory-block-size=1M", "--memory-total-size=102400G", "--memory-oper="+oper, "--max-time="+maxTime, "--memory-access-mode=seq", "run")
	}
	output, err := command.CombinedOutput()
	return string(output), err
}

// SysBenchTest 使用 sysbench 进行内存测试
// https://github.com/spiritLHLS/ecs/blob/641724ccd98c21bb1168e26efb349df54dee0fa1/ecs.sh#L2143
func SysBenchTest(language string) string {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	if !hasRootPermission() {
		if language == "en" {
			fmt.Println("Current system detected no root permission")
		} else {
			fmt.Println("当前检测到系统无root权限")
		}
		return simpleMemoryTest(language)
	}
	var result string
	comCheck := exec.Command("sysbench", "--version")
	output, err := comCheck.CombinedOutput()
	if err == nil {
		version := string(output)
		var (
			totalSize                                                string
			testReadOps, testReadSpeed, testWriteOps, testWriteSpeed float64
			mibReadFlag, mibWriteFlag                                bool
		)
		processResult := func(result string) (float64, float64, bool) {
			var ops, speed float64
			var mibFlag bool
			tempList := strings.Split(result, "\n")
			if len(tempList) > 0 {
				for _, line := range tempList {
					if strings.Contains(line, "total size") {
						totalSize = strings.TrimSpace(strings.Split(line, ":")[1])
						if strings.Contains(totalSize, "MiB") {
							mibFlag = true
						}
					} else if strings.Contains(line, "per second") || strings.Contains(line, "ops/sec") {
						temp1 := strings.Split(line, "(")
						if len(temp1) == 2 {
							temp2 := strings.Split(strings.TrimSpace(temp1[1]), " ")
							if len(temp2) >= 2 {
								value, err := strconv.ParseFloat(strings.TrimSpace(temp2[0]), 64)
								if err == nil {
									ops = value
								}
							}
						}
					} else if strings.Contains(line, "MB/sec") || strings.Contains(line, "MiB/sec") {
						temp1 := strings.Split(line, "(")
						if len(temp1) == 2 {
							temp2 := strings.Split(strings.TrimSpace(temp1[1]), " ")
							if len(temp2) >= 2 {
								value, err := strconv.ParseFloat(strings.TrimSpace(temp2[0]), 64)
								if err == nil {
									speed = value
								}
							}
						}
					}
				}
			}
			return ops, speed, mibFlag
		}
		readResult, err := runSysBenchCommand("1", "read", "5", version)
		if err != nil {
			if EnableLoger {
				Logger.Info(fmt.Sprintf("Error running read test: %v %s\n", strings.TrimSpace(readResult), err.Error()))
			}
			return simpleMemoryTest(language)
		} else {
			testReadOps, testReadSpeed, mibReadFlag = processResult(readResult)
		}
		time.Sleep(700 * time.Millisecond)
		writeResult, err := runSysBenchCommand("1", "write", "5", version)
		if err != nil {
			if EnableLoger {
				Logger.Info(fmt.Sprintf("Error running write test: %v %s\n", strings.TrimSpace(writeResult), err.Error()))
			}
			return simpleMemoryTest(language)
		} else {
			testWriteOps, testWriteSpeed, mibWriteFlag = processResult(writeResult)
		}
		if mibWriteFlag {
			testWriteSpeed = testWriteSpeed / 1048576 * 1000000
		}
		if language == "en" {
			result += "Single Seq Write Speed: "
		} else {
			result += "单线程顺序写速度: "
		}
		testWriteSpeedStr := strconv.FormatFloat(testWriteSpeed, 'f', 2, 64)
		if testWriteOps > 1000 {
			testWriteOpsStr := strconv.FormatFloat(testWriteOps/1000.0, 'f', 2, 64)
			result += testWriteSpeedStr + " MB/s(" + testWriteOpsStr + "K IOPS, 5s)\n"
		} else {
			testWriteOpsStr := strconv.FormatFloat(testWriteOps, 'f', 0, 64)
			result += testWriteSpeedStr + " MB/s(" + testWriteOpsStr + " IOPS, 5s)\n"
		}
		// 读
		if mibReadFlag {
			testReadSpeed = testReadSpeed / 1048576.0 * 1000000.0
		}
		if language == "en" {
			result += "Single Seq Read  Speed: "
		} else {
			result += "单线程顺序读速度: "
		}
		testReadSpeedStr := strconv.FormatFloat(testReadSpeed, 'f', 2, 64)
		if testReadOps > 1000 {
			testReadOpsStr := strconv.FormatFloat(testReadOps/1000.0, 'f', 2, 64)
			result += testReadSpeedStr + " MB/s(" + testReadOpsStr + "K IOPS, 5s)\n"
		} else {
			testReadOpsStr := strconv.FormatFloat(testReadOps, 'f', 0, 64)
			result += testReadSpeedStr + " MB/s(" + testReadOpsStr + " IOPS, 5s)\n"
		}
	} else {
		if EnableLoger {
			Logger.Info("cannot match sysbench command: " + err.Error())
		}
		return simpleMemoryTest(language)
	}
	return result
}

// execDDTest 执行dd命令测试内存IO，并回传结果和测试错误
func execDDTest(ifKey, ofKey, bs, blockCount string, isWrite bool) (string, error) {
	_ = isWrite
	var tempText string
	var cmd2 *exec.Cmd
	ddCmd, ddPath, err := dd.GetDD()
	defer dd.CleanDD(ddPath)
	if err != nil {
		return "", err
	}
	if ddCmd == "" {
		return "", fmt.Errorf("execDDTest: ddCmd is NULL")
	}
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

func parseUsageTime(s string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(strings.Fields(strings.TrimSpace(s))[0]), 64)
}

func parseIOSpeed(s string) (string, string, error) {
	parts := strings.Fields(strings.TrimSpace(s))
	if len(parts) < 2 {
		return "", "", fmt.Errorf("IO速度格式错误")
	}
	return parts[0], parts[1], nil
}

func formatIOPS(iops, usageTime float64) string {
	if iops >= 1000 {
		return fmt.Sprintf("%.2fK IOPS, %.2fs", iops/1000, usageTime)
	}
	return fmt.Sprintf("%.2f IOPS, %.2fs", iops, usageTime)
}

// parseOutput 解析结果
func parseOutput(tempText, language string, records float64) (string, error) {
	_ = language
	var result string
	lines := strings.Split(tempText, "\n")
	for _, line := range lines {
		if strings.Contains(line, "bytes") || strings.Contains(line, "字节") {
			var separator string
			if strings.Contains(line, "bytes") {
				separator = ","
			} else {
				separator = "，"
			}
			parts := strings.Split(line, separator)
			if len(parts) < 3 || len(parts) > 4 {
				return "", fmt.Errorf("意外的dd输出格式")
			}
			timeIndex := len(parts) - 2
			speedIndex := len(parts) - 1
			usageTime, err := parseUsageTime(parts[timeIndex])
			if err != nil {
				return "", err
			}
			ioSpeed, ioSpeedFlat, err := parseIOSpeed(parts[speedIndex])
			if err != nil {
				return "", err
			}
			iops := records / usageTime
			iopsText := formatIOPS(iops, usageTime)
			result += fmt.Sprintf("%-30s\n", strings.TrimSpace(ioSpeed)+" "+ioSpeedFlat+"("+iopsText+")")
		}
	}
	return result, nil
}

// 检查是否为真正的内存文件系统
func isRealMemoryFS(path string) bool {
	// 对于Linux，只有/dev/shm才是真正的tmpfs内存文件系统
	if runtime.GOOS == "linux" {
		return path == "/dev/shm"
	}
	// 对于macOS，RAM磁盘是真正的内存存储
	if runtime.GOOS == "darwin" {
		return strings.Contains(path, "/Volumes/RAMDisk")
	}
	return false
}

// 创建macOS RAM磁盘
func createMacOSRamDisk() (string, string, error) {
	// 创建512MB的RAM磁盘
	cmd := exec.Command("hdiutil", "attach", "-nomount", "ram://1048576") // 512MB = 1048576 * 512 bytes
	output, err := cmd.Output()
	if err != nil {
		return "", "", err
	}
	diskName := strings.TrimSpace(string(output))
	// 格式化RAM磁盘
	mountPoint := "/Volumes/RAMDisk"
	cmd = exec.Command("diskutil", "erasevolume", "HFS+", "RAMDisk", diskName)
	if err := cmd.Run(); err != nil {
		// 如果格式化失败，清理已创建的磁盘
		exec.Command("diskutil", "eject", diskName).Run()
		return "", "", err
	}
	return mountPoint, diskName, nil
}

// 卸载macOS RAM磁盘
func unmountMacOSRamDisk(mountPoint, diskName string) error {
	var lastErr error
	// 通过挂载点卸载
	cmd := exec.Command("diskutil", "unmount", mountPoint)
	if err := cmd.Run(); err != nil {
		lastErr = err
		// 强制卸载挂载点
		cmd = exec.Command("diskutil", "unmount", "force", mountPoint)
		if err := cmd.Run(); err != nil {
			lastErr = err
			// 方法3: 通过设备名弹出
			if diskName != "" {
				cmd = exec.Command("diskutil", "eject", diskName)
				if err := cmd.Run(); err != nil {
					lastErr = err
				} else {
					return nil // 成功
				}
			}
			// 通过hdiutil分离
			if diskName != "" {
				cmd = exec.Command("hdiutil", "detach", diskName)
				if err := cmd.Run(); err != nil {
					lastErr = err
				} else {
					return nil // 成功
				}
			}
		} else {
			return nil // 成功
		}
	} else {
		return nil // 成功
	}
	return lastErr
}

// 获取内存文件系统目录
func getMemoryDir() (string, func(), error) {
	switch runtime.GOOS {
	case "darwin":
		// macOS: 创建RAM磁盘
		ramDisk, diskName, err := createMacOSRamDisk()
		if err != nil {
			// 如果创建RAM磁盘失败，返回错误
			return "", nil, fmt.Errorf("failed to create RAM disk: %v", err)
		}
		cleanup := func() {
			if err := unmountMacOSRamDisk(ramDisk, diskName); err != nil {
				fmt.Printf("Warning: Failed to cleanup RAM disk: %v\n", err)
			}
		}
		return ramDisk, cleanup, nil
	case "linux":
		// Linux: 检查/dev/shm是否存在且可用
		if _, err := os.Stat("/dev/shm"); err == nil {
			// 验证/dev/shm确实是tmpfs
			return "/dev/shm", func() {}, nil
		}
		return "", nil, fmt.Errorf("/dev/shm not available")

	default:
		return "", nil, fmt.Errorf("unsupported OS for memory testing")
	}
}

func DDTest(language string) string {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	if !hasRootPermission() {
		if language == "en" {
			fmt.Println("Current system detected no root permission")
		} else {
			fmt.Println("当前检测到系统无root权限")
		}
		return simpleMemoryTest(language)
	}
	// 获取内存文件系统目录
	memoryDir, cleanup, err := getMemoryDir()
	if err != nil {
		if EnableLoger {
			Logger.Info(fmt.Sprintf("Failed to get memory directory: %v", err))
		}
		return simpleMemoryTest(language)
	}
	// 验证确实是真正的内存文件系统
	if !isRealMemoryFS(memoryDir) {
		if EnableLoger {
			Logger.Info(fmt.Sprintf("Directory %s is not a real memory filesystem", memoryDir))
		}
		cleanup() // 清理已创建的资源
		return simpleMemoryTest(language)
	}
	defer cleanup()
	var result string
	var tempText string
	var records float64 = 1024.0
	testFileName := filepath.Join(memoryDir, fmt.Sprintf("testfile_%d.test", time.Now().UnixNano()))
	readTestFileName := fmt.Sprintf("/tmp/testfile_read_%d.test", time.Now().UnixNano())
	defer func() {
		os.Remove(testFileName)
		os.Remove(readTestFileName)
	}()
	// Write test
	sizes := []string{"1024", "128"}
	writeSuccess := false
	for _, size := range sizes {
		os.Remove(testFileName)
		tempText, err = execDDTest("/dev/zero", testFileName, "1M", size, true)
		if EnableLoger {
			Logger.Info("Write test:")
			Logger.Info(tempText)
		}
		if err == nil {
			writeSuccess = true
			break
		}
		os.Remove(testFileName)
	}
	if !writeSuccess {
		if EnableLoger {
			Logger.Info(fmt.Sprintf("Error running write test: %v %s\n", strings.TrimSpace(tempText), err.Error()))
		}
		return simpleMemoryTest(language)
	}
	if err == nil || strings.Contains(tempText, "No space left on device") {
		writeResult, parseErr := parseOutput(tempText, language, records)
		if parseErr == nil {
			if language == "en" {
				result += "Single Seq Write Speed: "
			} else {
				result += "单线程顺序写速度: "
			}
			result += writeResult
		} else {
			if EnableLoger {
				Logger.Info(fmt.Sprintf("Error parsing write test: %v\n", parseErr.Error()))
			}
			return simpleMemoryTest(language)
		}
	} else {
		if EnableLoger {
			Logger.Info(fmt.Sprintf("Error running write test: %v %s\n", strings.TrimSpace(tempText), err.Error()))
		}
		return simpleMemoryTest(language)
	}
	// Read test
	readSuccess := false
	for _, size := range sizes {
		tempText, err = execDDTest(testFileName, "/dev/null", "1M", size, false)
		if err != nil || strings.Contains(tempText, "Invalid argument") || strings.Contains(tempText, "Permission denied") {
			os.Remove(readTestFileName)
			tempText, err = execDDTest(testFileName, readTestFileName, "1M", size, false)
		}
		if EnableLoger {
			Logger.Info("Read test:")
			Logger.Info(tempText)
		}
		if err == nil {
			readSuccess = true
			break
		}
		os.Remove(readTestFileName)
	}
	if !readSuccess {
		if EnableLoger {
			Logger.Info(fmt.Sprintf("Error running read test: %v %s\n", strings.TrimSpace(tempText), err.Error()))
		}
		return simpleMemoryTest(language)
	}
	if err == nil || strings.Contains(tempText, "No space left on device") {
		readResult, parseErr := parseOutput(tempText, language, records)
		if parseErr == nil {
			if language == "en" {
				result += "Single Seq Read  Speed: "
			} else {
				result += "单线程顺序读速度: "
			}
			result += readResult
		} else {
			if EnableLoger {
				Logger.Info(fmt.Sprintf("Error parsing read test: %v\n", parseErr.Error()))
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

// StreamTest 使用 stream 进行内存测试
func StreamTest(language string) string {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
		Logger.Info("Running STREAM memory test")
	}

	// Try different stream binary names based on architecture
	streamBinaries := []string{
		"./stream-linux-amd64",
		"./stream",
		"stream-linux-amd64",
		"stream",
	}

	var streamCmd string
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
			Logger.Warn("STREAM binary not found, falling back to DD test")
		}
		return "" // Return empty to indicate fallback needed
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
