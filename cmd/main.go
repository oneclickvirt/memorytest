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
	memorytestFlag.StringVar(&testMethod, "m", "", "Specific Test Method (stream, dd, sysbench, winsat, or auto)")
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
			// stream → winsat → dd
			res = memory.StreamTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				if language == "en" {
					res = "STREAM test failed, switching to Winsat for testing.\n"
				} else {
					res = "STREAM测试失败，切换使用Winsat进行测试。\n"
				}
				res += memory.WinsatTest(language)
				if res == "" || strings.TrimSpace(res) == "" {
					if language == "en" {
						res += "Winsat test failed, switching to DD for testing.\n"
					} else {
						res += "Winsat测试失败，切换使用DD进行测试。\n"
					}
					res += memory.WindowsDDTest(language)
				}
			}
		case "dd":
			// dd → winsat → stream
			res = memory.WindowsDDTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				if language == "en" {
					res = "DD test failed, switching to Winsat for testing.\n"
				} else {
					res = "DD测试失败，切换使用Winsat进行测试。\n"
				}
				res += memory.WinsatTest(language)
				if res == "" || strings.TrimSpace(res) == "" {
					if language == "en" {
						res += "Winsat test failed, switching to STREAM for testing.\n"
					} else {
						res += "Winsat测试失败，切换使用STREAM进行测试。\n"
					}
					res += memory.StreamTest(language)
				}
			}
		case "sysbench":
			// sysbench → stream → winsat → dd (Windows不支持sysbench)
			if language == "en" {
				res = "Sysbench is not supported on Windows, switching to STREAM for testing.\n"
			} else {
				res = "Windows下不支持Sysbench，切换使用STREAM进行测试。\n"
			}
			res += memory.StreamTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				if language == "en" {
					res += "STREAM test failed, switching to Winsat for testing.\n"
				} else {
					res += "STREAM测试失败，切换使用Winsat进行测试。\n"
				}
				res += memory.WinsatTest(language)
				if res == "" || strings.TrimSpace(res) == "" {
					if language == "en" {
						res += "Winsat test failed, switching to DD for testing.\n"
					} else {
						res += "Winsat测试失败，切换使用DD进行测试。\n"
					}
					res += memory.WindowsDDTest(language)
				}
			}
		case "winsat":
			// winsat → stream → dd (已经用过winsat，不再重复)
			res = memory.WinsatTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				if language == "en" {
					res = "Winsat test failed, switching to STREAM for testing.\n"
				} else {
					res = "Winsat测试失败，切换使用STREAM进行测试。\n"
				}
				res += memory.StreamTest(language)
				if res == "" || strings.TrimSpace(res) == "" {
					if language == "en" {
						res += "STREAM test failed, switching to DD for testing.\n"
					} else {
						res += "STREAM测试失败，切换使用DD进行测试。\n"
					}
					res += memory.WindowsDDTest(language)
				}
			}
		case "auto":
			// auto → stream → winsat → dd
			res = memory.StreamTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				if language == "en" {
					res = "STREAM test failed, switching to Winsat for testing.\n"
				} else {
					res = "STREAM测试失败，切换使用Winsat进行测试。\n"
				}
				res += memory.WinsatTest(language)
				if res == "" || strings.TrimSpace(res) == "" {
					if language == "en" {
						res += "Winsat test failed, switching to DD for testing.\n"
					} else {
						res += "Winsat测试失败，切换使用DD进行测试。\n"
					}
					res += memory.WindowsDDTest(language)
				}
			}
		default:
			// 其他方法 → stream → winsat → dd
			res = memory.StreamTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				if language == "en" {
					res = "STREAM test failed, switching to Winsat for testing.\n"
				} else {
					res = "STREAM测试失败，切换使用Winsat进行测试。\n"
				}
				res += memory.WinsatTest(language)
				if res == "" || strings.TrimSpace(res) == "" {
					if language == "en" {
						res += "Winsat test failed, switching to DD for testing.\n"
					} else {
						res += "Winsat测试失败，切换使用DD进行测试。\n"
					}
					res += memory.WindowsDDTest(language)
				}
			}
		}
	} else {
		switch testMethod {
		case "stream":
			res = memory.StreamTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				if language == "en" {
					res = "STREAM test failed, switching to Sysbench for testing.\n"
				} else {
					res = "STREAM测试失败，切换使用Sysbench进行测试。\n"
				}
				res += memory.SysBenchTest(language)
				if res == "" || strings.TrimSpace(res) == "" {
					if language == "en" {
						res = "Sysbench test failed, switching to DD for testing.\n"
					} else {
						res = "Sysbench测试失败，切换使用DD进行测试。\n"
					}
					res += memory.DDTest(language)
				}
			}
		case "dd":
			res = memory.DDTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				if language == "en" {
					res = "DD test failed, switching to STREAM for testing.\n"
				} else {
					res = "DD测试失败，切换使用STREAM进行测试。\n"
				}
				res += memory.StreamTest(language)
				if res == "" || strings.TrimSpace(res) == "" {
					if language == "en" {
						res = "STREAM test failed, switching to Sysbench for testing.\n"
					} else {
						res = "STREAM测试失败，切换使用Sysbench进行测试。\n"
					}
					res += memory.SysBenchTest(language)
				}
			}
		case "sysbench":
			res = memory.SysBenchTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				if language == "en" {
					res = "Sysbench test failed, switching to STREAM for testing.\n"
				} else {
					res = "Sysbench测试失败，切换使用STREAM进行测试。\n"
				}
				res += memory.StreamTest(language)
				if res == "" || strings.TrimSpace(res) == "" {
					if language == "en" {
						res = "STREAM test failed, retrying Sysbench for testing.\n"
					} else {
						res = "STREAM测试失败，重试使用Sysbench进行测试。\n"
					}
					res += memory.SysBenchTest(language)
					if res == "" || strings.TrimSpace(res) == "" {
						if language == "en" {
							res = "Sysbench retry failed, switching to DD for testing.\n"
						} else {
							res = "Sysbench重试失败，切换使用DD进行测试。\n"
						}
						res += memory.DDTest(language)
					}
				}
			}
		case "auto":
			// Priority: stream > sysbench > dd (with mbw fallback built into each)
			res = memory.StreamTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				// Stream failed or not available, try Sysbench
				res = memory.SysBenchTest(language)
				if res == "" || strings.TrimSpace(res) == "" {
					// Sysbench failed, try DD as final fallback
					res = memory.DDTest(language)
				}
			}
		case "winsat":
			// Winsat is only supported on Windows
			if language == "en" {
				res = "Winsat is only supported on Windows, switching to STREAM for testing.\n"
			} else {
				res = "Winsat仅在Windows上支持，切换使用STREAM进行测试。\n"
			}
			res += memory.StreamTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				if language == "en" {
					res = "STREAM test failed, switching to Sysbench for testing.\n"
				} else {
					res = "STREAM测试失败，切换使用Sysbench进行测试。\n"
				}
				res += memory.SysBenchTest(language)
				if res == "" || strings.TrimSpace(res) == "" {
					if language == "en" {
						res = "Sysbench test failed, switching to DD for testing.\n"
					} else {
						res = "Sysbench测试失败，切换使用DD进行测试。\n"
					}
					res += memory.DDTest(language)
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
