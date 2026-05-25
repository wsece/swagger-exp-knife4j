// main.go 程序入口：调用 cmd.Execute() 启动 Cobra 命令行。
package main

import "swagger-exp-knife4j/cmd"

// main 将进程参数交给 cmd 包解析并执行对应子命令。
func main() {
	cmd.Execute()
}
