# 扩展模块目录

在此目录放置自定义 Go 包，并在 `main.go` 中空白导入以注册：

```go
import _ "swagger-exp-knife4j/extensions/myplugin"
```

完整契约见在线文档 **模块开发与扩展接口**（`docs/sphinx/source_md/module-development.md`）。

参考实现：`extensions/example/`（默认未链接，需自行添加上述 import 后 `go build`）。
