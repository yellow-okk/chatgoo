# Chatgoo 前端开发文档

> 本文档面向前端开发者，目标是描述 Chatgoo 在线聊天应用前端需要实现的**功能、页面、接口对接方式与数据结构**。
> 文档刻意**脱离具体 UI 框架**（不绑定 React / Vue / Angular / Svelte 等），开发者可基于本文档选用任意前端技术栈进行实现。

---

## 目录

1. [项目概述](#1-项目概述)
2. [技术栈建议（框架无关）](#2-技术栈建议框架无关)
3. [后端服务基础信息](#3-后端服务基础信息)
4. [统一响应与错误处理](#4-统一响应与错误处理)
5. [认证机制与 Token 管理](#5-认证机制与-token-管理)
6. [数据结构定义](#6-数据结构定义)
7. [通用基础设施实现要求](#7-通用基础设施实现要求)
8. [功能模块与接口对接](#8-功能模块与接口对接)
9. [WebSocket 实时通信](#9-websocket-实时通信)
10. [页面规划与路由结构](#10-页面规划与路由结构)
11. [各页面功能详细说明](#11-各页面功能详细说明)
12. [全局状态设计建议](#12-全局状态设计建议)
13. [开发约定与注意事项](#13-开发约定与注意事项)
14. [未实现 / 待后端补全的能力](#14-未实现--待后端补全的能力)

---

## 1. 项目概述

Chatgoo 是一个在线聊天应用，前端需要实现的核心能力包括：

- **用户账号体系**：注册、登录、个人资料管理。
- **个人用户管理**：查看 / 编辑个人资料、头像、昵称、性别、签名、地区等；搜索用户。
- **好友系统**：好友申请、审批、拒绝、删除、好友分组管理、好友列表。
- **群组系统**：创建群、查看 / 编辑群资料、解散群、成员管理（加人、踢人、查看成员）。
- **会话与消息**：会话列表、消息收发、消息历史（分页）、已读回执、未读数统计、会话免打扰。
- **实时通信**：基于 WebSocket 的消息推送、会话订阅 / 退订、心跳保活。
- **文件能力**：获取文件元信息（图片 / 文件消息引用）。

---

## 2. 技术栈建议（框架无关）

文档不强制具体框架，但前端工程需具备以下能力，选型时应予满足：

| 能力 | 说明 |
|------|------|
| 组件化渲染 | 任意组件框架（React/Vue/Angular/Svelte）或纯原生 JS |
| 路由 | 需支持嵌套路由、路由守卫（鉴权拦截） |
| HTTP 客户端 | 支持 request/response 拦截器、统一错误处理（fetch / axios 等） |
| 全局状态管理 | 用于存储当前用户、Token、会话列表、未读数、WebSocket 状态等 |
| WebSocket 客户端 | 原生 `WebSocket` API 即可，需支持自动重连 |
| 构建工具 | Vite / Webpack / Rollup 等任意 |
| 类型系统 | **强烈建议**使用 TypeScript（或等效类型标注），便于对接本文档的数据结构 |
| 持久化 | `localStorage` / `sessionStorage` / IndexedDB 任选，用于持久化 Token 与用户基本信息 |

---

## 3. 后端服务基础信息

### 3.1 服务地址

| 环境 | 地址 |
|------|------|
| 开发环境 | `http://localhost:8000` |
| WebSocket | `ws://localhost:8000/ws?token=<JWT>` |

> 部署时通过环境变量 / 构建期配置注入实际地址。

### 3.2 API 前缀

所有 REST API 统一前缀：`/api/v1`

### 3.3 请求约定

- **请求体**：`Content-Type: application/json`（除文件上传外，目前后端未提供上传接口，详见第 14 节）。
- **路径参数**：URL 中 `{xxx}` 形式的路径参数需在请求时替换。
- **查询参数**：以 `?key=value` 形式拼接。
- **认证头**：除 `/auth/register`、`/auth/login` 外，所有接口必须在请求头携带：

  ```
  Authorization: Bearer <JWT_TOKEN>
  ```

### 3.4 跨域

后端已开启 CORS，允许任意来源（`Access-Control-Allow-Origin: *`），允许方法：`GET, POST, PUT, DELETE, OPTIONS`，允许头：`Content-Type, Authorization`。开发期无需额外代理即可直连。

---

## 4. 统一响应与错误处理

### 4.1 成功响应结构

所有接口返回统一 JSON 结构：

```jsonc
{
  "code": 0,            // 业务码，0 表示成功
  "message": "success", // 提示信息
  "data": { /* 业务数据，可能为 null、对象、数组 */ }
}
```

> 当 `data` 为 `null` 时，JSON 中可能省略 `data` 字段（`omitempty`）。前端取值时需做空安全处理。

### 4.2 错误响应结构

错误时 HTTP 状态码非 2xx，body 形如：

```jsonc
{
  "code": 40101,        // 业务错误码
  "message": "invalid or expired token"
}
```

或中间件直接返回的纯 JSON 字符串（同样包含 `code` 与 `message` 字段）。

### 4.3 业务错误码

| code | 含义 | HTTP 状态 | 前端建议处理 |
|------|------|-----------|--------------|
| `0` | 成功 | 200 | 正常处理 data |
| `40001` | 参数错误 | 400 | 表单校验提示 |
| `40101` | 未授权（缺失/无效/过期 Token） | 401 | **清除本地登录态，跳转登录页** |
| `40301` | 禁止访问（无权限） | 403 | 提示无权限 |
| `40401` | 资源不存在 | 404 | 提示并回退 |
| `40901` | 冲突（通用） | 409 | 提示冲突信息 |
| `50001` | 服务器内部错误 | 500 | 提示稍后重试 |
| `41001` | 用户名已存在 | 409 | 注册页用户名冲突提示 |
| `41002` | 用户名或密码错误 | 401 | 登录页凭据错误提示 |
| `41003` | 会话不存在 | 404 | 会话失效提示 |
| `41004` | 非会话参与者 | 403 | 无权操作提示 |

### 4.4 HTTP 客户端拦截器要求

建议在 HTTP 客户端中实现：

1. **请求拦截器**：自动注入 `Authorization: Bearer <token>`（从持久化存储读取）。
2. **响应拦截器**：
   - 解析统一结构，提取 `data` 返回给业务层；
   - 当 `code !== 0` 或 HTTP 状态非 2xx 时，进入错误处理流程；
   - 拦截 `40101`：清空 Token、跳转登录页；
   - 其他错误：抛出包含 `code` 与 `message` 的错误对象，由调用方或全局 toast 处理。

---

## 5. 认证机制与 Token 管理

### 5.1 登录流程

1. 用户提交用户名 + 密码 → `POST /api/v1/auth/login`。
2. 成功后返回 `{ user, token }`：
   - `token`：JWT，**有效期 72 小时**（由后端配置，前端不应硬编码该时长）。
   - `user`：当前用户完整资料。
3. 前端将 `token` 与 `user` 持久化（如 `localStorage`），并写入内存全局状态。
4. 后续所有受保护接口携带该 token。

### 5.2 Token 失效处理

- 收到 `code=40101` 时，认定 Token 失效（过期或非法）。
- 清除持久化的 Token 与用户信息。
- 跳转登录页，并记录来源路由以便登录后回跳。

### 5.3 Token 续期

- 当前后端**未提供刷新 Token 接口**，Token 过期后需用户重新登录。
- 建议在 Token 即将过期前（如剩余 1 小时）提示用户重新登录，避免操作中断。

### 5.4 路由守卫

- 受保护路由在进入前检查 Token 存在性；缺失则重定向至登录页。
- 登录 / 注册页在已登录状态下应重定向至主界面。

---

## 6. 数据结构定义

> 以下使用 TypeScript 接口描述数据结构，非 TS 项目可视为类型注释，按字段名映射即可。所有时间字段为 ISO 8601 字符串（后端 `TIMESTAMPTZ` 序列化为 RFC3339）。

### 6.1 用户

```ts
interface User {
  user_id: number;
  username: string;
  nickname: string;
  avatar_url: string;
  gender: 0 | 1 | 2;   // 0未知 1男 2女
  signature: string;
  region: string;
  birthday: string | null; // ISO date
  status: 0 | 1 | 2;       // 0离线 1在线 2离开
  last_login_at: string | null;
  created_at: string;
  updated_at: string;
}
```

> 注：`password_hash` 字段后端不返回（`json:"-"`）。

### 6.2 好友分组

```ts
interface FriendGroup {
  group_id: number;
  user_id: number;
  group_name: string;
  created_at: string;
}
```

### 6.3 好友关系

```ts
interface FriendRelation {
  relation_id: number;
  user_id: number;
  friend_id: number;
  group_id: number | null;
  remark: string;
  status: 0 | 1 | 2 | 3; // 0待处理 1已接受 2已拒绝 3已拉黑
  applied_at: string;
  approved_at: string | null;
  created_at: string;
}
```

### 6.4 群组

```ts
interface GroupInfo {
  group_id: number;
  group_name: string;
  owner_id: number;
  avatar_url: string;
  announcement: string;
  max_members: number; // 默认 200
  created_at: string;
  updated_at: string;
}

interface GroupMember {
  id: number;
  group_id: number;
  user_id: number;
  role: 0 | 1 | 2; // 0成员 1管理员 2群主
  join_at: string;
}
```

### 6.5 会话

```ts
interface ChatSession {
  session_id: number;
  session_type: 1 | 2; // 1单聊 2群聊
  target_user_id: number | null; // 单聊时为对方 user_id
  group_id: number | null;        // 群聊时为 group_id
  last_message_id: number | null;
  last_message_at: string | null;
  created_at: string;
  updated_at: string;
}

interface SessionParticipant {
  id: number;
  session_id: number;
  user_id: number;
  last_read_msg_id: number;
  muted: boolean;
  joined_at: string;
}
```

### 6.6 消息

```ts
interface Message {
  message_id: number;
  session_id: number;
  sender_id: number;
  message_type: 1 | 2 | 3 | 4; // 1文本 2图片 3文件 4语音
  content: string;
  file_id: number | null;
  reply_to_msg_id: number | null;
  sent_at: string;
}
```

### 6.7 文件

```ts
interface File {
  file_id: number;
  uploader_id: number;
  file_name: string;
  file_url: string;
  file_size: number; // 字节
  file_type: string;
  mime_type: string;
  created_at: string;
}
```

### 6.8 未读数

`GET /api/v1/messages/unread-count` 返回的 `data` 为按会话统计的未读数映射：

```ts
type UnreadCount = Record<number, number>; // { [session_id]: count }
```

---

## 7. 通用基础设施实现要求

### 7.1 HTTP 客户端封装

实现一个统一的请求函数（命名示例 `request`），签名建议：

```ts
function request<T>(method: 'GET'|'POST'|'PUT'|'DELETE', path: string, opts?: {
  params?: Record<string, any>;   // 查询参数
  body?: any;                     // 请求体
}): Promise<T>;                   // 直接返回 data 部分
```

职责：
- 自动拼接 baseURL + path；
- GET 时把 `params` 序列化为 querystring；
- 非 GET 时把 `body` 序列化为 JSON 并设置 `Content-Type`；
- 注入 `Authorization` 头；
- 解包响应：成功返回 `data`，失败抛出带 `code` / `message` 的错误。

### 7.2 WebSocket 客户端封装

实现一个单例 WS 客户端，需具备：

- **连接**：`connect(token)`，拼接 `ws://host/ws?token=<token>`。
- **重连**：连接断开时指数退避重连（如 1s/2s/4s/8s，上限 30s），重连成功后重新订阅当前会话。
- **心跳**：定时发送 `{ "type": "ping", "data": null }`，收到 `pong` 视为链路正常；建议 25~30 秒一次（后端 ping 间隔 30s，读超时 60s）。
- **消息分发**：收到消息按 `type` 字段分发到对应处理器（见第 9 节）。
- **会话订阅**：进入某会话时发送 `subscribe_session`，离开时发送 `unsubscribe_session`。
- **状态暴露**：对外暴露连接状态（连接中 / 已连接 / 已断开），便于 UI 显示。

### 7.3 持久化存储

至少持久化以下内容：
- `token`：JWT。
- `user`：当前用户基本信息（用于首屏渲染前的占位）。

> 会话列表、消息记录建议缓存在内存 / IndexedDB，不强制持久化。

---

## 8. 功能模块与接口对接

### 8.1 认证模块

#### 8.1.1 注册

- **接口**：`POST /api/v1/auth/register`
- **请求体**：

  ```ts
  interface RegisterRequest {
    username: string; // 至少 3 字符
    password: string; // 至少 6 字符
    nickname?: string; // 可选，为空时后端默认取 username
  }
  ```

- **响应 data**：

  ```ts
  { user: User; token: string }
  ```

- **错误码**：`41001` 用户名已存在；`40001` 用户名/密码长度不合法。
- **前端行为**：注册成功后等同登录，直接持久化 Token 并跳转主界面。

#### 8.1.2 登录

- **接口**：`POST /api/v1/auth/login`
- **请求体**：

  ```ts
  { username: string; password: string }
  ```

- **响应 data**：

  ```ts
  { user: User; token: string }
  ```

- **错误码**：`41002` 用户名或密码错误。

### 8.2 个人用户管理模块

> 个人用户管理是核心模块之一，涵盖当前用户的资料查看、编辑，以及用户搜索。

#### 8.2.1 获取当前用户资料

- **接口**：`GET /api/v1/users/profile`
- **响应 data**：`User`
- **使用场景**：登录后刷新本地缓存、个人中心页展示。

#### 8.2.2 更新当前用户资料

- **接口**：`PUT /api/v1/users/profile`
- **请求体**：

  ```ts
  interface UpdateProfileRequest {
    nickname: string;
    avatar_url: string;
    gender: 0 | 1 | 2;
    signature: string;
    region: string;
  }
  ```

- **响应 data**：`null`
- **说明**：
  - 后端仅接受上述 5 个字段，`username`、`user_id`、`birthday` 等当前不可通过此接口修改（详见第 14 节）。
  - `avatar_url` 为字符串 URL，当前后端**未提供文件上传接口**，前端需通过外部图床 / 预置头像 / 第三方对象存储获取 URL 后填入（详见第 14 节）。
- **前端行为**：成功后刷新本地 `user` 状态并提示已保存。

#### 8.2.3 搜索用户

- **接口**：`GET /api/v1/users/search?keyword=<keyword>`
- **响应 data**：`User[]`
- **使用场景**：添加好友前的查找步骤。
- **说明**：当前后端 `SearchUser` **返回 `nil, nil`（未实现）**，前端可先实现 UI 与接口调用，待后端补全后联调。详见第 14 节。
- **前端行为**：展示搜索结果列表，每项提供“加为好友”入口。

### 8.3 好友管理模块

#### 8.3.1 好友列表

- **接口**：`GET /api/v1/friends`
- **响应 data**：`FriendRelation[]`（`status=1` 已接受的记录）
- **前端行为**：按 `group_id` 分组展示，或在无分组时归入“未分组”。

#### 8.3.2 申请好友

- **接口**：`POST /api/v1/friends/apply`
- **请求体**：

  ```ts
  interface ApplyFriendRequest {
    friend_id: number;
    remark?: string;
    group_id?: number | null;
    message?: string; // 申请留言
  }
  ```

- **响应 data**：`null`
- **错误码**：`40001` 不能添加自己 / 已是好友 / 申请已存在 / 目标用户不存在。

#### 8.3.3 审批好友申请

- **接口**：`POST /api/v1/friends/approve`
- **请求体**：`{ relation_id: number }`
- **响应 data**：`null`

#### 8.3.4 拒绝好友申请

- **接口**：`POST /api/v1/friends/reject`
- **请求体**：`{ relation_id: number }`
- **响应 data**：`null`

#### 8.3.5 待处理好友申请列表

- **接口**：`GET /api/v1/friends/requests`
- **响应 data**：`FriendRelation[]`（`status=0` 待处理，且 `friend_id` 为当前用户的记录）
- **前端行为**：在“新朋友”入口展示，支持同意 / 拒绝操作。

#### 8.3.6 删除好友

- **接口**：`DELETE /api/v1/friends/{friendID}`
- **路径参数**：`friendID` = 对方用户 ID
- **响应 data**：`null`

#### 8.3.7 好友分组列表

- **接口**：`GET /api/v1/friend-groups`
- **响应 data**：`FriendGroup[]`

#### 8.3.8 创建好友分组

- **接口**：`POST /api/v1/friend-groups`
- **请求体**：`{ group_name: string }`
- **响应 data**：`FriendGroup`

#### 8.3.9 删除好友分组

- **接口**：`DELETE /api/v1/friend-groups/{groupID}`
- **响应 data**：`null`
- **说明**：后端外键约束为 `ON DELETE SET NULL`，删除分组后，该分组下的好友关系 `group_id` 将被置空（归入未分组）。

### 8.4 群组管理模块

#### 8.4.1 创建群

- **接口**：`POST /api/v1/groups`
- **请求体**：

  ```ts
  interface CreateGroupRequest {
    group_name: string;
    avatar_url?: string;
    announcement?: string;
    max_members?: number; // 不传或 0 时后端默认 200
  }
  ```

- **响应 data**：`GroupInfo`
- **说明**：创建者自动成为群主（`role=2`），后端会自动创建对应群聊会话。

#### 8.4.2 获取群信息

- **接口**：`GET /api/v1/groups/{groupID}`
- **响应 data**：`GroupInfo`

#### 8.4.3 更新群信息

- **接口**：`PUT /api/v1/groups/{groupID}`
- **请求体**：

  ```ts
  {
    group_name: string;
    avatar_url: string;
    announcement: string;
    max_members: number;
  }
  ```

- **响应 data**：`null`
- **权限**：仅群主可操作（后端校验，非群主返回 `40301` / `ErrNotGroupOwner`）。

#### 8.4.4 解散群

- **接口**：`DELETE /api/v1/groups/{groupID}`
- **响应 data**：`null`
- **权限**：仅群主可操作。

#### 8.4.5 群成员列表

- **接口**：`GET /api/v1/groups/{groupID}/members`
- **响应 data**：`GroupMember[]`

#### 8.4.6 添加群成员

- **接口**：`POST /api/v1/groups/{groupID}/members`
- **请求体**：`{ user_id: number }`
- **响应 data**：`null`
- **错误码**：`40001` 该用户已是群成员。
- **说明**：当前后端未做“邀请人权限校验”，任何登录用户均可调用，但建议前端在 UI 上仅对群成员 / 管理员暴露此入口。

#### 8.4.7 移除群成员

- **接口**：`DELETE /api/v1/groups/{groupID}/members/{userID}`
- **响应 data**：`null`
- **说明**：后端未严格校验操作者权限，前端应自行控制 UI 入口（建议仅群主 / 管理员可见）。

### 8.5 会话管理模块

#### 8.5.1 会话列表

- **接口**：`GET /api/v1/sessions`
- **响应 data**：`ChatSession[]`
- **前端行为**：
  - 按 `last_message_at` 倒序展示；
  - 单聊会话（`session_type=1`）需根据 `target_user_id` 拉取对方资料展示头像 / 昵称；
  - 群聊会话（`session_type=2`）需根据 `group_id` 拉取群信息展示；
  - 结合未读数 `messages/unread-count` 展示红点。

#### 8.5.2 会话详情

- **接口**：`GET /api/v1/sessions/{sessionID}`
- **响应 data**：`ChatSession`

#### 8.5.3 会话免打扰

- **接口**：`PUT /api/v1/sessions/{sessionID}/mute`
- **请求体**：`{ muted: boolean }`
- **响应 data**：`null`
- **前端行为**：在会话右键菜单 / 设置面板中提供开关；免打扰开启后，新消息不弹窗 / 不响铃，但仍计入未读数。

### 8.6 消息模块

#### 8.6.1 发送消息

- **接口**：`POST /api/v1/messages`
- **请求体**：

  ```ts
  interface SendMessageRequest {
    session_id: number;
    message_type: 1 | 2 | 3 | 4; // 1文本 2图片 3文件 4语音
    content: string;              // 文本类型必填，非文本可填文件名等描述
    file_id?: number | null;      // 图片/文件/语音类型关联的文件 ID
    reply_to_msg_id?: number | null;
  }
  ```

- **响应 data**：`Message`（含服务端生成的 `message_id` 与 `sent_at`）
- **错误码**：`41003` 会话不存在；`41004` 非会话参与者；`40001` 文本消息内容为空。
- **前端行为**：
  - 发送时先在 UI 临时展示（乐观更新），收到响应后用真实 `message_id` 替换临时 ID；
  - 后端会通过 WebSocket 向会话内所有在线成员广播 `new_message`，前端发送方自身也会收到，需做去重（依据 `message_id`）。

#### 8.6.2 消息历史（分页）

- **接口**：`GET /api/v1/messages/history?session_id=<id>&before_id=<id>&limit=<n>`
- **查询参数**：
  - `session_id`（必填）：会话 ID
  - `before_id`（可选）：返回该 ID 之前的消息，用于向上翻页
  - `limit`（可选）：每页条数，默认 50
- **响应 data**：`Message[]`（按时间倒序返回，前端展示时需正序渲染）
- **前端行为**：
  - 进入会话首次拉取最新一页；
  - 滚动到顶部时，以当前最早消息 ID 作为 `before_id` 继续向上加载；
  - 注意去重，避免 WebSocket 推送与历史拉取重复。

#### 8.6.3 标记已读

- **接口**：`POST /api/v1/messages/read`
- **请求体**：`{ session_id: number; message_id: number }`
- **响应 data**：`null`
- **前端行为**：进入会话、或会话内收到新消息时调用，更新该会话未读数为 0。

#### 8.6.4 未读数统计

- **接口**：`GET /api/v1/messages/unread-count`
- **响应 data**：`Record<string, number>`，键为 `session_id`（字符串形式），值为未读条数
- **前端行为**：
  - 应用启动时拉取一次；
  - 收到 `new_message` WebSocket 推送时，若该会话非当前激活会话，则本地未读数 +1；
  - 进入会话并标记已读后，清零该会话未读数；
  - 汇总所有会话未读数显示在角标。

### 8.7 文件模块

#### 8.7.1 获取文件信息

- **接口**：`GET /api/v1/files/{fileID}`
- **响应 data**：`File`
- **使用场景**：渲染图片消息、提供文件消息下载链接。
- **说明**：`file_url` 为可访问的文件 URL（后端存储路径，前端可直接用于展示 / 下载）。

---

## 9. WebSocket 实时通信

### 9.1 连接

- **地址**：`ws://<host>/ws?token=<JWT>`
- **认证**：通过 query 参数 `token` 传递 JWT（与 REST 接口的 Bearer 头不同）。
- **连接时机**：用户登录成功后立即建立；应用进入后台 / 用户主动退出时关闭。
- **协议**：文本帧（JSON 字符串）。

### 9.2 客户端 → 服务端 消息

所有上行消息结构：

```ts
interface ClientWSMessage {
  type: string;
  data: any;
}
```

支持的消息类型：

#### 9.2.1 心跳 ping

```json
{ "type": "ping", "data": null }
```

服务端响应：

```json
{ "type": "pong", "data": null }
```

#### 9.2.2 订阅会话

进入某会话页面时发送：

```json
{ "type": "subscribe_session", "data": { "session_id": 123 } }
```

#### 9.2.3 退订会话

离开会话页面时发送：

```json
{ "type": "unsubscribe_session", "data": { "session_id": 123 } }
```

> 只有订阅了某会话，才会收到该会话的 `new_message` 推送。

### 9.3 服务端 → 客户端 消息

所有下行消息结构：

```ts
interface ServerWSMessage {
  type: string;
  data: any;
}
```

#### 9.3.1 新消息推送

```json
{
  "type": "new_message",
  "data": {
    "message_id": 456,
    "session_id": 123,
    "sender_id": 789,
    "message_type": 1,
    "content": "hello",
    "file_id": null,
    "reply_to_msg_id": null,
    "sent_at": "2026-07-06T10:00:00Z"
  }
}
```

`data` 即完整的 `Message` 对象。

#### 9.3.2 心跳响应 pong

见 9.2.1。

### 9.4 前端处理流程

收到 `new_message` 时：

1. 解析 `data.session_id`：
   - 若为当前激活会话：直接追加到消息列表底部，并调用 `POST /messages/read` 标记已读；
   - 若非激活会话：将该会话未读数 +1，更新会话列表的 `last_message_at` / 预览内容，触发角标 / 红点更新；
   - 若该会话被免打扰：不弹窗不响铃，但仍计入未读数。
2. 依据 `message_id` 去重（与本地已存在消息 / 乐观发送的临时消息比对）。
3. 滚动到底部（激活会话场景）。

### 9.5 连接状态处理

- **断开**：UI 上提示“连接断开，正在重连…”。
- **重连成功**：重新订阅当前激活会话；拉取一次未读数与最新会话列表，补偿断线期间可能丢失的消息。
- **Token 失效**：WebSocket 握手会失败（连接被拒绝），此时应清空登录态跳转登录页。

---

## 10. 页面规划与路由结构

以下路由以路径形式描述，与具体框架无关；命名仅作参考，可按团队习惯调整。

### 10.1 路由总览

| 路由 | 是否需登录 | 说明 |
|------|------------|------|
| `/login` | 否 | 登录页 |
| `/register` | 否 | 注册页 |
| `/` | 是 | 主界面（含会话列表 + 聊天区，三栏布局） |
| `/chat/:sessionId` | 是 | 主界面，并打开指定会话 |
| `/contacts` | 是 | 通讯录（好友 + 群组） |
| `/contacts/friends` | 是 | 好友列表（可分组） |
| `/contacts/groups` | 是 | 群组列表 |
| `/contacts/requests` | 是 | 新的朋友（待处理申请） |
| `/profile` | 是 | 个人资料查看页 |
| `/profile/edit` | 是 | 个人资料编辑页 |
| `/settings` | 是 | 设置页（含退出登录、免打扰总开关等） |
| `/users/:userId` | 是 | 他人资料页（来自好友 / 搜索结果） |
| `/groups/:groupId` | 是 | 群信息页（含成员列表、群管理入口） |
| `/groups/:groupId/edit` | 是 | 群信息编辑页（仅群主） |
| `/groups/create` | 是 | 创建群页 |
| `/friends/apply/:userId` | 是 | 申请好友页（填写备注 / 分组 / 留言） |

### 10.2 主界面布局

主界面建议采用经典三栏布局：

```
┌─────────────┬──────────────────────┬──────────────────┐
│  导航栏      │  会话列表            │  聊天区           │
│  (侧边)      │  (中栏左)            │  (中栏右)         │
│             │                      │                  │
│  - 消息     │  搜索框              │  消息头部         │
│  - 通讯录   │  会话项列表          │  消息历史         │
│  - 个人     │  (含未读红点)         │  输入区           │
│  - 设置     │                      │                  │
└─────────────┴──────────────────────┴──────────────────┘
```

- 导航栏为最左侧固定窄列，切换“消息 / 通讯录 / 个人 / 设置”。
- 会话列表与聊天区在“消息”视图下并存；点击会话项右侧展示对应聊天区。
- 移动端可折叠为单栏切换（列表 → 聊天 → 返回列表）。

---

## 11. 各页面功能详细说明

### 11.1 登录页 `/login`

- 表单字段：用户名、密码。
- 校验：用户名非空，密码非空。
- 提交：调用 `POST /auth/login`。
- 成功：持久化 Token + user，跳转 `/`。
- 失败：根据 `code` 提示（`41002` → “用户名或密码错误”）。
- 底部链接：跳转注册页。

### 11.2 注册页 `/register`

- 表单字段：用户名、昵称（可选）、密码、确认密码。
- 校验：
  - 用户名 ≥ 3 字符；
  - 密码 ≥ 6 字符；
  - 两次密码一致。
- 提交：调用 `POST /auth/register`。
- 成功：等同登录，跳转 `/`。
- 失败：`41001` → “用户名已存在”。

### 11.3 主界面 - 消息视图 `/` & `/chat/:sessionId`

#### 会话列表区

- 进入时拉取 `GET /sessions` + `GET /messages/unread-count`。
- 每项展示：
  - 头像（单聊取对方头像，群聊取群头像）；
  - 名称（单聊取对方昵称 / 备注，群聊取群名）；
  - 最后一条消息预览（可拉取该会话最新一条消息，或由前端维护本地缓存）；
  - 时间戳；
  - 未读红点（来自未读数映射）；
  - 免打扰图标。
- 交互：
  - 点击进入聊天区，路由更新为 `/chat/:sessionId`；
  - 右键 / 长按菜单：免打扰开关、删除会话（注：后端暂无删除会话接口，详见第 14 节）。
- 实时更新：WebSocket 收到 `new_message` 时，更新对应会话项的预览、时间、未读数，并上浮至顶部。

#### 聊天区

- 顶部：对方名称 / 群名，右侧操作按钮（查看资料 / 群信息、搜索消息等）。
- 消息历史区：
  - 进入会话拉取 `GET /messages/history?session_id=…&limit=50`；
  - 滚动到顶部向上加载（`before_id` = 当前最早消息 ID）；
  - 自己发送的消息靠右，他人靠左，群聊显示发送者昵称与头像；
  - 支持消息类型渲染：文本、图片（`file_id` → `GET /files/{id}` 获取 `file_url` 展示）、文件（提供下载链接）、语音（提供播放按钮）；
  - 支持引用消息展示（`reply_to_msg_id` 对应消息内容预览）。
- 输入区：
  - 文本输入框（支持多行、Enter 发送、Shift+Enter 换行）；
  - 工具按钮：表情、图片、文件、语音（UI 占位即可，具体能力依赖后端文件上传补全）；
  - 发送按钮：调用 `POST /messages`，乐观更新。
- 进入会话后：
  - 调用 `POST /messages/read` 标记最新消息已读；
  - WebSocket 发送 `subscribe_session`；
  - 离开时发送 `unsubscribe_session`。

### 11.4 通讯录 - 好友列表 `/contacts/friends`

- 拉取 `GET /friends` + `GET /friend-groups`。
- 按分组折叠展示，每个好友项展示头像、昵称（或备注）、签名。
- 点击好友 → 跳转 `/users/:friendId`。
- 每项操作：发消息（跳转至对应单聊会话）、删除好友、移动到其他分组（注：后端暂无“修改好友分组”接口，详见第 14 节）。
- 顶部按钮：添加好友（跳转搜索）。

### 11.5 通讯录 - 群组列表 `/contacts/groups`

- 群组列表数据来源：当前后端**无“我加入的群列表”接口**，需通过会话列表中 `session_type=2` 的项反推（`group_id` → `GET /groups/{id}`）。详见第 14 节。
- 每项展示群头像、群名、群公告摘要、成员数。
- 点击 → 跳转 `/groups/:groupId`。
- 顶部按钮：创建群（跳转 `/groups/create`）。

### 11.6 通讯录 - 新的朋友 `/contacts/requests`

- 拉取 `GET /friends/requests`。
- 列表展示每条申请：申请人头像、昵称、申请留言、申请时间。
- 操作：同意（`POST /friends/approve`）、拒绝（`POST /friends/reject`）。
- 操作后从列表移除该条目。

### 11.7 个人资料查看页 `/profile`

- 拉取 `GET /users/profile`（或直接使用全局缓存的 user）。
- 展示：头像、用户名、昵称、性别、签名、地区、注册时间、最后登录时间。
- 操作：编辑资料（跳转 `/profile/edit`）、退出登录（跳转设置或直接弹窗确认）。

### 11.8 个人资料编辑页 `/profile/edit`

- 表单字段：昵称、头像 URL、性别（下拉/单选 0/1/2）、签名、地区。
- 头像编辑：
  - 由于后端无文件上传接口，提供两种方案：
    1. 预置头像库供用户选择；
    2. 允许用户粘贴图片 URL；
    3. 接入第三方图床 / 对象存储（前端直传，拿到 URL 后填入）。
- 校验：昵称非空。
- 提交：`PUT /users/profile`，成功后更新全局 user 状态并返回 `/profile`。

### 11.9 设置页 `/settings`

- 通用设置：
  - 消息提示音开关（本地存储）；
  - 桌面通知开关（请求浏览器 Notification 权限）；
  - 主题（亮 / 暗，本地存储）。
- 账号操作：
  - 退出登录：清空 Token / user，关闭 WebSocket，跳转 `/login`。
- 关于：版本号、后端服务地址。

### 11.10 他人资料页 `/users/:userId`

- 用于查看好友或搜索到的用户。
- 数据来源：当前后端**无“按 ID 查询用户”的独立接口**（`SearchUser` 未实现），建议：
  - 若该用户在好友列表中，从本地好友数据取；
  - 否则提示接口待补全或通过搜索接口查找。详见第 14 节。
- 展示：头像、昵称、用户名、性别、签名、地区。
- 操作：
  - 发消息（若是好友，跳转对应单聊会话）；
  - 添加好友（若非好友，跳转 `/friends/apply/:userId`）；
  - 删除好友（若是好友）。

### 11.11 群信息页 `/groups/:groupId`

- 拉取 `GET /groups/{id}` + `GET /groups/{id}/members`。
- 展示：群头像、群名、群公告、群主、成员数、成员列表（分页或全部，按角色排序：群主 > 管理员 > 成员）。
- 操作：
  - 进入群聊（跳转对应群会话）；
  - 编辑群信息（仅群主可见，跳转 `/groups/:groupId/edit`）；
  - 解散群（仅群主，二次确认后 `DELETE /groups/{id}`）；
  - 邀请成员（弹出用户选择器，调用 `POST /groups/{id}/members`）；
  - 退出群（注：后端 `RemoveGroupMember` 可用于退群，传入自己的 userID；但需注意群主退群后端未做特殊处理，详见第 14 节）。

### 11.12 群信息编辑页 `/groups/:groupId/edit`

- 表单字段：群名、群头像 URL、群公告、最大成员数。
- 提交：`PUT /groups/{id}`。
- 仅群主可进入（前端路由层校验）。

### 11.13 创建群页 `/groups/create`

- 表单字段：群名（必填）、群头像 URL、群公告、最大成员数（可选，默认 200）。
- 可选：选择初始成员（从好友列表多选，创建后批量调用 `POST /groups/{id}/members`）。
- 提交：`POST /groups`，成功后跳转群信息页或群会话。

### 11.14 申请好友页 `/friends/apply/:userId`

- 表单字段：备注（可选）、好友分组（下拉，来自 `GET /friend-groups`）、申请留言（可选）。
- 提交：`POST /friends/apply`。
- 成功后返回来源页并提示“申请已发送”。

---

## 12. 全局状态设计建议

建议将以下切片纳入全局状态管理（实现方式不限）：

| 切片 | 内容 | 用途 |
|------|------|------|
| `auth` | `token`、当前 `user`、登录状态 | 鉴权、首屏渲染、用户信息展示 |
| `sessions` | 会话列表、当前激活会话 ID | 会话列表展示、聊天区切换 |
| `messages` | 按 `session_id` 分桶的消息数组、加载状态、是否还有更多历史 | 聊天区渲染、分页加载 |
| `unread` | `{ [session_id]: count }`、总未读数 | 角标、红点 |
| `friends` | 好友列表、好友分组、待处理申请 | 通讯录、申请审批 |
| `groups` | 群组信息缓存、群成员缓存 | 群信息页、群聊展示 |
| `users` | 用户资料缓存（`{ [user_id]: User }`） | 避免重复拉取他人资料 |
| `ws` | 连接状态、重连次数 | 状态提示、重连控制 |
| `ui` | 主题、提示音开关、桌面通知开关 | 个性化设置 |

> 消息记录较大时，建议结合 IndexedDB 做持久化缓存，避免内存膨胀。

---

## 13. 开发约定与注意事项

### 13.1 字段命名

- 后端 JSON 字段统一为 **snake_case**（如 `user_id`、`session_id`、`avatar_url`），前端在数据层应保持一致，避免无谓的转换开销；仅在视图层展示时按需格式化。

### 13.2 数值类型

- 所有 ID（`user_id`、`session_id`、`message_id` 等）为 `number`，但值域可能超过 JS 安全整数（`2^53`）。后端为 `BIGSERIAL`，常规业务量下不会超限，但若考虑极端情况，可在 HTTP 层将数字 ID 解析为字符串处理。一般业务可忽略。

### 13.3 时间格式

- 所有时间为 ISO 8601 字符串（RFC3339），前端按本地时区格式化展示。

### 13.4 错误提示

- 统一通过全局 toast / message 组件展示错误，依据 `code` 给出友好文案。
- 表单错误就近在字段下方展示。

### 13.5 加载状态

- 列表首次加载显示骨架屏 / loading；
- 翻页加载显示顶部 / 底部 loading 条；
- 按钮提交时禁用并显示 loading，避免重复提交。

### 13.6 空状态

- 会话列表为空：引导添加好友 / 创建群；
- 消息历史为空：提示“开始对话吧”；
- 搜索无结果：提示“未找到用户”。

### 13.7 安全

- Token 存储于 `localStorage`（注意 XSS 风险，需做好内容转义与 CSP）；
- 用户输入的文本消息在渲染时**必须转义**，防止 XSS（尤其消息内容、群公告、个人签名）；
- 不要在前端代码中硬编码 JWT 密钥或后端机密信息。

### 13.8 性能

- 长会话消息列表采用虚拟滚动，避免大量 DOM 节点；
- 图片消息懒加载；
- 会话列表项避免不必要的重渲染（使用稳定 key，如 `session_id`）。

### 13.9 移动端适配

- 三栏布局在窄屏（< 768px）下退化为单栏，通过返回按钮在列表与聊天区切换；
- 触摸交互：长按会话弹出菜单、下拉刷新等。

---

## 14. 未实现 / 待后端补全的能力

文档编写时基于当前后端代码，以下能力**后端尚未实现或未提供对应接口**，前端在开发时需注意：

| 能力 | 现状 | 前端应对 |
|------|------|----------|
| 用户搜索 `GET /users/search` | `SearchUser` 返回 `nil, nil`，未实现 | 实现 UI 与调用，待后端补全；可临时用本地好友列表模拟搜索 |
| 按 ID 查询用户 | 无独立接口 | 从本地缓存 / 好友列表取数据；待后端补全 |
| 文件上传 | 无上传接口，仅有 `GET /files/{id}` | 头像、图片、文件消息依赖外部图床 / 对象存储获取 URL，或待后端补全上传接口 |
| 修改好友分组 / 备注 | 无接口 | 好友分组仅支持增删，不支持把好友移到其他分组；备注在申请时设置后不可改 |
| 删除会话 | 无接口 | 会话列表不支持删除，仅支持免打扰；前端可做“隐藏”本地操作 |
| 退出群（主动） | `RemoveGroupMember` 可复用 | 前端调用 `DELETE /groups/{id}/members/{自身userID}` 退群；但群主退群行为后端未特殊处理，建议前端限制群主先转让再退群 |
| 群角色变更（升管理员 / 降级） | 无接口 | 群成员角色展示只读 |
| 消息撤回 / 编辑 | 无接口 | 不实现 |
| 消息删除 | 无接口 | 不实现 |
| @ 提及 / 消息通知 | 无接口 | 不实现 |
| 消息已读回执详情（谁读了） | 仅 `MarkRead` 接口 | 不展示“谁已读”，仅维护自己侧的未读数 |
| 用户在线状态实时推送 | 后端 `Hub.IsOnline` 存在但未通过 WS 推送状态变更 | 不展示实时在线状态，或仅展示静态 `status` 字段 |

> 前端在开发过程中如发现新增接口或字段变化，以最新后端代码 / 接口文档为准；本节列出的事项建议与后端开发者确认排期。

---

## 附录：接口速查表

| 模块 | 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|------|
| 认证 | POST | `/api/v1/auth/register` | 否 | 注册 |
| 认证 | POST | `/api/v1/auth/login` | 否 | 登录 |
| 用户 | GET | `/api/v1/users/profile` | 是 | 获取当前用户资料 |
| 用户 | PUT | `/api/v1/users/profile` | 是 | 更新当前用户资料 |
| 用户 | GET | `/api/v1/users/search?keyword=` | 是 | 搜索用户（待实现） |
| 好友 | GET | `/api/v1/friends` | 是 | 好友列表 |
| 好友 | POST | `/api/v1/friends/apply` | 是 | 申请好友 |
| 好友 | POST | `/api/v1/friends/approve` | 是 | 同意申请 |
| 好友 | POST | `/api/v1/friends/reject` | 是 | 拒绝申请 |
| 好友 | GET | `/api/v1/friends/requests` | 是 | 待处理申请 |
| 好友 | DELETE | `/api/v1/friends/{friendID}` | 是 | 删除好友 |
| 好友分组 | GET | `/api/v1/friend-groups` | 是 | 分组列表 |
| 好友分组 | POST | `/api/v1/friend-groups` | 是 | 创建分组 |
| 好友分组 | DELETE | `/api/v1/friend-groups/{groupID}` | 是 | 删除分组 |
| 群组 | POST | `/api/v1/groups` | 是 | 创建群 |
| 群组 | GET | `/api/v1/groups/{groupID}` | 是 | 群信息 |
| 群组 | PUT | `/api/v1/groups/{groupID}` | 是 | 更新群信息（群主） |
| 群组 | DELETE | `/api/v1/groups/{groupID}` | 是 | 解散群（群主） |
| 群组 | GET | `/api/v1/groups/{groupID}/members` | 是 | 成员列表 |
| 群组 | POST | `/api/v1/groups/{groupID}/members` | 是 | 添加成员 |
| 群组 | DELETE | `/api/v1/groups/{groupID}/members/{userID}` | 是 | 移除成员 |
| 会话 | GET | `/api/v1/sessions` | 是 | 会话列表 |
| 会话 | GET | `/api/v1/sessions/{sessionID}` | 是 | 会话详情 |
| 会话 | PUT | `/api/v1/sessions/{sessionID}/mute` | 是 | 免打扰开关 |
| 消息 | POST | `/api/v1/messages` | 是 | 发送消息 |
| 消息 | GET | `/api/v1/messages/history?session_id=&before_id=&limit=` | 是 | 历史消息 |
| 消息 | POST | `/api/v1/messages/read` | 是 | 标记已读 |
| 消息 | GET | `/api/v1/messages/unread-count` | 是 | 未读数 |
| 文件 | GET | `/api/v1/files/{fileID}` | 是 | 文件信息 |
| WebSocket | WS | `/ws?token=<JWT>` | 是（query） | 实时通信 |

---

*文档完。如有疑问，请对照后端源码 `internal/handler/`、`internal/service/`、`internal/model/` 与 `internal/router/router.go`。*
