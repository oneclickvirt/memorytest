package main

import (
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	. "github.com/oneclickvirt/defaultset"
	"github.com/oneclickvirt/memorytest/memory"
)

func main() {
	go func() {
		http.Get("https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Foneclickvirt%2Fmemorytest&count_bg=%2323E01C&title_bg=%23555555&icon=sonarcloud.svg&icon_color=%23E7E7E7&title=hits&edge_flat=false")
	}()
	fmt.Println(Green("项目地址:"), Yellow("https://github.com/oneclickvirt/memorytest"))
	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "show version")
	languagePtr := flag.String("l", "", "Language parameter (en or zh)")
	testMethodPtr := flag.String("m", "", "Specific Test Method (sysbench or dd)")
	flag.Parse()
	if showVersion {
		fmt.Println(memory.MemoryTestVersion)
		return
	}
	var language, res, testMethod string
	if *languagePtr == "" {
		language = "zh"
	} else {
		language = strings.ToLower(*languagePtr)
	}
	if *testMethodPtr == "" || *testMethodPtr == "sysbench" {
		testMethod = "sysbench"
	} else if *testMethodPtr == "dd" {
		testMethod = "dd"
	}
	if runtime.GOOS == "windows" {
		res = memory.WinsatTest(language)
	} else {
		if testMethod == "sysbench" {
			res = memory.SysBenchTest(language)
			if res == "" {
				res = "sysbench test failed, switch to use dd test.\n"
				res += memory.DDTest(language)
			}
		} else if testMethod == "dd" {
			res = memory.DDTest(language)
		}
	}
	fmt.Println("--------------------------------------------------")
	fmt.Printf(res)
	fmt.Println("--------------------------------------------------")
}
