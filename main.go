package main

import (
	"flag"
	"fmt"
	"log-compare/config"
	"log-compare/reporter"
	"os"
	"path/filepath"
	"time"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// 预处理：-output 不带值时，插入空字符串作为占位
	outputExplicit, osArgs := preprocessOutputFlag(os.Args)

	configPath := flag.String("config", "conf/log_info.json", "配置文件路径")
	stdDir := flag.String("stddir", "", "标准日志目录")
	logDir := flag.String("logdir", "", "待对比日志目录")
	outputPath := flag.String("output", "", "报告输出文件路径（可选）")
	showVersion := flag.Bool("version", false, "显示版本信息")
	os.Args = osArgs
	flag.Parse()

	if *showVersion {
		fmt.Printf("log-compare %s (built: %s)\n", Version, BuildTime)
		os.Exit(0)
	}

	if *configPath == "" || *stdDir == "" || *logDir == "" {
		fmt.Println("用法: log-compare -config <配置文件> -stddir <标准日志目录> -logdir <待对比日志目录> [-output <报告文件>]")
		os.Exit(1)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "配置验证失败: %v\n", err)
		os.Exit(1)
	}

	overall, err := Run(cfg, *stdDir, *logDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "对比失败: %v\n", err)
		os.Exit(1)
	}

	reporter.PrintTerminal(overall.LogTypeResults)

	// 输出报告：未指定 -output 则不输出，-output 无参数则按时间生成文件名
	if outputExplicit {
		outPath := *outputPath
		if outPath == "" {
			outPath = "report_" + time.Now().Format("20060102150405") + ".txt"
		}
		if err := reporter.WriteFile(overall.LogTypeResults, outPath); err != nil {
			fmt.Fprintf(os.Stderr, "写入报告失败: %v\n", err)
			os.Exit(1)
		}
		absPath, _ := filepath.Abs(outPath)
		fmt.Printf("\n详细报告已写入: %s\n", absPath)
	}

	// 计算差异总数：字段差异记录 + 仅标准端有 + 仅待对比端有
	totalDiff := 0
	for _, r := range overall.LogTypeResults {
		totalDiff += r.RecordsWithDiff + r.LogOnlyMatchKeys + r.StdOnlyMatchKeys
	}

	if totalDiff > 0 {
		if totalDiff > 255 {
			totalDiff = 255
		}
		os.Exit(totalDiff)
	}
}

// preprocessOutputFlag 预处理 -output 参数
// 如果 -output 存在但后面没有跟值（是最后一个参数或下一个参数以 - 开头），
// 则在 args 中 -output 后插入空字符串作为占位，使 flag.Parse 能正常工作。
// 返回值：outputExplicit 表示 -output 是否被显式指定，osArgs 是处理后的参数列表。
func preprocessOutputFlag(args []string) (outputExplicit bool, newArgs []string) {
	for i, arg := range args {
		if arg == "-output" || arg == "--output" {
			outputExplicit = true
			if i+1 >= len(args) || (len(args[i+1]) > 0 && args[i+1][0] == '-') {
				// -output 后没有值，插入空字符串
				newArgs = make([]string, 0, len(args)+1)
				newArgs = append(newArgs, args[:i+1]...)
				newArgs = append(newArgs, "")
				newArgs = append(newArgs, args[i+1:]...)
				return outputExplicit, newArgs
			}
		}
	}
	return outputExplicit, args
}
