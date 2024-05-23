package memorytest

import (
	"fmt"
	"os/exec"
)

// sysbench --test=memory --num-threads=1 --memory-block-size=1M --memory-total-size=102400G --memory-oper=read --max-time=5 --memory-access-mode=seq run 2>&1

// sysbench --test=memory --num-threads=1 --memory-block-size=1M --memory-total-size=102400G --memory-oper=write --max-time=5 --memory-access-mode=seq run 2>&1

func runSysBenchCommand(numThreads, oper, maxTime string) (string, error) {
	command := exec.Command("sysbench", "--test=memory", "--num-threads="+numThreads, "--memory-block-size=1M", "--memory-total-size=102400G", "--memory-oper="+oper, "--max-time="+maxTime, "--memory-access-mode=seq", "run", "2>&1")
	output, err := command.CombinedOutput()
	return string(output), err
}

func SysBenchMemoryTest() {
	fmt.Println("Running read test...")
	readResult, err := runSysBenchCommand("1", "read", "5")
	if err != nil {
		fmt.Printf("Error running read test: %v\n", err)
	} else {
		fmt.Println("Read test result:")
		fmt.Println(readResult)
	}
	fmt.Println("Running write test...")
	writeResult, err := runSysBenchCommand("1", "write", "5")
	if err != nil {
		fmt.Printf("Error running write test: %v\n", err)
	} else {
		fmt.Println("Write test result:")
		fmt.Println(writeResult)
	}
}
