# version 版本信息

## 摘要 {#abstract}

`version` 子命令用于查询当前可执行文件的**发行版本**与**构建溯源**元数据。字段定义位于 `internal/version/version.go`；未通过构建脚本注入时，Git 与构建相关项默认为 `dev`。

## 命令形式 {#usage}

```bash
swagger-exp-knife4j version
```

## 输出字段 {#fields}

固定五行，写入 **stdout**（便于脚本解析）：

```text
swagger-exp-knife4j
Version:    2.1.2
Git commit: dev
Built:      dev
Build env:  dev
```

| 输出行 | 源变量 | 含义 |
|--------|--------|------|
| `Version:` | `Version` | 发行版本号 |
| `Git commit:` | `GitHash` | 构建时 Git 提交（短哈希或 `dev`） |
| `Built:` | `GoBuildTime` | 构建时间戳 |
| `Build env:` | `GoBuildEnv` | 构建环境标识（如 `go1.26 windows/amd64`） |

## 与发布构建的关系 {#build}

发布流水线可通过 `go build -ldflags` 覆盖上述变量，例如：

```bash
go build -ldflags "\
  -X swagger-exp-knife4j/internal/version.GitHash=$(git rev-parse --short HEAD) \
  -X swagger-exp-knife4j/internal/version.GoBuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -X swagger-exp-knife4j/internal/version.GoBuildEnv=$(go version)" \
  -o swagger-exp-knife4j .
```

未注入时，`version` 仍可运行，仅溯源字段显示为 `dev`。

[安装与编译](installation.md) · [首页](index.md)
