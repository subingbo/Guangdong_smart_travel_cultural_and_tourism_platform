# 小程序本地前后端联调指南

小程序已从微信云开发切换到本地 Go 后端（`server-go`，默认 `http://127.0.0.1:8081`）。
所有数据与 AI 能力（路线、游记、导游、文案、海报、菜单 OCR/翻译）都通过 HTTP + JWT 调用后端，
可直接在微信开发者工具里完成前后端联调、截图与录屏。

## 1. 启动数据库（Docker）

```bash
docker compose up -d mysql
```

首次启动会自动执行 `sql/schema.sql` 与 `sql/seed.sql`（含 `demo/demo123`、`alice/demo123` 种子账号）。
验证：

```bash
docker exec travel_guangdong-mysql-1 mysql -utravelgd -ptravelgd travel_gd -e "SELECT COUNT(*) FROM scenic_images;"
```

## 2. 启动后端

```bash
# Git Bash
scripts/dev-backend.sh
# PowerShell
powershell -ExecutionPolicy Bypass -File scripts\dev-backend.ps1
```

脚本会加载根目录 `.env`（`DASHSCOPE_API_KEY` 等），设置本地数据库连接与 `UPLOAD_DIR=.local-uploads`，
然后 `go run .`。看到 `travel-gd listening on :8081` 即启动成功。

冒烟验证：

```bash
curl -s http://127.0.0.1:8081/api/scenics
TOKEN=$(curl -s -X POST http://127.0.0.1:8081/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"demo","password":"demo123"}' | python -c "import sys,json;print(json.load(sys.stdin)['data']['token'])")
curl -s http://127.0.0.1:8081/api/me -H "Authorization: Bearer $TOKEN"
```

## 3. 打开小程序

1. 微信开发者工具「导入项目」，选择仓库根目录 `D:\code\Travel_Guangdong`（appid 已配置）。
2. 详情 → 本地设置 → 勾选「不校验合法域名…」（`project.private.config.json` 已设 `urlCheck: false`）。
3. 设置 → 安全设置 → 开启「服务端口」（自动截图脚本需要）。
4. 编译后即可登录、生成路线、发布游记等；「网络」面板能看到所有 `127.0.0.1:8081` 请求。

## 4. 关键实现

| 能力 | 位置 |
| --- | --- |
| 后端地址 | `miniprogram/config/env.ts`（`BASE_URL`，上线时改为 HTTPS 域名） |
| 请求封装（token 注入 / 401 跳登录 / 统一 toast） | `miniprogram/utils/request.ts` |
| 上传封装（`wx.uploadFile` + token） | `miniprogram/utils/upload.ts` |
| 登录态与设备 ID | `miniprogram/utils/auth.ts` |
| API 类型层 | `miniprogram/services/api.ts` |
| 登录页（微信一键登录 + 账号密码） | `miniprogram/pages/user/login/` |

后端新增接口：

- `POST /api/auth/wechat-login` — 配 `WX_APPID/WX_SECRET` 走真实 `jscode2session`；未配置时用 `deviceId` 派生稳定开发账号。
- `POST /api/ai/copy` — 商家营销文案
- `POST /api/ai/poster` + `GET /api/ai/poster/:taskId` — 文生图异步任务（DashScope wanx）
- `POST /api/ai/menu-ocr` — 菜单识别（`AI_VISION_MODEL`，默认 `qwen3-vl-plus`）
- `POST /api/ai/menu-translate` — 菜单翻译
- `POST /api/ai/dish-explain` — 菜品文化卡片 / 讲解
- `GET /uploads/*` — 本地静态图片（生产由 nginx 托管）

## 5. 自动截图

```bash
node scripts/mp-screenshots.js          # 全流程驱动 + 截图到 docs/screenshots/
SKIP_SHOTS=1 node scripts/mp-screenshots.js   # 只驱动不截图（验证流程）
```

前提：开发者工具已打开项目、服务端口开启、**桌面未锁屏**（Windows 锁屏时只能截到锁屏画面）。
脚本会优先 `connect ws://127.0.0.1:9420`，失败则尝试 `automator.launch`。

> 已知兼容性：部分开发者工具版本不会回包 `navigateTo/switchTab` 与 `App.captureScreenshot`，
> 脚本已做容错（忽略导航 ack、截图失败时回退 PowerShell 窗口截图）。
> 若窗口截图得到白屏/锁屏画面，请解锁桌面后重跑。

自动启动 IDE（服务端口已开启时）：

```bash
"/c/Program Files (x86)/Tencent/微信web开发者工具/cli.bat" auto \
  --project "D:\code\Travel_Guangdong" --auto-port 9420
# 若弹出升级提示挡住主窗口：
powershell -ExecutionPolicy Bypass -File scripts/close-devtools-dialogs.ps1
```

## 6. 环境变量（根目录 .env）

```ini
DASHSCOPE_API_KEY=sk-...      # 必填（文本 + 视觉 + 文生图）
AI_MODEL=qwen-plus            # 文本模型
# AI_VISION_MODEL=qwen3-vl-plus   # 视觉模型（默认值，可覆盖）
AMAP_WEB_KEY=...              # 路线几何 / 天气（可缺省，降级直线）
# WX_APPID=...                # 配置后微信登录走真实 code2session
# WX_SECRET=...
JWT_SECRET=...                # 建议本地显式设置
```

## 7. 常见问题

| 现象 | 处理 |
| --- | --- |
| 所有请求 `request:fail` | 后端未启动；或 BASE_URL 端口与后端不一致 |
| 中文乱码 | 重新初始化数据库（`docker compose down -v && docker compose up -d mysql`），SQL 已加 `SET NAMES utf8mb4` |
| 502 `Model not exist` | 当前 DashScope 账号未开通该模型，改用 `AI_VISION_MODEL=qwen3-vl-plus`（或 `qwen-vl-ocr`） |
| 图片 404 | `UPLOAD_DIR` 未设置或 `/uploads` 路由被 nginx 接管；本地联调请用 `scripts/dev-backend.sh` |
| 微信登录 401 | 配了 `WX_APPID/WX_SECRET` 但 code 无效；本地演示可留空走开发模式 |
