# 智能伴游系统 - 动态路线与知识库替代方案

本方案旨在替代原有的静态 `docs/Route` 设计，通过接入云函数和 AI 模型（如通义千问 Qwen 或 DeepSeek），实现**免费且动态**的路线规划与智能导游功能。

## 一、 核心架构变更

| 功能模块 | 原设计 (静态/Mock) | 新方案 (动态 AI) |
| :--- | :--- | :--- |
| **路线规划** | 前端写死或读取固定 JSON | 用户输入偏好 -> 云函数调用 AI 生成 JSON -> 存入数据库 -> 地图渲染 |
| **智能导游** | 预设问答对 (FAQ) | 实时聊天窗口 -> 云函数调用 AI (RAG/Prompt) -> 实时回答 |
| **知识库** | 静态数据库集合 | AI 模型内置知识 + 动态检索 (可选) |

## 二、 技术实现细节

### 1. 云函数: `aiService`

复用了现有的 `aiService` 云函数，扩展了以下能力：

- **Generate Smart Route (`generate_smart_route`)**:
  - **输入**: `duration` (天数), `interests` (兴趣标签), `pace` (节奏), `location` (地点).
  - **处理**: 构建 Prompt 让 AI 生成符合 JSON Schema 的行程单。
  - **输出**: 保存到 `routes` 集合并返回 `routeId`。
  - **模型**: 推荐使用 `qwen-plus` 或 `deepseek-v3` (通过兼容接口) 以保证 JSON 格式的稳定性。

- **AI Guide Chat (`ai_guide_chat`)**:
  - **输入**: `question` (问题), `history` (聊天记录).
  - **处理**: 维护上下文，扮演导游角色进行回答。
  - **输出**: 文本回复。

### 2. 前端页面改造

- **`pages/user/routeInput`**:
  - 新增了标签选择逻辑 (时长、兴趣、节奏)。
  - 点击"生成"调用云函数，获取 `routeId` 后跳转。

- **`pages/user/map`**:
  - 改造为动态读取数据库。
  - 使用 `<map>` 组件渲染 AI 生成的 `polyline` (路线) 和 `markers` (景点)。
  - 底部时间轴动态展示每日行程。

- **`pages/user/guide`**:
  - 实现了真实的聊天界面，对接云函数。

## 三、 配置与免费额度说明

本方案默认兼容 **阿里云百炼 (Aliyun Bailian)** 接口 (OpenAI 兼容格式)，目前提供大量免费额度。

### 1. 获取 API Key
1. 访问 [阿里云百炼控制台](https://bailian.console.aliyun.com/).
2. 开通模型服务 (通义千问 Qwen-Turbo/Plus).
3. 创建 API Key.

### 2. 配置云函数
在微信开发者工具 -> 云开发 -> 云函数 -> `aiService` -> 配置:
- 添加环境变量: `DASHSCOPE_API_KEY` = `你的API_KEY`

> **替代选项**: 如果想使用 DeepSeek，可寻找支持 OpenAI 格式的 DeepSeek 供应商 (如 SiliconFlow)，修改 `aiService/index.js` 中的 `CHAT_URL` 和模型名称即可。

## 四、 数据结构 (Routes Collection)

AI 生成并存入数据库的结构如下：

```json
{
  "_id": "...",
  "_openid": "user_openid",
  "routeName": "广州历史美食三日游",
  "duration": "3 Days",
  "waypoints": [
    {
      "day": 1,
      "name": "陈家祠",
      "latitude": 23.126,
      "longitude": 113.245,
      "description": "岭南建筑艺术的明珠...",
      "stayTime": "2 hours"
    },
    // ...
  ],
  "polyline": ["23.126,113.245", "23.110,113.250"],
  "createdAt": "Timestamp"
}
```

## 五、 后续优化建议
1. **流式输出**: 当前采用同步等待，生成长路线可能需要 10-20 秒。后续可升级为 SSE 流式输出提升体验。
2. **知识库增强**: 可将 `admin/knowledge` 中的数据向量化，在 `ai_guide_chat` 中作为 Context 传入，实现更精准的 RAG。
