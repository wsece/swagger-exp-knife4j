# Web 报告站

## 概述 {#overview}

`report server` 提供基于 HTTP 的本地报告界面（默认 `http://127.0.0.1:7171/`），用于查询 SQLite 中的探测记录、浏览 `output/` 下的 OpenAPI 会话，并通过嵌入页使用 Knife4j 调试。响应相似度、列表去重及大 body 预览规则见 [响应智能分析](smart-analysis.md)。

## 界面结构 {#layout}

| 区域 | 功能 |
|------|------|
| 左侧 Hosts | 按 host 汇总 DB 记录；选中后加载该 host 的 API 列表 |
| 左侧 API Docs | 枚举 `output/` 下 `api-docs.json` 会话 |
| 中间表格 | 探测记录列表；工具栏提供 F5/F4 类操作（见 §3） |
| 右侧详情 | 单条记录的 Request/Response（Burp 式分栏） |
| 顶部（可折叠） | 提交扫描任务（单 URL / 多行 / 文件） |

## 中间栏：记录列表 {#table}

### 3.1 列与表头排序

默认列：时间、Method、API 路径、状态码、未授权风险标记（非 401 显示 `!`）。点击表头可在时间、方法、路径、状态、风险字段间切换升/降序（与工具栏排序正交）。

### 3.2 工具栏

在已加载 Host 记录时显示：

| 控件 | 规格 ID | 行为 |
|------|---------|------|
| 仅最新请求 | F5 | 按 `(Method, API path)` 去重，保留 `when` 最新记录；计数格式 `n / m APIs（仅最新）` |
| 按相似响应排序 | F4 | 按 `similarity_cluster`、`similarity_distance`、Method+路径 排序 |

两控件可叠加。去重仅影响前端展示，不修改数据库。相似度字段由 `GET /api/hosts/{host}/records` 提供；详见 [响应相似度分析](smart-analysis.md) §4.5。

### 3.3 数据来源

```text
GET /api/hosts/{host}/records
```

响应为 `RecordJSON` 数组，嵌入 `SwaggerAPIRecord`、`packet`（`HTTPExchange`）、`unauthorized_risk`、`similarity_cluster` 等字段。

## 右侧：报文详情 {#detail}

选中表格行后展示：

| 项 | 说明 |
|----|------|
| 布局 | 上 Request、下 Response；各支持 Raw / Pretty |
| 数据 | 优先 `packet_json`；缺失时由 legacy 字段组装 |
| Body 限制 | Raw 预览 ≤ 64 KiB；Pretty JSON 解析 ≤ 256 KiB；超限显示截断与总长度（§4.6） |
| 渲染 | 行切换经 `requestAnimationFrame` 异步更新 DOM |

## 扫描任务下发 {#scan-tasks}

| 方式 | 对应后端 |
|------|----------|
| 单 URL | `POST /api/scan/tasks`（JSON） |
| 多行 URL | 同上 |
| 文件上传 | `multipart/form-data` |

任务异步执行，写入当前 server 配置的 `--db-uri` 与 `--api-doc-path`。完成后重新选择 Host 即可加载新记录。

## Knife4j 集成 {#knife4j}

**入口：** API Docs 列表「Open」，或 `/knife4j-doc.html?session=<session_id>`（`session_id` 通常为 `host__scope`）。

**代理路径：** 调试请求发往同源 `http://127.0.0.1:7171/knife4j/<session>/try/...`，由服务端转发至 manifest 中的真实 API 地址。

**目的：** 避免浏览器对跨域 `POST application/json` 触发 CORS 预检（OPTIONS）；多数目标 API 不处理 OPTIONS 将导致 405，而 CLI 探测无此约束。

## HTTP 接口索引 {#urls}

| 路径 | 方法 | 说明 |
|------|------|------|
| `/` | GET | 报告 SPA |
| `/api/hosts` | GET | Host 汇总 |
| `/api/hosts/{host}/records` | GET | 记录列表 |
| `/api/records/{id}` | GET | 单条记录 |
| `/api/sessions` | GET | api-docs 会话 |
| `/api/docs/{sessionID}/preview` | GET | OpenAPI 预览元数据 |
| `/api/scan/tasks` | POST | 创建扫描任务 |
| `/knife4j-doc.html` | GET | Knife4j 壳页 |

## 推荐操作流程 {#workflow}

```text
1. scan single|file --write-db
2. report server --db-uri <同扫描> --api-doc-path <同 output>
3. 浏览器打开 / ，选择 Host
4. （可选）启用「仅最新请求」
5. （可选）启用「按相似响应排序」
6. 选择行查看 Request/Response
7. （可选）API Docs → Knife4j 经同源代理调试
```

## 异常与空状态 {#empty}

| 现象 | 处理 |
|------|------|
| 无 Host | 执行带 `--write-db` 的扫描，或 Web 下发任务 |
| 无 API Docs | 确认 `--api-doc-path` 与扫描 `output` 一致 |
| 相似排序无效 | 确认存在响应体或 `response_sim_hash`；失败/空 body 不参与聚类 |
| Knife4j OPTIONS 405 | 确认请求路径为 `/knife4j/.../try/...`；重建并强刷缓存 |

[响应智能分析](smart-analysis.md) · [report 命令](commands-report.md) · [首页](index.md)
