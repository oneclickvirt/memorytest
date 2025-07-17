//go:build (linux && amd64) || (linux && 386) || (linux && arm64) || (linux && riscv64) || (linux && mips64) || (linux && mips64le) || (linux && ppc64le) || (windows && amd64) || (windows && 386)

package memory

// #include "mbw.h"
import "C"

import (
	"fmt"
	"strings"
	"unsafe"
)

func simpleMemoryTest(language string) string {
	testSizes := []uint64{1024, 512, 256, 128, 64, 32}
	var results [3]C.struct_TestResult
	// var successfulSize uint64
	var success bool
	for _, size := range testSizes {
		ret := C.run_memory_test(C.ulonglong(size), (*C.struct_TestResult)(unsafe.Pointer(&results[0])))
		if ret == 0 {
			// successfulSize = size
			success = true
			break
		}
	}
	if !success {
		if language == "en" {
			return "Memory test execution failed: insufficient memory\n"
		} else {
			return "内存测试执行失败：可用内存不足\n"
		}
	}
	var result strings.Builder
	if language == "en" {
		result.WriteString(fmt.Sprintf("Memory Copy Speed (MEMCPY)   : %10.2f MB/s \n", float64(results[0].speed)))
		result.WriteString(fmt.Sprintf("Memory Copy Speed (DUMB)     : %10.2f MB/s \n", float64(results[1].speed)))
		result.WriteString(fmt.Sprintf("Memory Copy Speed (MCBLOCK)  : %10.2f MB/s \n", float64(results[2].speed)))
	} else {
		result.WriteString(fmt.Sprintf("内存复制速度(读+写) (MEMCPY)   : %10.2f MB/s \n", float64(results[0].speed)))
		result.WriteString(fmt.Sprintf("内存复制速度(读+写) (DUMB)     : %10.2f MB/s \n", float64(results[1].speed)))
		result.WriteString(fmt.Sprintf("内存复制速度(读+写) (MCBLOCK)  : %10.2f MB/s \n", float64(results[2].speed)))
	}
	return result.String()
}
