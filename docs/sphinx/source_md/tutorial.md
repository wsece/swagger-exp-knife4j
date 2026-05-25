# 快速入门



## 第一步：扫描单个 URL {#scan-url}

1、可直接扫描单个 url；结束后终端会输出 **`Scan Finished`** 摘要块（`-q` 安静模式仍会显示）：

```text
2026-05-20 12:23:58  Scan Finished
	Input: https://your-target.com/doc.html
	Result: API=12 | Request[ok]=12 | Request[skip]=0 | Request[fail]=0 | Unauthorized=3
	DumpJson: output/your-target.com/swagger/v1/api-docs.json
```

| 字段 | 含义 |
|------|------|
| `Input` | 输入 URL |
| `API` | OpenAPI 待测接口数 |
| `Request[ok]` / `Request[skip]` / `Request[fail]` | 成功 / 跳过 / 失败 |
| `Unauthorized` | 非 401 响应数（仅 >0 时显示） |
| `DumpJson` | 规范 JSON 落盘路径（失败为 `none`） |

完整说明见 [scan 扫描 · 终端输出](commands-scan.md)。

```bash
swagger-exp-knife4j scan single -u https://your-target.com/doc.html
# 安静：只看最终摘要
swagger-exp-knife4j scan single -u https://your-target.com/doc.html -q
# 排障：每个 API、未授权候选详情
swagger-exp-knife4j scan single -u https://your-target.com/doc.html -D
```

- `-u`：Swagger JSON 或 Knife4j / Swagger UI 页面地址  

2、可将结果存储到数据库。

```bash
swagger-exp-knife4j scan single -u https://your-target.com/doc.html --write-db
swagger-exp-knife4j scan single -u https://your-target.com/doc.html --write-csv
swagger-exp-knife4j scan single -u https://your-target.com/doc.html --write-jsonl
```

- `--write-db`：结果写入数据库 ，才能使用 web 端查看（默认 `sqlite://swagger-scan.sqlite3` ）
- `--write-csv`：结果写入 csv 文件（默认名称 result.csv）
- `--write-jsonl`：结果写入 jsonl 文件（默认名称 result.jsonl）



## 第二步：结果查看 {#result-view}

1、可在控制台查看简要扫描信息：

```
swagger-exp-knife4j report list
```

2、可在 Web 管理端查看统计信息或下发新任务：

启动服务，默认连接 `sqlite://swagger-scan.sqlite3` 及 `./output` 下的转储文件；可用 `--db-uri`、`--api-doc-path` 覆盖。

```bash
swagger-exp-knife4j report server
swagger-exp-knife4j report server --db-uri sqlite://swagger-scan.sqlite3 --api-doc-path ./output
```

浏览器访问：**http://127.0.0.1:7171/**

- 左侧选 **Host** 加载探测记录表  
- 中栏可选：**仅最新请求**（按 Method+路径去重）、**按相似响应排序**（基于 SimHash 聚类字段）  
- 选中行后右侧查看 Request/Response（body 超过 64 KiB 时仅预览片段）  
- **API Docs**：**Open Knife4j** 直连目标域名（便于抓包）；**Proxy mode** 经本机代理（CORS 异常时用）  
- 顶部 **下发扫描任务**（默认折叠）提交后续扫描  

规格说明：[响应智能分析](smart-analysis.md)；界面说明：[Web 报告站](web-report.md)。



## 批量扫描 {#batch}

`targets.txt` 每行一个 URL，`#` 为注释：

```text
https://a.example.com/v3/api-docs
https://b.example.com/doc.html
```

```bash
swagger-exp-knife4j scan file -f targets.txt --write-db
```

每个 URL 结束会单独输出 **`Scan Finished`** 块（`Input` 为该行 URL）；全部结束后输出 **`Batch Scan Finished`** 汇总。Info 日志用 `──────── scan target ────────` 与 `progress=[n/N]` 区分当前目标。

## 带 Cookie / 代理 {#cookie-proxy}

```bash
swagger-exp-knife4j scan single -u https://api.example.com/v3/api-docs \
  -H "Authorization: Bearer <token>" \
  -b "session=xxx" \
  -x http://127.0.0.1:8080 \
  -P 4 \
  --write-db
```

## 下一步阅读 {#next}

| 需求 | 文档 |
|------|------|
| 全部 scan 参数 | [scan 扫描](commands-scan.md) |
| Web 报告与相似度 | [Web 报告站](web-report.md) · [响应智能分析](smart-analysis.md) |
| Cursor MCP | [mcp 服务](commands-mcp.md) |

[上一章：安装](installation.md) · [首页](index.md)
