package memory

import (
	"context"
	"testing"
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
