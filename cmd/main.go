package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	. "github.com/oneclickvirt/defaultset"
	"github.com/oneclickvirt/memorytest/memory"
)

type cliOptions struct {
	help, version, jsonOutput, log bool
	language, testMethod           string
	sizeBytes                      int64
	iterations                     int
	timeout                        time.Duration
}

func parseCLI(args []string) (cliOptions, error) {
	opts := cliOptions{}
	fs := newFlagSet(&opts, io.Discard)
	if err := fs.Parse(args); err != nil {
		return opts, err
	}
	if opts.sizeBytes < 0 || opts.iterations < 0 || opts.timeout < 0 {
		return opts, fmt.Errorf("size, iterations, and timeout must not be negative")
	}
	return opts, nil
}

func newFlagSet(opts *cliOptions, output io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet("memorytest", flag.ContinueOnError)
	fs.SetOutput(output)
	fs.BoolVar(&opts.help, "h", false, "Show help information")
	fs.BoolVar(&opts.version, "v", false, "show version")
	fs.StringVar(&opts.language, "l", "", "Language parameter (en or zh)")
	fs.StringVar(&opts.testMethod, "m", "", "Specific Test Method (stream, dd, sysbench, winsat, or auto)")
	fs.BoolVar(&opts.log, "log", false, "Enable logging")
	fs.BoolVar(&opts.jsonOutput, "json", false, "Print the Go structured memory result as JSON")
	fs.BoolVar(&opts.jsonOutput, "structured", false, "Print the Go structured memory result as JSON")
	fs.Int64Var(&opts.sizeBytes, "size", 0, "Structured working-set size in bytes")
	fs.IntVar(&opts.iterations, "iterations", 0, "Structured benchmark iterations")
	fs.DurationVar(&opts.timeout, "timeout", 0, "Structured benchmark timeout (for example 30s)")
	return fs
}

func printCLIHelp(program string) {
	fmt.Printf("Usage: %s [options]\n", program)
	newFlagSet(&cliOptions{}, os.Stdout).PrintDefaults()
}

func selectCLIAction(opts cliOptions) string {
	if opts.help {
		return "help"
	}
	if opts.version {
		return "version"
	}
	if opts.jsonOutput {
		return "structured"
	}
	return "legacy"
}

func main() {
	opts, err := parseCLI(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	memory.EnableLoger = opts.log
	action := selectCLIAction(opts)
	if action == "help" || action == "version" {
		printLegacyHeader()
		if action == "help" {
			printCLIHelp(os.Args[0])
			return
		}
		fmt.Println(memory.MemoryTestVersion)
		return
	}
	if action == "structured" {
		if strings.TrimSpace(opts.testMethod) != "" {
			fmt.Fprintln(os.Stderr, "-m/--test-method is only supported by legacy output")
			os.Exit(2)
		}
		config := memory.DefaultBenchmarkConfig()
		if opts.sizeBytes > 0 {
			config.WorkingSetBytes = int(min(opts.sizeBytes, int64(512<<20)))
		}
		if opts.iterations > 0 {
			config.Iterations = opts.iterations
		}
		ctx := context.Background()
		cancel := func() {}
		if opts.timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, opts.timeout)
		}
		result, runErr := memory.RunBenchmark(ctx, config)
		cancel()
		encoded, marshalErr := json.Marshal(result)
		if marshalErr != nil {
			fmt.Fprintln(os.Stderr, marshalErr)
			return
		}
		fmt.Println(string(encoded))
		if runErr != nil {
			return
		}
		return
	}
	printLegacyHeader()
	language, testMethod := opts.language, opts.testMethod
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
	fmt.Print(res)
	fmt.Println("--------------------------------------------------")
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
	}
}

func printLegacyHeader() {
	go func() {
		http.Get("https://hits.spiritlhl.net/memorytest.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false")
	}()
	fmt.Println(Green("Repo:"), Yellow("https://github.com/oneclickvirt/memorytest"))
}
