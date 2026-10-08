# 实时代码协作编辑后端 Demo

当前是第 1 步骨架：HTTP 健康检查和独立的 WebSocket echo。文件 Room、认证、编辑和权限协议尚未接入。

## 运行

需要 Go 1.26 或更新版本。在项目根目录执行：

```powershell
go run .
```

服务默认监听 `127.0.0.1:8080`。可用 `ADDR` 环境变量更改地址。另开终端检查：

```powershell
curl.exe http://127.0.0.1:8080/healthz
```

预期返回 `{"status":"ok"}`。按 Ctrl+C 停止服务。

## WebSocket echo

连接 `ws://127.0.0.1:8080/ws/echo`，发送文本帧（例如 `{"type":"ping","text":"hello"}`），服务端会原样回送消息类型和内容；二进制帧也会原样回送。它只用于验证 WebSocket 连接和收发，不是协作编辑协议。

可以先在浏览器打开 `http://127.0.0.1:8080/healthz`，再在该页面的开发者工具控制台运行：

```javascript
const ws = new WebSocket("ws://127.0.0.1:8080/ws/echo");
ws.onmessage = event => console.log(event.data);
ws.onopen = () => ws.send(JSON.stringify({ type: "ping", text: "hello" }));
```

Echo 保留 WebSocket 库默认的同源检查。设计文档中的 `/ws/files/{fileId}` 将用于后续的首帧认证和文件订阅，不与 echo 共用协议。

## 测试

```powershell
go test ./...
```
