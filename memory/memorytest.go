package memory

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

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
	cmd2 = exec.Command("sudo", "dd", "if="+ifKey, "of="+ofKey, "bs="+bs, "count="+blockCount)
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

// parseOutput 解析结果
func parseOutput(tempText, language string, records float64) (string, error) {
	var result string
	tp1 := strings.Split(tempText, "\n")
	for _, t := range tp1 {
		if strings.Contains(t, "bytes") || strings.Contains(t, "字节") {
			var tp2 []string
			if strings.Contains(t, "bytes") {
				tp2 = strings.Split(t, ",")
			} else {
				tp2 = strings.Split(t, "，")
			}
			if len(tp2) == 4 {
				usageTime, err := strconv.ParseFloat(strings.TrimSpace(strings.Split(strings.TrimSpace(tp2[2]), " ")[0]), 64)
				if err != nil {
					return "", err
				}
				ioSpeed := strings.Split(strings.TrimSpace(tp2[3]), " ")[0]
				ioSpeedFlat := strings.Split(strings.TrimSpace(tp2[3]), " ")[1]
				iops := records / usageTime
				var iopsText string
				if iops >= 1000 {
					iopsText = strconv.FormatFloat(iops/1000, 'f', 2, 64) + "K IOPS, " + strconv.FormatFloat(usageTime, 'f', 2, 64) + "s"
				} else {
					iopsText = strconv.FormatFloat(iops, 'f', 2, 64) + " IOPS, " + strconv.FormatFloat(usageTime, 'f', 2, 64) + "s"
				}
				result += fmt.Sprintf("%-30s", strings.TrimSpace(ioSpeed)+" "+ioSpeedFlat+"("+iopsText+")") + "\n"
			}
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
		if err == nil {
			break
		}
		if EnableLoger && err != nil {
			Logger.Info(err.Error())
		}
		os.Remove("/dev/shm/testfile.test")
	}
	if err == nil {
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
		if EnableLoger && err != nil {
			Logger.Info(err.Error())
		}
		if err == nil {
			break
		}
		os.Remove("/tmp/testfile_read.test")
	}
	if err == nil {
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
