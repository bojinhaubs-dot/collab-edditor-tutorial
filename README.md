# 实时代码协作编辑系统

项目架构与 MVP 范围见 [docs/proposal.md](docs/proposal.md)。当前完成 Go Gin 服务骨架、健康检查，以及内存版文档创建和查询 API；协同编辑、权限和持久化尚未实现。

## 目录职责

- `cmd/api/`、`internal/httpapi/`：Go 业务服务入口及 HTTP API；后续负责身份、文档、成员和权限，不维护权威 Y.Doc。
- `frontend/`：后续放置 React、CodeMirror 6 与 Yjs 客户端。
- `collaboration/`：后续放置 Node.js Yjs 协同服务；它负责权威 Y.Doc、同步、逐次写入授权与持久化确认。

## 运行与验证

在仓库根目录运行：

```powershell
go run ./cmd/api
```

默认监听 `127.0.0.1:8080`；可用 `ADDR` 环境变量修改地址。另开终端执行：

```powershell
Invoke-RestMethod http://127.0.0.1:8080/health
go test ./...
```

`GET /health` 返回 `{"status":"ok"}`，仅表示 Go HTTP 服务正在运行，不代表数据库或协同服务就绪。

## 文档 API（A2 内存版）

`POST /documents` 接收 JSON `{"title":"main.go","content":"package main"}`，返回 HTTP 201 和含 `id`、`title`、`content` 的文档。`title` 不能为空白；`content` 必须是字符串，可以是空字符串。无效请求返回 400。

`GET /documents/:id` 返回已创建文档；不存在时返回 404。文档目前只保存在当前 Go 进程的内存中，服务重启后会丢失；接口尚无身份认证和权限控制，不可用于满足方案中的安全与恢复验收。
