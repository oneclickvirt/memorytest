package memorytest

import (
	"fmt"
	"os/exec"
	"strings"
)

func runSysBenchCommand(numThreads, oper, maxTime, version string) (string, error) {
	// version 1.0.16
	// 读测试
	// sysbench --test=memory --num-threads=1 --memory-block-size=1M --memory-total-size=102400G --memory-oper=read --max-time=5 --memory-access-mode=seq run 2>&1
	// 写测试
	// sysbench --test=memory --num-threads=1 --memory-block-size=1M --memory-total-size=102400G --memory-oper=write --max-time=5 --memory-access-mode=seq run 2>&1
	// version 1.0.18
	// 读测试
	// sysbench memory --num-threads=1 --memory-block-size=1M --memory-total-size=102400G --memory-oper=read --time=5 --memory-access-mode=seq run 2>&1
	// 写测试
	// sysbench memory --num-threads=1 --memory-block-size=1M --memory-total-size=102400G --memory-oper=write --time=5 --memory-access-mode=seq run 2>&1
	var command *exec.Cmd
	if strings.Contains(version, "1.0.18") || strings.Contains(version, "1.0.19") || strings.Contains(version, "1.0.20") {
		command = exec.Command("sysbench", "memory", "--num-threads="+numThreads, "--memory-block-size=1M", "--memory-total-size=102400G", "--memory-oper="+oper, "--time="+maxTime, "--memory-access-mode=seq", "run")
	} else {
		command = exec.Command("sysbench", "--test=memory", "--num-threads="+numThreads, "--memory-block-size=1M", "--memory-total-size=102400G", "--memory-oper="+oper, "--max-time="+maxTime, "--memory-access-mode=seq", "run")
	}
	output, err := command.CombinedOutput()
	return string(output), err
}

func SysBenchMemoryTest() string {
	var result string
	comCheck := exec.Command("sysbench", "--version")
	output, err := comCheck.CombinedOutput()
	if err == nil {
		version := string(output)
		readResult, err := runSysBenchCommand("1", "read", "5", version)
		if err != nil {
			fmt.Printf("Error running read test: %v\n", strings.TrimSpace(readResult))
		} else {
			tempList := strings.Split(readResult, "\n")
			if len(tempList) > 0 {
				// https://github.com/spiritLHLS/ecs/blob/641724ccd98c21bb1168e26efb349df54dee0fa1/ecs.sh#L2143
				for _, line := range tempList {
					var totalSize, testScore, testSpeed string
					if strings.Contains(line, "total size") {
						totalSize = strings.TrimSpace(strings.Split(line, ":")[1])
					} else if strings.Contains(line, "per second") || strings.Contains(line, "ops/sec") {
						temp1 := strings.Split(line, "(")
						if len(temp1) == 2 {
							temp2 := strings.Split(temp1[1], " ")
							if len(temp2) >= 2 {
								testScore = temp2[0]
							}
						}
					} else if strings.Contains(line, "MB/sec") || strings.Contains(line, "MiB/sec") {
						temp1 := strings.Split(line, "(")
						if len(temp1) == 2 {
							temp2 := strings.Split(temp1[1], " ")
							if len(temp2) >= 2 {
								testSpeed = temp2[0]
							}
						}
					}
					fmt.Println(totalSize, testScore, testSpeed)
				}
			}
			fmt.Println(readResult)
		}
		fmt.Println("Running write test...")
		writeResult, err := runSysBenchCommand("1", "write", "5", version)
		if err != nil {
			fmt.Printf("Error running write test: %v\n", strings.TrimSpace(writeResult))
		} else {
			fmt.Println(writeResult)
		}
	}
	return result
}
