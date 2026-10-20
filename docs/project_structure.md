# 南粤文化入境游 AI 智能伴游系统 - 项目结构文档

## 1. 项目概述

本项目是一个基于微信小程序开发的南粤文化入境游 AI 智能伴游系统。旨在为入境外国游客提供无障碍的旅游体验，并赋能本地商家。项目采用了 **TDesign** 组件库进行 UI 开发，支持多语言、AI 问答、路线规划等功能。

## 2. 目录结构说明

```
d:\NewWsDemo
├── cloudfunctions/             # 云函数目录
│   ├── initDB/                 # 数据库初始化脚本
│   ├── login/                  # 用户登录获取 OpenID
│   ├── seedData/               # 测试数据注入脚本
│   ├── generateCopy/           # AI 文案生成 (模拟，已被 aiService 替代)
│   └── aiService/              # AI 聚合服务 (Chat/Copy/Guide: Qwen-Turbo, Route: Qwen-Plus, Poster: Wanxiang V2.6, Menu: Qwen-MT-Image, OCR: Qwen-OCR, 关键词: Qwen-Plus, 文化卡片: Qwen-Plus)
├── docs/                       # 文档目录
│   ├── prd.md                  # 产品需求文档
│   ├── Smart_Travel_System_Design.md # 智能伴游系统设计文档 (动态版)
│   ├── cloudbase_ai_rules.md   # CloudBase AI 开发规范指南
│   └── cloudbase_mcp_guide.md  # CloudBase MCP 工具指南
├── miniprogram/                # 小程序源码目录
│   ├── custom-tab-bar/         # 自定义底部导航栏组件 (支持 i18n)
│   ├── pages/                  # 页面目录
│   │   ├── admin/              # 管理员端页面
│   │   │   └── knowledge/  # 知识库管理
│   │   ├── index/          # 首页
│   │   ├── merchant/       # 商家端页面
│   │   │   ├── index/      # 商家助手主页
│   │   │   ├── generate/   # AI 文案生成
│   │   │   ├── chatbot/    # AI 客服助手
│   │   │   └── menuScan/   # 菜单拍照翻译
│   │   ├── user/           # 用户端页面
│   │   │   ├── guide/      # AI 导游 (RAG 问答)
│   │   │   ├── map/        # 路线地图展示
│   │   │   ├── profile/    # 个人中心
│   │   │   ├── routeInput/ # 路线规划输入
│   │   │   └── travelBlog/ # 智慧游记
│   │   └── logs/           # 日志页面 (模板自带)
│   ├── utils/              # 工具函数目录
│   │   ├── i18n.ts         # 多语言字典与管理工具
│   ├── app.json            # 全局配置 (页面路径、窗口表现、TabBar、组件声明)
│   ├── app.ts              # 小程序入口逻辑
│   └── app.wxss            # 全局样式
├── prototype/              # HTML 原型演示目录 (用于设计验证)
├── typings/                # TypeScript 类型定义
├── project.config.json     # 微信开发者工具项目配置
└── package.json            # NPM 依赖配置 (主要依赖 tdesign-miniprogram)
```

## 3. 核心模块功能

### 3.1 用户端 (User)
- **首页 (`pages/index`)**: 展示轮播图、功能入口网格、今日推荐。
- **路线规划 (`pages/user/routeInput` & `map`)**:
  - 输入：选择天数、出行方式(驾车/步行/公交/骑行)、兴趣标签、游玩节奏。
  - 处理：调用 AI 实时生成结构化行程，并结合腾讯地图 API 计算路线。
  - 输出：在地图上展示规划路线、景点标记及时间轴详情。
- **智慧游记 (`pages/user/travelBlog`)**: 社区游记流，包含 "AI 写游记" 功能入口 (AI 生成)。
- **AI 导游 (`pages/user/guide`)**: 基于 RAG 和上下文记忆的 AI 实时问答导游。
- **个人中心 (`pages/user/profile`)**:
  - **用户信息**: 展示头像、昵称，支持头像上传与编辑。
- **数据统计**: 实时展示收藏 (Favorites)、历史路线 (History)、游记 (My Posts) 数量。
- **海报历史**: 读取 `poster_generations` 集合，展示用户生成海报记录与预览。
- **语言切换**: 支持中英文界面切换，全局生效。
- **开发工具**: 内置数据库初始化和数据注入工具，方便调试。
- **其他入口**: 包含商家端/管理端入口。

### 3.2 商家端 (Merchant)
- **商家助手 (`pages/merchant/index`)**: 商家功能聚合页，展示数据概览。
- **文案生成 (`pages/merchant/generate`)**: 
  - 输入：中文原文。
  - 参数：选择目标语言、文案风格、海报风格。
  - 输出：AI 生成多语种营销文案；独立的海报生成功能（支持自定义提示词和历史查看）。
- **AI 客服 (`pages/merchant/chatbot`)**: 模拟 AI 自动回复游客咨询的对话界面。
- **菜单翻译 (`pages/merchant/menuScan`)**: 
  - 实现真实相机实时取景 (`<camera>` 组件)。
  - 集成 **Qwen-MT-Image** 进行图像翻译。
  - 集成 **Qwen-OCR** 进行文字识别，并触发关键词抽取与文化卡片生成。
  - 支持拍照/相册选图、自动翻译、结果预览。
  - 支持保存原图和翻译后的图片至系统相册。

### 3.3 管理端 (Admin)
- **知识库管理 (`pages/admin/knowledge`)**: 对景点、美食、文化等知识条目进行增删改查。

### 3.4 公共组件
- **Custom TabBar (`custom-tab-bar/`)**:
  - 自定义底部导航栏，实现了 5 个 Tab 的布局。
  - Tab 顺序：Route (路线) -> Blog (游记) -> Home (首页/居中) -> Guide (导游) -> Profile (我的)。
  - 首页图标采用了突出的圆形设计。

## 4. 技术栈与依赖

- **框架**: 微信小程序原生框架 (Native)
- **语言**: TypeScript, WXML, WXSS
- **UI 组件库**: `tdesign-miniprogram`
- **样式**: 自定义 WXSS 结合 TDesign 组件样式

## 5. 开发注意事项

1. **依赖安装**:
   项目使用了 npm 包，下载代码后需运行 `npm install`，并在微信开发者工具中执行 "工具 -> 构建 npm"。

2. **TabBar 配置**:
   `app.json` 中的 `tabBar` 配置为 `custom: true`，实际渲染由 `custom-tab-bar` 组件接管，但 `list` 字段仍需保留以保证兼容性。

3. **图片资源**:
   目前的图片资源主要使用 Unsplash 等外链，真机调试时需确保网络通畅。

4. **地图组件**:
   当前地图页面使用图片模拟 (`<image>`)，实际生产环境中应替换为微信原生的 `<map>` 组件并对接地图 SDK。
