#!/usr/bin/env python3
"""Regenerate index.rst from docs/sphinx/source_md/*.md (canonical doc source)."""
from __future__ import annotations

from pathlib import Path

HERE = Path(__file__).resolve().parent
SOURCE_DIR = HERE / "source_md"

SKIP_NAMES = {"README.md"}

# Sphinx toctree order (filename without .md)
DOC_ORDER = [
    "index",
    "introduction",
    "installation",
    "docker",
    "tutorial",
    "commands-scan",
    "commands-report",
    "commands-mcp",
    "commands-version",
    "smart-analysis",
    "web-report",
    "faq",
    "mcp-tools",
    "module-development",
    "building-site",
]


def collect_markdown_stems() -> list[str]:
    stems: list[str] = []
    for path in sorted(SOURCE_DIR.glob("*.md")):
        if path.name in SKIP_NAMES:
            continue
        stems.append(path.stem)
    return stems


def write_index_rst(stems: list[str]) -> None:
    order = [s for s in DOC_ORDER if s in stems]
    for s in stems:
        if s not in order:
            order.append(s)

    lines = [
        "swagger-exp-knife4j 文档",
        "========================",
        "",
        "面向 **Swagger / Knife4j / OpenAPI** 的接口发现与探测工具。",
        "",
        ".. toctree::",
        "   :maxdepth: 2",
        "   :caption: 目录",
        "",
    ]
    titles = {
        "index": "欢迎",
        "introduction": "项目概览",
        "installation": "安装与编译",
        "docker": "Docker 部署",
        "tutorial": "快速入门",
        "commands-scan": "scan 扫描",
        "commands-report": "report 报告",
        "commands-mcp": "mcp 服务",
        "commands-version": "version 版本",
        "smart-analysis": "响应相似度分析",
        "web-report": "Web 报告站",
        "faq": "常见问题",
        "mcp-tools": "MCP 工具契约",
        "module-development": "模块开发与扩展",
        "building-site": "构建文档站",
    }
    for stem in order:
        if stem == "index":
            continue
        title = titles.get(stem, stem)
        lines.append(f"   {title} <source_md/{stem}>")
    lines.append("")
    # RST :doc: must be in index.rst (not included index.md). Place before include so the toctree is parsed first.
    if "smart-analysis" in stems:
        lines.extend([
            "技术分析",
            "--------",
            "",
            "响应相似度（SimHash/汉明分组）、Web 列表去重与排序、报文详情 body 上限见 :doc:`source_md/smart-analysis`。",
            "",
        ])
    if "module-development" in stems:
        lines.extend([
            "二次开发",
            "--------",
            "",
            "需要自定义扫描逻辑、CLI 或 MCP 工具时，见 :doc:`source_md/module-development`",
            "（``pkg/extension`` + ``extensions/``）。",
            "",
        ])
    lines.append(".. include:: source_md/index.md")
    lines.append("   :parser: myst_parser.sphinx_")
    lines.append("")

    (HERE / "index.rst").write_text("\n".join(lines), encoding="utf-8")


def main() -> None:
    if not SOURCE_DIR.is_dir():
        raise SystemExit(f"Missing {SOURCE_DIR} — add Markdown pages there first.")
    stems = collect_markdown_stems()
    write_index_rst(stems)
    print(f"Regenerated index.rst from {len(stems)} file(s) in {SOURCE_DIR}")


if __name__ == "__main__":
    main()
