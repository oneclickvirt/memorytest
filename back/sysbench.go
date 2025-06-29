package memory

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Config 定义测试配置
type Config struct {
	NumThreads int
	BlockSize  int64
	TotalSize  int64
	TestTime   int
	Operation  string
	AccessMode string
}

// Stats 定义延迟统计
type Stats struct {
	Min         float64
	Max         float64
	Avg         float64
	Percentile95 float64
	Sum         float64
	Ops         int64
	Duration    time.Duration
}

// Result 定义测试结果
type Result struct {
	Operations uint64
	Latencies  []float64
	StartTime  time.Time
	EndTime    time.Time
}

// MemoryTest 定义内存测试结构
type MemoryTest struct {
	config Config
	buffer []byte
}

// NewMemoryTest 创建新的内存测试实例
func NewMemoryTest(config Config) *MemoryTest {
	return &MemoryTest{
		config: config,
		buffer: make([]byte, config.BlockSize),
	}
}

// Run 执行测试
func (m *MemoryTest) Run() (*Stats, error) {
	// 打印测试配置
	m.printConfig()
	result := m.runTest()
	stats := m.calculateStats(result)
	m.printResults(stats)
	return stats, nil
}

func (m *MemoryTest) printConfig() {
	fmt.Printf("Running memory speed test with the following options:\n")
	fmt.Printf("  block size: %dKiB\n", m.config.BlockSize/1024)
	fmt.Printf("  total size: %dMiB\n", m.config.TotalSize/1024/1024)
	fmt.Printf("  operation: %s\n", m.config.Operation)
	fmt.Printf("  threads: %d\n", m.config.NumThreads)
	fmt.Printf("  access mode: %s\n\n", m.config.AccessMode)
}

func (m *MemoryTest) runTest() Result {
	var wg sync.WaitGroup
	result := Result{
		Latencies: make([]float64, 0),
		StartTime: time.Now(),
	}
	// 启动工作线程
	for i := 0; i < m.config.NumThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			deadline := time.Now().Add(time.Duration(m.config.TestTime) * time.Second)
			for time.Now().Before(deadline) {
				start := time.Now()
				if m.config.Operation == "read" {
					// 顺序读取
					for j := int64(0); j < m.config.BlockSize; j += 8 {
						_ = m.buffer[j:j+8]
					}
				} else {
					// 顺序写入
					for j := int64(0); j < m.config.BlockSize; j += 8 {
						copy(m.buffer[j:j+8], []byte{1,2,3,4,5,6,7,8})
					}
				}
				latency := float64(time.Since(start).Microseconds()) / 1000.0 // 转换为毫秒
				result.Latencies = append(result.Latencies, latency)
				atomic.AddUint64(&result.Operations, 1)
			}
		}()
	}
	wg.Wait()
	result.EndTime = time.Now()
	return result
}

func (m *MemoryTest) calculateStats(result Result) *Stats {
	if len(result.Latencies) == 0 {
		return &Stats{}
	}
	// 计算延迟统计
	sort.Float64s(result.Latencies)
	var sum float64
	for _, lat := range result.Latencies {
		sum += lat
	}
	p95Index := int(float64(len(result.Latencies)) * 0.95)
	return &Stats{
		Min:         result.Latencies[0],
		Max:         result.Latencies[len(result.Latencies)-1],
		Avg:         sum / float64(len(result.Latencies)),
		Percentile95: result.Latencies[p95Index],
		Sum:         sum,
		Ops:         int64(result.Operations),
		Duration:    result.EndTime.Sub(result.StartTime),
	}
}

func (m *MemoryTest) printResults(stats *Stats) {
	duration := float64(stats.Duration.Seconds())
	opsPerSec := float64(stats.Ops) / duration
	fmt.Printf("Total operations: %d (%.2f per second)\n", stats.Ops, opsPerSec)
	// 计算传输速率
	mibTransferred := float64(stats.Ops) * float64(m.config.BlockSize) / (1024 * 1024)
	mibPerSec := mibTransferred / duration
	fmt.Printf("%.2f MiB transferred (%.2f MiB/sec)\n", mibTransferred, mibPerSec)
	fmt.Printf("\nLatency (ms):\n")
	fmt.Printf("         min:                              %.2f\n", stats.Min)
	fmt.Printf("         avg:                              %.2f\n", stats.Avg)
	fmt.Printf("         max:                              %.2f\n", stats.Max)
	fmt.Printf("         95th percentile:                  %.2f\n", stats.Percentile95)
	fmt.Printf("         sum:                              %.2f\n", stats.Sum)
}

func init() {
	// 设置最大线程数
	runtime.GOMAXPROCS(runtime.NumCPU())
}