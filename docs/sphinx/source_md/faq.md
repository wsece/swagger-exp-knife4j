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

**Q：`report server` 提示端口 7171 已被占用？**  
A：通常为本机仍在运行的旧 `report server` 或其它进程。换端口启动，例如 `report server --port 7172`；或结束占用进程后重试。Windows 可查：`netstat -ano | findstr :7171`。

**Q：Knife4j 里先出现 OPTIONS，返回 405，但命令行 POST 正常？**  
A：浏览器 **CORS 预检**：Knife4j 页在 `http://127.0.0.1:7171`，**Open Knife4j（直连）** 会向 `https://目标站` 发跨域 POST/JSON，可能先 OPTIONS 而目标 API 不支持。改用 **Proxy mode** 或 `report server --knife4j-proxy`，Network 中应看到 `http://127.0.0.1:7171/knife4j/.../try/...`。详见 [Web 报告站 · Knife4j](web-report.md#knife4j)。

**Q：Knife4j 调试时 F12 里不是目标域名，抓包不方便？**  
A：默认应使用 **Open Knife4j**（直连），Network 显示真实 API 域名。若仍为 `127.0.0.1:7171/knife4j/.../try/...`，说明当前为代理模式——关闭该标签，改点 **Open Knife4j** 而非 **Proxy mode**；并 Ctrl+F5 强刷。

**Q：Knife4j 里请求 404，打到 127.0.0.1？**  
A：多出现在 **Proxy mode**。确认 Network 路径为 `/knife4j/<session>/try/...`；检查 `output` 下 manifest 的 `json_url` 是否正确。若已用直连仍 404，核对 OpenAPI 中 path 与目标环境是否一致。

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
