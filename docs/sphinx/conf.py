# Sphinx configuration — confdir is docs/sphinx; Markdown sources are synced to source_md/.
from __future__ import annotations

import sys
from datetime import datetime
from pathlib import Path

CONF_DIR = Path(__file__).resolve().parent
if str(CONF_DIR) not in sys.path:
    sys.path.insert(0, str(CONF_DIR))

project = "swagger-exp-knife4j"
author = "swagger-exp-knife4j"
copyright = f"{datetime.now().year}, {author}"

extensions = [
    "myst_parser",
    "sphinx.ext.autodoc",
    "sphinx.ext.viewcode",
    "sphinx_copybutton",
]

templates_path = ["_templates"]
exclude_patterns = ["_build", "Thumbs.db", ".DS_Store", "README.md"]

source_suffix = {
    ".rst": "restructuredtext",
    ".md": "markdown",
}

root_doc = "index"
master_doc = "index"

language = "zh_CN"

pygments_style = "sphinx"

html_theme = "sphinx_rtd_theme"
html_static_path = ["_static"]
html_css_files = ["custom.css"]
html_title = "swagger-exp-knife4j 文档"
html_short_title = "swagger-exp-knife4j"
html_show_sphinx = True
html_show_copyright = True

html_theme_options = {
    "navigation_depth": 4,
    "collapse_navigation": False,
    "sticky_navigation": True,
    "includehidden": True,
    "titles_only": False,
}

myst_enable_extensions = [
    "colon_fence",
    "deflist",
    "linkify",
    "substitution",
    "tasklist",
    "dollarmath",
]
myst_heading_anchors = 3
myst_fence_as_directive = ["note", "warning", "tip", "important", "caution"]

copybutton_prompt_text = r">>> |\.\.\. |\$ |In \[\d*\]: | {2,5}\.\.\."
copybutton_prompt_is_regexp = True
