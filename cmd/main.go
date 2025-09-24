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
		http.Get("https://hits.spiritlhl.net/memorytest.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false")
	}()
	fmt.Println(Green("Repo:"), Yellow("https://github.com/oneclickvirt/memorytest"))
	var showVersion, help bool
	var language, testMethod string
	memorytestFlag := flag.NewFlagSet("cputest", flag.ContinueOnError)
	memorytestFlag.BoolVar(&help, "h", false, "Show help information")
	memorytestFlag.BoolVar(&showVersion, "v", false, "show version")
	memorytestFlag.StringVar(&language, "l", "", "Language parameter (en or zh)")
	memorytestFlag.StringVar(&testMethod, "m", "", "Specific Test Method (stream or dd or sysbench or winsat)")
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
	// Parse and normalize test method
	testMethod = strings.ToLower(testMethod)
	if testMethod == "" {
		// Use automatic priority: stream > dd > sysbench (with mbw fallback)
		testMethod = "auto"
	}
	if runtime.GOOS == "windows" {
		switch testMethod {
		case "stream":
			if language == "en" {
				res = "STREAM is not supported on Windows, using Winsat for testing.\n"
			} else {
				res = "Windows下不支持STREAM，使用Winsat进行测试。\n"
			}
			res += memory.WinsatTest(language)
		case "dd":
			res = memory.WindowsDDTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				if language == "en" {
					res = "DD test failed, switching to Winsat for testing.\n"
				} else {
					res = "DD测试失败，切换使用Winsat进行测试。\n"
				}
				res += memory.WinsatTest(language)
			}
		case "sysbench":
			if language == "en" {
				res = "Sysbench is not supported on Windows, using Winsat for testing.\n"
			} else {
				res = "Windows下不支持Sysbench，使用Winsat进行测试。\n"
			}
			res += memory.WinsatTest(language)
		default:
			// For auto or winsat or any other method
			if testMethod != "winsat" && testMethod != "auto" {
				if language == "en" {
					res = "Detected host is Windows, using Winsat for testing.\n"
				} else {
					res = "检测到主机为Windows，使用Winsat进行测试。\n"
				}
			}
			res += memory.WinsatTest(language)
		}
	} else {
		switch testMethod {
		case "stream":
			res = memory.StreamTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				if language == "en" {
					res = "STREAM test failed, switching to DD for testing.\n"
				} else {
					res = "STREAM测试失败，切换使用DD进行测试。\n"
				}
				res += memory.DDTest(language)
			}
		case "dd":
			res = memory.DDTest(language)
		case "sysbench":
			res = memory.SysBenchTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				if language == "en" {
					res = "Sysbench test failed, switching to DD for testing.\n"
				} else {
					res = "Sysbench测试失败，切换使用DD进行测试。\n"
				}
				res += memory.DDTest(language)
			}
		case "auto":
			// Priority: stream > dd > sysbench (with mbw fallback built into each)
			res = memory.StreamTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				// Stream failed or not available, try DD
				res = memory.DDTest(language)
				if res == "" || strings.TrimSpace(res) == "" {
					// DD failed, try sysbench as final fallback
					res = memory.SysBenchTest(language)
				}
			}
		default:
			res = "Unsupported test method"
		}
	}
	fmt.Println("--------------------------------------------------")
	fmt.Printf(res)
	fmt.Println("--------------------------------------------------")
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
	}
}
