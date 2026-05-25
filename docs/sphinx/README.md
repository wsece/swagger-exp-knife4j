# Sphinx 文档工程

项目说明书使用 [Sphinx](https://www.sphinx-doc.org/) + [MyST](https://myst-parser.readthedocs.io/) + [Read the Docs 主题](https://sphinx-rtd-theme.readthedocs.io/) 构建。

## 目录说明

| 路径 | 说明 |
|------|------|
| `conf.py` | Sphinx 配置 |
| `sync_md.py` | 根据 `source_md/*.md` 重新生成 `index.rst`（toctree） |
| `index.rst` | 首页与左侧目录（由 `sync_md.py` 生成） |
| `source_md/` | **文档源稿**（在此编辑 `*.md`） |
| `_build/html/` | HTML 输出 |

## 环境

```bash
pip install -r docs/sphinx/requirements.txt
```

## 构建与预览

```bash
# Linux / macOS
cd docs/sphinx
make html
make serve    # http://127.0.0.1:8000/

# Windows
cd docs\sphinx
make.bat html
make.bat serve
```

或从仓库根目录：

```powershell
.\scripts\build-docs.ps1
```

```bash
./scripts/build-docs.sh
```

## 修改左侧目录

编辑 `sync_md.py` 中的 `DOC_ORDER` 与 `write_index_rst()` 里的标题映射，然后重新 `make html`。
