package memory

import (
	"fmt"
	"testing"
)

func TestM(t *testing.T) {
	// res := SysBenchTest("zh")
	// res := DDTest("zh")
	res := StreamTest("zh")
	fmt.Println(res)
}
