# scan 扫描

## 命令结构 {#structure}

```text
swagger-exp-knife4j scan [子命令] [scan 级参数]
```

| 子命令 | 作用 |
|--------|------|
| `single` | 扫描单个 URL，必填 `-u` / `--url` |
| `file` | 从文件批量扫描，必填 `-f` / `--file` |

子命令共享下方 **scan 持久化参数** 与 **HTTP 参数**（在 `scan` 上声明，对 single / file 均生效）。

---

## scan single {#scan-single}

### 功能

对一条 Swagger / Knife4j 入口 URL 执行完整流水线：解析 JSON 地址 → 保存 api-docs → 分析 paths → 自动请求 GET/POST → 可选写入 DB/CSV/JSONL。

### 示例

```bash
# 仅统计，不落库
swagger-exp-knife4j scan single -u https://example.com/swagger.json

# 写入默认 SQLite
swagger-exp-knife4j scan single -u https://example.com/doc.html --write-db

# 指定库与输出目录
swagger-exp-knife4j scan single -u https://example.com/v2/api-docs \
  --write-db-uri sqlite://./data/scan.sqlite3 \
  --output-dir ./output

# 仅转储 OpenAPI JSON，不发起接口探测
swagger-exp-knife4j scan single -u https://example.com/doc.html --docs-only
```

### 子命令参数

| 参数 | 简写 | 必填 | 说明 |
|------|------|------|------|
| `--url` | `-u` | 是 | 目标 URL |

---

## scan file {#scan-file}

### 功能

按行读取 URL 文件并依次执行与 `single` 相同的流程；指定 `--write-csv` / `--write-jsonl` 时，**所有 URL 的结果写入同一文件**——第一个 URL 创建文件，后续 URL **追加** 行（不会覆盖前一个 URL 的记录）。

### 目标文件格式

- 每行一个 URL  
- 空行忽略  
- `#` 开头为注释  

### 示例

```bash
swagger-exp-knife4j scan file -f targets.txt --write-db
```

---

## scan 共享参数：输出 {#scan-output}

| 参数 | 默认 | 说明 |
|------|------|------|
| `--write-db` | false | 写入 SQLite；成功响应写入 SimHash（`response_sim_hash`）及分组字段（`response_group_id`） |
| `--write-db-uri` | — | 覆盖数据库 URI，如 `sqlite://xxx.sqlite3` |
| `--write-db-enable-debug` | false | 打印 GORM SQL |
| `--write-csv` | false | 写 CSV，默认 `result.csv` |
| `--write-csv-file` | — | 指定 CSV 路径 |
| `--write-jsonl` | false | 写 JSONL，默认 `result.jsonl` |
| `--write-jsonl-file` | — | 指定 JSONL 路径 |
| `--output-dir` | `output` | api-docs 根目录：`{host}/{scope}/api-docs.json` |
| `--docs-only` | false | 仅解析并转储 OpenAPI JSON 到 `--output-dir`；跳过自动 API 请求；`--write-*` 不生效 |

默认数据库 URI：`sqlite://swagger-scan.sqlite3`

写库流程中的指纹与分组算法见 [响应智能分析](smart-analysis.md)。

---

## scan 共享参数：HTTP {#scan-http}

| 参数 | 简写 | 默认 | 说明 |
|------|------|------|------|
| `--header` | `-H` | — | 请求头，可重复，格式 `Key: Value` |
| `--user-agent` | `-A` | 内置 Chrome UA | User-Agent |
| `--cookie` | `-b` | — | Cookie 或 Cookie 文件路径，可重复 |
| `--proxy` | `-x` | — | HTTP 代理 |
| `--delay` | — | 0 | 每次 API 请求间隔 |
| `--max-timeout` | `-m` | 0 | 单次请求总超时，0 不限制 |
| `--connect-timeout` | — | 30s | TCP 连接超时 |
| `--parallel` | `-P` | 1 | 并发 worker 数 |

---

## 全局日志参数（根命令） {#scan-logging}

`scan` 继承根命令的日志开关（对 `scan single` / `scan file` 均生效）：

| 参数 | 说明 |
|------|------|
| `-q` / `--quiet` | 静默模式：隐藏 **Info** 进度（解析 URL、保存 api-docs、探测起止、写出路径等）；**不**隐藏最终 **`Scan Finished`** 摘要块 |
| `-D` / `--debug-log` | 调试模式：额外输出每个 API/参数、未授权候选接口等 **Debug** 详情 |

日志输出到 **stderr**；`report list` 的表格数据在 **stdout**，便于管道重定向。

### 日志分级（实现见 `pkg/log`）

| 级别 | 典型内容 | `-q` 时 |
|------|----------|---------|
| **Print** | 最终 **`Scan Finished`** / **`Batch Scan Finished`** 摘要块 | 仍显示 |
| **Info** | 解析 OpenAPI、api-docs 落盘、探测开始/结束、CSV/DB 路径 | 隐藏 |
| **Debug** | 每个 API 参数、未授权候选（host/path/status） | 仅 `-D` |
| **Warn** | 未配置 `--write-*`、批量部分失败 | 隐藏 |

## 终端输出说明

### 进度行（Info，`-q` 隐藏）

扫描过程中按步骤输出对齐日志，例如：

```text
2026-05-20 12:23:58  INFO  Found SwaggerUI  | page: https://example.com/index.html
2026-05-20 12:23:58  INFO  Found OpenAPI    | json: https://example.com/swagger/v1/swagger.json
2026-05-20 12:23:58  INFO  Dump OpenAPI     | file: output/example.com/swagger/v1/api-docs.json
2026-05-20 12:23:58  INFO  Found API number | Api:12
2026-05-20 12:23:58  INFO  Request automate | swagger: ... | parallel=1 | delay=0s
2026-05-20 12:23:58  INFO  Request finished | total=12 | ok=12 | skip=0 | fail=0
```

直连 OpenAPI JSON 时无 `Found SwaggerUI` 行。启用 `--write-db` 等后另有 `Write database` 等行。

`--docs-only` 时无 `Request automate` / `Request finished`，改为 `Docs only | skipped automated API requests`；`Scan Finished` 中 `Request[*]` 均为 0。

### 最终摘要（`Scan Finished`，`-q` 仍显示）

```text
2026-05-20 12:23:58  Scan Finished
	Input: https://example.com/index.html
	Result: API=12 | Request[ok]=12 | Request[skip]=0 | Request[fail]=0 | Unauthorized=3
	DumpJson: output/example.com/swagger/v1/api-docs.json
```

| 字段 | 含义 |
|------|------|
| `API` | OpenAPI 待测接口数 |
| `Request[ok]` | 已收到 HTTP 状态且无错误 |
| `Request[skip]` | 跳过（非 GET/POST 等） |
| `Request[fail]` | 构建或网络失败 |
| `Unauthorized` | 非 401 响应数（仅 >0 时显示） |
| `DumpJson` | 本地 `api-docs.json` 路径 |

`scan file` 对每个 URL 重复上述进度 + `Scan Finished`；全部结束后输出 **`Batch Scan Finished`** 汇总。未配置 `--write-*` 时附加 **WARN** 提示。

解析失败（找不到 Swagger JSON 等）同样输出 **`Scan Finished`** 块，`Result` 为可读错误说明，`DumpJson: none`，不再单独打印 `scan error:` 行。

### 其它输出

- 未授权接口的逐条列表仅在 **`-D`** 下以 Debug 输出；`-q` **不会**单独打印未授权统计行。
- 启用 `--write-db` / `--write-csv` / `--write-jsonl` 时，写出路径在 **Info**（`-q` 隐藏）。
- 数据库中单条 API 一行，含方法、路径、状态码、响应分组（SimHash）及 `packet_json` 抓包字段。

[上一章：快速入门](tutorial.md) · [report 报告](commands-report.md) · [首页](index.md)
