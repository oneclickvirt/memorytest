package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/oneclickvirt/memorytest/memory"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	report, err := memory.RunBenchmark(ctx, memory.BenchmarkConfig{WorkingSetBytes: 32 << 20, Iterations: 3})
	if err != nil && ctx.Err() == nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(string(encoded))
}
