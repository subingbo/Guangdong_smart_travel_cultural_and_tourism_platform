# 生产部署说明（Azure 单机）

把 `web/` + `server-go/` + MySQL 部署到 **20.194.43.11**（Ubuntu 22.04，2C1G）。
该机已有一个无关服务 `seaside`（nginx:80 + Java:8090 + Node:5000 + MySQL + Redis），本部署与它完全隔离、只共用 nginx 与 MySQL 实例。

## 架构

```
公网 :8091 ──> 主机 nginx（新增 sites-enabled/travel-gd.conf）
                 ├─ /            /opt/travel-gd/web        (web/dist 静态文件)
                 └─ /api/   ──>  127.0.0.1:18080           (Go 后端，仅回环)
                                       │
                                       ├─> DashScope qwen-plus (AI 路线/游记/导游)
                                       ├─> 高德 Web服务（有 AMAP_WEB_KEY 时）
                                       ├─> 腾讯地图（无高德 Web服务 key 时的回退）
                                       └─> 127.0.0.1:3306  MySQL schema `travel_gd`
                                                           专属用户 `travelgd_app`
```

浏览器端地图使用**高德 JS API 2.0**（GCJ-02，与数据库坐标一致），
key 在构建时通过 `web/.env.local` 注入：`VITE_AMAP_JS_KEY` + `VITE_AMAP_JS_SECURITY_CODE`。

## 隔离边界

| 资源 | 本服务 | 既有服务 | 隔离方式 |
|---|---|---|---|
| MySQL | schema `travel_gd`，用户 `travelgd_app`@localhost/127.0.0.1 | schema `seaside`，用户 `seaside` | 独立 schema + 仅授权本 schema（无全局权限） |
| Redis | **不使用**（项目代码无 Redis 依赖） | db0 | 无交集 |
| nginx | `:8091` 独立 server 块 | `:80` default_server | 只新增文件，不修改 `seaside.conf` |
| 进程 | `travel-gd.service`，用户 `travelgd` | `seaside-*.service`，用户 `sususu` | 独立 unit / 独立 OS 用户 |
| 目录 | `/opt/travel-gd` | `/opt/seaside` | 独立目录 |

## 文件布局（服务器）

```
/opt/travel-gd/
  bin/travel-gd-server          # linux/amd64 静态二进制
  web/                          # web/dist 构建产物
  etc/app.env                   # 0600 root:root，含 DB 密码 / JWT / AI keys
/etc/systemd/system/travel-gd.service
/etc/nginx/sites-available/travel-gd.conf  -> sites-enabled/travel-gd.conf
/var/log/nginx/travel-gd.{access,error}.log
```

## 首次部署

前置：本机安装 Go 1.25+ 与 Node 20+；目标机可 SSH（`sususu`，免密 sudo）。

```bash
# 1. 构建
cd server-go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o /tmp/travel-gd-server .
cd ../web
# 先在 web/.env.local 写入 VITE_AMAP_JS_KEY / VITE_AMAP_JS_SECURITY_CODE（gitignored）
npm ci && npm run build          # VITE_API_BASE_URL 必须为空（同源 /api）

# 2. 打包并上传（schema/seed + 二进制 + dist + 单元/nginx 配置 + bootstrap.sh）
# 3. 在目标机解包目录执行（API key 从本地未跟踪的 .env 读取，不落仓库）：
DASHSCOPE_API_KEY=... TENCENT_MAP_API_KEY=... bash bootstrap.sh
```

`bootstrap.sh` 是幂等的：重复执行会重置应用文件与 `app.env`（生成新的 DB 密码与 JWT），
但只在 `users` 表为空时导入种子数据。

## 更新部署

```bash
# 后端
cd server-go && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o /tmp/travel-gd-server .
scp /tmp/travel-gd-server sususu@20.194.43.11:/tmp/
ssh sususu@20.194.43.11 'sudo install -o travelgd -g travelgd -m 0755 /tmp/travel-gd-server /opt/travel-gd/bin/travel-gd-server && sudo systemctl restart travel-gd'

# 前端
cd web && npm run build
tar czf - -C dist . | ssh sususu@20.194.43.11 'sudo rm -rf /opt/travel-gd/web/* && sudo tar xzf - -C /opt/travel-gd/web && sudo chown -R travelgd:travelgd /opt/travel-gd/web'
```

## 验证

```bash
curl -I http://127.0.0.1:8091/                      # 200，index.html
curl    http://127.0.0.1:8091/api/scenics           # 种子景点数据
curl -X POST http://127.0.0.1:8091/api/auth/login \
     -H 'Content-Type: application/json' \
     -d '{"username":"demo","password":"demo123"}'  # 200 + token
curl -H "Authorization: Bearer <token>" http://127.0.0.1:8091/api/me
sudo mysql -N -e "SHOW GRANTS FOR 'travelgd_app'@'localhost';"   # 仅 travel_gd.*
sudo journalctl -u travel-gd -n 50 --no-pager
```

外网入口：`http://20.194.43.11:8091`（UFW 已放行 8091；Azure NSG 实测已可访问）。

## 回滚

```bash
sudo systemctl disable --now travel-gd
sudo rm /etc/nginx/sites-enabled/travel-gd.conf
sudo nginx -t && sudo systemctl reload nginx
# 数据库默认保留；确认不再需要后再手动 DROP DATABASE travel_gd;
```

既有 `seaside` 站点全程未修改，回滚不影响它。

## 已知问题

- **高德 key 类型**：高德控制台里 "Web端(JS API)" 的 key 不能调用 Web服务 API（会返回
  `USERKEY_PLAT_NOMATCH`）。浏览器地图用 JS API key；后端路线几何需要另建一条
  **Web服务** key 并配置 `AMAP_WEB_KEY`。未配置时后端自动回退到腾讯地图（坐标同为 GCJ-02，
  前端高德底图显示无偏移）。
- **AI 模型**：统一使用 `AI_MODEL=qwen-plus`。同一账号的 `qwen-turbo` 目前返回
  `Throttling.RateQuota`（429），因此不要把模型改回 qwen-turbo。
- **地图瓦片**：国内由高德 CDN 提供，无需再依赖 OpenStreetMap。
- **种子账号**：`demo` / `demo123`、`alice` / `demo123`（`sql/seed.sql` 已修正 bcrypt 哈希）。
- **无健康检查接口**：用 `GET /api/scenics` 作为探活。

## 安全约定

- `app.env` 权限 `0600 root:root`，只由 systemd 读取；**不要提交到仓库**。
- 仓库内 `deploy/app.env.example` 只含占位符。
- 后端只监听 `127.0.0.1:18080`，公网只能经 nginx:8091 访问。
