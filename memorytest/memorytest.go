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
		fmt.Println("Running read test...")
		readResult, err := runSysBenchCommand("1", "read", "5", version)
		if err != nil {
			fmt.Printf("Error running read test: %v\n", err)
			fmt.Println(readResult)
		} else {
			fmt.Println("Read test result:")
			fmt.Println(readResult)
		}
		fmt.Println("Running write test...")
		writeResult, err := runSysBenchCommand("1", "write", "5", version)
		if err != nil {
			fmt.Printf("Error running write test: %v\n", err)
		} else {
			fmt.Println("Write test result:")
			fmt.Println(writeResult)
		}
	}
	return result
}
