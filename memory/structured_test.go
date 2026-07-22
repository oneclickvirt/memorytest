package memory

import (
	"context"
	"testing"
	"time"
)

func TestRunBenchmarkStructured(t *testing.T) {
	result, err := RunBenchmark(context.Background(), BenchmarkConfig{WorkingSetBytes: 2 << 20, Iterations: 2})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != BenchmarkOK || result.SequentialReadMBps <= 0 || result.SequentialWriteMBps <= 0 || result.CopyMBps <= 0 || result.RandomLatencyNS <= 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestRunBenchmarkCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := RunBenchmark(ctx, BenchmarkConfig{WorkingSetBytes: 1 << 20, Iterations: 1})
	if err == nil || result.Status != BenchmarkCanceled {
		t.Fatalf("status=%s err=%v", result.Status, err)
	}
}

func TestRunBenchmarkDeadline(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	result, err := RunBenchmark(ctx, DefaultBenchmarkConfig())
	if err == nil || result.Status != BenchmarkTimeout || result.Error != "timeout" {
		t.Fatalf("unexpected timeout result: result=%+v err=%v", result, err)
	}
}
