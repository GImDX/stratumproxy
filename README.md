# stratumproxy

A TLS proxy designed for mining software to connect to mining pools. It intercepts and optionally modifies specific mining protocol messages (e.g., `mining.authorize`, `mining.submit`, `mining.subscribe`) for authentication obfuscation, load balancing, or credential masking.  
一个为矿工软件设计的 TLS 代理，允许其通过中间层连接至矿池。该程序可拦截并（可选地）修改特定挖矿协议消息，用于认证信息替换、负载均衡或用户身份伪装。

## ✨ Features / 功能特性

- ✅ Acts as a secure TLS proxy between miner and pool  
  在矿机与矿池之间作为 TLS 安全代理
- ✅ Supports automatic replacement of `mining.authorize`, `mining.submit`, `mining.subscribe` usernames  
  自动替换这些消息中的用户名
- ✅ Preserves user-specific suffixes (e.g., `.worker_id`) during username replacement  
  替换时保留用户名后缀（如 `.worker_id`）
- ✅ Logs and forwards all JSON-RPC messages (optional debug mode)  
  可选调试模式记录所有 JSON-RPC 消息日志
- ✅ Automatically terminates connections on EOF or errors  
  连接断开或发生错误时自动关闭连接

## 🔧 Configuration / 配置参数

| Flag 参数 | Description 描述 | Default 默认值 |
|-----------|------------------|----------------|
| `--server-pem` | Path to TLS certificate file / TLS 证书路径 | `./server.pem` |
| `--server-key` | Path to TLS private key / TLS 私钥路径 | `./server.key` |
| `--listen-addr` | Address to listen for miner connections / 本地监听地址 | `:9999` |
| `--server-addr` | Remote mining pool address / 矿池地址 | `:1177` |
| `--replaced-user` | Username prefix before `.` to replace / 要替换用户名的前缀部分 | 示例值 |
| `--replaced-password` | Password to replace in `mining.authorize` / 要替换的密码 | `pyi114514` |

## 🏗️ Build / 编译

Ensure Go is installed (version 1.18+ recommended).  
请确保已安装 Go（建议版本 1.18+）。

```bash
go build -o stratumproxy
```

## 🚀 Usage / 使用示例

```bash
./stratumproxy \
  --server-pem server.pem \
  --server-key server.key \
  --listen-addr :34010 \
  --server-addr 43.134.68.141:10250 \
  --replaced-user "YOUR_REAL_MINER_USERNAME" \
  --replaced-password "PASSWORD"
```

Place your TLS certificate and key in the same directory or provide paths explicitly.  
请将 TLS 证书和私钥文件放在程序目录或通过参数指定路径。

## 🧪 Example Flow / 数据流示例

Miner sends / 矿机发送：

```json
{
  "id": 1,
  "method": "mining.authorize",
  "params": ["9i9m9AxmqgBUBD6G.worker1", "x"]
}
```

Proxy replaces / 代理修改后：

```json
{
  "id": 1,
  "method": "mining.authorize",
  "params": ["pyrin:qq0240xcnlk52jt4t007gwe97hnr33g5knx9kkgarmm0p9ghm9sg68qrakyf2.worker1", "pyi114514"]
}
```

## 🔍 Debug Mode / 调试模式

To enable raw data logging:  
若要启用调试日志：

```go
const debug = true
```

## 📌 Notes / 注意事项

- Only the part before `.` is replaced. Worker suffix is preserved.  
  程序仅替换用户名中 `.` 前的部分，保留后缀。
- TLS certificate verification is skipped (for debug/dev).  
  跳过 TLS 证书校验。
- Keep the proxy running if your miner uses persistent connections.  
  若矿机需要持续连接，请保持代理持续运行。

## 📃 License / 许可协议

MIT License. Provided as-is. Use at your own risk.  
MIT 协议，使用风险自负。
