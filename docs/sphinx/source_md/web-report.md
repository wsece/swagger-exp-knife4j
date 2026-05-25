# Web 报告站

## 概述 {#overview}

`report server` 提供基于 HTTP 的本地报告界面（默认 `http://127.0.0.1:7171/`），用于查询 SQLite 中的探测记录、浏览 `output/` 下的 OpenAPI 会话，并通过嵌入页使用 Knife4j 调试。响应相似度、列表去重及大 body 预览规则见 [响应智能分析](smart-analysis.md)。

## 界面结构 {#layout}

| 区域 | 功能 |
|------|------|
| 左侧 Hosts | 按 host 汇总 DB 记录；选中后加载该 host 的 API 列表 |
| 左侧 API Docs | 枚举 `output/` 下 `api-docs.json` 会话；提供 **Open Knife4j** / **Proxy mode** 入口 |
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

(knife4j)=
## Knife4j 集成

### 入口

| 控件 | 位置 | 行为 |
|------|------|------|
| **Open Knife4j** | API Docs 列表 / OpenAPI 预览区 | 新标签打开 Knife4j，**默认直连目标 API 域名**（便于 Burp 等抓包改包） |
| **Proxy mode** | 同上（第二按钮） | 新标签打开 Knife4j，调试请求经报告站同源代理转发 |
| **Open** / **Proxy** | 左侧 API Docs 每条会话 | 与上表等价，分别对应直连 / 代理 |

等价 URL：

- 直连：`/knife4j-doc.html?session=<session_id>`
- 代理：`/knife4j-doc.html?session=<session_id>&proxy=1`

`session_id` 通常为 `host__scope`（与 `output/` 目录对应）。

### 两种模式对比

| 项 | Open Knife4j（直连，默认） | Proxy mode（同源代理） |
|----|---------------------------|------------------------|
| F12 / Network 请求 URL | 目标站，如 `https://api.example.com/...` | 本机，如 `http://127.0.0.1:7171/knife4j/<session>/try/...` |
| 适用场景 | Burp、Charles 等系统代理抓包；需看到真实 Host | 目标未配 CORS，浏览器 POST/JSON 触发 OPTIONS 405 |
| OpenAPI `servers` | manifest 解析出的真实 origin | 相对路径 `/knife4j/<session>/try` |
| CLI 全局默认 | 是（`--knife4j-proxy` 未指定） | `report server --knife4j-proxy` 或点 **Proxy mode** |

**直连原理：** OpenAPI 规范中 `servers` 写入真实 API 地址；Knife4j 页内脚本不再把 try-it-out 请求改写为 `/knife4j/.../try`。

**代理原理：** 浏览器仅访问 `127.0.0.1:7171`，由 `report server` 根据 manifest 的 `json_url` / `input_url` 转发至上游；可吞掉浏览器 OPTIONS 预检，避免部分 API 返回 405。

### 选型建议

```text
需要抓包改包、Network 显示目标域名  → Open Knife4j
浏览器报 CORS / OPTIONS 405       → Proxy mode 或 report server --knife4j-proxy
```

修改 Knife4j 行为后请 **Ctrl+F5 强刷** 页面，避免旧脚本缓存。

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
7. （可选）API Docs → **Open Knife4j**（直连）或 **Proxy mode**（CORS 回退）
```

## 异常与空状态 {#empty}

| 现象 | 处理 |
|------|------|
| 无 Host | 执行带 `--write-db` 的扫描，或 Web 下发任务 |
| 无 API Docs | 确认 `--api-doc-path` 与扫描 `output` 一致 |
| 相似排序无效 | 确认存在响应体或 `response_sim_hash`；失败/空 body 不参与聚类 |
| `report server` 端口占用 | 关闭旧进程或 `report server --port 7172`；详见 [report 命令](commands-report.md) |
| Knife4j CORS / OPTIONS 405（直连） | 改用 **Proxy mode** 或启动时加 `--knife4j-proxy` |
| Knife4j 404（代理模式） | 确认 Network 为 `/knife4j/<session>/try/...`；检查 manifest 的 `json_url` |

[响应智能分析](smart-analysis.md) · [report 命令](commands-report.md) · [首页](index.md)
