# log-compare

[![Go Reference](https://pkg.go.dev/badge/github.com/TrailHuang/log-compare.svg)](https://pkg.go.dev/github.com/TrailHuang/log-compare)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

一个用于对比"标准日志目录"与"待对比日志目录"的 Go 工具与库。

它既可作为 **命令行工具** 直接使用，也可以作为 **Go 库** 被其他项目引用，
对其执行日志对比、字段校验、差异统计等操作。

---

## 安装（命令行工具）

```bash
go install github.com/TrailHuang/log-compare/cmd/log-compare@latest
```

或者从源码构建：

```bash
git clone https://github.com/TrailHuang/log-compare.git
cd log-compare
make build
./bin/log-compare --help
```

## 作为库被引用

模块路径：`github.com/TrailHuang/log-compare`

```bash
go get github.com/TrailHuang/log-compare
```

最小用法示例：

```go
import (
    "github.com/TrailHuang/log-compare/config"
    "github.com/TrailHuang/log-compare/logcompare"
    "github.com/TrailHuang/log-compare/reporter"
)

cfg, err := config.Load("conf/log_info.json")
if err != nil {
    log.Fatal(err)
}
if err := cfg.Validate(); err != nil {
    log.Fatal(err)
}

overall, err := logcompare.Run(cfg, "path/to/standard", "path/to/log")
if err != nil {
    log.Fatal(err)
}

// 终端可视化输出
reporter.PrintTerminal(overall.LogTypeResults)

// 或自行遍历每个日志类型的对比结果
for name, r := range overall.LogTypeResults {
    fmt.Printf("%s: total=%d, diff=%d, logOnly=%d, stdOnly=%d\n",
        name, r.TotalLogRecords, r.RecordsWithDiff,
        r.LogOnlyMatchKeys, r.StdOnlyMatchKeys)
}
```

更多公开 API 见 [`pkg.go.dev`](https://pkg.go.dev/github.com/TrailHuang/log-compare)。

## 命令行用法

```text
用法: log-compare -config <配置文件> -stddir <标准日志目录> -logdir <待对比日志目录> [-output <报告文件>]
```

常用参数：

- `-config`  配置文件路径（默认 `conf/log_info.json`）
- `-stddir`  标准日志目录
- `-logdir`  待对比日志目录
- `-output`  报告输出文件路径（不指定则不输出）
- `-version` 显示版本信息

退出码：`0` 表示无差异，`1` 表示存在差异或错误。

## 项目结构

```
.
├── cmd/                 # 命令行入口
├── logcompare/          # 作为库的公开 API（Run 等）
├── config/              # 配置加载与校验
├── reader/              # 日志读取（txt / tar）
├── matcher/             # 按匹配键分组与匹配
├── comparator/          # 字段级差异比对
├── validator/           # 必填字段校验
├── reporter/            # 终端 / 文件报告输出
├── model/               # 公共数据结构
├── conf/                # 示例配置
└── testdata/            # 测试数据
```

## 开发

```bash
go vet ./...
go test ./...
make build        # 构建 linux/amd64
make build-all    # 构建 linux/amd64 与 linux/arm64
make package      # 打包发布物
```

## License

[MIT](./LICENSE)