package memory

import (
	"context"
	"errors"
	"runtime"
	"time"
)

type BenchmarkStatus string

const (
	BenchmarkOK          BenchmarkStatus = "ok"
	BenchmarkCanceled    BenchmarkStatus = "canceled"
	BenchmarkTimeout     BenchmarkStatus = "timeout"
	BenchmarkUnavailable BenchmarkStatus = "unavailable"
)

type BenchmarkConfig struct {
	WorkingSetBytes int
	Iterations      int
}

type BenchmarkResult struct {
	SchemaVersion       string          `json:"schema_version"`
	Status              BenchmarkStatus `json:"status"`
	WorkingSetBytes     int             `json:"working_set_bytes"`
	Iterations          int             `json:"iterations"`
	SequentialReadMBps  float64         `json:"sequential_read_mbps"`
	SequentialWriteMBps float64         `json:"sequential_write_mbps"`
	CopyMBps            float64         `json:"copy_mbps"`
	RandomLatencyNS     float64         `json:"random_latency_ns"`
	DurationMS          int64           `json:"duration_ms"`
	Error               string          `json:"error,omitempty"`
}

func DefaultBenchmarkConfig() BenchmarkConfig {
	return BenchmarkConfig{WorkingSetBytes: 32 << 20, Iterations: 3}
}

func RunBenchmark(ctx context.Context, config BenchmarkConfig) (result BenchmarkResult, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	defaults := DefaultBenchmarkConfig()
	if config.WorkingSetBytes <= 0 {
		config.WorkingSetBytes = defaults.WorkingSetBytes
	}
	if config.Iterations <= 0 {
		config.Iterations = defaults.Iterations
	}
	// Keep standard runs bounded even when called with untrusted API input.
	if config.WorkingSetBytes > 512<<20 {
		config.WorkingSetBytes = 512 << 20
	}
	if config.Iterations > 20 {
		config.Iterations = 20
	}
	result = BenchmarkResult{SchemaVersion: "goecs.memory/v1", Status: BenchmarkOK, WorkingSetBytes: config.WorkingSetBytes, Iterations: config.Iterations}
	started := time.Now()
	defer func() {
		result.DurationMS = time.Since(started).Milliseconds()
		if recovered := recover(); recovered != nil {
			result.Status = BenchmarkUnavailable
			result.Error = "memory allocation failed"
			err = errors.New(result.Error)
		}
	}()
	if err := ctx.Err(); err != nil {
		result.Status, result.Error = benchmarkStop(err)
		return result, err
	}
	count := max(config.WorkingSetBytes/8, 1024)
	source, destination := make([]uint64, count), make([]uint64, count)
	for i := range source {
		source[i] = uint64(i*2 + 1)
	}
	bytesPerPass := float64(count * 8)

	writeDuration, err := timeOperation(ctx, config.Iterations, func() {
		for i := range destination {
			destination[i] = uint64(i) + 1
		}
	})
	if err != nil {
		return canceledMemoryResult(result, err)
	}
	result.SequentialWriteMBps = megabytesPerSecond(bytesPerPass*float64(config.Iterations), writeDuration)

	var checksum uint64
	readDuration, err := timeOperation(ctx, config.Iterations, func() {
		var sum uint64
		for _, value := range source {
			sum += value
		}
		checksum ^= sum
	})
	if err != nil {
		return canceledMemoryResult(result, err)
	}
	result.SequentialReadMBps = megabytesPerSecond(bytesPerPass*float64(config.Iterations), readDuration)

	copyDuration, err := timeOperation(ctx, config.Iterations, func() { copy(destination, source) })
	if err != nil {
		return canceledMemoryResult(result, err)
	}
	result.CopyMBps = megabytesPerSecond(bytesPerPass*float64(config.Iterations), copyDuration)

	step := 8191
	for step >= count || commonDivisor(step, count) != 1 {
		step -= 2
		if step < 1 {
			step = 1
			break
		}
	}
	accesses := count * config.Iterations
	for i := range source {
		source[i] = uint64((i + step) % count)
	}
	index := 0
	randomStarted := time.Now()
	for access := 0; access < accesses; access++ {
		if access&0xffff == 0 {
			if err := ctx.Err(); err != nil {
				return canceledMemoryResult(result, err)
			}
		}
		index = int(source[index])
		checksum ^= uint64(index)
	}
	result.RandomLatencyNS = float64(time.Since(randomStarted).Nanoseconds()) / float64(accesses)
	runtime.KeepAlive(checksum)
	runtime.KeepAlive(source)
	runtime.KeepAlive(destination)
	return result, nil
}

func timeOperation(ctx context.Context, iterations int, operation func()) (time.Duration, error) {
	started := time.Now()
	for iteration := 0; iteration < iterations; iteration++ {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		operation()
	}
	return time.Since(started), nil
}

func canceledMemoryResult(result BenchmarkResult, err error) (BenchmarkResult, error) {
	result.Status, result.Error = benchmarkStop(err)
	return result, err
}

func benchmarkStop(err error) (BenchmarkStatus, string) {
	if errors.Is(err, context.Canceled) {
		return BenchmarkCanceled, "canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return BenchmarkTimeout, "timeout"
	}
	return BenchmarkUnavailable, "benchmark_failed"
}

func megabytesPerSecond(bytes float64, duration time.Duration) float64 {
	if duration <= 0 {
		return 0
	}
	return bytes / (1024 * 1024) / duration.Seconds()
}

func commonDivisor(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
