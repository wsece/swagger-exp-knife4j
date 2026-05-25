#!/usr/bin/env python3
"""Serve built HTML docs with UTF-8 charset for plain-text source files."""
from __future__ import annotations

import argparse
import http.server
import socketserver
from pathlib import Path

HERE = Path(__file__).resolve().parent
DEFAULT_DIR = HERE / "_build" / "html"


class UTF8Handler(http.server.SimpleHTTPRequestHandler):
    extensions_map = {
        **http.server.SimpleHTTPRequestHandler.extensions_map,
        ".txt": "text/plain; charset=utf-8",
        ".md": "text/plain; charset=utf-8",
    }


def main() -> None:
    p = argparse.ArgumentParser(description="Serve Sphinx _build/html")
    p.add_argument("--port", type=int, default=8000)
    p.add_argument("--directory", type=Path, default=DEFAULT_DIR)
    args = p.parse_args()
    directory = args.directory.resolve()
    if not directory.is_dir():
        raise SystemExit(f"directory not found: {directory}")

    handler = UTF8Handler
    handler.directory = str(directory)  # type: ignore[attr-defined]

    with socketserver.TCPServer(("", args.port), handler) as httpd:
        print(f"Serving {directory} at http://127.0.0.1:{args.port}/")
        httpd.serve_forever()


if __name__ == "__main__":
    main()
