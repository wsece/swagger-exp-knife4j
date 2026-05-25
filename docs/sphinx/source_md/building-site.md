# 构建文档静态站

本站使用 [Sphinx](https://www.sphinx-doc.org/) + [MyST Parser](https://myst-parser.readthedocs.io/) + [Read the Docs 主题](https://sphinx-rtd-theme.readthedocs.io/) 构建。

Sphinx 工程目录：**`docs/sphinx/`**（配置、Makefile、构建脚本）；源稿 Markdown 在 **`docs/sphinx/source_md/`**。

## 环境 {#env}

- Python 3.9+
- pip

```bash
pip install -r requirements.txt
# 等价于 pip install -r docs/sphinx/requirements.txt
```

## 本地预览（开发文档时推荐） {#serve}

```bash
cd docs/sphinx
make html
make serve
```

Windows：

```bat
cd docs\sphinx
make.bat html
make.bat serve
```

浏览器访问 **http://127.0.0.1:8000/** 。修改 `docs/sphinx/source_md/*.md` 后重新执行 `make html`（会先运行 `sync_md.py` 更新 `index.rst`）。

```{note} 查看页面源码
构建结束后会运行 `postprocess_sourcelinks.py`，将 `_sources/*.txt` 包装为带 `charset=utf-8` 的 HTML，避免中文乱码。请使用 `make serve` / `make.bat serve`（`serve_docs.py`）预览，不要直接双击打开 `_build/html` 下的文件。
```

## 构建静态文件 {#build}

```bash
cd docs/sphinx
make html
# 或：python sync_md.py && sphinx-build -b html . _build/html && python postprocess_sourcelinks.py
```

HTML 产物在 **`docs/sphinx/_build/html/`**，可部署到 Nginx、GitHub Pages、对象存储等。

## GitHub Pages {#github-pages}

仓库 `.github/workflows/docs.yml` 使用 Sphinx 构建并发布：

1. GitHub 仓库 **Settings → Pages** → Source 选 **GitHub Actions**
2. 推送 `main`/`master` 后自动部署
3. 自定义域名在 Pages 设置中配置 CNAME；必要时在 `docs/sphinx/conf.py` 调整 `html_baseurl`

## 修改目录 {#nav-edit}

1. 编辑 **`docs/sphinx/sync_md.py`** 中的 `DOC_ORDER` 与 `write_index_rst()` 标题映射
2. 执行 `python sync_md.py` 或 `make html` 重新生成 `index.rst`
3. 样式微调：**`docs/sphinx/_static/custom.css`**



[Docker 部署](docker.md) · [首页](index.md)
