# Minimal makefile for Sphinx documentation in docs/sphinx/
SPHINXOPTS    ?=
SPHINXBUILD   ?= sphinx-build
SOURCEDIR     = .
BUILDDIR      = _build

.PHONY: help sync html clean serve

help:
	@$(SPHINXBUILD) -M help "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS)

sync:
	python sync_md.py

html: sync
	@$(SPHINXBUILD) -M html "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS)
	python postprocess_sourcelinks.py

clean:
	rm -rf "$(BUILDDIR)"
	@$(SPHINXBUILD) -M clean "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) || true

serve: html
	@echo Open http://127.0.0.1:8000/ after server starts
	python serve_docs.py --port 8000 --directory "$(BUILDDIR)/html"
