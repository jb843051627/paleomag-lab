# Docker 打包说明

镜像由 `benzhi.Dockerfile` 构建，基础镜像固定为 `golang:1.22-bookworm`，使用纯 Go SQLite 驱动，不依赖外部数据库服务。

```bash
./build_benzhi_docker.sh <image-name> linux/amd64
./build_benzhi_docker.sh <image-name> linux/arm64
docker run --rm -p 8080:8080 benzhi/<image-name>:latest
```
