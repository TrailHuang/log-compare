package main

import (
	"flag"
	"fmt"
	"log-compare/config"
	"log-compare/reporter"
	"os"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	configPath := flag.String("config", "", "配置文件路径")
	stdDir := flag.String("stddir", "", "标准日志目录")
	logDir := flag.String("logdir", "", "待对比日志目录")
	outputPath := flag.String("output", "", "报告输出文件路径（可选）")
	showVersion := flag.Bool("version", false, "显示版本信息")
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

	if *outputPath != "" {
		if err := reporter.WriteFile(overall.LogTypeResults, *outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "写入报告失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\n详细报告已写入: %s\n", *outputPath)
	}

	// 计算差异总数
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
