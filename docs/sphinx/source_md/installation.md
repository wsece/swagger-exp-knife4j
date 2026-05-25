# 安装与编译

## 环境要求 {#env}

| 项 | 要求 |
|----|------|
| Go | 1.21+（推荐与 `go.mod` 一致的最新稳定版） |
| 操作系统 | Windows / Linux / macOS |
| 网络 | 扫描阶段需访问目标 URL |

## 获取源码 {#clone}

```bash
git clone https://github.com/wsece/swagger-exp-knife4j.git
```

## 编译 {#build}

```bash
go build -o swagger-exp-knife4j
```

Windows 下得到 `swagger-exp-knife4j.exe`，下文统一写作 `swagger-exp-knife4j`。

## 验证安装 {#verify}

```bash
swagger-exp-knife4j version
swagger-exp-knife4j --help
swagger-exp-knife4j scan single --help
```

`version` 输出发行号与构建溯源字段，规格见 [version 命令](commands-version.md)。

## 运行测试（可选） {#test}

```bash
go test ./...
```

## Docker 部署 {#docker}

使用 Docker Compose 启动报告站、在容器内执行扫描，见 **[Docker 部署](docker.md)**。

```bash
docker compose up -d --build
# http://127.0.0.1:7171/
```

## 数据文件位置 {#paths}

| 路径 | 说明 |
|------|------|
| `swagger-scan.sqlite3` | 默认 SQLite（`--write-db`） |
| `output/` | 默认转储的 api-docs 目录（`--output-dir`） |
| `result.csv` / `result.jsonl` | 默认 CSV / JSONL（启用对应 write 标志时） |

首次使用可无上述文件；扫描后会自动创建。

## 常见问题 {#install-faq}

- **无法覆盖 exe**：Windows 下请先关闭正在运行的 `report server` 或 MCP 子进程再编译。  
- **SQLite 路径**：URI 形如 `sqlite://swagger-scan.sqlite3` 或 `sqlite:///绝对路径/xxx.sqlite3`。

[上一章：项目概览](introduction.md) · [下一章：快速入门](tutorial.md) · [首页](index.md)
