#!/usr/bin/env python3
"""Wrap Sphinx _sources/*.txt as UTF-8 HTML pages and fix sourcelink hrefs.

Raw .txt sources have no charset; browsers on Windows often guess GBK and show mojibake.
"""
from __future__ import annotations

import html
import re
import sys
from pathlib import Path

_SOURCE_LINK = re.compile(r'(_sources/[^"]+)\.txt"')

HERE = Path(__file__).resolve().parent
HTML_ROOT = HERE / "_build" / "html"
SOURCES_ROOT = HTML_ROOT / "_sources"


def txt_to_html(txt_path: Path) -> Path:
    text = txt_path.read_text(encoding="utf-8")
    out_path = txt_path.with_suffix(".html")
    title = html.escape(txt_path.name)
    body = html.escape(text)
    doc = (
        "<!DOCTYPE html>\n"
        '<html lang="zh-CN">\n'
        "<head>\n"
        '  <meta charset="utf-8" />\n'
        f"  <title>{title}</title>\n"
        "  <style>pre{white-space:pre-wrap;word-wrap:break-word;font-family:Consolas,monospace;}</style>\n"
        "</head>\n"
        "<body>\n"
        f"<pre>{body}</pre>\n"
        "</body>\n"
        "</html>\n"
    )
    out_path.write_text(doc, encoding="utf-8")
    return out_path


def patch_sourcelinks(html_path: Path) -> bool:
    content = html_path.read_text(encoding="utf-8")
    updated = _SOURCE_LINK.sub(r'\1.html"', content)
    if updated == content:
        return False
    html_path.write_text(updated, encoding="utf-8")
    return True


def main() -> int:
    if not SOURCES_ROOT.is_dir():
        print(f"skip: {SOURCES_ROOT} not found (run sphinx-build first)", file=sys.stderr)
        return 0

    wrapped = 0
    for txt in SOURCES_ROOT.rglob("*.txt"):
        txt_to_html(txt)
        wrapped += 1

    patched = 0
    for page in HTML_ROOT.rglob("*.html"):
        if "_sources" in page.parts:
            continue
        if patch_sourcelinks(page):
            patched += 1

    print(f"postprocess_sourcelinks: {wrapped} source wrap(s), {patched} page link(s) patched")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
