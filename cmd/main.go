package main

import (
	"fmt"
	"flag"
	"net/http"
	"runtime"
	. "github.com/oneclickvirt/memoryTest/defaultset"
	"github.com/oneclickvirt/memoryTest/memorytest"
)

func main() {
	go func() {
		http.Get("https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Foneclickvirt%2FmemoryTest&count_bg=%2323E01C&title_bg=%23555555&icon=sonarcloud.svg&icon_color=%23E7E7E7&title=hits&edge_flat=false")
	}()
	fmt.Println(Green("项目地址:"), Yellow("https://github.com/oneclickvirt/memoryTest"))
	languagePtr := flag.String("l", "", "Language parameter (en or zh)")
	testMethodPtr := flag.String("m", "", "Specific Test Method (sysbench or dd)")
	flag.Parse()
	var language, res, testMethod string
	var isMultiCheck bool
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
		res = memorytest.WinsatTest(language)
	} else {
		if testMethod == "sysbench" {
			res = memorytest.SysBenchTest(language)
			if res == "" {
				res = "sysbench test failed, switch to use dd test.\n"
				res += memorytest.DDTest(language)
			}
		} else if testMethod == "dd" {
			res = memorytest.DDTest(language)
		}
	}
	fmt.Println("--------------------------------------------------")
	fmt.Printf(res)
	fmt.Println("--------------------------------------------------")
}
