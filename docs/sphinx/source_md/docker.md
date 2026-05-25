# Docker 部署

使用 Docker / Docker Compose 运行 **Web 报告站**（`report server`），数据持久化在卷 `swagger-data`（挂载到容器内 `/data`）。

## 环境要求 {#docker-env}

| 项 | 要求 |
|----|------|
| Docker | 20.10+ |
| Docker Compose | v2（`docker compose`） |

## 一键启动报告站 {#docker-compose}

在项目根目录：

```bash
docker compose up -d --build
```

浏览器访问：**http://127.0.0.1:7171/**

| 路径（容器内） | 说明 |
|----------------|------|
| `/data/swagger-scan.sqlite3` | 默认 SQLite |
| `/data/output/` | api-docs 转储目录 |

修改宿主机映射端口（可选）：

```bash
REPORT_PORT=8080 docker compose up -d
```

停止并保留数据：

```bash
docker compose down
```

删除数据卷（清空库与 output）：

```bash
docker compose down -v
```

## 仅构建镜像 {#docker-build}

```bash
docker build -t swagger-exp-knife4j:latest .
```

## 在容器内执行扫描 {#docker-scan}

扫描需访问**外网目标**，与报告站共用同一数据卷：

```bash
# 单 URL 扫描并写入库
docker compose run --rm report scan single \
  -u https://example.com/doc.html \
  --write-db \
  --db-uri sqlite://swagger-scan.sqlite3 \
  --output-dir ./output

# 批量（先将 targets.txt 放到宿主机并挂载，或复制进容器）
docker compose run --rm report scan file \
  -f /data/targets.txt \
  --write-db \
  --db-uri sqlite://swagger-scan.sqlite3 \
  --output-dir ./output
```

批量文件示例：在宿主机创建 `./data/targets.txt` 后：

```bash
mkdir -p data
echo "https://example.com/v3/api-docs" > data/targets.txt
docker compose run --rm \
  -v "${PWD}/data:/data" \
  report scan file -f /data/targets.txt --write-db \
  --db-uri sqlite://swagger-scan.sqlite3 --output-dir ./output
```

扫描完成后刷新 **http://127.0.0.1:7171/** 即可查看结果（`report` 服务需已 `up`）。

## 仅用 docker run {#docker-run}

```bash
docker build -t swagger-exp-knife4j:latest .

docker volume create swagger-exp-data

# 报告站（后台）
docker run -d --name swagger-report \
  -p 7171:7171 \
  -v swagger-exp-data:/data \
  swagger-exp-knife4j:latest

# 一次性扫描
docker run --rm \
  -v swagger-exp-data:/data \
  swagger-exp-knife4j:latest \
  scan single -u https://example.com/doc.html \
  --write-db --db-uri sqlite://swagger-scan.sqlite3 --output-dir ./output
```

## 自定义命令 {#docker-custom}

镜像 `ENTRYPOINT` 为 `swagger-exp-knife4j`，可覆盖 `CMD` 运行任意子命令：

```bash
docker compose run --rm report report list \
  --db-uri sqlite://swagger-scan.sqlite3

docker compose run --rm report --help
```

> **MCP**（`mcp serve`）依赖 stdio，不适合作为长期运行的 Docker 服务；请在宿主机或 IDE 中直接启动二进制。

## 镜像说明 {#docker-image}

| 项 | 值 |
|----|-----|
| 基础镜像 | `alpine:3.20`（运行期） |
| 工作目录 | `/data` |
| 监听 | `0.0.0.0:7171`（容器内须绑定全接口） |
| 运行用户 | `app`（uid 1000） |

## 常见问题 {#docker-faq}

**Q：页面无数据？**  
A：先执行带 `--write-db` 的 `scan`，且 `--db-uri`、`--output-dir` 与 report 一致（Compose 默认均为 `/data` 下路径）。

**Q：Knife4j 调试请求失败？**  
A：浏览器访问的是真实目标 API，与是否 Docker 无关；确认目标可达、授权与网络策略。

**Q：健康检查失败？**  
A：确认 `7171` 未被占用；`docker compose logs report` 查看启动日志。

[安装与编译](installation.md) · [Web 报告站](web-report.md) · [首页](index.md)
