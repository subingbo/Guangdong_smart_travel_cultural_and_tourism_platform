# 微信小程序云开发 AI 智能伴游系统设计文档 (动态版)

## 一、项目概述

本项目基于微信小程序云开发平台，结合腾讯位置服务与 **通义千问/DeepSeek AI 模型**，实现一个全动态的智能 AI 伴游系统。系统不再依赖静态预设数据，而是通过 AI 实时生成个性化路线规划、提供即时导游问答，并自动生成智慧游记。

## 二、技术架构

### 2.1 技术栈
- **前端框架**: 微信小程序原生开发 (WXML/WXSS/TS)
- **地图服务**: 腾讯地图 SDK (用于渲染标记点、路线与定位)
- **后端服务**: 腾讯云开发 CloudBase (云函数、云数据库、云存储)
- **AI 能力**: 通义千问 Qwen-Plus / DeepSeek (通过云函数 `aiService` 调用)
- **存储**: 云开发数据库 (NoSQL)

### 2.2 架构流程
1. **用户输入**: 在前端选择旅行偏好 (时长、兴趣、节奏)。
2. **AI 生成**: 云函数调用大模型，生成符合 JSON Schema 的详细行程单。
3. **数据存储**: 行程单存入云数据库 `routes` 集合。
4. **前端渲染**: 小程序从数据库拉取行程，解析坐标点，在地图上绘制路线与标记。
5. **智能交互**: 用户在旅途中通过聊天界面与 AI 导游实时问答 (RAG/Prompt)。

## 三、数据库设计

### 3.1 核心集合设计

#### 用户信息 (users)
```javascript
{
  _id: '自动生成',
  openid: '微信openid',
  nickname: '用户昵称',
  avatar: '头像URL',
  createdAt: '创建时间'
}
```

#### 路线规划 (routes) - **核心动态数据**
```javascript
{
  _id: '自动生成',
  _openid: '用户openid', // 权限控制
  routeName: 'AI生成的路线名称 (e.g. 广州历史美食探索之旅)',
  duration: '3 Days', // 总时长
  totalDistance: '25km', // 预估总距离
  waypoints: [
    {
      day: 1, // 第几天
      name: '陈家祠', // 地点名称
      latitude: 23.126, // 纬度 (GCJ-02)
      longitude: 113.245, // 经度 (GCJ-02)
      description: '岭南建筑艺术的明珠...', // AI生成的简短介绍
      stayTime: '2 hours', // 建议停留时间
      address: '广州市荔湾区中山七路恩龙里34号' // (可选)
    },
    // ... 更多途经点
  ],
  polyline: [
    "23.126,113.245", // 坐标点串，用于绘制路线轨迹
    "23.127,113.246",
    // ...
  ],
  originalRequest: { // 记录用户当时的请求参数
    interests: ['history', 'food'],
    pace: 'relaxed'
  },
  createdAt: '创建时间'
}
```

#### 游记 (travelBlogs)
```javascript
{
  _id: '自动生成',
  userId: '用户ID',
  routeId: '关联的路线ID',
  title: '游记标题',
  content: 'AI生成的游记内容',
  images: ['图片URL列表'],
  createdAt: '创建时间'
}
```

## 四、核心模块实现

### 4.1 个性化路线定制模块 (`pages/user/routeInput`)

**功能**: 用户选择偏好，触发 AI 生成路线。

#### 页面结构 (WXML)
```xml
<view class="page">
  <!-- 时长选择 -->
  <view class="card">
    <view class="section-title">
      <t-icon name="time" /> {{t.route.duration}}
    </view>
    <view class="tags-container">
      <block wx:for="{{durations}}" wx:key="value">
        <t-tag 
          variant="{{item.selected ? 'dark' : 'light'}}" 
          bind:change="onDurationChange"
          class="tag-item"
        >{{item.label}}</t-tag>
      </block>
    </view>
  </view>

  <!-- 兴趣标签 -->
  <view class="card">
    <view class="section-title">
      <t-icon name="heart" /> {{t.route.interests}}
    </view>
    <view class="tags-container">
      <block wx:for="{{interestTags}}" wx:key="value">
        <t-tag 
          variant="{{item.selected ? 'dark' : 'light'}}" 
          icon="{{item.icon}}"
          bind:change="onInterestChange"
          class="tag-item"
        >{{item.label}}</t-tag>
      </block>
    </view>
  </view>

  <!-- 节奏选择 -->
  <view class="card">
    <view class="section-title">
      <t-icon name="dashboard" /> {{t.route.pace}}
    </view>
    <view class="pace-selector">
      <block wx:for="{{paces}}" wx:key="value">
        <view 
          class="pace-option {{item.selected ? 'active' : ''}}" 
          bindtap="onPaceChange"
        >{{item.label}}</view>
      </block>
    </view>
  </view>

  <!-- 生成按钮 -->
  <view class="btn-area">
    <t-button theme="primary" block loading="{{loading}}" bindtap="onGenerate">
      {{t.route.generate}}
    </t-button>
  </view>
</view>
```

#### 逻辑实现 (TS)
```typescript
async onGenerate() {
  // 1. 收集用户选择
  const duration = this.data.durations.find(d => d.selected)?.value || 1;
  const interests = this.data.interestTags.filter(t => t.selected).map(t => t.value);
  const pace = this.data.paces.find(p => p.selected)?.value || 'balanced';

  // 2. 校验
  if (interests.length === 0) {
    wx.showToast({ title: '请选择至少一个兴趣', icon: 'none' });
    return;
  }

  this.setData({ loading: true });

  try {
    // 3. 调用云函数 aiService
    const res = await wx.cloud.callFunction({
      name: 'aiService',
      data: {
        action: 'generate_smart_route', // 指定动作
        payload: {
          duration,
          interests,
          pace,
          location: 'Guangzhou' // 默认城市，后续可扩展
        }
      }
    });

    const result = res.result as any;
    if (result.success) {
      // 4. 跳转到地图页，传递生成的 routeId
      wx.navigateTo({
        url: `/pages/user/map/index?routeId=${result.routeId}`
      });
    } else {
      throw new Error(result.message);
    }
  } catch (error) {
    wx.showToast({ title: '路线生成失败，请重试', icon: 'none' });
  } finally {
    this.setData({ loading: false });
  }
}
```

---

### 4.2 动态地图路径规划模块 (`pages/user/map`)

**功能**: 从数据库加载 AI 生成的路线数据，并在地图上可视化展示。

#### 页面结构 (WXML)
```xml
<view class="page">
  <!-- 地图组件 -->
  <view class="map-container">
    <map 
      id="routeMap"
      longitude="{{center.longitude}}" 
      latitude="{{center.latitude}}" 
      scale="12" 
      markers="{{markers}}" 
      polyline="{{polyline}}"
      show-location
      style="width: 100%; height: 50vh;"
    ></map>
  </view>

  <!-- 路线标题 -->
  <view class="header">
    <view class="title-card">
      <view class="title">{{routeData.routeName}}</view>
      <view class="subtitle">{{routeData.duration}} • {{routeData.totalDistance}}</view>
    </view>
  </view>

  <!-- 行程时间轴 -->
  <view class="bottom-sheet">
    <scroll-view scroll-y class="timeline-scroll">
      <view class="timeline">
        <block wx:for="{{routeData.waypoints}}" wx:key="name">
          <view class="timeline-item">
            <view class="timeline-icon bg-blue">
              <t-icon name="map" />
            </view>
            <view class="timeline-content">
              <view class="timeline-title">{{item.name}}</view>
              <view class="timeline-time">Day {{item.day}} • 建议停留 {{item.stayTime}}</view>
              <view class="timeline-desc">{{item.description}}</view>
            </view>
          </view>
        </block>
      </view>
    </scroll-view>
  </view>
</view>
```

#### 逻辑实现 (TS)
```typescript
async loadRoute(routeId: string) {
  wx.showLoading({ title: '加载路线中...' });
  try {
    // 1. 从云数据库获取路线详情
    const db = wx.cloud.database();
    const res = await db.collection('routes').doc(routeId).get();
    const route = res.data;

    this.setData({ routeData: route });

    // 2. 渲染地图标记点 (Markers)
    const markers = route.waypoints.map((wp, index) => ({
      id: index,
      latitude: wp.latitude,
      longitude: wp.longitude,
      title: wp.name,
      callout: {
        content: wp.name,
        display: 'ALWAYS',
        padding: 8,
        borderRadius: 4
      }
    }));

    // 3. 渲染路线轨迹 (Polyline)
    // 如果 AI 返回了 polyline 坐标串则直接使用，否则简单连接各个 waypoint
    let points = [];
    if (route.polyline && route.polyline.length > 0) {
      points = route.polyline.map(p => {
        const [lat, lng] = p.split(',').map(Number);
        return { latitude: lat, longitude: lng };
      });
    } else {
      points = route.waypoints.map(wp => ({
        latitude: wp.latitude,
        longitude: wp.longitude
      }));
    }

    this.setData({
      markers,
      polyline: [{
        points,
        color: '#0052D9',
        width: 4,
        arrowLine: true
      }],
      // 设置中心点为第一个景点
      center: {
        latitude: route.waypoints[0].latitude,
        longitude: route.waypoints[0].longitude
      }
    });

  } catch (error) {
    wx.showToast({ title: '加载失败', icon: 'none' });
  } finally {
    wx.hideLoading();
  }
}
```

---

### 4.3 AI 智能导游模块 (`pages/user/guide`)

**功能**: 提供实时 AI 问答，替代静态 FAQ。

#### 页面结构 (WXML)
```xml
<view class="guide-container">
  <!-- 聊天记录区域 -->
  <scroll-view class="chat-area" scroll-y scroll-into-view="{{toView}}">
    <block wx:for="{{messages}}" wx:key="index">
      <view class="message {{item.isUser ? 'user' : 'ai'}}" id="msg-{{index}}">
        <view class="avatar">
          <image src="{{item.isUser ? userAvatar : aiAvatar}}" />
        </view>
        <view class="bubble">
          <text>{{item.text}}</text>
        </view>
      </view>
    </block>
  </scroll-view>
  
  <!-- 输入框 -->
  <view class="input-area">
    <input 
      value="{{inputValue}}" 
      bindinput="onInputChange" 
      bindconfirm="onSend" 
      placeholder="问问导游..." 
    />
    <button bindtap="onSend" disabled="{{loading}}">发送</button>
  </view>
</view>
```

#### 逻辑实现 (TS)
```typescript
async onSend() {
  const text = this.data.inputValue;
  if (!text.trim()) return;
  
  // 1. 添加用户消息到界面
  const newMessages = [...this.data.messages, { text, isUser: true }];
  this.setData({
    messages: newMessages,
    inputValue: '',
    loading: true
  });
  
  // 2. 添加 "Thinking..." 占位符
  this.setData({
    messages: [...newMessages, { text: 'Thinking...', isUser: false, isThinking: true }]
  });

  try {
    // 3. 调用云函数 aiService
    // 提取最近的聊天记录作为上下文
    const history = this.data.messages
      .filter(m => !m.isThinking)
      .slice(-6)
      .map(m => ({
        role: m.isUser ? 'user' : 'assistant',
        content: m.text
      }));

    const res = await wx.cloud.callFunction({
      name: 'aiService',
      data: {
        action: 'ai_guide_chat',
        payload: {
          question: text,
          history: history
        }
      }
    });

    // 4. 更新 AI 回复
    const result = res.result as any;
    const cleanMessages = this.data.messages.filter(m => !m.isThinking);
    
    if (result.success) {
      this.setData({
        messages: [...cleanMessages, { text: result.data.content, isUser: false }]
      });
    }
  } catch (error) {
    // 错误处理...
  } finally {
    this.setData({ loading: false });
  }
}
```

---

### 4.4 智慧游记生成模块 (`pages/user/travelBlog`)

**功能**: 上传图片，结合路线信息，调用 AI 生成游记。

1. **上传图片**: 将用户选择的照片上传至云存储。
2. **获取路线**: 读取当前路线的景点信息。
3. **生成内容**: 调用 `aiService` (action: `generate_copy`)，传入图片上下文或路线信息。
4. **保存游记**: 将生成的标题、内容及图片链接存入 `travelBlogs` 集合。

## 五、环境配置

为了使上述功能正常工作，需在微信开发者工具中配置云函数环境变量：

1. 打开云开发控制台 -> 云函数 -> `aiService` -> 配置。
2. 添加环境变量：
   - `DASHSCOPE_API_KEY`: 您的阿里云百炼 API Key (或兼容 OpenAI 格式的其他 Key)。
   - `TENCENT_MAP_API_KEY`: 腾讯地图 API Key (用于路径规划服务)。
3. 重新部署云函数。

---

## 六、未来规划 (Phase 2): 引入 MCP 协议升级实时决策系统

为了进一步提升系统的智能性与实时性，计划在下一阶段引入 **MCP (Model Context Protocol)** 协议，使 AI 能够主动调用外部工具，获取实时数据并执行复杂决策。

### 6.1 核心升级目标
- **实时数据接入**: AI 可直接调用地图 API 获取实时路况、调用天气 API 获取实时天气。
- **动态决策**: 根据实时天气（如下雨）或路况（拥堵），AI 主动建议调整行程（如推荐室内景点）。
- **外部服务联动**: 联动打车、订票等外部服务（Mock 或真实 API）。

### 6.2 架构调整
- **MCP Server**: 在云函数中部署或集成 MCP Server，作为 AI 与外部工具的中间层。
- **Tool Definitions**: 定义标准的工具接口，如 `get_weather(location)`, `get_traffic(route)`, `search_places(keyword)`.
- **AI Agent**: 升级 `aiService` 中的 Prompt，使其具备 Tool Calling 能力（ReAct 模式）。

### 6.3 预期场景
1. **天气预警**: 用户规划路线时，AI 检测到当日有雨，主动提示“周六预报有雨，建议将越秀公园行程改为广东省博物馆”，并自动更新路线。
2. **动态避堵**: 导游对话中，用户问“现在去广州塔堵车吗？”，AI 调用地图 API 查看路况并回答“目前猎德大桥拥堵，建议乘坐地铁 3 号线前往”。

---

**注**: 本文档替代原有的 `docs/Route` 设计，全面采用动态 AI 生成方案，实现了更灵活、更智能的用户体验，同时利用免费的 AI 模型额度降低了运营成本。
