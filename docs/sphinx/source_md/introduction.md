# 项目概览

## 介绍 {#intro}

**swagger-exp-knife4j** 是面向 **Swagger / OpenAPI / Knife4j** 类 API 文档的 **自动化测试工具**，核心用于接口资产梳理与未授权访问漏洞的批量检测。

1. 支持接口文档直链、网页地址两种接入方式，自动解析获取标准 OpenAPI 规范源地址
2. 批量抓取接口文档，按域名、业务维度分层归档，本地持久化存储为 `api-docs.json`（按 host / scope 分目录）
3. 智能解析接口路由，自动填充请求参数，对 GET、POST 等常用接口批量发起请求，探测未授权访问风险
4. 全量留存请求与响应数据，支持 SQLite、CSV、JSONL 多格式结果存储
5. 内置可视化报告站点，兼容 Knife4j 原生页面，便捷完成接口查看、在线调试与结果复盘

> 适用于授权范围内企业内部接口资产盘点、未授权访问漏洞粗筛、批量接口资产合规扫描、接口自动化巡检等场景。

## 特色 {#features}

| 能力              | 说明                                                         |
| ----------------- | ------------------------------------------------------------ |
| 多源文档接入      | 兼容 `swagger.json`、`/v3/api-docs`、Knife4j、doc.html 等各类接口文档入口地址 |
| HTTP 可配置       | 自定义请求头、Cookie、网络代理、请求并发数、超时时间，用法兼容 curl 常用参数 |
| 标准化记录        | 完整记录接口元数据、原始请求报文、响应报文，结构化 JSON 存储，适配 Burp 联动格式 |
| 可视化 Web 管理端 | Host 维度记录查询；Burp 式 Request/Response 详情；列表去重与相似度排序；大 body 分片预览 |
| 响应相似度分析 | 入库 SimHash 与汉明分组（`internal/islazy`）；报告 API 聚类字段；见 [响应智能分析](smart-analysis.md) |
| Web 下发扫描      | 报告页内提交单 URL / 批量 / 文件任务                         |
| MCP 对接          | 开放标准 MCP 调用接口，支持大模型直接调用扫描，或 AI 分析结果 |
| 多端运行模式      | 适配 SKILL 的 CLI + 可视化 Web 服务双端运行                  |
| 接口拓展          | 面向二次开发者与 AI 预留拓展接口，相关约束位于文档目录 "模块开发与拓展" |

## 项目架构 {#architecture}

### 总体分层

```text
┌─────────────────────────────────────────────────────────────┐
│  入口层                                                      │
│  main.go  →  cmd/（Cobra：version / scan / report / mcp）     │
└───────────────────────────┬─────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────┐
│  编排层                                                      │
│  pkg/scanrun        单次/批量扫描流水线                        │
│  pkg/config         CLI 全局与子命令参数（Options）            │
└───────────────────────────┬─────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
┌───────────────┐  ┌───────────────┐  ┌───────────────────┐
│ pkg/scanner   │  │ pkg/writers   │  │ pkg/reportserver  │
│ 解析 URL       │  │ SQLite 写入   │  │ Web 报告 + API     │
│ 保存 api-docs  │  │               │  │ Knife4j 反向代理   │
│ 统计 paths     │  │               │  │ 页面下发扫描任务    │
│ 并发探测请求    │  │               │  │                   │
│ Burp 风格抓包  │  │               │  │                   │
└───────┬───────┘  └───────┬───────┘  └─────────┬─────────┘
        │                  │                    │
        └──────────────────┼────────────────────┘
                           ▼
              ┌────────────────────────┐
              │ pkg/models             │
              │ pkg/database（GORM）    │
              │ internal/islazy        │
              │   批量目标文件解析        │
              └────────────────────────┘
                           │
              ┌────────────▼────────────┐
              │ pkg/mcpserver（stdio）   │
              │ 暴露 swagger_scan 等工具  │
              └─────────────────────────┘
```

### 目录与职责

| 路径 | 职责 |
|------|------|
| `main.go` | 进程入口，调用 `cmd.Execute()` |
| `cmd/` | Cobra 命令树：`version`、`scan single\|file`、`report list\|server`、`mcp serve`；解析参数并调用下层 |
| `internal/version/` | 发行版本与构建元数据（`version` 子命令输出源） |
| `pkg/config/` | `Options`：URL、HTTP（头/Cookie/代理/并发/超时）、输出目录、数据库 URI 等 |
| `pkg/scanrun/` | **扫描编排**：`RunSingle` / `RunFile` / `RunURLs`；CLI 与 MCP 共用 |
| `pkg/scanner/` | **扫描核心**：解析 Knife4j/HTML → OpenAPI URL；落盘 `api-docs.json`；解析 paths；自动 GET/POST 探测；`http_exchange` 抓包 |
| `pkg/writers/` | 将 `SwaggerAPIRecord` 写入 SQLite |
| `pkg/database/` | GORM 模型与迁移 |
| `pkg/models/` | `SwaggerAPIRecord`（含 `packet_json`、SimHash 分组等） |
| `pkg/reportserver/` | 本地 HTTP 服务：记录查询、OpenAPI 预览、Knife4j 代理、`POST /api/scan/tasks` |
| `pkg/reportserver/static/` | 嵌入的前端报告页（`index.html`、Knife4j 静态资源） |
| `pkg/mcpserver/` | MCP stdio 服务：`swagger_scan`、`swagger_list_hosts`、`swagger_list_sessions` |
| `pkg/extension/` | **扩展注册表**：`ScanHook` / `ScanWriter` / `CommandExtension` / `MCPTool` |
| `extensions/` | 用户自定义模块目录（空白导入 `main.go` 启用） |
| `pkg/log/` | 统一日志（Print / Info / Debug；CLI `-q`、`-D`） |
| `internal/islazy/` | 批量目标解析；**响应 SimHash / 汉明聚类 / 相似度排序键**；供写库与 Web 报告复用 |
| `output/` | 运行时生成：`{host}/{scope}/api-docs.json` |
| `docs/sphinx/source_md/` | 文档源稿（`*.md`）；Sphinx 工程在 `docs/sphinx/` |
| `Dockerfile` / `docker-compose.yml` | 容器化部署（默认 `report server` + 数据卷 `/data`） |

### 扫描数据流（scan single）

```text
URL（JSON 或 doc.html）
  → swagger_url_resolve     解析真实 OpenAPI JSON 地址
  → swagger_docs_save       写入 output/{host}/{scope}/api-docs.json
  → swagger_api_statistics  展开 paths，生成待测 API 列表
  → swagger_api_request_exec  并发请求（可配置 -H/-b/-x/-P/-m）
  → writers / CSV / JSONL     可选持久化（DB 写入时 SimHash 自动分组 response_group_id）
  → 终端 Scan Finished 摘要块（Input / Result / DumpJson，见 commands-scan §终端输出）
```

### 报告与 MCP 数据流

```text
report server
  → Store：读 SQLite 记录 + 枚举 output/ 下 api-docs 会话
  → 浏览器：记录表（去重/相似度排序）/ 报文详情 / Knife4j（同源反向代理）
  → Web 下发任务 → scan_jobs → 复用 pkg/scanrun

mcp serve
  → swagger_scan           同 scan single 流水线
  → swagger_list_hosts     读 DB（需先 write_db）
  → swagger_list_sessions  读 output/ 目录
```

### 技术栈

| 类别 | 选型 |
|------|------|
| 语言 | Go（见 `go.mod`） |
| CLI | [Cobra](https://github.com/spf13/cobra) |
| ORM / DB | GORM + SQLite |
| Web | 标准库 `net/http`，静态资源 `go:embed` |
| MCP | [mcp-go](https://github.com/mark3labs/mcp-go)（stdio） |
| 说明文档站点 | Sphinx 模型（Read the Docs 主题） |

## 使用边界 {#limits}

- 仅用于**已获得授权**的目标；请遵守法律法规与单位安全制度。
- 自动探测可能对目标产生压力，请合理设置 `--delay`、`-P` 并发。
- Knife4j 调试页依赖本机 `report server`；默认 **Open Knife4j** 为浏览器直连目标 API，可选 **Proxy mode** 同源转发（见 [Web 报告站](web-report.md#knife4j)）。

## 与相关项目关系 {#relations}

- 基于 Swagger 生态（OpenAPI 2/3）与 Knife4j 文档页习惯用法设计。
- 命令行使用 [Cobra](https://github.com/spf13/cobra)；持久化使用 GORM + SQLite。

[返回文档首页](index.md)