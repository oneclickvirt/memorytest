package memory

import (
	"fmt"
	"io"
	"os"
	"os/exec"
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
		// 统一的结果处理函数
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
		// 读测试
		readResult, err := runSysBenchCommand("1", "read", "5", version)
		if err != nil {
			if EnableLoger {
				Logger.Info(fmt.Sprintf("Error running read test: %v %s\n", strings.TrimSpace(readResult), err.Error()))
			}
			return ""
		} else {
			testReadOps, testReadSpeed, mibReadFlag = processResult(readResult)
		}
		time.Sleep(700 * time.Millisecond)
		// 写测试
		writeResult, err := runSysBenchCommand("1", "write", "5", version)
		if err != nil {
			if EnableLoger {
				Logger.Info(fmt.Sprintf("Error running write test: %v %s\n", strings.TrimSpace(writeResult), err.Error()))
			}
			return ""
		} else {
			testWriteOps, testWriteSpeed, mibWriteFlag = processResult(writeResult)
		}
		// 计算和匹配格式
		// 写
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
		return ""
	}
	return result
}

// execDDTest 执行dd命令测试内存IO，并回传结果和测试错误
func execDDTest(ifKey, ofKey, bs, blockCount string, isWrite bool) (string, error) {
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

func DDTest(language string) string {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	var result string
	var err error
	var tempText string
	var records float64 = 1024.0
	// Write test
	// sudo dd if=/dev/zero of=/dev/shm/testfile.test bs=1M count=1024
	sizes := []string{"1024", "128"}
	for _, size := range sizes {
		tempText, err = execDDTest("/dev/zero", "/dev/shm/testfile.test", "1M", size, true)
		defer os.Remove("/dev/shm/testfile.test")
		if EnableLoger {
			Logger.Info("Write test:")
			Logger.Info(tempText)
		}
		if err == nil {
			break
		}
		os.Remove("/dev/shm/testfile.test")
	}
	if err == nil || strings.Contains(tempText, "No space left on device") {
		writeResult, err := parseOutput(tempText, language, records)
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
			return ""
		}
	} else {
		if EnableLoger {
			Logger.Info(fmt.Sprintf("Error running write test: %v %s\n", strings.TrimSpace(tempText), err.Error()))
		}
		return ""
	}
	// Read test
	for _, size := range sizes {
		tempText, err = execDDTest("/dev/shm/testfile.test", "/dev/null", "1M", size, false)
		if err != nil || strings.Contains(tempText, "Invalid argument") || strings.Contains(tempText, "Permission denied") {
			tempText, _ = execDDTest("/dev/shm/testfile.test", "/tmp/testfile_read.test", "1M", size, false)
			defer os.Remove("/tmp/testfile_read.test")
		}
		if EnableLoger {
			Logger.Info("Read test:")
			Logger.Info(tempText)
		}
		if err == nil {
			break
		}
		os.Remove("/tmp/testfile_read.test")
	}
	if err == nil || strings.Contains(tempText, "No space left on device") {
		readResult, err := parseOutput(tempText, language, records)
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
			return ""
		}
	} else {
		if EnableLoger {
			Logger.Info(fmt.Sprintf("Error running read test: %v %s\n", strings.TrimSpace(tempText), err.Error()))
		}
		return ""
	}
	return result
}

// WinsatTest 通过 winsat 测试内存读写
func WinsatTest(language string) string {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	var result string
	cmd := exec.Command("winsat", "mem")
	output, err := cmd.Output()
	if err != nil {
		if EnableLoger {
			Logger.Info(fmt.Sprintf("Error running winsat command: %v %s\n", strings.TrimSpace(string(output)), err.Error()))
		}
		return ""
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
	sizes := []string{"1024", "128", "64"}  // 添加更小的尺寸以防内存不足
	for _, size := range sizes {
		tempText, err = execDDTest("NUL", testWriteFile, "1M", size, true)
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
		writeResult, err := parseOutput(tempText, language, records)
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
			return ""
		}
	} else {
		if EnableLoger {
			Logger.Info(fmt.Sprintf("Error running write test: %v %s\n", strings.TrimSpace(tempText), err.Error()))
		}
		return ""
	}
	// Read test - 在Windows上从临时文件读取到NUL
	for _, size := range sizes {
		tempText, err = execDDTest(testWriteFile, "NUL", "1M", size, false)
		if err != nil || strings.Contains(tempText, "Invalid argument") || strings.Contains(tempText, "Permission denied") {
			tempText, _ = execDDTest(testWriteFile, testReadFile, "1M", size, false)
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
		readResult, err := parseOutput(tempText, language, records)
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
			return ""
		}
	} else {
		if EnableLoger {
			Logger.Info(fmt.Sprintf("Error running read test: %v %s\n", strings.TrimSpace(tempText), err.Error()))
		}
		return ""
	}
	
	return result
}

// execWindowsDDTest 在Windows环境执行dd命令测试内存IO
func execWindowsDDTest(ifKey, ofKey, bs, blockCount string, isWrite bool) (string, error) {
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