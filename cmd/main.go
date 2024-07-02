package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
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
	var showVersion, help bool
	var language, testMethod string
	memorytestFlag := flag.NewFlagSet("cputest", flag.ContinueOnError)
	memorytestFlag.BoolVar(&help, "h", false, "Show help information")
	memorytestFlag.BoolVar(&showVersion, "v", false, "show version")
	memorytestFlag.StringVar(&language, "l", "", "Language parameter (en or zh)")
	memorytestFlag.StringVar(&testMethod, "m", "", "Specific Test Method (sysbench or dd)")
	memorytestFlag.BoolVar(&memory.EnableLoger, "log", false, "Enable logging")
	memorytestFlag.Parse(os.Args[1:])
	if help {
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		memorytestFlag.PrintDefaults()
		return
	}
	if showVersion {
		fmt.Println(memory.MemoryTestVersion)
		return
	}
	var res string
	if language == "" {
		language = "zh"
	} else {
		language = strings.ToLower(language)
	}
	if testMethod == "" || strings.ToLower(testMethod) == "sysbench" {
		testMethod = "sysbench"
	} else if strings.ToLower(testMethod) == "dd" {
		testMethod = "dd"
	}
	if runtime.GOOS == "windows" {
		if testMethod != "winsat" && testMethod != "" {
			res = "Detected host is Windows, using Winsat for testing.\n"
		}
		res += memory.WinsatTest(language)
	} else {
		switch testMethod {
		case "sysbench":
			res = memory.SysBenchTest(language)
			if res == "" {
				res = "sysbench test failed, switch to use dd test.\n"
				res += memory.DDTest(language)
			}
		case "dd":
			res = memory.DDTest(language)
		default:
			res = "Unsupported test method"
		}
	}
	fmt.Println("--------------------------------------------------")
	fmt.Printf(res)
	fmt.Println("--------------------------------------------------")
}
