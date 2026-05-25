# 欢迎访问 swagger-exp-knife4j 在线文档

面向 **Swagger / Knife4j / OpenAPI** 的接口发现与探测工具：解析文档、并发探测、结果持久化（SQLite / CSV / JSONL）。入库阶段计算响应 SimHash 并分组；Web 报告支持列表去重、相似度排序及报文详情预览。

```{note} 如何查阅
左侧 **目录** 在首页与各章节页面；使用顶部 **搜索框** 全文检索。
```

## 命令速查

```text
swagger-exp-knife4j
├── version                 # 发行版本与构建元数据
├── scan
│   ├── single -u <URL>     # 扫描单个目标
│   └── file -f <文件>      # 批量扫描
├── report
│   ├── list                # 终端查看数据库记录
│   └── server              # Web 报告（默认 :7171）
└── mcp
    └── serve               # MCP stdio 服务
```

## 典型工作流

```text
scan（写入 DB + output） → report server（浏览 / Knife4j） → 可选 mcp serve（AI 辅助）
```

## 快速开始

```bash
go build -o swagger-exp-knife4j .
swagger-exp-knife4j version
swagger-exp-knife4j scan single -u https://example.com/doc.html --write-db
swagger-exp-knife4j report server --api-doc-path ./output
```

浏览器打开 [http://127.0.0.1:7171/](http://127.0.0.1:7171/) 查看报告与 Knife4j。更多章节见**左侧主目录**。