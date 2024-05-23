package memorytest

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
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
		// 测试两次取平均值
		for count := 0; count < 2; count++ {
			// 读测试
			readResult, err := runSysBenchCommand("1", "read", "5", version)
			if err != nil {
				fmt.Printf("Error running read test: %v\n", strings.TrimSpace(readResult))
			} else {
				tempList := strings.Split(readResult, "\n")
				if len(tempList) > 0 {
					for _, line := range tempList {
						if strings.Contains(line, "total size") {
							totalSize = strings.TrimSpace(strings.Split(line, ":")[1])
							if strings.Contains(totalSize, "MiB") {
								mibReadFlag = true
							}
						} else if strings.Contains(line, "per second") || strings.Contains(line, "ops/sec") {
							temp1 := strings.Split(line, "(")
							if len(temp1) == 2 {
								temp2 := strings.Split(strings.TrimSpace(temp1[1]), " ")
								if len(temp2) >= 2 {
									value, err := strconv.ParseFloat(strings.TrimSpace(temp2[0]), 64)
									if err == nil {
										testReadOps += value
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
										testReadSpeed += value
									}
								}
							}
						}
					}
				}
			}
			// 写测试
			writeResult, err := runSysBenchCommand("1", "write", "5", version)
			if err != nil {
				fmt.Printf("Error running write test: %v\n", strings.TrimSpace(writeResult))
			} else {
				tempList := strings.Split(readResult, "\n")
				if len(tempList) > 0 {
					// https://github.com/spiritLHLS/ecs/blob/641724ccd98c21bb1168e26efb349df54dee0fa1/ecs.sh#L2143
					for _, line := range tempList {
						if strings.Contains(line, "total size") {
							totalSize = strings.TrimSpace(strings.Split(line, ":")[1])
							if strings.Contains(totalSize, "MiB") {
								mibWriteFlag = true
							}
						} else if strings.Contains(line, "per second") || strings.Contains(line, "ops/sec") {
							temp1 := strings.Split(line, "(")
							if len(temp1) == 2 {
								temp2 := strings.Split(temp1[1], " ")
								if len(temp2) >= 2 {
									value, err := strconv.ParseFloat(strings.TrimSpace(temp2[0]), 64)
									if err == nil {
										testWriteOps += value
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
										testWriteSpeed += value
									}
								}
							}
						}
					}
				}
			}
		}
		// 计算和匹配格式
		if mibReadFlag {
			testReadSpeed = testReadSpeed / 2.0 / 1048576.0 * 1000000.0
		}
		if language == "en" {
			result += "  Single Read Speed: "
		} else {
			result += "  单线程读速度: "
		}
		testReadSpeedStr := strconv.FormatFloat(testReadSpeed, 'f', 2, 64)
		if testReadOps > 1000 {
			testReadOpsStr := strconv.FormatFloat(testReadOps/1000.0, 'f', 2, 64)
			result += testReadSpeedStr + " MB/s(" + testReadOpsStr + "K IOPS)\n"
		} else {
			testReadOpsStr := strconv.FormatFloat(testReadOps, 'f', 0, 64)
			result += testReadSpeedStr + " MB/s(" + testReadOpsStr + " IOPS)\n"
		}
		if mibWriteFlag {
			testWriteSpeed = testWriteSpeed / 2.0 / 1048576 * 1000000
		}
		if language == "en" {
			result += "  Single Write Speed: "
		} else {
			result += "  单线程写速度: "
		}
		testWriteSpeedStr := strconv.FormatFloat(testWriteSpeed, 'f', 2, 64)
		if testWriteOps > 1000 {
			testWriteOpsStr := strconv.FormatFloat(testWriteOps/1000.0, 'f', 2, 64)
			result += testWriteSpeedStr + " MB/s(" + testWriteOpsStr + "K IOPS)\n"
		} else {
			testWriteOpsStr := strconv.FormatFloat(testWriteOps, 'f', 0, 64)
			result += testWriteSpeedStr + " MB/s(" + testWriteOpsStr + " IOPS)\n"
		}
	} else {
		return ""
	}
	return result
}

// DDTest 通过 dd 测试内存读写
func DDTest(language string) string {
	var result string
	return result
}

// WinsatTest 通过 winsat 测试内存读写
func WinsatTest(language string) string {
	var result string
	return result
}
