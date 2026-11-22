# 小程序演示脚本（截图 / 录屏分镜）

> 目标：用一条完整动线展示「广东智慧文旅」小程序与本地 Go 后端的前后端联调。
> 录屏前按下方「录制前检查」准备好环境，然后按分镜顺序操作。每节末尾的「展示点」用于旁白或字幕。

## 录制前检查（2 分钟）

1. **后端已启动**：终端执行 `scripts/dev-backend.sh`（或 `scripts\dev-backend.ps1`），看到 `travel-gd listening on :8081`。
2. **数据库已启动**：`docker compose ps` 中 `mysql` 为 healthy。
3. **开发者工具**：
   - 打开项目根目录 `D:\code\Travel_Guangdong`；
   - 详情 → 本地设置 → 勾选「不校验合法域名、web-view（业务域名）、TLS 版本以及 HTTPS 证书」（`project.private.config.json` 已设置 `urlCheck: false`）；
   - 设置 → 安全设置 → 开启「服务端口」（自动截图脚本需要）。
4. **清一次登录态**（可选，保证演示从游客开始）：开发者工具 → 清缓存 → 清除数据缓存。
5. 建议窗口布局：开发者工具左侧模拟器，右侧「网络」面板（Network），底部「Console」。录屏时切到 iPhone 14 Pro 像素。

## 分镜 1：首页（0:00–0:20）

| 操作 | 预期结果 | 讲解点 |
| --- | --- | --- |
| 编译并启动 | 首页轮播图显示 5 张广东景点图（来自 `GET /api/scenics`） | 轮播数据来自本地 Go 后端，不是云开发 |
| 切换到「网络」面板，点击下拉刷新（重新进入首页触发 onShow） | 看到 `GET http://127.0.0.1:8081/api/scenics` 200 | 前后端联调的关键证据：请求地址是本地 8081 |

**展示点**：轮播标题随中/英切换；网络面板中的请求域名。

## 分镜 2：登录（0:20–0:50）

| 操作 | 预期结果 | 讲解点 |
| --- | --- | --- |
| 右下角「个人中心」→ 点登录按钮进入登录页 | 登录页展示「登录 / 注册」切换 | 双通道登录设计 |
| 点「微信一键登录」 | Toast「登录成功」并返回上一页 | 本地未配置微信 AppSecret 时，后端用设备 ID 派生稳定游客身份（开发模式），演示中途身份不丢 |
| 网络面板确认 | `POST /api/auth/wechat-login` 200，响应含 token | JWT 由后端签发，后续请求自动带 `Authorization: Bearer` |
| （备选）切到「注册」，用 `demo` / `demo123` 登录 | 登录成功 | 种子账号来自 `sql/seed.sql` |

**展示点**：登录后个人信息页出现昵称/头像；`utils/request.ts` 自动注入 token。

## 分镜 3：AI 路线规划（0:50–1:40）

| 操作 | 预期结果 | 讲解点 |
| --- | --- | --- |
| 底部 Tab「路线」 | 表单页，顶部地图定位到当前位置 | 地图为微信原生 map 组件，无需第三方 SDK |
| 选择「1 天」+「美食」兴趣 +「适中」节奏 +「驾车」，点「生成」 | 按钮 loading，约 10–40 秒后跳转地图页 | 请求 `POST /api/routes/generate`，后端调用 qwen-plus 生成行程 + 高德路径规划 |
| 地图页展示 | 途经点标记 + 绿色路线 + 底部可拖拽行程卡片 | 路线几何来自高德 Web 服务，无 key 时降级为直线 |
| 拖动底部卡片 | 卡片高度在 15vh / 40vh / 90vh 吸附 | 交互细节 |
| 网络面板 | `POST /api/routes/generate`（耗时较长）→ 跳转后 `GET /api/routes/:id` | 生成与查询分离，路线落库 |

**展示点**：路线名称、途经点数量、总距离；后端 `routes` 表中新增记录。

## 分镜 4：AI 导游（1:40–2:10）

| 操作 | 预期结果 | 讲解点 |
| --- | --- | --- |
| 底部 Tab「导游」 | 欢迎语 + 地图 | 欢迎语本地生成，不浪费 token |
| 输入「广州早茶有哪些必点菜？」，发送 | 出现思考气泡，数秒后返回四大天王介绍 | `POST /api/guide/chat`，后端检索知识库后拼入系统提示词 |
| 连续追问一句「哪里能吃到？」 | 记住上下文（后端会话 ID 持久化） | sessionId 保存在页面 data，后端 `guide_sessions` 落库 |

## 分镜 5：AI 游记（2:10–3:00）

| 操作 | 预期结果 | 讲解点 |
| --- | --- | --- |
| 底部 Tab「游记」 | 社区 feed 显示已有游记（来自 `GET /api/blogs`） | 真数据，非硬编码 |
| 点上传区选 1–3 张图，输入主题「周末夜游珠江看灯光秀」，点「AI 起草游记」 | 出现可编辑草稿（标题 + 正文） | `POST /api/blogs/draft`，qwen 生成标题与正文 |
| 修改标题后点「发布游记」 | Toast「已发布」，草稿清空，feed 顶部出现新游记 | 先 `POST /api/upload?kind=blog` 传图，再 `POST /api/blogs` 发布 |
| 网络面板 | 依次看到 upload → blogs 两条请求 | 图片 URL 由后端 `/uploads` 静态路由提供 |

**展示点**：新游记封面图能正常显示（证明 `/uploads` 静态托管 + `fileUrl()` 拼接生效）。

## 分镜 6：个人中心（3:00–3:30）

| 操作 | 预期结果 | 讲解点 |
| --- | --- | --- |
| 底部 Tab「个人中心」 | 昵称、头像、统计（收藏/行程/游记/获赞） | `GET /api/me` + `GET /api/me/stats` + 收藏/路线/游记列表 |
| 点头像 → 选择图片 | 上传后头像更新 | `POST /api/upload?kind=avatar` + `PUT /api/me` |
| 下滑到「我的收藏」「我的游记」 | 显示真实数据 | 收藏与游记均来自本地 MySQL |
| 切换语言（中/英） | 页面文案与列表名称切换 | i18n + 后端中英字段 |

## 分镜 7：商家 AI 文案与海报（3:30–4:30）

| 操作 | 预期结果 | 讲解点 |
| --- | --- | --- |
| 首页「商家中心」→「AI 文案」 | 文案页 | — |
| 输入「粤香茶楼招牌虾饺皇，皮薄馅足」，勾选中文，点「生成文案」 | 数秒后出现营销文案卡片 | `POST /api/ai/copy` |
| 选海报风格「Realistic」，点「生成海报」 | 约 10–30 秒后出现海报图 | `POST /api/ai/poster` 创建异步任务，`GET /api/ai/poster/:taskId` 轮询（DashScope wanx） |
| 点海报图 | 全屏预览 | — |

**展示点**：海报是真实文生图结果（DashScope OSS 地址）。

## 分镜 8：菜单拍照识别 / 翻译 / 文化卡片（4:30–5:30）

| 操作 | 预期结果 | 讲解点 |
| --- | --- | --- |
| 首页「菜单翻译」 | 相机取景框 | 开发者工具里点「相册」选择 `docs/demo-menu.png`（仓库自带测试菜单图） |
| 点「Culture Card」 | 底部抽屉打开，显示识别出的菜品关键词 chips | `POST /api/ai/menu-ocr`（qwen3-vl-plus 视觉模型） |
| 点任一菜品 chip | 弹出文化卡片（介绍 + 标签） | `POST /api/ai/dish-explain kind=card` |
| 长按 chip | 弹出简要讲解卡片 | `POST /api/ai/dish-explain kind=explain` |
| 点「Translate」→ 选 English | 抽屉列出「原名 → 英文译名 + 备注」 | `POST /api/ai/menu-translate`，同步返回 |
| 点「Save」 | 可保存原图到相册 | 本地相册权限 |

**展示点**：六道粤菜全部识别 + 英译正确；网络面板三次 AI 请求。

## 分镜 9：AI 客服（5:30–5:50）

| 操作 | 预期结果 | 讲解点 |
| --- | --- | --- |
| 商家中心 →「AI 客服」 | 对话页 | — |
| 问「你们几点营业？」 | 数秒后回复 | 复用 `POST /api/guide/chat`，独立会话 |

## 录屏结束（5:50–6:00）

- 切到后端终端，展示 `GET /api/scenics` / `POST /api/routes/generate` 等访问日志；
- 可选：展示 MySQL 中 `routes`、`travel_blogs` 新增记录，证明数据落库。

## 自动截图（可选）

```bash
# 1) 确保开发者工具已打开项目且「服务端口」已开启
# 2) 桌面处于解锁状态（锁屏时 Windows 截图只能拿到锁屏画面）
node scripts/mp-screenshots.js
```

截图输出到 `docs/screenshots/`，编号顺序与上表一致。若脚本在 `navigateTo/switchTab` 处提示 ack timeout 属正常（该版本 IDE 不回包但页面已跳转）。

## 常见问题

| 现象 | 处理 |
| --- | --- |
| 请求失败 `request:fail` | 后端未启动或端口不是 8081；重新执行 `scripts/dev-backend.sh` |
| 登录弹「请先登录」 | 清缓存后重新微信一键登录 |
| AI 接口 502 `AI provider not configured` | 根目录 `.env` 缺少 `DASHSCOPE_API_KEY` |
| AI 接口 502 `Model not exist` | 账号未开通该模型；视觉模型默认 `qwen3-vl-plus`，可在 `.env` 设 `AI_VISION_MODEL` 覆盖 |
| 图片不显示 | 确认后端 `UPLOAD_DIR` 可写且 `/uploads` 路由可访问 |
| 地图空白 | 首次进入需允许定位；开发者工具可在「模拟器 → 位置」设置广州 |
