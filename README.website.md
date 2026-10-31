# Travel Guangdong - Website (前后端分离)

原微信小程序（[miniprogram](miniprogram) / [cloudfunctions](cloudfunctions)）保留不动。网站部分把游客主流程拆成了三个服务，接口合同见 [docs/api/openapi.yaml](docs/api/openapi.yaml)。

```
web/          React + Vite + Ant Design + Leaflet（游客网站）
server-java/  Spring Boot 3 + MyBatis + JWT（正式后端，端口 8080）
server-go/    Gin + sqlx（Go 学习对照版，端口 8081，接口完全相同）
sql/          MySQL schema 与种子数据
nginx/        网站托管 + /api 反向代理
```

学习对照文档：[server-go/JAVA_VS_GO.md](server-go/JAVA_VS_GO.md)。

## 快速启动（Docker Compose）

```powershell
Copy-Item .env.example .env   # 填入 DASHSCOPE_API_KEY 等真实值
docker compose up -d --build
# 网站入口 http://localhost （Nginx 托管前端并把 /api 转给 Java 后端）
```

MySQL 首次启动会自动执行 `sql/schema.sql` 和 `sql/seed.sql`。
种子账号：`demo` / `demo123`。

## 生产部署（Azure 单机）

已部署到 `http://20.194.43.11:8091`：Go 后端（仅回环 `:18080`）+ nginx 独立端口 `:8091`，
复用宿主机 MySQL 但使用独立 schema `travel_gd` 与专属用户，Redis 未使用。
与同机既有服务共用 nginx，互不影响。部署、更新与回滚见 [deploy/README.md](deploy/README.md)。

## 本地开发（不用 Docker）

1. 启动本机 MySQL 并执行 `sql/schema.sql`、`sql/seed.sql`。
2. 后端（二选一或都起）：

   ```powershell
   # Java
   cd server-java; mvn spring-boot:run          # :8080
   # Go
   cd server-go; go run .                        # :8081
   ```

3. 前端：

   ```powershell
   cd web; npm install; npm run dev              # http://localhost:5173
   ```

   Vite 开发代理默认转发到 8080；改 `VITE_API_BASE_URL` 可指向 Go 后端。

## 功能范围（本期）

- 首页：轮播、快捷入口、今日推荐
- 路线：AI 生成个性化行程（Qwen），Leaflet 地图展示 polyline
- 智慧游记：浏览、发布、点赞、AI 起草
- AI 导游：多轮对话，检索知识库拼入提示词（轻量 RAG）
- 个人中心：注册 / 登录（JWT + bcrypt）、收藏、统计、中英切换

未迁移（后续期）：商家端、商城、菜单 OCR、管理后台。
