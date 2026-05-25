# mcp 服务

通过 **Model Context Protocol (stdio)** 把扫描能力暴露给 Cursor、Claude Desktop 等客户端。

> 浏览器 Knife4j / Web 报告请用 `report server`，**不在 MCP 内提供**。

## 启动 {#start}

```bash
swagger-exp-knife4j mcp serve \
  --db-uri sqlite://swagger-scan.sqlite3 \
  --api-doc-path ./output
```

| 参数 | 默认 | 说明 |
|------|------|------|
| `--db-uri` | `sqlite://swagger-scan.sqlite3` | `swagger_scan` 等工具的默认库 |
| `--api-doc-path` | `output` | 默认 api-docs 目录 |

## Cursor 配置示例 {#cursor}

项目根目录参考 `mcp.cursor.json.example`：

```json
{
  "mcpServers": {
    "swagger-exp-knife4j": {
      "command": "swagger-exp-knife4j",
      "args": ["mcp", "serve"],
      "cwd": "/path/to/swagger-exp-knife4j"
    }
  }
}
```

修改后于 Cursor：**设置 → MCP** 中重载服务。

## 工具一览 {#tools}

| 工具 | 作用 | 典型前置 |
|------|------|----------|
| `swagger_scan` | 扫描单个 URL | 无 |
| `swagger_list_hosts` | 数据库 host 汇总 | 曾 scan 且 write_db |
| `swagger_list_sessions` | 列出 output 下 api-docs | 曾 scan |

依赖关系：

```text
swagger_scan → output/  → swagger_list_sessions
            → DB       → swagger_list_hosts
```

## 详细契约 {#contract}

参数类型、返回 JSON 字段、调用顺序见 **[mcp-tools.md](mcp-tools.md)**（面向大模型与集成开发）。

## 全局选项 {#global-options}

与 CLI 共用根命令参数（在启动 MCP 子进程时一般不需交互）：

| 参数 | 说明 |
|------|------|
| `-q` / `--quiet` | 静默 Info；`mcp serve` 启动时已等价静默 |
| `-D` / `--debug-log` | 输出 Debug（含扫描流水线细节；MCP 子进程一般不用） |

`swagger_scan` 返回 JSON 含 `request_ok`、`request_failed`、`unauthorized_count`，与 CLI **`Scan Finished`** 摘要语义一致（字段名不同），见 [mcp-tools.md](mcp-tools.md)。

[首页](index.md) · [MCP 详细契约](mcp-tools.md)
