# 南粤文化入境游数字化生态系统  
## 微信小程序（云开发）完整产品需求文档 PRD

---

# 一、项目概述

## 1.1 项目名称
南粤文化入境游 AI 智能伴游系统

## 1.2 项目定位

本项目是一个基于微信云开发的微信小程序，面向：

- 入境外国游客（需求侧）
- 本地文旅商家（供给侧）

通过 AI 技术解决语言障碍、文化理解障碍与内容生产难题，构建“游览-理解-分享-消费”闭环。

---

# 二、系统架构说明

## 技术结构

- 前端：微信小程序
- 后端：微信云开发
  - 云函数
  - 云数据库
  - 云存储
- 第三方能力：
  - 大模型 API（文案生成 / 路线规划 / 问答 / 生图 - 阿里云百炼 Qwen/Wanxiang）
  - OCR 识别
  - 地图服务
  - 天气接口
  - 翻译接口（腾讯云翻译/阿里云翻译，或大模型翻译）

## 2.3 国际化 (Internationalization)

本项目支持全站中英文（及其他扩展语言）即时切换。

- **核心实现**: 
  - `utils/i18n.ts` 集中管理多语言字典。
  - `wx.getStorageSync('language')` 持久化存储用户语言偏好。
  - 页面 `onShow` 生命周期钩子中初始化语言设置。
- **覆盖范围**:
  - 底部导航栏 (TabBar)
  - 首页、个人中心、商家中心
  - AI 导游对话、AI 文案生成页面
  - 静态文案及部分动态数据（如时间显示）


---

# 三、整体页面结构设计

## 3.1 小程序整体结构

采用 **底部 TabBar 导航结构**。

### 底部导航栏设计（5个Tab）

1. 伴游路线 (Route)
2. 智慧游记 (Blog)
3. 首页 (Home) - 居中突出
4. AI 导游 (Guide)
5. 个人中心 (Profile)

---

## 3.2 底部导航设计规范

- 首页位于正中间
- 首页图标尺寸大于其他图标
- 首页采用突出圆形按钮设计
- 选中状态高亮
- 未选中状态灰色
- 所有 Tab 支持状态保持

---

## 3.3 线上商城 (Shop)

### 页面路径
pages/shop/index

### 入口建议

- 首页快捷入口（与路线/游记/导游并列）
- 个人中心快捷入口（“优惠券/订单/购物车”方向）
- 景点详情/攻略内容页中的“购票/购买特产”按钮

### 页面结构（自上而下）

- 顶部搜索：t-input + t-button
- Banner：Swiper + t-image
- 分类入口：t-grid（门票/特产/文创/优惠券）
- 推荐区：并排 2 张推荐卡片（一个文创 + 一个盲盒），包含图片、标题、价格、标签、加购按钮
- 底部结算栏：合计/件数 + 去结算

### 设计约束

- 组件统一使用 TDesign（t-input、t-button、t-card、t-tag、t-grid、t-cell-group 等）
- 预览与截图使用：`docs/all_pages_preview.html`

---

# 四、首页设计

## 页面路径
pages/index/index

---

## 4.1 首页功能定位

- 展示南粤风景名胜
- 提供视觉沉浸感
- 提供功能快捷入口
- 作为系统主入口

---

## 4.2 页面结构设计（自上而下）

---

### ① 顶部欢迎区域

内容：

- 主标题：Welcome to Guangdong
- 副标题：Explore Lingnan Culture with AI
- 支持中英文切换按钮

---

### ② 南粤风景轮播图区域（核心视觉）

组件：Swiper 轮播组件

数据来源：

- 云数据库 collection：scenic_images
- 图片存储在云存储中

字段包括：

- image_url
- title_cn
- title_en
- location
- tag
- description_cn
- description_en

---

### 轮播规则

- 自动循环播放
- 每3秒自动切换
- 支持手动滑动
- 显示底部圆点指示器
- 支持点击跳转

---

### 点击轮播图交互

1. 获取当前图片对应景点信息
2. 跳转至景点详情页
3. 展示文化介绍 + 推荐理由

---

### ③ 快捷功能入口区

布局：2行2列网格

入口包括：

- 个性化路线
- 菜单翻译
- 智慧游记
- AI 导游

点击跳转对应页面。

---

### ④ 今日推荐区域

展示：

- 随机推荐1-2个景点
- 显示图片
- 简短英文介绍
- 点击跳转详情页

---

## 4.3 首页加载逻辑

1. 页面加载
2. 从数据库获取 scenic_images 数据
3. 从云存储获取图片
4. 渲染轮播图
5. 异常时显示默认图片

---

# 五、功能模块设计

---

# 1. 商家 AI 助手模块

---

## 1.1 多语种内容生产

### 页面路径
pages/merchant/generate/index

---

### 页面结构设计

① 页面标题区  
② 中文输入区（textarea，500字限制）  
③ 目标语言选择区（多选：English, Japanese, Korean）  
④ 文案风格选择区（单选：Instagram, Twitter, Facebook）  
⑤ 生成文案按钮（主按钮，居中，调用云函数 `aiService` action: `generate_copy`）  
⑥ 结果展示区（默认隐藏，展示多语言结果）  
⑦ 海报生成区（常驻显示）：
   - 海报风格选择（单选：Default, Cartoon, Realistic, Cyberpunk）
   - 海报提示词调整区（可编辑 Prompt，保留历史修改）
   - 生成海报按钮（调用 `aiService` action: `generate_poster`）
⑧ 海报结果展示区（默认隐藏，生成后展示在独立卡片中）

### 功能逻辑

1. **输入**: 商家输入中文原文（如菜品介绍）。
2. **选择参数**: 
   - 目标语言（如：英语、日语、韩语）
   - 文案风格（如：社交媒体风、正式风）
3. **生成文案**: 调用阿里云 Qwen 模型生成多语言文案。
4. **生成海报**: 
   - 用户选择海报风格或输入自定义提示词。
   - 用户点击"生成海报"按钮（无需等待文案生成）。
   - 调用 `aiService` (action: `generate_poster`) 提交异步任务。
   - 前端轮询任务状态直至完成。
   - 阿里云 Wan2.6-t2i 模型生成图片并保存至云存储。
   - 记录生成历史至数据库 `poster_generations` 表。
5. **历史查询**: 用户可在个人中心查看所有生成的海报记录。

---

### 交互逻辑

1. 输入中文内容  
2. 选择语言与风格  
3. 点击生成文案 -> 展示文案结果  
4. 选择海报风格或输入提示词  
5. 点击生成海报 -> 显示loading -> 展示海报结果  

---

## 1.2 AI Chatbot 商家问答助手

### 页面路径
pages/merchant/chatbot/index

---

### 页面结构

① 顶部商家信息区  
② 聊天记录滚动区  
③ 常见问题快捷按钮区  
④ 底部输入区（语言选择 + 输入框 + 发送按钮）

---

### 交互逻辑

1. 输入问题  
2. 选择语言  
3. 点击发送  
4. 显示用户消息  
5. 获取AI回复 (调用云函数 `aiService` - Qwen-Turbo)  
6. 自动滚动到底部  

---

## 1.3 菜单拍照翻译 + 文化解释

### 页面路径
pages/merchant/menuScan/index

---

### 页面结构

① 图片上传区（拍照/相册）  
② OCR识别结果区  
③ 关键词抽取与词云区  
④ 文化卡片展示区  
⑤ 翻译语言选择区  
⑥ 翻译按钮  

---

### 交互流程

1. 上传图片  
2. OCR 识别文字  
3. 展示原文  
4. 提取关键词并展示词云  
5. 点击关键词生成文化卡片  
6. 选择语言  
7. 点击翻译  
8. 展示翻译与文化解释  

---

# 2. 智能伴游模块

---

## 2.1 个性化路线定制 (Dynamic Route Planning)

### 页面路径
pages/user/routeInput/index

---

### 页面结构

① 时长选择区 (1天/2天/3天/5+天)  
② 出行方式选择区 (驾车/步行/骑行/公交)  
③ 兴趣标签区 (美食/历史/自然/艺术/购物)  
④ 游玩节奏区 (休闲/适中/紧凑)  
⑤ 生成路线按钮  

---

### 交互逻辑

1. 选择时长、出行方式、兴趣、节奏  
2. 点击生成按钮  
3. 调用云函数 `aiService` (action: `generate_smart_route`)，传入 `mode` 参数  
4. AI 根据偏好实时生成结构化行程单  
5. 获取 `routeId` 并跳转至地图页面  

---

## 2.2 地图路径规划页面

### 页面路径
pages/user/map/index

---

### 页面结构

① 全屏地图 (展示 AI 生成的路线 Polyline 和 Marker)  
② 路线信息卡片 (标题/总时长/总距离)  
③ 底部行程时间轴 (每日景点/建议停留时间)  

---

### 交互逻辑

- 自动缩放地图以展示完整路线  
- 点击 Marker 显示景点详情气泡  
- 滑动底部时间轴查看具体行程  
- (未来规划) 支持实时天气预警与路线动态调整  

---

## 2.3 智慧游记生成

### 页面路径
pages/user/travelBlog/index

---

### 页面结构

① 图片上传区  
② AI 生成按钮  
③ 游记预览卡片  
④ 分享操作区  

---

### 交互流程

1. 上传图片  
2. 点击生成  
3. 调用 `aiService` (action: `generate_copy`)  
4. AI 根据图片和当前路线信息生成生动游记  
5. 展示结果并支持保存/分享  

---

## 2.4 AI 导游 (AI Guide Chat)

### 页面路径
pages/user/guide/index

---

### 页面结构

① 聊天记录滚动区  
② 底部输入框  
③ 快捷提问词  

---

### 交互逻辑

1. 用户输入问题 (语音/文字)  
2. 系统提取最近聊天记录作为上下文  
3. 调用 `aiService` (action: `ai_guide_chat`)  
4. AI 扮演导游角色，结合当前景点信息进行回复  
5. 支持多轮对话与上下文记忆  

---

## 2.5 个人中心

### 页面路径
pages/user/profile/index

---

### 页面结构

① 顶部用户信息区（头像、昵称、标签）  
② 核心数据统计区（收藏、历史、游记、优惠券）  
③ 功能菜单列表（我的计划、语言切换、商家入口、管理入口等）  
④ 底部操作区（退出登录、开发调试工具）  

---

### 交互逻辑

1. **用户登录/注册**：点击头像或按钮进行登录，自动获取微信 OpenID 并创建/更新用户数据。
2. **头像编辑**：登录状态下点击头像旁边的编辑按钮，支持上传图片至云存储 `/avatars` 目录并更新个人资料。
3. **数据统计**：登录后自动加载并显示用户的收藏数、路线历史数、游记发布数。
4. **海报历史**：根据用户 OpenID 查询 `poster_generations`，展示生成历史与预览入口。
5. **语言切换**：点击语言切换选项，支持中英文切换，即时更新整个小程序界面语言。
6. **开发工具**：提供数据库初始化 (initDB) 和测试数据注入 (seedData) 的快捷入口。

---

# 3. 线上商城模块

---

## 3.1 线上商城首页 (Shop Home)

### 页面路径
pages/shop/index

---

### 页面结构

① 顶部搜索区（t-input + 购物车入口 t-button）  
② 活动 Banner 区（t-swiper / t-image）  
③ 商品分类区（t-grid / t-grid-item）  
④ 热销商品区（t-card 商品卡片 + t-button “加购”）  
⑤ 底部结算条（t-card + t-button “去结算”）  

---

### 交互逻辑

1. 仅做页面与布局展示，用于方案评审与策划书截图
2. 后续接入：商品列表、购物车、下单支付、优惠券与订单列表

---

# 六、后台知识库管理模块

---

## 页面路径
pages/admin/knowledge/index

---

### 页面结构

① 顶部操作区（上传Excel / 新增条目）  
② 数据表格区  
③ 编辑 / 删除按钮  

---

### 字段包括

- 中文名称  
- 英文名称  
- 分类标签  
- 文化解释（中英）  
- 是否室内  
- 建议游览时间  

---

### 交互逻辑

- 支持增删改查  
- 支持批量导入  
- 数据实时更新  

---

# 七、数据库结构设计

---

## collection：scenic_images

- image_url  
- title_cn  
- title_en  
- location  
- tag  
- description_cn  
- description_en  

---

## collection：places

- name_cn  
- name_en  
- lat  
- lng  
- tags  
- indoor  
- duration  
- description_cn  
- description_en  

---

## collection：knowledge

- title_cn  
- title_en  
- content_cn  
- content_en  
- tags  

---

## collection：chat_logs

- merchant_id  
- user_id  
- message_list  

---

## collection：users

- openid (唯一标识)
- nickName
- avatarUrl
- createdAt
- updatedAt

---

## collection：favorites

- _openid (系统自动添加)
- type (target type: place/route/blog)
- targetId
- title
- image
- createdAt

---

## collection：routes (New)

- _openid  
- routeName  
- duration  
- totalDistance  
- waypoints (Array: {name, lat, lng, description, stayTime})  
- polyline (Array: ["lat,lng"])  
- originalRequest (Object)  
- createdAt  

---

## collection：user_routes (Deprecated, use 'routes')

- _openid (系统自动添加)
- title
- days
- spots (Array)
- createdAt

---

## collection：travel_blogs

- _openid (系统自动添加)
- title
- content
- images (Array)
- createdAt

---

## collection：poster_generations

- openid
- userId
- nickName
- avatarUrl
- prompt
- imageUrl
- originalUrl
- size
- model
- taskId
- imageCount
- createdAt

---

# 八、非功能需求

- 页面响应时间 ≤ 3 秒  
- 首页图片加载 ≤ 2 秒  
- 所有请求必须有 loading 状态  
- 错误必须提示  
- 支持中英文界面切换  
- 地图操作流畅  
- 权限分级控制  

---

# 九、交付成果

- 可运行微信小程序  
- 云数据库结构  
- 500-1000条双语知识库  
- 演示视频  
- 商家合作证明  
- 用户访问截图  
- 推广账号截图  

---

# 十、验收标准

- 首页轮播图正常循环  
- 图片来自云存储  
- 首页按钮居中突出  
- 所有模块可跳转  
- 路线规划可演示  
- AI问答准确  
- 页面交互流畅  

---

（完）
