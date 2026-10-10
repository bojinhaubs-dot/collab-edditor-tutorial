# 实时代码协作编辑系统

项目架构与 MVP 范围见 [docs/proposal.md](docs/proposal.md)。当前完成 Go Gin 健康检查、文档创建与查询 API，以及 A3 的 MySQL 持久化。前端、协同编辑和权限尚未实现。

## 目录职责

- `cmd/api/`、`internal/httpapi/`：Go 服务入口及 HTTP API。
- `internal/database/`：MySQL 连接；`internal/model/`：GORM 文档模型；`internal/repository/`：文档数据库读写。
- `frontend/`：后续放置 React、CodeMirror 6 与 Yjs 客户端。
- `collaboration/`：后续放置 Node.js Yjs 协同服务；它负责权威 Y.Doc、同步、逐次写入授权与持久化确认。

## 运行

需要 Go 1.26、PowerShell 7 和 Docker Compose。在仓库根目录执行：

```powershell
$env:MYSQL_ROOT_PASSWORD = Read-Host -Prompt 'Local MySQL root password' -MaskInput
$env:MYSQL_PASSWORD = Read-Host -Prompt 'Local MySQL app password' -MaskInput
docker compose up -d mysql
docker compose ps
$env:MYSQL_DSN = "collab:$($env:MYSQL_PASSWORD)@tcp(127.0.0.1:3307)/collab?charset=utf8mb4&parseTime=True&loc=Local"
go run ./cmd/api
```

Go 服务默认监听 `127.0.0.1:8080`；可用 `ADDR` 环境变量修改。服务启动时连接 MySQL，并通过 GORM `AutoMigrate` 初始化 `documents` 表；缺少 `MYSQL_DSN`、连接失败或迁移失败时会打印错误并退出。MySQL 首次启动可能需要几秒钟，此时可等待容器就绪后重试 `go run`。Compose 将 MySQL 映射到本机 `127.0.0.1:3307`，数据保存在 `mysql_data` 卷中。密码只保存在本地环境变量中；如果该卷已用其他密码初始化，请输入原有密码。

## 文档 API

`POST /documents` 接收 JSON `{"title":"main.go","content":"package main"}`，返回 HTTP 201 和含 `id`、`title`、`content` 的文档。`title` 不能为空白；`content` 必须是字符串，可以是空字符串。无效请求返回 400，数据库保存失败返回 500。

`GET /documents/:id` 返回已创建文档；无效 ID 返回 400，不存在返回 404，数据库读取失败返回 500。`GET /health` 返回 `{"status":"ok"}`，仅表示 HTTP 服务运行，不代表数据库仍可用。API 尚无身份认证和权限控制；本阶段保存的是文档初始内容，不包含未来的 Yjs 更新日志。

## A3 测试方法

在另一个 PowerShell 终端、仓库根目录执行：

```powershell
go test ./...
go test ./tests -run 'TestDatabaseConnectionFailure|TestDatabaseFailureResponses' -v
$env:MYSQL_PASSWORD = Read-Host -Prompt 'Same MySQL app password' -MaskInput
$env:MYSQL_TEST_DSN = "collab:$($env:MYSQL_PASSWORD)@tcp(127.0.0.1:3307)/collab?charset=utf8mb4&parseTime=True&loc=Local"
go test ./tests -run TestMySQLDocumentSurvivesReconnect -v
```

最后一项需要先运行 `docker compose up -d mysql`；测试会通过一个数据库连接创建文档，再关闭连接、重新连接并查询相同 ID。没有设置 `MYSQL_TEST_DSN` 时该集成测试会标记为 **SKIP**，不能据此声称 MySQL 持久化已通过验证。

可手动验证真实 Go 进程重启：

```powershell
$created = Invoke-RestMethod -Uri http://127.0.0.1:8080/documents -Method Post -ContentType 'application/json' -Body '{"title":"main.go","content":"package main"}'
$created
Invoke-RestMethod "http://127.0.0.1:8080/documents/$($created.id)"
```

记下 `id`，在运行 Go 服务的终端按 `Ctrl+C`，再运行 `go run ./cmd/api`，随后重新执行 `GET /documents/<记录的id>`。仍返回原始 `title` 和 `content` 才说明重启后可查询。完成后可运行 `docker compose down`；不要加 `-v`，否则会删除保存文档的数据库卷。
