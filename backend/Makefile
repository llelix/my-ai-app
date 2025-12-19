# AI知识库查询应用 - 后端 Makefile

.PHONY: help dev build swagger clean test deps

# 默认目标
help:
	@echo "AI知识库查询应用 - 后端管理命令"
	@echo ""
	@echo "可用命令:"
	@echo "  dev      - 启动开发服务器"
	@echo "  build    - 构建生产版本"
	@echo "  swagger  - 生成Swagger文档"
	@echo "  clean    - 清理构建文件"
	@echo "  test     - 运行测试"
	@echo "  deps     - 安装依赖"
	@echo "  docker   - 启动PostgreSQL数据库"

# 安装依赖
deps:
	go mod tidy
	go mod download

# 生成Swagger文档
swagger:
	swag init -g ./cmd/server/main.go -o ./docs --parseDependency --parseInternal --quiet

# 启动开发服务器
dev: swagger
	cd cmd/server && go run .

# 构建生产版本
build: swagger
	go build -o ai-knowledge-app cmd/server/main.go

# 运行测试
test:
	go test ./...

# 清理构建文件
clean:
	rm -f ai-knowledge-app
	rm -rf docs/

# 启动PostgreSQL数据库
docker:
	docker-compose up -d postgres

# 停止PostgreSQL数据库
docker-stop:
	docker-compose down

# 查看数据库日志
docker-logs:
	docker-compose logs -f postgres