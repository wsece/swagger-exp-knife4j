# swagger-exp-knife4j

![swagger-exp-knife4j](https://socialify.git.ci/wsece/swagger-exp-knife4j/image?description=1&descriptionEditable=The%20Swagger%20interface%20based%20on%20Knife4j%20automates%20testing%20of%20unauthorized%20tools&font=Jost&forks=1&issues=1&language=1&name=1&owner=1&pattern=Solid&stargazers=1&theme=Dark)



## 前言

一款 Swagger / Knife4j / OpenAPI 接口自动发现与检测工具。

swagger-exp-knife4j，基于 Swagger（OpenAPI 2/3）与 Knife4j 文档页习惯用法设计，核心是用于批量自动化扫描接口文档是否存在未授权访问漏洞。

支持双端（Web、CLI ）跨平台独立运行，做到团队协作。扫描结果支持 web 报告可视化呈现，自研相似度算法完成接口响应数据的梳理，自动评估风险等级，burp 风格展示请求响应，可联动 Knife4j 文档实现在线接口请求，也能导出多格式 API 文档。

支持 MCP 调用，内置 3 款工具完成扫描到报告输出的流水线，定义了 MCP 工具契约及扩展接口规范，减少你二开时与 AI 交互 Token 消耗。

## 技术栈

| 类别     | 选型                                                   |
| -------- | ------------------------------------------------------ |
| 语言     | Go（见 `go.mod`）                                      |
| CLI      | [Cobra](https://github.com/spf13/cobra)                |
| ORM / DB | GORM + SQLite                                          |
| Web      | 标准库 `net/http`，静态资源 `go:embed`                 |
| MCP      | [mcp-go](https://github.com/mark3labs/mcp-go)（stdio） |

## 特色

| 能力               | 说明                                                         |
| ------------------ | ------------------------------------------------------------ |
| 数据可视化 Web     | Web 管理端，支持 Host 维度记录查询；Burp 式 Request/Response 详情；列表去重与相似度排序；大 body 分片预览 |
| Web 下发扫描       | 报告页内提交单 URL / 批量 / 文件任务                         |
| 多端运行模式       | 适配 SKILL 的 CLI + 可视化 Web 服务双端运行                  |
| 多源文档接入       | 兼容 `swagger.json`、`/v3/api-docs`、Knife4j、doc.html 等各类接口文档入口地址 |
| 响应相似度分析算法 | 响应相似度分析，入库 SimHash 与汉明分组（`internal/islazy`），报告 API 聚类字段 |
| 请求体系可配置     | 自定义请求头、Cookie、网络代理、请求并发数、超时时间，用法兼容 curl 常用参数 |
| MCP 对接           | 开放标准 MCP 调用接口，支持大模型直接调用扫描，或 AI 分析结果 |
| 标准化结果记录     | 完整记录接口元数据、原始请求报文、响应报文，结构化 JSON 存储，适配 Burp 联动格式 |
| 接口拓展           | 面向二次开发者与 AI 预留拓展接口，相关约束位于文档目录 “模块开发与拓展” |

## 命令列表

三大模块：scan、report、mcp

```bash
Usage:
  swagger-exp-knife4j [command]

Available Commands:
  help        Help about any command
  mcp         Model Context Protocol (MCP) server for AI clients
  report      View stored Swagger scan reports
  scan        Scan Swagger API interfaces
  version     Show build version information

Flags:
  -D, --debug-log   Show debug logging
  -h, --help        help for swagger-exp-knife4j
  -q, --quiet       Silence (almost all) logging

```



scan 模块子命令：

```bash
## This module[scan] scans Swagger/Knife4j/OpenAPI targets and auto-detect interfaces.
## Subcommands: single (single URL), file (URL list file).
## Flags for write config, HTTP options, output directory and more are shown below.

Usage:
  swagger-exp-knife4j scan [command]

Examples:
swagger-exp-knife4j scan single -u https://example.com/doc.html --write-db
swagger-exp-knife4j scan file -f targets.txt --write-db

Available Commands:
  file        Scan multiple targets listed in a file
  single      Scan a single URL target

Flags:
      --connect-timeout duration   Max wait for TCP connect (default 30s)
  -b, --cookie stringArray         Cookie string or file (repeatable), e.g. -b "session=abc"
      --delay duration             Sleep between each API request (e.g. 100ms, 1s)
      --docs-only                  Only resolve and dump OpenAPI JSON to --output-dir; skip automate
d API requests and --write-* outputs
  -H, --header stringArray         Custom request header (repeatable), e.g. -H "Authorization: Beare
r xxx"
  -h, --help                       help for scan
  -m, --max-timeout duration       Per-request timeout (0 = unlimited)
      --output-dir string          Base directory for scan output ({host}/{scope}/api-docs.json) (de
fault "output")
  -P, --parallel int               Concurrent API request workers (default 1)
  -x, --proxy string               HTTP proxy URL, e.g. -x http://127.0.0.1:8080
  -A, --user-agent string          User-Agent string, e.g. -A "Mozilla/5.0"
      --write-csv                  Write scan results to CSV (default result.csv)
      --write-csv-file string      CSV file path (overrides --write-csv default)
      --write-db                   Write scan results to database (default sqlite://swagger-scan.sql
ite3)
      --write-db-enable-debug      Show the database query debug logging
      --write-db-uri string        Database URI (overrides --write-db default)
      --write-jsonl                Write scan results to JSONL (default result.jsonl)
      --write-jsonl-file string    JSONL file path (overrides --write-jsonl default)

Global Flags:
  -D, --debug-log   Show debug logging
  -q, --quiet       Silence (almost all) logging

Use "swagger-exp-knife4j scan [command] --help" for more information about a command.

```

report 模块子命令：

```bash
## This module[report] enables report viewing and displays scanned results.。

Usage:
  swagger-exp-knife4j report [command]

Examples:
swagger-exp-knife4j report list
swagger-exp-knife4j report server

Available Commands:
  list        List API scan records from the database
  server      Start local web server to browse scan results

Flags:
      --db-uri string   Swagger scan database URI (e.g. sqlite://swagger-scan.sqlite3) (default "sqlite://swagger-scan.
sqlite3")
  -h, --help            help for report

Global Flags:
  -D, --debug-log   Show debug logging
  -q, --quiet       Silence (almost all) logging

Use "swagger-exp-knife4j report [command] --help" for more information about a command.

```

mcp 模块子命令：

```bash
## This module[mcp] is used to start the MCP service for invocation by large language models.

Usage:
  swagger-exp-knife4j mcp [command]

Examples:
swagger-exp-knife4j mcp serve

Available Commands:
  serve       Run MCP server on stdio (for Cursor / Claude Desktop)

Flags:
  -h, --help   help for mcp

Global Flags:
  -D, --debug-log   Show debug logging
  -q, --quiet       Silence (almost all) logging

Use "swagger-exp-knife4j mcp [command] --help" for more information about a command.

```



## 快速开始

```bash
# 编译项目
go build -o swagger-exp-knife4j .

# 扫描，-u 指定swaggerUI页或 OpenAPI JSON 页，对GET/POST接口发起自动化探测，--write-db 写入默认数据库
swagger-exp-knife4j scan single -u https://example.com/doc.html --write-db -q

# 打开本地 http 服务查看报告，支持 Web 端批量提交扫描任务。
swagger-exp-knife4j report server
```

浏览器打开 http://127.0.0.1:7171/

- 左侧选 **Host** 加载探测记录表（数据源为 sqliteDB ）
- 中栏可选：**仅最新请求**（按 Method+路径去重）、**按相似响应排序**（基于 SimHash 聚类字段）
- 选中行后右侧查看 Request/Response（body 超过 64 KiB 时仅预览片段）
- **API Docs** 可预览 OpenAPI（数据源为 ./output）；**Open Knife4j** 经报告站同源代理转发请求，页面的操作参考官网 https://doc.xiaominfo.com/docs/blog
- 顶部 **下发扫描任务**（默认折叠）可提交新的扫描

![pic_report](./docs/pic_report.jpg)

![pic_report2](./docs/pic_report2.jpg)

![pic_knife4j](./docs/pic_knife4j.jpg)

## Docker 部署

```bash
docker compose up -d --build
```

数据卷持久化 SQLite 与 `output/`。扫描示例见 [docs/docker.md](docs/docker.md)。

## 文档

项目白皮书文档 :  [wiki_swagger-exp-knife4j](https://wsece.github.io/wiki_swagger-exp-knife4j/)

## 许可与免责

如果您下载使用即表明您信任本工具，在使用时造成的损失和伤害，我们不承担任何责任。使用本工具的过程中存在任何非法行为需自行承担相应后果，我们将不承担任何法律及连带责任。您的下载、安装、使用等行为即视为您已阅读并同意上述协议的约束


![Star History Chart](https://api.star-history.com/svg?repos=wsece/swagger-exp-knife4j&type=Date)

