# 常见问题

## 安装与运行 {#install}

**Q：编译提示无法写入 exe？**  
A：关闭正在运行的 `report server` 或占用该文件的 MCP 进程后重新 `go build`。

**Q：`report server` 启动失败，数据库打不开？**  
A：检查 `--db-uri` 是否正确；SQLite 首次访问会自动建库。Windows 绝对路径建议 `sqlite:///D:/path/to/db.sqlite3`。

---

## 扫描 {#scan}

**Q：解析不到 swagger.json？**  
A：确认 URL 可访问；Knife4j/HTML 页需能拉取到页面内容。可换直连 JSON 地址试 `-u https://host/v3/api-docs`。

**Q：大量 401 / 超时？**  
A：配置 `-H` / `-b` 认证信息，调整 `-m` 超时与 `-P` 并发，必要时加 `--delay`。

**Q：不写库只有统计？**  
A：需加 `--write-db` 或 `--write-db-uri` 才会写入 SQLite；否则终端仍会输出 **`Scan Finished`** 摘要块（`Result` 含 `API`、`Request[ok]` 等，`DumpJson` 为落盘路径）。

**Q：`-q` 安静模式还看什么？**  
A：每个目标仍输出 **`Scan Finished`** 块（`scan file` 有几个 URL 就有几块），以及最后的 **`Batch Scan Finished`**。解析进度（Info）、`scan target` 分隔线、未授权逐条列表均隐藏；未授权**数量**在 `Result` 的 `Unauthorized=` 中。

**Q：`Request[ok]` 和 `Unauthorized` 怎么算？**  
A：`Request[ok]`：请求已发出且收到 HTTP 状态（无 Error）。`Request[fail]`：跳过非 GET/POST、构建或网络错误。`Unauthorized`：状态码不为 0 且 **不是 401** 的响应条数。MCP `swagger_scan` 返回 JSON 字段名不同（如 `request_ok`），语义一致，见 [mcp-tools.md](mcp-tools.md)。

---

## Web 报告 {#web-report}

**Q：打开 http://127.0.0.1:7171 报错 `Cannot read properties of null`？**  
A：升级至最新构建（已修复空 sessions 返回 null）；强刷浏览器缓存。

**Q：左侧没有 Host / API Docs？**  
A：先完成至少一次带 `--write-db` 的扫描，且 `report server` 的 `--api-doc-path`、`--db-uri` 与扫描时一致。

**Q：Knife4j 里先出现 OPTIONS，返回 405，但命令行 POST 正常？**  
A：这是浏览器 **CORS 预检**：页面在 `http://127.0.0.1:7171`，若请求直接发到 `https://目标站`，跨域 + JSON 会先发 OPTIONS；目标 API 常不支持 OPTIONS。本工具已把调试请求改走同源代理 `/knife4j/<session>/try/...`。**请重新编译并强刷** Knife4j 页（Ctrl+F5）；Network 里应看到 `POST http://127.0.0.1:7171/knife4j/.../try/api/...`，而不是直连 `https://目标站`。

**Q：Knife4j 里请求 404，打到 127.0.0.1？**  
A：确认 Network 中路径为 `/knife4j/<session>/try/...`；检查 `output` 下 manifest 的 `json_url` 是否正确。

**Q：同一接口在中栏出现多行？**  
A：数据库按扫描批次累积历史记录，未做自动去重。启用中栏 **仅最新请求** 后，按 `Method + API path` 仅展示 `when` 最新的一条；计数为 `n / m APIs（仅最新）`。见 [响应相似度分析](smart-analysis.md) §4.5（列表排序与去重）。

**Q：「按相似响应排序」无明显分组？**  
A：依赖 `response_sim_hash` 或可用 `response_body`（`--write-db` 且探测成功）。无 body 的失败记录不参与聚类。汉明阈值与限制见 [响应相似度分析](smart-analysis.md)「限制与误差」一节。

**Q：打开详情后页面长时间无响应？**  
A：旧版本可能将完整大 body 写入 DOM。当前实现 Raw 预览上限 64 KiB（§4.6）。请使用最新构建并强刷缓存（Ctrl+F5）。

---

## MCP {#mcp}

**Q：Cursor 里看不到工具？**  
A：确认 `mcp serve` 路径正确、已重载 MCP；查看 Cursor MCP 日志是否有启动错误。

**Q：list_hosts 为空？**  
A：先调用 `swagger_scan` 且 `write_db=true`，或指定与扫描相同的 `db_uri`。

---

## 合法与合规 {#legal}

**Q：能否扫描任意公网地址？**  
A：仅在你**已获得书面授权**的范围内使用；未经授权的扫描可能违法。

[返回首页](index.md)
