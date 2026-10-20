# 数据库表结构文档 (Database Schema)

本文档记录了项目当前使用的云数据库集合（Collection）结构及示例数据。所有集合均包含 `_id`（自动生成）、`createdAt`（创建时间）和 `updatedAt`（更新时间）字段。

## 1. 用户表 (users)
存储用户信息，用于登录与鉴权。

| 字段名 | 类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| **_id** | String | 唯一标识符 (OpenID 或自动生成) | `W5v...` |
| **nickName** | String | 用户昵称 | `Traveler A` |
| **avatarUrl** | String | 头像链接 | `https://i.pravatar.cc/150?img=1` |
| **openid** | String | 微信 OpenID (登录时写入) | `oW...` |

---

## 2. 景点轮播图表 (scenic_images)
用于首页轮播图展示及景点基础信息。

| 字段名 | 类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| **image_url** | String | 图片链接 | `https://...` |
| **title_cn** | String | 中文标题 | `越秀山晨景` |
| **title_en** | String | 英文标题 | `Morning at Yuexiu Hill` |
| **location** | String | 地点 | `Guangzhou` |
| **tag** | String | 标签 | `nature` |
| **description_cn** | String | 中文描述 | `清晨的越秀山空气清新...` |
| **description_en** | String | 英文描述 | `Fresh morning air...` |

---

## 3. 景点详情表 (places)
存储景点、商圈或地标的详细地理位置与属性信息。

| 字段名 | 类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| **name_cn** | String | 中文名称 | `越秀公园` |
| **name_en** | String | 英文名称 | `Yuexiu Park` |
| **lat** | Number | 纬度 | `23.1387` |
| **lng** | Number | 经度 | `113.2644` |
| **tags** | Array | 标签列表 | `['park', 'history']` |
| **indoor** | Boolean | 是否室内 | `false` |
| **duration** | Number | 建议游玩时长(分钟) | `120` |
| **description_cn** | String | 中文描述 | `广州最大的综合性公园之一` |
| **description_en** | String | 英文描述 | `One of the largest...` |

---

## 4. 文化知识库表 (knowledge)
用于 AI 导游 (RAG) 检索及文化展示。

| 字段名 | 类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| **title_cn** | String | 中文标题 | `醒狮文化` |
| **title_en** | String | 英文标题 | `Lion Dance` |
| **content_cn** | String | 中文内容 | `醒狮是岭南传统民俗...` |
| **content_en** | String | 英文内容 | `Lion dance is a Lingnan...` |
| **tags** | Array | 关联标签 | `['culture', 'festival']` |

---

## 5. 聊天记录表 (chat_logs)
存储商家 AI 助手或用户与 AI 导游的对话历史。

| 字段名 | 类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| **merchant_id** | String | 商家ID | `merchant_demo_1` |
| **user_id** | String | 用户ID | `user_demo_1` |
| **message_list** | Array | 消息列表 | `[{ from: 'user', text: 'Hello', ts: 16... }]` |

**message_list 结构:**
- `from`: 发送者 ('user' | 'bot')
- `text`: 消息内容
- `ts`: 时间戳

---

## 6. 智慧游记表 (travel_blogs)
存储用户发布的游记内容。

| 字段名 | 类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| **title** | String | 游记标题 | `A Day in Yuexiu Park` |
| **content** | String | 正文内容 | `The park is huge...` |
| **images** | Array | 图片链接列表 | `['https://...']` |
| **author** | Object | 作者简要信息 | `{ nickName: '...', avatarUrl: '...' }` |
| **likes** | Number | 点赞数 | `12` |

---

## 7. 路线规划表 (routes) - New
存储 AI 实时生成的个性化路线数据。

| 字段名 | 类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| **_openid** | String | 用户OpenID | `oW...` |
| **routeName** | String | 路线名称 | `广州历史美食探索之旅` |
| **duration** | String | 行程时长 | `3 Days` |
| **totalDistance** | String | 总距离 | `25km` |
| **waypoints** | Array | 途经点列表 | `[{name: '...', lat: 23..., stayTime: '2h'}]` |
| **polyline** | Array | 路线坐标串 | `['23.1,113.2', ...]` |
| **originalRequest** | Object | 原始请求参数 | `{interests: ['history'], mode: 'driving'}` |

---

## 8. 用户路线表 (user_routes) - Deprecated
(旧版) 存储用户生成或收藏的个性化游玩路线。

| 字段名 | 类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| **days** | Number | 行程天数 | `1` |
| **tags** | Array | 偏好标签 | `['history', 'food']` |
| **route_nodes** | Array | 路线节点列表 | `[{ name: 'Chen Clan...', time: '09:00' }]` |

**route_nodes 结构:**
- `name`: 地点名称
- `time`: 建议时间

---

## 9. 海报生成记录表 (poster_generations)
存储用户生成海报的历史记录及云存储地址。

| 字段名 | 类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| **openid** | String | 微信 OpenID | `oW...` |
| **userId** | String | 用户表 ID | `W5v...` |
| **nickName** | String | 用户昵称 | `Traveler A` |
| **avatarUrl** | String | 头像链接 | `cloud://...` |
| **prompt** | String | 海报提示词 | `夏日饮品宣传海报...` |
| **imageUrl** | String | 云存储图片 fileID | `cloud://...` |
| **style** | String | 风格 (Default/Cartoon/Realistic) | `Cartoon` |
| **originalUrl** | String | 模型原始图片 URL | `https://...` |
| **size** | String | 图片尺寸 | `1280*1280` |
| **model** | String | 模型名称 | `wan2.6-t2i` |
| **taskId** | String | 异步任务 ID | `0385dc79-...` |
| **imageCount** | Number | 生成张数 | `1` |

---

## 10. OCR 记录表 (ocr_records)
存储每次 OCR 识别的原始文本结果。

| 字段名 | 类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| **imageUrl** | String | 图片 URL 或 fileID | `cloud://...` |
| **rawText** | String | OCR 原文 | `菜单内容...` |

---

## 11. 关键词集合 (keywords)
存储 OCR 文本的关键词提取结果。

| 字段名 | 类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| **sourceTextSnippet** | String | 源文本片段 | `菜单前100字符...` |
| **keywords** | Array | 关键词列表 | `[{word, category, importance}]` |

---

## 12. 文化卡片集合 (culture_cards)
存储关键词生成的文化卡片内容。

| 字段名 | 类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| **_openid** | String | 用户 OpenID | `oW...` |
| **keyword** | String | 关键词 | `双皮奶` |
| **cardData** | Object | 卡片内容 | `{ title, description, cultural_significance, category, tags }` |

---

## 13. 词条解释缓存表 (word_content)
存储菜单词条解释的缓存数据。

| 字段名 | 类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| **word** | String | 词条 | `dim sum` |
| **lang** | String | 语言 | `en` |
| **data** | Object | 解释内容 | `{ title, definition, culture_note, quote, tags }` |
| **view_count** | Number | 查看次数 | `3` |

---

## 14. 模型配置表 (model_config)
存储模型切换与参数配置。

| 字段名 | 类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| **configKey** | String | 配置键 | `ocr_model` |
| **configValue** | String | 配置值 | `qwen-vl-ocr-latest` |
| **scope** | String | 作用域 | `ocr` |

---

## 15. AI 对话历史表 (ai_bot_chat_history_5hobd2b)
用于 AI 对话模块的历史记录存储。

| 字段名 | 类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| **sessionId** | String | 会话 ID | `sess_...` |
| **messages** | Array | 消息列表 | `[{ role, content, ts }]` |

---

## 16. AI 对话关系表 (conversation_relation_5hobd2b)
用于 AI 对话的会话关系存储。

| 字段名 | 类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| **sessionId** | String | 会话 ID | `sess_...` |
| **relations** | Array | 关系列表 | `[{ from, to }]` |

---

## 17. 广州风景集合 (GuangZhou_lanscape)
用于存储广州风景相关数据，目前集合为空，字段待补充。
