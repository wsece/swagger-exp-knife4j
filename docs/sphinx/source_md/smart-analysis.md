# 响应相似度与报告分析

## 摘要 {#abstract}

本文档描述 swagger-exp-knife4j 在扫描持久化与 Web 报告环节采用的**响应体相似度**处理机制。实现位于 `internal/islazy`，由 `pkg/writers`（入库分组）与 `pkg/reportserver`（列表聚类、前端排序）调用。该机制用于：

- 对探测响应进行近似重复检测（非字节级相等）；
- 在数据库中维护响应分组标识；
- 在 Web 界面支持按相似度排序与按接口去重展示。

默认参数下无需用户额外配置；行为随 `--write-db` 与 `report server` 自动启用。

## 术语 {#terms}

| 术语 | 定义 |
|------|------|
| 响应指纹 | 对 HTTP 响应 body 计算的 64 位 SimHash，存储为 `s:<hex>` |
| 汉明距离 | 两个 8 字节指纹之间不同比特位的计数，取值 0–64 |
| 响应组 | 入库时分配的 `response_group_id`，组内指纹与组代表距离 ≤ 阈值 |
| 接口键 | Web 去重使用的 `(Method, API path)`，与 OpenAPI path 字段对齐 |
| 聚类编号 | API 返回的 `similarity_cluster`，用于前端排序，可与 DB 组号合并逻辑 |

## 系统边界 {#scope}

**包含：**

- 成功探测且 body 非空的响应指纹计算；
- SQLite 写入阶段的增量分组（`HashGrouper`）；
- `GET /api/hosts/{host}/records` 返回 `similarity_cluster` 及可选的现场指纹补算；
- 浏览器端「仅最新请求」「按相似响应排序」及 Burp 详情区的 body 长度限制。

**不包含：**

- 请求参数级去重（query/body 差异仍视为同一路径）；
- 语义级 NLP 相似度（仅基于规范化 token 的 SimHash）；
- 将现场补算的指纹或聚类结果写回数据库。

## 功能规格总览 {#spec-overview}

| 编号 | 功能 | 阶段 | 实现要点 |
|------|------|------|----------|
| F1 | 响应指纹 | 入库 | `ResponseSimHash` → `response_sim_hash` |
| F2 | 响应分组 | 入库 | `HashGrouper.Assign`，阈值 `DefaultResponseHammingThreshold = 5` |
| F3 | 列表聚类 | API | `AssignResponseSimilarityClusters` |
| F4 | 相似度排序 | Web | 排序键：簇号 → `similarity_distance` → Method+路径 |
| F5 | 接口去重 | Web | 每接口键保留 `when` 最大记录，`id` 次之 |
| F6 | 大 body 预览 | Web | Raw 预览 64KB；Pretty JSON 解析上限 256KB |

## 算法说明 {#algorithms}

### 4.1 响应体规范化 {#normalize}

在计算指纹前对 body 执行规范化（`NormalizeResponseBody`）：

1. 去除首尾空白；
2. 若内容为合法 JSON：反序列化后重新 `Marshal` 为紧凑 JSON（消除字段顺序与空白差异）；
3. 否则：按空白符折叠为 token 序列（`strings.Fields`）。

**设计依据：** 探测结果中大量响应为 JSON；规范化可降低因序列化差异导致的指纹漂移，使 F2/F3 分组更稳定。

### 4.2 SimHash 指纹 {#simhash}

对规范化后的 token 列表执行 64 维 SimHash（`simHash64`）：

- 每个 token 经 FNV-1a 映射为 64 位值，按位对权重向量累加；
- 最终指纹为 8 字节大端序整数，对外编码为 `s:` + 十六进制字符串。

**与密码学哈希的区别：** MD5/SHA 对单比特差异敏感，不适用于「近似相同」归类；SimHash 配合汉明距离用于**局部相似**判定。

### 4.3 入库分组 {#db-grouping}

`SwaggerDbWriter` 在事务内对每条成功响应：

1. 计算 `simHash := ResponseSimHash(response)`；
2. 调用 `HashGrouper.Assign`：
   - 若与已有组代表汉明距离 ≤ 5：复用该 `response_group_id`，写入 `similarity_distance`；
   - 否则：分配新 `response_group_id`（单调递增），该记录距离记为 0。

组代表指纹在内存中维护；进程启动时可通过 `LoadFromDB` 从已有记录恢复。

**参数：** `internal/islazy/group.go` 中 `DefaultResponseHammingThreshold = 5`。阈值越大，合并越激进，误合并风险上升。

### 4.4 列表聚类（API）{#api-cluster}

`similarityClustersForRecords` 对同一 host 的全量记录：

1. 优先采用已存在的 `response_group_id` 映射为簇；
2. 对 `response_group_id = 0` 的记录，按与 F2 相同阈值对指纹做贪心合并；
3. 无指纹（空 body）记录簇号为 0。

指纹来源：优先 `response_sim_hash`；缺失时对 `response_body` 调用 `ResponseSimHash`（仅用于本次 API 响应，不持久化）。

### 4.5 列表排序与去重 {#web-logic}

**按相似响应排序（F4）：** 客户端对 `visibleRecords()` 按 `similarity_cluster`、`similarity_distance`、Method+路径 字典序排序；与 F3 字段一致。

**仅最新请求（F5）：** 客户端以 `recordDedupeKey = UPPER(Method) + "\n" + api` 分组，保留 `when` 最大者；`when` 相等时保留较大 `id`。计数区显示 `shown / total APIs（仅最新）`。

两功能可组合：先 F5 过滤，再 F4 排序。

### 4.6 详情区 body 限制 {#body-limits}

| 项 | 上限 | 行为 |
|----|------|------|
| Raw body 渲染 | 64 KiB | 超出部分不写入 DOM，展示截断说明与总长度 |
| Pretty JSON | 256 KiB | 超出则回退 Raw 切片 |
| 单条 Header 值 | 2 KiB | 截断显示 |
| 行切换渲染 | — | `requestAnimationFrame` 推迟 DOM 更新，降低主线程阻塞 |

## 数据模型 {#fields}

| 字段 | 类型 | 写入方 | 说明 |
|------|------|--------|------|
| `response_sim_hash` | string | 扫描入库 | `s:<hex>`，8 字节 SimHash |
| `response_group_id` | uint | 扫描入库 | 分组 ID，0 表示未分组 |
| `similarity_distance` | int | 扫描入库 | 与组代表指纹的汉明距离 |
| `similarity_cluster` | int | 报告 API | 列表聚类编号，仅 HTTP JSON |
| `packet_json` / `packet` | text / object | 扫描 / API | HTTP 交换报文，Burp 布局数据源 |

## 典型用例 {#use-cases}

| 用例 | 操作路径 | 预期结果 |
|------|----------|----------|
| 多次扫描同一 host | Web：启用 F5 | 表格每接口键一行，历史仍存库 |
| 分析统一错误页 | Web：F5 + F4 | 同接口最新记录按响应形态相邻排列 |
| 入库后离线分析 | SQL / 导出 | 可按 `response_group_id` 聚合统计 |
| 扩展自定义 Writer | Go：`islazy.ResponseSimHash` | 与内置分组语义一致 |

## 限制与误差 {#limitations}

- **HTML / 含动态字段的文本：** SimHash 对 token 级变化敏感，动态内容可能导致同模板响应被划入不同组。
- **短响应：** token 过少时指纹区分度下降，可能过度合并。
- **失败或无 body 的记录：** 不参与 F2/F3；F4 排序时落入簇 0 或末尾。
- **汉明阈值固定为 5：** 未提供运行时配置；调整需修改源码常量或扩展。
- **F5 不区分 query：** `/api/x?a=1` 与 `/api/x?a=2` 若 path 相同则视为同一接口键（与 OpenAPI path 设计一致）。

## 程序接口（内部库）{#api-go}

| 函数 | 包路径 | 说明 |
|------|--------|------|
| `NormalizeResponseBody` | `internal/islazy` | 规范化 body |
| `ResponseSimHash` | 同上 | 返回 8 字节指纹 |
| `FormatResponseSimHash` / `ParseResponseSimHash` | 同上 | `s:` 编解码 |
| `HammingDistance` | 同上 | 指纹距离 |
| `ResponseBodyHammingDistance` | 同上 | 两段 body 的距离 |
| `AssignResponseSimilarityClusters` | 同上 | 批量簇编号 |
| `BuildSimilaritySortKeys` | 同上 | 排序键构造 |

二次开发见 [模块开发与扩展](module-development.md)。

## 参考文献与实现路径 {#refs}

| 组件 | 源文件 |
|------|--------|
| SimHash | `internal/islazy/simhash.go` |
| 分组器 | `internal/islazy/group.go`, `internal/islazy/hamming.go` |
| 列表聚类 / 排序键 | `internal/islazy/similarity_sort.go` |
| DB 写入 | `pkg/writers/swagger_db.go` |
| API 封装 | `pkg/reportserver/store.go` |
| 前端 | `pkg/reportserver/static/index.html` |

## 相关文档 {#related}

- [Web 报告站](web-report.md) — 界面布局与操作  
- [report 报告](commands-report.md) — `report server`  
- [scan 扫描](commands-scan.md) — `--write-db`  

[返回首页](index.md)
