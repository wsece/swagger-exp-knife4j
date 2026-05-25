swagger-exp-knife4j 文档
========================

面向 **Swagger / Knife4j / OpenAPI** 的接口发现与探测工具。

.. toctree::
   :maxdepth: 2
   :caption: 目录

   项目概览 <source_md/introduction>
   安装与编译 <source_md/installation>
   Docker 部署 <source_md/docker>
   快速入门 <source_md/tutorial>
   scan 扫描 <source_md/commands-scan>
   report 报告 <source_md/commands-report>
   mcp 服务 <source_md/commands-mcp>
   version 版本 <source_md/commands-version>
   响应相似度分析 <source_md/smart-analysis>
   Web 报告站 <source_md/web-report>
   常见问题 <source_md/faq>
   MCP 工具契约 <source_md/mcp-tools>
   模块开发与扩展 <source_md/module-development>
   构建文档站 <source_md/building-site>

技术分析
--------

响应相似度（SimHash/汉明分组）、Web 列表去重与排序、报文详情 body 上限见 :doc:`source_md/smart-analysis`。

二次开发
--------

需要自定义扫描逻辑、CLI 或 MCP 工具时，见 :doc:`source_md/module-development`
（``pkg/extension`` + ``extensions/``）。

.. include:: source_md/index.md
   :parser: myst_parser.sphinx_
