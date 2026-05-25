# swagger-exp-knife4j MCP 工具说明（供大模型 / 开发者）

启动方式：

```bash
swagger-exp-knife4j mcp serve --db-uri sqlite://swagger-scan.sqlite3 --api-doc-path ./output
```

Cursor 配置见项目根目录 `mcp.cursor.json.example`。

---

## 全局约定 {#conventions}

| 项 | 说明 |
|----|------|
| 传输 | stdio（子进程 stdin/stdout） |
| 成功返回 | `CallToolResult` 的 `content[0].text` 为 **JSON 字符串** |
| 失败返回 | `isError: true`，`text` 为错误说明（非 JSON） |
| 默认 DB | `sqlite://swagger-scan.sqlite3` |
| 默认 output | `output/` |

---

## 工具依赖关系 {#deps}

```text
swagger_scan
    ├─► 写入 output/  ──► swagger_list_sessions 可读
    └─► 写入 DB（write_db）──► swagger_list_hosts 可读

swagger_list_sessions  不依赖 DB，但依赖磁盘上已有 api-docs
swagger_list_hosts     不依赖 output，但依赖 DB 里有扫描记录
```

| 你想做的事 | 应调用的 tool | 前置条件 |
|------------|---------------|----------|
| 扫描新 URL | `swagger_scan` | 无 |
| 看扫过哪些 host | `swagger_list_hosts` | 曾 `swagger_scan` 且 `write_db=true`（或 DB 已有数据） |
| 看有哪些 api-docs 会话 id | `swagger_list_sessions` | `output/` 下已有 `api-docs.json` |
| 浏览器里调试 API | **无** | 请用户运行 `report server` |

---

## swagger_scan {#swagger-scan}

**用途**：完整扫描单个 Swagger/Knife4j 目标（等同 `scan single -u <url>`）。

**何时用**：用户提供新的文档地址；需要自动请求接口并持久化。

**参数**

| 名称 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `url` | string | 是 | — | Swagger JSON 或 UI 入口 URL |
| `output_dir` | string | 否 | `output` | api-docs 落盘目录 |
| `write_db` | boolean | 否 | false | 是否写入 SQLite |
| `db_uri` | string | 否 | 见上 | 数据库 URI；仅 `db_uri` 也会写入 |
| `write_csv` | boolean | 否 | false | 写 CSV |
| `csv_file` | string | 否 | `result.csv` | CSV 路径 |
| `write_jsonl` | boolean | 否 | false | 写 JSONL |
| `jsonl_file` | string | 否 | `result.jsonl` | JSONL 路径 |
| `parallel` | number | 否 | 1 | API 探测并发数 |

**返回类型**：`string`（JSON 对象文本）

```json
{
  "input_url": "https://api.example.com/index.html",
  "resolved_json_url": "https://api.example.com/swagger/v1/swagger.json",
  "api_docs_path": "output/api.example.com/swagger/v1/api-docs.json",
  "path_count": 120,
  "request_count": 85,
  "request_ok": 80,
  "request_failed": 5,
  "unauthorized_count": 12,
  "wrote_db": true,
  "wrote_csv": false,
  "wrote_jsonl": false,
  "db_uri": "sqlite://swagger-scan.sqlite3"
}
```

| 字段 | 说明 |
|------|------|
| `path_count` | 文档中 API 数量 |
| `request_count` | 探测请求总数 |
| `request_ok` | 成功收到 HTTP 响应的次数 |
| `request_failed` | 跳过、构建错误或网络失败的次数 |
| `unauthorized_count` | 状态码 **非 401** 的响应数（可能未授权） |

**内部调用链**（非 MCP 工具，仅供理解）：

1. `scanner.ResolveSwaggerJSONURL`
2. `scanner.SaveAPIDocsJSON`
3. `scanner.AnalyzeSwaggerAPI`
4. `scanner.AutoRequestAllAPI`
5. 可选 `writers.NewSwaggerDbWriter` / CSV / JSONL

---

## swagger_list_hosts {#swagger-list-hosts}

**用途**：从数据库列出 host 汇总（等同 `report list`）。

**何时用**：查询历史扫描、记录数、风险条数。

**前置**：数据库已有记录 → 通常先 `swagger_scan` + `write_db=true`。

**参数**

| 名称 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `db_uri` | string | 否 | 服务默认 | SQLite/其它 GORM URI |

**返回类型**：`string`（JSON **数组**）

```json
[
  {
    "host": "api.example.com",
    "record_count": 85,
    "last_scanned": "2026-05-19T12:00:00+08:00",
    "input_urls": ["https://api.example.com/index.html"],
    "risk_count": 12,
    "doc_session_id": "api.example.com__swagger__v1"
  }
]
```

---

## swagger_list_sessions {#swagger-list-sessions}

**用途**：列出 `output/` 下所有 `api-docs.json` 会话。

**何时用**：需要 `session` id 打开 Knife4j 文档页、确认磁盘上有哪些 spec。

**前置**：目录内已有 scan 产物；通常先 `swagger_scan`。

**参数**

| 名称 | 类型 | 必填 | 默认 | 说明 |
|------|------|------|------|------|
| `api_doc_path` | string | 否 | `output` | 与 scan 的 `output_dir` 一致 |

**返回类型**：`string`（JSON **数组**）

```json
[
  {
    "id": "api.example.com__swagger__v1",
    "host": "api.example.com",
    "scope": "swagger/v1",
    "label": "api.example.com / swagger/v1",
    "scanned_at": "2026-05-19T12:00:00+08:00",
    "input_url": "https://api.example.com/index.html",
    "json_url": "https://api.example.com/swagger/v1/swagger.json",
    "record_count": 85,
    "spec_url": "/api/docs/api.example.com__swagger__v1/openapi.json"
  }
]
```

---

## 代码位置 {#source}

| 文件 | 内容 |
|------|------|
| `pkg/mcpserver/tools_schema.go` | 注入 MCP 的说明常量 |
| `pkg/mcpserver/tools_register.go` | 注册 3 个 tool |
| `pkg/mcpserver/server.go` | handler 实现与内部 Go 函数映射 |
| `pkg/scanrun/single.go` | `swagger_scan` 实际逻辑 |

修改说明后重新编译即可，客户端重新加载 MCP 后生效。
