// tools_schema.go: MCP tool names, descriptions, and parameter JSON Schema constants (for LLM and human docs).
package mcpserver

// The constants below are injected into MCP server.instructions and each tool/argument description
// so the model can decide whether to call, which tool, what to pass, and what to expect before tools/call.

const serverInstructions = `
你是 swagger-exp-knife4j 的 MCP 助手。只能通过下列 3 个 tool 操作，不要臆造其它函数名。

══════════════════════════════════════════════════════════════
全局约定
══════════════════════════════════════════════════════════════
• 成功：tools/call 返回 type=text，body 为 JSON 字符串（UTF-8）。
• 失败：返回 isError=true，body 为人类可读错误信息（非 JSON）。
• 默认数据库：sqlite://swagger-scan.sqlite3
• 默认 api-docs 目录：output/（结构 output/{host}/{scope}/api-docs.json）

══════════════════════════════════════════════════════════════
推荐流程（必读：B 常依赖 A 的副作用）
══════════════════════════════════════════════════════════════
流程 1 — 用户给了一个新 Swagger/Knife4j 地址，要探测接口：
  ① 调用 swagger_scan（必填 url；建议 write_db=true）
  ② 可选 swagger_list_sessions 确认 output 下已生成 api-docs
  ③ 可选 swagger_list_hosts 查看 DB 里 host 统计

流程 2 — 用户问「之前扫过哪些站 / 有多少接口 / 风险数量」：
  ① 必须先有数据库记录 → 若从未 scan 且 write_db，先 swagger_scan
  ② 再 swagger_list_hosts（读 DB，不发起 HTTP 探测）

流程 3 — 用户问「磁盘上有哪些 OpenAPI 文档 / session id」：
  ① 若 output 为空，先对目标 swagger_scan
  ② 再 swagger_list_sessions（只读文件，不读 DB）

流程 4 — 用户要在浏览器里调试接口、看 Knife4j UI：
  → 不要调用 MCP；请用户在本机执行 CLI：report server（MCP 不提供此能力）

禁止：
• 未 scan 就期望 list_hosts 有数据（除非用户已有历史 sqlite 文件）
• 用 list_sessions 代替 scan（list 只列举已有文件，不会抓取新目标）

══════════════════════════════════════════════════════════════
工具索引
══════════════════════════════════════════════════════════════
| tool名 | 何时用 | 前置条件 |
| swagger_scan | 扫描/探测一个新 URL | 无 |
| swagger_list_hosts | 查 DB 中已扫描 host 汇总 | DB 存在且有记录（通常先 scan+write_db） |
| swagger_list_sessions | 查 output 下 api-docs 会话 | output 目录已有 scan 产物 |
`

const toolSwaggerScanDesc = `
【工具名】swagger_scan
【用途】对单个 Swagger JSON 或 Swagger UI / Knife4j 页面执行完整扫描流水线（等同 CLI: scan single -u <url>）。

【何时调用】
• 用户提供了新的 swagger/knife4j/index.html 或 swagger.json 地址
• 需要保存 api-docs、自动请求文档中的 GET/POST 接口、写入 DB/CSV/JSONL
• 在调用 swagger_list_hosts 之前，若数据库尚无该目标数据

【何时不要调用】
• 仅想列出已有扫描结果 → 用 swagger_list_hosts 或 swagger_list_sessions
• 需要 Web UI 调试 → 引导用户使用 report server CLI

【内部步骤（只读说明，非独立 tool）】
1) ResolveSwaggerJSONURL(input) → resolved_json_url
2) SaveAPIDocsJSON → output/{host}/{scope}/api-docs.json
3) AnalyzeSwaggerAPI → 统计 path 数量
4) AutoRequestAllAPI → 对每个 GET/POST 发探测请求
5) 可选写入 DB / CSV / JSONL

【返回类型】成功时 text 内容为 JSON 对象（见下）；失败时 isError + 错误字符串。

【返回 JSON 字段说明】
{
  "input_url": string,           // 你传入的 url
  "resolved_json_url": string,   // 实际拉取的 OpenAPI/Swagger JSON 地址
  "api_docs_path": string,       // 本地保存的 api-docs.json 绝对/相对路径
  "path_count": number,          // 文档中 path 条目数（API 数量）
  "request_count": number,       // 实际发起的 GET/POST 探测次数
  "request_ok": number,          // 探测成功（收到 HTTP 状态，无传输/构建错误）
  "request_failed": number,      // 探测失败（跳过、构建错误或网络错误）
  "unauthorized_count": number,    // 非 401 的响应数（可能未授权访问）
  "wrote_db": boolean,
  "wrote_csv": boolean,
  "wrote_jsonl": boolean,
  "db_uri": string,              // 仅 wrote_db=true 时
  "csv_file": string,
  "jsonl_file": string
}
`

const argScanURL = `
【参数】url
【类型】string
【必填】是
【说明】Swagger 文档入口。可为：
  - 直接 swagger/openapi JSON URL（如 https://host/swagger/v1/swagger.json）
  - Knife4j/Swagger UI 页面（如 https://host/doc.html、/index.html）
服务端会自动解析出真实 json 地址。
`

const argScanOutputDir = `
【参数】output_dir
【类型】string
【必填】否
【默认】output
【说明】api-docs.json 与 manifest 的落盘根目录。之后 swagger_list_sessions 默认也读此目录。
`

const argScanWriteDB = `
【参数】write_db
【类型】boolean
【必填】否
【默认】false
【说明】为 true 时将本次每个 API 探测结果写入 SQLite（见 db_uri）。若后续要用 swagger_list_hosts，应设为 true。
`

const argScanDbURI = `
【参数】db_uri
【类型】string
【必填】否
【默认】未设置 write_db 时忽略；write_db=true 时用服务器默认 sqlite://swagger-scan.sqlite3
【说明】显式指定数据库 URI 时也会写入（即使 write_db=false）。例：sqlite://swagger-scan.sqlite3
`

const argScanWriteCSV = `
【参数】write_csv
【类型】boolean
【必填】否
【默认】false
【说明】为 true 时写入 CSV，默认文件名 result.csv（可用 csv_file 覆盖）。
`

const argScanCSVFile = `
【参数】csv_file
【类型】string
【必填】否
【说明】CSV 路径。write_csv=true 时默认 result.csv；也可仅传 csv_file 不写 write_csv。
`

const argScanWriteJSONL = `
【参数】write_jsonl
【类型】boolean
【必填】否
【默认】false
【说明】为 true 时写入 JSONL，默认 result.jsonl。
`

const argScanJSONLFile = `
【参数】jsonl_file
【类型】string
【必填】否
【说明】JSONL 路径。write_jsonl=true 时默认 result.jsonl。
`

const argScanParallel = `
【参数】parallel
【类型】number（整数）
【必填】否
【默认】1
【说明】自动探测 API 时的并发 worker 数。大于 1 加快扫描，可能增加目标站压力。
`

const toolSwaggerListHostsDesc = `
【工具名】swagger_list_hosts
【用途】从扫描结果数据库列出已扫描过的 host 汇总（等同 CLI: report list）。

【何时调用】
• 用户问「扫过哪些域名」「某 host 有多少条 API 记录」「风险接口数量」
• 已在其它会话执行过 swagger_scan 且 write_db（或 DB 文件已存在）

【前置条件 / 依赖】
• 必须先有 sqlite（或配置的 db_uri）且库内有 swagger_api_records 数据。
• 若库为空 → 先对目标 URL 调用 swagger_scan，并 write_db=true。
• 不读取 output 目录；与 swagger_list_sessions 互补。

【返回类型】成功时 text 为 JSON 数组 HostSummary[]；失败 isError。

【返回 JSON 数组元素】
{
  "host": string,              // 域名
  "record_count": number,      // 该 host 下 API 探测记录条数
  "last_scanned": string,      // RFC3339 时间
  "input_urls": string[],      // 出现过的入口 URL
  "risk_count": number,        // 非 401 的响应条数（粗略未授权风险指标）
  "doc_session_id": string     // 关联的 api-docs session id（若有）
}
`

const argListHostsDbURI = `
【参数】db_uri
【类型】string
【必填】否
【默认】sqlite://swagger-scan.sqlite3（服务器启动参数 --db-uri）
【说明】要查询的扫描数据库。必须与 swagger_scan 写入时使用的库一致。
`

const toolSwaggerListSessionsDesc = `
【工具名】swagger_list_sessions
【用途】遍历本地 output 目录，列出已保存的 api-docs 会话（每个 host+scope 一条）。

【何时调用】
• 用户问「有哪些 api-docs」「session id 是什么」「openapi 文件在哪」
• 准备打开 report server / Knife4j 文档页，需要 session id（格式 host__scope）

【前置条件 / 依赖】
• 目录下需已有 scan 生成的 output/{host}/{scope}/api-docs.json。
• 通常先对目标 swagger_scan；本工具不会下载新文档。
• 不查数据库；要 DB 统计用 swagger_list_hosts。

【返回类型】成功时 text 为 JSON 数组 DocSession[]；失败 isError。

【返回 JSON 数组元素】
{
  "id": string,           // session id，用于 /knife4j-doc.html?session=
  "host": string,
  "scope": string,        // 如 swagger/v1
  "label": string,        // 展示用
  "scanned_at": string,   // RFC3339
  "input_url": string,
  "json_url": string,
  "record_count": number, // 来自 DB 的 host 记录数（DB 不存在时可能为 0）
  "spec_url": string      // 报告服务内 openapi 路径（仅供参考）
}
`

const argListSessionsAPIDocPath = `
【参数】api_doc_path
【类型】string
【必填】否
【默认】output（服务器 --api-doc-path）
【说明】与 swagger_scan 的 output_dir 应对齐，否则列不出刚扫描的会话。
`
