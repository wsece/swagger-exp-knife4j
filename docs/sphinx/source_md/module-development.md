# 模块开发与扩展接口

本文档面向**二次开发者**与 **AI 编程助手**：说明如何在 **swagger-exp-knife4j** 上增加自定义功能（扫描钩子、CLI 子命令、MCP 工具、自定义输出），包括**必须实现的接口**、**类型与返回值**、**可用依赖**与**目录约定**。

```{note} AI 助手使用说明
实现新功能前请通读「扩展点总览」与对应接口章节；仅使用 `go.mod` 已有依赖；在 `extensions/<name>/` 新建包并在 `main.go` 空白导入；不要修改 `pkg/scanner` 核心解析逻辑，优先用 `ScanHook` 或调用 `scanrun.RunSingle`。
```

## 扩展点总览 {#overview}

| 扩展点 | 接口 | 注册函数 | 生效时机 |
|--------|------|----------|----------|
| 扫描流水线钩子 | `extension.ScanHook` | `extension.RegisterScanHook` | 每次 `scan single` / MCP `swagger_scan` / 批量扫描中的单次 `RunSingle` |
| 自定义扫描输出 | `extension.ScanWriter` | `extension.RegisterScanWriter` | 探测完成后、内置 CSV/JSONL/DB 写入之前 |
| CLI 子命令 | `extension.CommandExtension` | `extension.RegisterCommand` | 进程启动时 `init()` 注册，`cmd.Execute()` 挂到根命令 |
| MCP 工具 | `extension.MCPTool` | `extension.RegisterMCPTool` | `mcp serve` 启动时挂到 MCP Server |

包路径：**`swagger-exp-knife4j/pkg/extension`**  
参考实现：**`extensions/example/`**（需在 `main.go` 中 `import _ "swagger-exp-knife4j/extensions/example"` 才会编入二进制）。

## 目录与启用方式 {#layout}

```text
swagger-exp-knife4j/
├── main.go                 # 在此空白导入你的扩展包
├── extensions/
│   └── myplugin/
│       ├── register.go     # init() 里 Register*
│       ├── scan_hook.go
│       └── command.go
├── pkg/
│   ├── extension/          # 接口与注册表（勿改业务逻辑，只扩展）
│   ├── scanrun/            # RunSingle 流水线（已调用扩展钩子）
│   ├── scanner/            # HTTP、OpenAPI 解析与探测
│   └── ...
```

**启用扩展（必须）：**

```go
// main.go
package main

import (
	_ "swagger-exp-knife4j/extensions/myplugin" // 触发 init 注册

	"swagger-exp-knife4j/cmd"
)

func main() {
	cmd.Execute()
}
```

```bash
go build -o swagger-exp-knife4j .
./swagger-exp-knife4j hello-myplugin   # 若注册了 CommandExtension
```

## 依赖与代码边界 {#deps}

### 允许使用的 Go 模块

仅使用仓库 **`go.mod` 已声明**的依赖，例如：

| 用途 | 包 |
|------|-----|
| CLI | `github.com/spf13/cobra` |
| 日志 | `swagger-exp-knife4j/pkg/log` |
| 扫描流水线 | `swagger-exp-knife4j/pkg/scanrun` |
| HTTP / OpenAPI | `swagger-exp-knife4j/pkg/scanner` |
| 数据模型 | `swagger-exp-knife4j/pkg/models` |
| 数据库写入 | `swagger-exp-knife4j/pkg/writers`、`gorm.io/gorm` |
| MCP | `github.com/mark3labs/mcp-go/mcp`、`.../server` |
| 标准库 | `context`、`encoding/json`、`fmt`、`net/http`、`os`、`path/filepath`、`time` 等 |

**不要**在扩展里 `go get` 新第三方库（除非合入主仓库 `go.mod`）。  
**不要**在 `init()` 里启动长期 goroutine 或监听端口（报告服务请用现有 `report server` 或独立命令）。

### 推荐放置代码的位置

| 需求 | 做法 |
|------|------|
| 改扫描流程某一步 | 实现 `ScanHook` |
| 额外落盘格式 | 实现 `ScanWriter` 或钩子内调用 `scanner.WriteReportToCSV` 等 |
| 新 CLI 功能 | `CommandExtension` + 内部调用 `scanrun.RunSingle` |
| 给 Cursor/Claude 用 | `MCPTool` |
| 改 Knife4j 解析规则 | 需改 `pkg/scanner`（非扩展接口，提 PR） |

## ScanHook：扫描流水线钩子 {#scan-hook}

### 接口定义

```go
type ScanHook interface {
    Name() string
    Phases() []ScanPhase
    OnScan(ctx *ScanContext) error
}
```

| 方法 | 类型 | 说明 |
|------|------|------|
| `Name()` | `string` | 全局唯一 ID，建议 `厂商.功能`，如 `acme.header-inject` |
| `Phases()` | `[]ScanPhase` | 只在这些阶段调用 `OnScan` |
| `OnScan` | `func(*ScanContext) error` | 返回非 `nil` **中止整次扫描**；应用 `fmt.Errorf("hook: %w", err)` 包装 |

### ScanPhase 枚举

| 常量 | 含义 |
|------|------|
| `PhaseBeforeResolve` | 解析 Swagger JSON URL 之前 |
| `PhaseAfterResolve` | 已得到 `ResolvedJSONURL` |
| `PhaseAfterSaveAPIDocs` | 已写入 `output/{host}/{scope}/api-docs.json` |
| `PhaseAfterAnalyze` | 已得到 `Stats`（待探测 API 列表） |
| `PhaseAfterProbe` | 已完成 `ProbeResults`（自动 GET/POST） |
| `PhaseBeforeWrite` | 内置 CSV/JSONL/DB 与 `ScanWriter` 之前 |
| `PhaseAfterWrite` | 全部写入成功之后 |

### ScanContext 字段（按阶段可用）

| 字段 | 类型 | 最早可用阶段 |
|------|------|----------------|
| `InputURL` | `string` | 全程 |
| `OutputDir` | `string` | 全程 |
| `HTTP` | `*scanner.HTTPOptions` | 全程（可改 Headers 等，需自行理解对后续请求的影响） |
| `ResolvedJSONURL` | `string` | `PhaseAfterResolve` 起 |
| `APIDocsPath` | `string` | `PhaseAfterSaveAPIDocs` 起 |
| `Stats` | `[]scanner.APIStatisticsInfo` | `PhaseAfterAnalyze` 起 |
| `ProbeResults` | `[]scanner.APIRequestResult` | `PhaseAfterProbe` 起 |
| `HTTPMeta` | `scanner.HTTPRequestMeta` | `PhaseBeforeWrite` 起 |
| `Abort` | `error` | 设为非 `nil` 时流水线立即以该错误结束 |

**只读建议：** 除 `HTTP`、自定义副作用外，尽量不要清空切片；不要并发写 `ScanContext`。

### 相关数据类型（pkg/scanner）

```go
// 单个待测 API
type APIStatisticsInfo struct {
    Method     string
    Path       string
    Parameters []APIAutoFillParam
}

// 单次自动探测结果
type APIRequestResult struct {
    Host, Method, Path, FullURL, FinalURL string
    RequestParams, RequestBody, Response  string
    RequestHeaders, ResponseHeaders       []HTTPHeaderKV
    StatusCode                            int
    ContentType                           string
    DurationMs                            int64
    Error                                 string
}

// HTTP 客户端配置（与 CLI -H/-A/-b/-x/-P 对应）
type HTTPOptions struct {
    Headers []string
    UserAgent, Proxy string
    Delay, RequestTimeout, ConnectTimeout time.Duration
    Parallel int
}
```

### 注册示例

```go
func init() {
    if err := extension.RegisterScanHook(&myHook{}); err != nil {
        panic(err)
    }
}

type myHook struct{}

func (myHook) Name() string { return "myteam.after-probe" }

func (myHook) Phases() []extension.ScanPhase {
    return []extension.ScanPhase{extension.PhaseAfterProbe}
}

func (myHook) OnScan(ctx *extension.ScanContext) error {
    for _, r := range ctx.ProbeResults {
        if r.StatusCode == 401 {
            return fmt.Errorf("unauthorized on %s", r.FullURL)
        }
    }
    return nil
}
```

## ScanWriter：自定义扫描输出 {#scan-writer}

```go
type ScanWriter interface {
    Name() string
    Write(ctx *ScanContext) error
}
```

- 在 **`PhaseBeforeWrite`** 阶段由 `scanrun` 调用（与 `ScanHook` 的 `PhaseBeforeWrite` 顺序：先所有 Hook，再所有 Writer，再内置 DB/CSV/JSONL）。
- `ctx` 中应已有 `Stats`、`ProbeResults`、`HTTPMeta`。
- 写入 DB 可复用：

```go
records := scanner.BuildSwaggerReportRecords(ctx.InputURL, ctx.ResolvedJSONURL, ctx.HTTPMeta, ctx.Stats, ctx.ProbeResults)
w, err := writers.NewSwaggerDbWriter("sqlite://out.sqlite3", false)
err = w.WriteRecords(records)
```

`models.SwaggerAPIRecord` 字段见 `pkg/models/swagger_scan.go`。

## CommandExtension：CLI 子命令 {#cli}

```go
type CommandExtension interface {
    CobraCommand() *cobra.Command
}
```

| 要求 | 说明 |
|------|------|
| `Use` | 子命令名，挂到**根命令**下，如 `my-scan` → `swagger-exp-knife4j my-scan` |
| `RunE` | 返回 `error` 表示失败；用 `pkg/log` 打日志 |
| 参数 | 用 `Flags()` 定义；**不要**依赖 `cmd` 包内未导出的 `opts` |

**调用内置扫描：**

```go
import "swagger-exp-knife4j/pkg/scanrun"

result, err := scanrun.RunSingle(scanrun.SingleParams{
    InputURL:  url,
    OutputDir: "output",
    HTTP:      &scanner.HTTPOptions{Parallel: 4},
    Writers: scanrun.WriterConfig{
        DbURI: "sqlite://swagger-scan.sqlite3",
    },
})
// result *scanrun.SingleResult — JSON 字段见下表
```

### SingleResult（JSON 友好）

| 字段 | 类型 | 含义 |
|------|------|------|
| `input_url` | string | 输入 URL |
| `resolved_json_url` | string | 解析后的 OpenAPI JSON |
| `api_docs_path` | string | 本地 api-docs 路径 |
| `path_count` | int | API 条数 |
| `request_count` | int | 探测请求总数 |
| `request_ok` | int | 探测成功数 |
| `request_failed` | int | 探测失败数 |
| `unauthorized_count` | int | 非 401 响应数（可能未授权） |
| `wrote_db` / `wrote_csv` / `wrote_jsonl` | bool | 是否写入 |
| `db_uri` / `csv_file` / `jsonl_file` | string | 输出路径 |

## MCPTool：MCP 工具 {#mcp}

```go
type MCPTool interface {
    Name() string
    Definition() mcp.Tool
    Handler(defaults extension.MCPDefaults) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
}
```

| 部分 | 约定 |
|------|------|
| `Name()` | 与 `Definition()` 中工具名一致，**snake_case**，勿与 `swagger_scan` 等冲突 |
| `Definition()` | `mcp.NewTool("my_tool", mcp.WithDescription("..."), mcp.WithString("url", mcp.Required()))` |
| `Handler` | 成功：`mcp.NewToolResultText(string)`；失败：`mcp.NewToolResultError(msg), nil` |
| 参数读取 | `req.RequireString("url")`、`req.GetString("k", default)`、`req.GetBool`、`req.GetFloat` |
| 默认值 | `defaults.DefaultDbURI`、`defaults.DefaultAPIDocPath` |

扫描类工具应委托 **`scanrun.RunSingle`**，与内置 `swagger_scan` 一致。

```go
func (t *myTool) Handler(def extension.MCPDefaults) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        _ = ctx
        url, err := req.RequireString("url")
        if err != nil {
            return mcp.NewToolResultError(err.Error()), nil
        }
        res, err := scanrun.RunSingle(scanrun.SingleParams{InputURL: url, OutputDir: def.DefaultAPIDocPath})
        if err != nil {
            return mcp.NewToolResultError(err.Error()), nil
        }
        text, err := scanrun.SingleResultJSON(res)
        if err != nil {
            return mcp.NewToolResultError(err.Error()), nil
        }
        return mcp.NewToolResultText(text), nil
    }
}
```

## 注册与错误处理 {#register}

```go
// register.go
func init() {
    must(extension.RegisterScanHook(&h{}))
    must(extension.RegisterCommand(&c{}))
    must(extension.RegisterMCPTool(&t{}))
    must(extension.RegisterScanWriter(&w{}))
}

func must(err error) {
    if err != nil {
        panic("myplugin: " + err.Error())
    }
}
```

| 规则 | 说明 |
|------|------|
| 重复 `Name()` | `Register*` 返回 `error`，`init` 中应 `panic` 或记录 |
| 中止扫描 | `OnScan` 返回 error，或设置 `ctx.Abort = err` |
| 日志 | 见下表；MCP `mcp serve` 默认 `EnableSilence()`（仅 `Print` 可见） |

扩展模块日志约定（`pkg/log`）：

| API | 用途 |
|-----|------|
| `log.Print` | 用户必须看到的结果（`-q` 仍输出） |
| `log.Info` | 正常进度（`-q` 隐藏） |
| `log.Debug` | 排障细节（需 `-D`） |
| `log.Warn` / `log.Error` | 配置提示与可继续错误 |
| 并发 | `RunFile` 顺序调用 `RunSingle`；Hook 内如需共享状态请自行 `sync.Mutex` |

## 内置流水线顺序（供 Hook 选型） {#pipeline}

```text
PhaseBeforeResolve
  → ResolveSwaggerJSONURL
PhaseAfterResolve
  → SaveAPIDocsJSON
PhaseAfterSaveAPIDocs
  → AnalyzeSwaggerAPI
PhaseAfterAnalyze
  → AutoRequestAllAPI
PhaseAfterProbe
  → HTTPMeta
PhaseBeforeWrite
  → [ScanWriter...] → CSV / JSONL / DB
PhaseAfterWrite
```

## 完整最小插件模板 {#template}

```go
// extensions/myplugin/register.go
package myplugin

import "swagger-exp-knife4j/pkg/extension"

func init() {
    must(extension.RegisterScanHook(&hook{}))
}

func must(err error) {
    if err != nil {
        panic(err)
    }
}
```

```go
// extensions/myplugin/hook.go
package myplugin

import "swagger-exp-knife4j/pkg/extension"

type hook struct{}

func (hook) Name() string { return "myplugin.noop" }
func (hook) Phases() []extension.ScanPhase {
    return []extension.ScanPhase{extension.PhaseAfterProbe}
}
func (hook) OnScan(ctx *extension.ScanContext) error { return nil }
```

## 测试与验证 {#test}

```bash
# 在 main.go 加入 import _ "swagger-exp-knife4j/extensions/myplugin"
go build -o swagger-exp-knife4j .
swagger-exp-knife4j scan single -u "https://example.com/v3/api-docs" --write-db
swagger-exp-knife4j mcp serve   # 若注册了 MCPTool
```

单元测试：在 `extensions/myplugin/*_test.go` 中对 `OnScan` 构造 `ScanContext` 填字段后直接调用（无需起完整 CLI）。

## 禁止事项 {#forbidden}

- 不要在扩展中替换或 fork `cmd.Execute` / `main` 逻辑
- 不要 import `cmd` 包（会循环依赖）
- 不要修改 `pkg/extension` 注册表以外的全局变量来传配置（用 flag / env）
- 不要在 Hook 里无超时地阻塞网络（使用 `HTTPOptions` 已有超时或自建带 `context` 的 client）
- 不要提交密钥；Cookie/Token 用 CLI `-b` / `-H` 传入

## 与 AI 协作的提示词片段 {#ai-prompt}

可将以下内容粘贴给 AI：

```text
在 swagger-exp-knife4j 上实现扩展：
1. 在 extensions/<name>/ 新建 Go 包，module 路径 swagger-exp-knife4j/extensions/<name>
2. 实现 pkg/extension 中的 ScanHook / CommandExtension / MCPTool / ScanWriter 之一
3. 在 register.go 的 init() 调用 extension.Register*
4. 在 main.go 添加 import _ "swagger-exp-knife4j/extensions/<name>"
5. 仅使用 go.mod 已有依赖；扫描逻辑复用 scanrun.RunSingle
6. 契约以 docs/sphinx/source_md/module-development.md 为准
```

[项目概览](introduction.md) · [MCP 工具契约](mcp-tools.md) · [构建文档站](building-site.md)
