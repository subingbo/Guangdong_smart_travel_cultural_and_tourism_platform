# Web 演示与截图清单

> 更新：2026-11-22。适用于本地 `http://localhost:5173`（Vite）+ Go 后端 `http://127.0.0.1:8081`。

## 1. 启动

```powershell
# MySQL（仓库根目录）
docker compose up -d mysql

# 后端
powershell -ExecutionPolicy Bypass -File scripts\dev-backend.ps1

# 前端（仓库根目录另开终端）
cd web
npm run dev
```

Vite 默认把 `/api` 和 `/uploads` 都代理到 `http://localhost:8081`。如果后端不在 8081：

```powershell
$env:API_PROXY_TARGET="http://127.0.0.1:8081"; npm run dev
```

健康检查：`http://127.0.0.1:8081/healthz`，正常返回 `{"status":"ok","agent":true}`。

## 2. 演示账号

- 游客：`demo` / `demo123`（林晓岚）
- 商家：`merchant` / `merchant123`
- 管理员：`admin` / `admin123`
- 社区用户：`maya`、`kenji`、`lena`、`jiaqi`、`qiming`、`sofia`，密码均为 `demo123`

右上角可以切换中英文；截图前建议先切到中文。

## 3. 截图页面

| 页面 | 路由 | 建议操作 |
| --- | --- | --- |
| 首页 | `/#/` | 等轮播加载完，截全页 |
| AI 路线表单 | `/#/route` | 选 2 天、美食/历史、适中，点生成 |
| 路线详情 + 地图 | `/#/routes/:id` | 等地图 polyline 出现后截图 |
| 游记列表 | `/#/blog` | 8 篇中英文游记、19 条评论；封面应全部显示 |
| 游记详情 | `/#/blog/:id` | 截大图、点赞、评论区 |
| 发布游记 | `/#/blog` → New Post | 上传图片、AI 起草、发布 |
| 文旅活动 | `/#/activities` | 5 场活动，首页也会推荐最近 3 场 |
| AI 导游 | `/#/guide` | 输入问题，截流式回复和方案卡片 |
| 商家 AI 工作台 | `/#/merchant` | 四个 Tab：文案 / 海报 / 拍照翻译 / 客服 |
| 商城 | `/#/shop` | 8 个商品，含评分、已售、精选评价；加购、购物车、提交预订单 |
| 我的订单 | `/#/orders` | 截订单号、状态、商品和金额 |
| 个人中心 | `/#/profile` | 头像、统计、我的文章、我的收藏 |
| 通知 | `/#/notifications` | 未读/全部筛选、标记已读 |
| 运营后台 | `/#/admin` | 用 admin 登录；知识库 / 商品 / 订单三个 Tab |

## 4. 推荐演示顺序

1. 首页 → 切换中文
2. 路线生成 → 路线详情地图
3. 游记列表（确认图片全部显示）→ 游记详情
4. AI 导游对话 → 方案确认
5. 商家工作台：生成文案 → 海报 → 菜单 OCR
6. 商城：加购 → 购物车 → 提交订单 → 我的订单
7. 后台：知识库编辑 / 商品编辑 / 订单状态修改
8. 个人中心与通知

## 5. 演示边界

- 商城是 **预订单交易**：生成 `pending` 订单、扣库存，到店出示订单号结算；暂不接微信支付/支付宝、物流和退款。
- AI 是真实接入：后端读取 `DASHSCOPE_API_KEY`，调用阿里云百炼 `qwen-plus`、`qwen3-vl-plus` 与通义万相；`/healthz` 返回 `"ai":true`。营销文案、菜单 OCR、AI 海报均已实测，海报异步生成通常 10–30 秒。
- 高德地图需要 `web/.env.local` 中的 `VITE_AMAP_JS_KEY` 与安全密钥；后端路线几何需要 `AMAP_WEB_KEY`。
- 小程序真机需要已备案 HTTPS 域名、微信合法域名和 `WX_APPID` / `WX_SECRET`，本地 `127.0.0.1` 只能用于开发者工具模拟器。
- 多实例部署时 Agent 只会在一台实例启用（MySQL advisory lock），其余实例 `/healthz` 返回 `agent:false`。

## 6. 关键接口

- 认证：`POST /api/auth/login`、`GET /api/me`
- 商城：`GET /api/shop/products`、`GET /api/shop/products/:id/reviews`、`GET/POST/PUT/DELETE /api/cart`、`POST /api/shop/orders`、`GET /api/shop/orders`
- 活动：`GET /api/activities`、`GET /api/activities/:id`
- 商家 AI：`POST /api/ai/copy`、`POST /api/ai/poster`、`GET /api/ai/poster/:taskId`、`POST /api/ai/menu-ocr`、`POST /api/ai/menu-translate`、`POST /api/ai/dish-explain`
- 后台：`/api/admin/knowledge`、`/api/admin/products`、`/api/admin/orders`（仅 admin）
- 上传：`POST /api/upload?kind=blog|avatar`；图片通过 `/uploads/...` 访问
