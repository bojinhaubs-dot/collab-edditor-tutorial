# 实时代码协作编辑系统

项目架构与 MVP 范围见 [docs/proposal.md](docs/proposal.md)。当前完成 Go Gin 文档 API、A3 的 MySQL 持久化、B1 的单用户代码编辑页，以及 B2 的 Yjs WebSocket 连接与状态显示。双人协同编辑和权限尚未实现。

## 目录职责

- `cmd/api/`、`internal/httpapi/`：Go 服务入口及 HTTP API。
- `internal/database/`：MySQL 连接；`internal/model/`：GORM 文档模型；`internal/repository/`：文档数据库读写。
- `frontend/`：React + CodeMirror 6 编辑页；B2 使用 Yjs v13 和 `y-websocket` 建立连接，但尚未将编辑器绑定到 `Y.Text`。
- `collaboration/`：Node.js Yjs 协同服务；当前只提供内存中的标准 Yjs 同步协议。逐次写入授权和可靠持久化属于后续任务。

## 运行

需要 Go 1.26、Node.js 20.19+ 或 22.12+、PowerShell 7 和 Docker Compose。在仓库根目录执行：

```powershell
$env:MYSQL_ROOT_PASSWORD = Read-Host -Prompt 'Local MySQL root password' -MaskInput
$env:MYSQL_PASSWORD = Read-Host -Prompt 'Local MySQL app password' -MaskInput
docker compose up -d mysql
docker compose ps
$env:MYSQL_DSN = "collab:$($env:MYSQL_PASSWORD)@tcp(127.0.0.1:3307)/collab?charset=utf8mb4&parseTime=True&loc=Local"
go run ./cmd/api
```

Go 服务默认监听 `127.0.0.1:8080`；可用 `ADDR` 环境变量修改。服务启动时连接 MySQL，并通过 GORM `AutoMigrate` 初始化 `documents` 表；缺少 `MYSQL_DSN`、连接失败或迁移失败时会打印错误并退出。MySQL 首次启动可能需要几秒钟，此时可等待容器就绪后重试 `go run`。Compose 将 MySQL 映射到本机 `127.0.0.1:3307`，数据保存在 `mysql_data` 卷中。密码只保存在本地环境变量中；如果该卷已用其他密码初始化，请输入原有密码。

在**第二个终端**启动协同服务（保持 Go API 运行）：

```powershell
cd collaboration
npm ci
npm start
```

协同服务默认监听 `ws://127.0.0.1:1234`，可用 `COLLAB_HOST` 和 `COLLAB_PORT` 修改监听地址和端口。默认只监听本机；B2 尚无鉴权，不应暴露到公网。这个服务可独立于 Go API 和 MySQL 启动。

在**第三个终端**启动前端：

```powershell
cd frontend
npm ci
npm run dev
```

前端开发服务器固定使用 `http://127.0.0.1:5173`，并将 `/documents` 请求代理到 `http://127.0.0.1:8080`。如果修改了 Go 服务的 `ADDR`，也需要相应修改 `frontend/vite.config.js` 中的代理目标；如果修改协同服务端口，需要设置前端环境变量 `VITE_COLLAB_WS_URL`，例如 `ws://127.0.0.1:1235`。前端可独立启动，但读取文档需要 Go API 和 MySQL 可用。

## 文档 API

`POST /documents` 接收 JSON `{"title":"main.go","content":"package main"}`，返回 HTTP 201 和含 `id`、`title`、`content` 的文档。`title` 不能为空白；`content` 必须是字符串，可以是空字符串。无效请求返回 400，数据库保存失败返回 500。

`GET /documents/:id` 返回已创建文档；无效 ID 返回 400，不存在返回 404，数据库读取失败返回 500。`GET /health` 返回 `{"status":"ok"}`，仅表示 HTTP 服务运行，不代表数据库仍可用。API 尚无身份认证和权限控制；本阶段保存的是文档初始内容，不包含未来的 Yjs 更新日志。

## B1 编辑页面与测试方法

在另一个终端、仓库根目录创建一个测试文档：

```powershell
$body = @{ title = 'main.go'; content = "package main`n`nfunc main() {}`n" } | ConvertTo-Json
$created = Invoke-RestMethod -Uri http://127.0.0.1:8080/documents -Method Post -ContentType 'application/json' -Body $body
$created
```

打开 `http://127.0.0.1:5173/?id=<返回的id>`（例如 `?id=1`），按以下步骤验收：

1. 页面显示 API 返回的标题、ID 和代码内容；Go 关键字有语法高亮。
2. 在编辑器中输入或删除字符，底部字符数随之变化，状态变为“本地修改未保存”。
3. 刷新页面，内容恢复为 API 中的初始内容，状态回到“尚无本地修改”。B1 没有保存接口。
4. 打开 `http://127.0.0.1:5173/?id=999999999`，应显示“文档不存在”；打开无 `id` 参数的首页，应显示 ID 提示。停止 Go 服务后刷新有效文档页，应显示加载错误。

自动检查命令（仓库根目录）：

```powershell
go test ./...
cd frontend
npm ci
npm run build
```

`npm run build` 只验证前端可打包；加载、语法高亮和编辑状态需按上面的浏览器步骤验证。B2 虽然建立了 Yjs 连接，但编辑器仍未绑定 `Y.Text`；编辑内容仅存于当前页面内存，不能视为已同步或保存。

## B2 WebSocket 连接与测试方法

先按“运行”一节启动 MySQL、Go API、协同服务和前端，再按 B1 步骤创建文档并打开 `http://127.0.0.1:5173/?id=<返回的id>`。

1. 页面文档栏应显示“已连接协同服务”。在浏览器开发者工具中打开 **Network → WS**，刷新页面；应看到 `ws://127.0.0.1:1234/document-<id>` 的 HTTP `101 Switching Protocols` 握手，以及 **Messages/Frames** 中的二进制 Yjs 同步消息。
2. 在运行协同服务的终端按 `Ctrl+C`，页面应显示“已断开，正在重连”。重新运行 `npm start`，页面应恢复“已连接协同服务”，无需刷新。
3. 在编辑器输入文字后刷新，页面仍会从 Go API 加载初始内容。B2 只验证连接与协议；双人编辑、权限控制、Yjs 更新持久化和“已保存”确认尚未实现。

自动验证命令：

```powershell
cd collaboration
npm ci
npm test
cd ../frontend
npm ci
npm run build
cd ..
go test ./...
```

`npm test` 使用真实 `y-websocket` 客户端连接临时 Node 服务，检查同步事件和断开后的重连。浏览器握手与页面状态仍需按上述手动步骤检查。`npm ci` 需要正常安装可选依赖；当前 `@y/websocket-server@0.1.1` 在导入同步工具时也会导入 `y-leveldb`，因此不要用 `--omit=optional` 安装协同服务。

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
