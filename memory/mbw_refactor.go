package memory

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
	"unsafe"
)

const (
	DefaultNrLoops   = 10
	MaxTests         = 3
	DefaultBlockSize = 262144
	TestMemcpy       = 0
	TestDumb         = 1
	TestMcblock      = 2
)

type TestResult struct {
	Speed float64
	Type  int
}

func makeArray(asize uint64) []int64 {
	runtime.GC()
	runtime.GC()
	a := make([]int64, asize)
	for t := uint64(0); t < asize; t++ {
		a[t] = 0xaa
	}
	runtime.GC()
	return a
}

//go:noinline
func memcpyOptimized(dst, src []int64) {
	copy(dst, src)
}

//go:noinline
func memcpyBlock(dst, src []byte, size uint64) {
	copy(dst[:size], src[:size])
}

//go:noinline
func dumbCopy(dst, src []int64, size uint64) {
	for i := uint64(0); i < size; i++ {
		dst[i] = src[i]
	}
}

func workerOnlyGo(asize uint64, a, b []int64, testType int, blockSize uint64) float64 {
	gcPercent := runtime.GOMAXPROCS(0)
	runtime.GC()
	debug.SetGCPercent(-1)
	defer func() {
		debug.SetGCPercent(gcPercent)
	}()
	longSize := uint64(unsafe.Sizeof(int64(0)))
	arrayBytes := asize * longSize
	var startTime, endTime time.Time
	switch testType {
	case TestMemcpy:
		copy(b[:min(1000, uint64(len(b)))], a[:min(1000, uint64(len(a)))])
		startTime = time.Now()
		memcpyOptimized(b, a)
		endTime = time.Now()
	case TestMcblock:
		srcPtr := unsafe.Pointer(&a[0])
		dstPtr := unsafe.Pointer(&b[0])
		srcBytes := (*[1 << 30]byte)(srcPtr)
		dstBytes := (*[1 << 30]byte)(dstPtr)
		memcpyBlock(dstBytes[:], srcBytes[:], min(blockSize, 1000))
		startTime = time.Now()
		var offset uint64
		remainingBytes := arrayBytes
		for remainingBytes >= blockSize {
			memcpyBlock(dstBytes[offset:], srcBytes[offset:], blockSize)
			offset += blockSize
			remainingBytes -= blockSize
		}
		if remainingBytes > 0 {
			memcpyBlock(dstBytes[offset:], srcBytes[offset:], remainingBytes)
		}
		endTime = time.Now()
	case TestDumb:
		dumbCopy(b, a, min(1000, asize))
		startTime = time.Now()
		dumbCopy(b, a, asize)
		endTime = time.Now()
	}
	duration := endTime.Sub(startTime)
	return duration.Seconds()
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func runMemoryTest(mt uint64) ([]TestResult, error) {
	longSize := uint64(unsafe.Sizeof(int64(0)))
	asize := 1024 * 1024 / longSize * mt
	blockSize := uint64(DefaultBlockSize)
	nrLoops := 5
	if asize*longSize < blockSize {
		return nil, fmt.Errorf("array size too small for block size")
	}
	a := makeArray(asize)
	b := makeArray(asize)
	runtime.KeepAlive(a)
	runtime.KeepAlive(b)
	results := make([]TestResult, MaxTests)
	for testno := 0; testno < MaxTests; testno++ {
		var teSum float64
		_ = workerOnlyGo(asize, a, b, testno, blockSize)
		for i := 0; i < nrLoops; i++ {
			te := workerOnlyGo(asize, a, b, testno, blockSize)
			teSum += te
		}
		results[testno].Speed = float64(mt) / teSum * float64(nrLoops)
		results[testno].Type = testno
	}
	runtime.KeepAlive(a)
	runtime.KeepAlive(b)
	return results, nil
}

func benchmarkMemoryTest(mt uint64) ([]TestResult, error) {
	oldGOMAXPROCS := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(oldGOMAXPROCS)
	oldGCPercent := debug.SetGCPercent(10)
	defer debug.SetGCPercent(oldGCPercent)
	runtime.GC()
	runtime.GC()
	return runMemoryTest(mt)
}

func simpleMemoryTestOnlyGoalng(language string) string {
	testSizes := []uint64{1024, 512, 256, 128, 64, 32}
	var results []TestResult
	var success bool
	for _, size := range testSizes {
		var err error
		results, err = benchmarkMemoryTest(size)
		if err == nil {
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
		result.WriteString(fmt.Sprintf("Memory Copy Speed (MEMCPY)   : %10.2f MB/s \n", results[0].Speed))
		result.WriteString(fmt.Sprintf("Memory Copy Speed (DUMB)     : %10.2f MB/s \n", results[1].Speed))
		result.WriteString(fmt.Sprintf("Memory Copy Speed (MCBLOCK)  : %10.2f MB/s \n", results[2].Speed))
	} else {
		result.WriteString(fmt.Sprintf("内存复制速度(读+写) (MEMCPY)   : %10.2f MB/s \n", results[0].Speed))
		result.WriteString(fmt.Sprintf("内存复制速度(读+写) (DUMB)     : %10.2f MB/s \n", results[1].Speed))
		result.WriteString(fmt.Sprintf("内存复制速度(读+写) (MCBLOCK)  : %10.2f MB/s \n", results[2].Speed))
	}
	return result.String()
}
