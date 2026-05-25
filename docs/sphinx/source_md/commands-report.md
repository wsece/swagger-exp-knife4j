# report 报告

## 命令结构 {#structure}

```text
swagger-exp-knife4j report [子命令] [--db-uri ...]
```

| 子命令 | 作用 |
|--------|------|
| `list` | 终端表格列出数据库中的探测记录 |
| `server` | 启动 Web 报告与 Knife4j（默认 `127.0.0.1:7171`） |

---

## report 共享参数 {#report-shared}

| 参数 | 默认 | 说明 |
|------|------|------|
| `--db-uri` | `sqlite://swagger-scan.sqlite3` | 扫描结果数据库，须与 scan 写入时一致 |

---

## report list {#report-list}

### 功能

按时间倒序输出：时间、入口 URL、方法、Host、API 路径、状态码等。

### 示例

```bash
swagger-exp-knife4j report list
swagger-exp-knife4j report list --db-uri sqlite://./data/scan.sqlite3
```

### 前置条件

需已执行过 `scan ... --write-db`（或库中已有历史数据）。

### 输出

- **stdout**：表格主体（便于 `report list > records.txt`）
- **stderr**：记录条数等 Info（`-q` 隐藏）

---

## report server {#report-server}

### 功能

提供本地 HTTP 服务：查询探测记录、报文详情（Request/Response）、列表去重与相似度排序、OpenAPI 预览、Knife4j 在线调试（直连目标或同源代理）、Web 端扫描任务提交。

### 示例

```bash
# 默认 127.0.0.1:7171
swagger-exp-knife4j report server

# 与 scan 对齐的数据路径
swagger-exp-knife4j report server \
  --db-uri sqlite://swagger-scan.sqlite3 \
  --api-doc-path ./output

# 端口被占用时换端口
swagger-exp-knife4j report server --port 7172

# 全局默认 Knife4j 走同源代理（CORS 回退，见 web-report §Knife4j）
swagger-exp-knife4j report server --knife4j-proxy
```

### 参数

| 参数 | 默认 | 说明 |
|------|------|------|
| `--host` | `127.0.0.1` | 监听地址 |
| `--port` | `7171` | 端口；若已被占用，程序会提示改用 `--port <其他端口>` |
| `--db-uri` | `sqlite://swagger-scan.sqlite3` | 数据库 |
| `--api-doc-path` | `./output` | api-docs 根目录 |
| `--knife4j-proxy` | false | 为 true 时，Knife4j「调试」默认经同源 `/knife4j/<session>/try/...` 转发（规避浏览器 CORS/OPTIONS）；未指定时 Web 页 **Open Knife4j** 为直连目标域名 |

### 启动日志

绑定成功后在 stderr 打印（`Print`，`-q` 仍可见）：

- `report server data loaded`：`hosts`、`sessions`、`db`、`output`
- `report server listening`：访问 URL（如 `http://127.0.0.1:7171/`）

端口冲突时**不会**打印 `listening`，并给出 `--port` 换端口提示（见 [常见问题](faq.md) §Web 报告）。

### 相关文档

界面说明见 [Web 报告站](web-report.md)；相似度与去重规格见 [响应智能分析](smart-analysis.md)。

[scan 扫描](commands-scan.md) · [响应智能分析](smart-analysis.md) · [Web 报告站](web-report.md) · [首页](index.md)
