# paleomag-lab

古地磁实验室试样实验流程服务。项目提供 Go HTTP 后端、SQLite 文件持久化和内嵌实验工作台页面。

## 本地运行

```bash
GOTOOLCHAIN=local go run .
```

默认监听 `:8080`，数据库文件为 `./paleomag.db`。可以通过 `PALEOMAG_DB` 和 `PALEOMAG_ADDR` 覆盖。

## 检查

```bash
go build ./...
go test -count=1 ./...
go run . --smoke-test
```
