# Travel Guangdong - Website (前后端分离)

原微信小程序（[miniprogram](miniprogram) / [cloudfunctions](cloudfunctions)）保留不动。网站部分把游客主流程拆成了三个服务，接口合同见 [docs/api/openapi.yaml](docs/api/openapi.yaml)。

```
web/          React + Vite + Ant Design + 高德 JS API（游客网站）
server-go/    Gin + sqlx（唯一运行后端，端口 8081；Web 与小程序共用）
server-java/  Spring Boot 3 + MyBatis + JWT（已冻结的学习对照版，端口 8080，仅覆盖早期接口）
sql/          MySQL schema 与种子数据
nginx/        网站托管 + /api、/uploads 反向代理
```

学习对照文档：[server-go/JAVA_VS_GO.md](server-go/JAVA_VS_GO.md)。

## 快速启动（Docker Compose）

```powershell
Copy-Item .env.example .env   # 填入 DASHSCOPE_API_KEY 等真实值
docker compose up -d --build
# 网站入口 http://localhost （Nginx 托管前端并把 /api 与 /uploads 转给 Go 后端）
```

MySQL 首次启动会自动执行 `sql/schema.sql` 和 `sql/seed.sql`。
种子账号：`demo` / `demo123`。

## 生产部署（Azure 单机）

已部署到 `http://20.194.43.11:8091`：Go 后端（仅回环 `:18080`）+ nginx 独立端口 `:8091`，
复用宿主机 MySQL 但使用独立 schema `travel_gd` 与专属用户，Redis 未使用。
与同机既有服务共用 nginx，互不影响。部署、更新与回滚见 [deploy/README.md](deploy/README.md)。

## 本地开发（不用 Docker）

1. 启动本机 MySQL 并执行 `sql/schema.sql`、`sql/seed.sql`。
2. 后端（Go，唯一运行后端）：

   ```powershell
   # 在仓库根目录执行；脚本会加载 .env、创建 .local-uploads 并启动 :8081
   powershell -ExecutionPolicy Bypass -File scripts\dev-backend.ps1
   ```

   Java 已冻结，仅作为学习对照保留，不再接入网站。

3. 前端：

   ```powershell
   cd web; npm install; npm run dev              # http://localhost:5173
   ```

   Vite 开发代理默认转发到 8081；`/uploads` 也会代理到 Go 后端，上传图片可直接显示。

## 功能范围（本期）

- 首页：轮播、快捷入口、今日推荐
- 路线：AI 生成个性化行程（Qwen），高德地图展示 polyline
- 智慧游记：浏览、发布、点赞、AI 起草
- AI 导游：多轮对话 + Agent 工具调用（含高德天气）+ 确定性行程求解 + SSE 流式 + 方案卡片确认落库
  （Go 后端；设计与压测见 [docs/agent/README.md](docs/agent/README.md)）
- 个人中心：注册 / 登录（JWT + bcrypt）、收藏、统计、中英切换
- 商家 AI 工作台：营销文案、AI 海报（万相异步轮询）、菜单 OCR / 翻译 / 文化卡片、商家客服
- 文旅活动：美食节、非遗手作、珠江夜游、摄影漫步、粤剧体验等 5 场活动
- 商城 MVP：8 个商品，含评分、已售、精选评价、购物车、预订单、库存扣减、我的订单；到店结算，不接真实支付
- 运营后台：知识库 CRUD、商品 CRUD、订单状态管理；仅 `admin` 角色可访问
- 社区内容：8 篇中英文游记、19 条真实评论、关注 / 点赞 / 通知
- 上传：博客配图 / 头像，开发环境通过 Vite `/uploads` 代理显示；种子景点、商品、活动图片位于 `web/public/seed/`

演示账号：

- 游客：`demo` / `demo123`
- 商家：`merchant` / `merchant123`
- 管理员：`admin` / `admin123`

商城支付、物流、优惠券核销、商家入驻与多商户结算仍未实现，属于后续期。
