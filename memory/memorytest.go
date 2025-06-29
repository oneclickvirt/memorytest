package memory

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	. "github.com/oneclickvirt/defaultset"
)

// hasRootPermission 检测是否有root权限
func hasRootPermission() bool {
	if runtime.GOOS == "windows" {
		return true // Windows环境直接返回true
	}
	return os.Getuid() == 0
}

// simpleMemoryTest 无root权限时模拟测试
func simpleMemoryTest(language string) string {
	if EnableLoger {
		Logger.Info("Running simple memory test without root permission")
	}
	sizes := []string{"1024", "128", "64"}
	var buffer []byte
	var allocatedSize int
	for _, s := range sizes {
		mb, err := strconv.Atoi(s)
		if err != nil {
			continue
		}
		sizeBytes := mb * 1024 * 1024
		func() {
			defer func() {
				if r := recover(); r != nil {
					buffer = nil
				}
			}()
			buffer = make([]byte, sizeBytes)
			allocatedSize = sizeBytes
		}()
		if buffer != nil && len(buffer) > 0 {
			break
		}
	}
	if buffer == nil || len(buffer) == 0 {
		return "Memory allocation failed. Cannot perform test.\n"
	}
	defer func() {
		buffer = nil
		runtime.GC()
	}()
	for i := 0; i < len(buffer); i++ {
		buffer[i] = 0
	}
	start := time.Now()
	for i := 0; i < len(buffer); i++ {
		buffer[i] = byte(i % 256)
	}
	writeTime := time.Since(start).Seconds()
	writeSpeed := float64(len(buffer)) / writeTime / 1024 / 1024
	var tmp byte
	for i := 0; i < len(buffer); i += 4096 {
		tmp += buffer[i]
	}
	start = time.Now()
	var sum uint64
	for i := 0; i < len(buffer); i++ {
		sum += uint64(buffer[i])
	}
	readTime := time.Since(start).Seconds()
	readSpeed := float64(len(buffer)) / readTime / 1024 / 1024
	if tmp > 255 {
		fmt.Println("Unlikely condition to keep tmp alive:", tmp)
	}
	var result string
	allocatedMB := float64(allocatedSize) / 1024 / 1024
	if language == "en" {
		result += fmt.Sprintf("Note: Running without root permission, using Go memory test\n")
		result += fmt.Sprintf("Test Buffer Size: %.1f MB\n", allocatedMB)
		result += fmt.Sprintf("Single Seq Write Speed: %.2f MB/s\n", writeSpeed)
		result += fmt.Sprintf("Single Seq Read  Speed: %.2f MB/s\n", readSpeed)
	} else {
		result += fmt.Sprintf("注意: 当前无root权限，使用Go内存测试\n")
		result += fmt.Sprintf("测试缓冲区大小: %.1f MB\n", allocatedMB)
		result += fmt.Sprintf("单线程顺序写速度: %.2f MB/s\n", writeSpeed)
		result += fmt.Sprintf("单线程顺序读速度: %.2f MB/s\n", readSpeed)
	}
	return result
}
