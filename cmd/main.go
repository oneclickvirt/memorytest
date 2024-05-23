package main

import (
	"fmt"

	"github.com/oneclickvirt/memoryTest/memorytest"
)

func main() {
	res := memorytest.SysBenchMemoryTest()
	fmt.Printf(res)
}
