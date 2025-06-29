package memory

// #cgo CFLAGS: -std=c99
// #include "memory.h"
import "C"
import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"unsafe"

	. "github.com/oneclickvirt/defaultset"
)

// hasRootPermission 检测是否有root权限
func hasRootPermission() bool {
	if runtime.GOOS == "windows" {
		return true // Windows环境直接返回true
	}
	return os.Getuid() == 0
}

// simpleMemoryTest CGO优化版本
func simpleMemoryTest(language string) string {
	if EnableLoger {
		Logger.Info("Running CGO-optimized memory test without root permission")
	}
	sizes := []string{"400", "512", "1024"}
	var allocatedSize int
	var writeSpeed, readSpeed float64
	pageSize := int(C.get_page_size())
	var buffer unsafe.Pointer
	for _, s := range sizes {
		mb, err := strconv.Atoi(s)
		if err != nil {
			continue
		}
		sizeBytes := mb * 1024 * 1024
		// 使用CGO分配对齐内存
		buffer = C.aligned_malloc(C.size_t(sizeBytes), C.size_t(pageSize))
		if buffer != nil {
			allocatedSize = sizeBytes
			break
		}
	}
	if buffer == nil {
		return "Memory allocation failed. Cannot perform test.\n"
	}
	defer func() {
		C.free(buffer)
		runtime.GC()
	}()
	// 清零内存
	C.memset(buffer, 0, C.size_t(allocatedSize))
	// 执行写入测试
	writeSpeed = float64(C.fast_memory_write_test(buffer, C.size_t(allocatedSize)))
	// 执行读取测试
	readSpeed = float64(C.fast_memory_read_test(buffer, C.size_t(allocatedSize)))
	var result string
	allocatedMB := float64(allocatedSize) / 1024 / 1024
	if language == "en" {
		result += fmt.Sprintf("Note: Running without root permission, using Go memory test\n")
		result += fmt.Sprintf("Test Buffer Size: %.1f MB\n", allocatedMB)
		result += fmt.Sprintf("Single Seq Write Speed: %.2f MB/s\n", writeSpeed)
		result += fmt.Sprintf("Single Seq Read Speed: %.2f MB/s\n", readSpeed)
	} else {
		result += fmt.Sprintf("注意: 当前无root权限，使用Go内存测试\n")
		result += fmt.Sprintf("测试缓冲区大小: %.1f MB\n", allocatedMB)
		result += fmt.Sprintf("单线程顺序写速度: %.2f MB/s\n", writeSpeed)
		result += fmt.Sprintf("单线程顺序读速度: %.2f MB/s\n", readSpeed)
	}
	return result
}
