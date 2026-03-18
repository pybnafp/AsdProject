# start-love-web 前端页面设计说明

## 1. 项目概览

`start-love-web` 是 AsdProject 的前端 Web 应用，面向 **星启诶艾（SCAI）** 孤独症早筛与智能体产品，提供：

- **PC 端**：完整对话、报告、单页资讯与登录能力
- **移动端**：独立路由与 Vant UI，新建会话、历史会话、我的报告、更多（个人/单页入口）

技术栈：**Vue 3** + **Vue Router 4** + **Pinia** + **Element Plus（PC）** + **Vant 4（移动端）**，样式使用 **Stylus**、**Tailwind CSS** 及主题变量，支持 **亮色/暗色主题**。

---

## 2. 技术栈与依赖

| 类别 | 技术/库 |
|------|--------|
| 框架 | Vue 3、Vue Router 4、Pinia |
| PC UI | Element Plus、@element-plus/icons-vue |
| 移动端 UI | Vant 4（Tabbar、NavBar、Form、Dialog、Uploader 等） |
| HTTP | Axios |
| 流式数据 | @microsoft/fetch-event-source（SSE） |
| 富文本/ Markdown | markdown-it、markdown-it-emoji、markdown-it-mathjax3、highlight.js、md-editor-v3 |
| 其它 | good-storage（持久化）、clipboard（复制）、qrcode、vue-next-wxlogin（微信登录）、animate.css、echarts、v3-waterfall 等 |

构建：Vue CLI 5、Webpack 5、Babel、ESLint。开发默认端口 **8888**，通过 `vue.config.js` 的 proxy 将 `/api` 转发到后端。

---

## 3. 目录结构

```
start-love-web/
├── public/                 # 静态资源
│   ├── index.html
│   ├── images/             # 图片（logo、登录背景、菜单图标等）
│   └── css/
│       └── wxlogin.css     # 微信登录组件样式
├── src/
│   ├── main.js             # 入口：Pinia、Element Plus、Vant 组件注册、全局样式
│   ├── App.vue              # 根组件：router-view、主题与 ResizeObserver 节流
│   ├── router.js            # 路由表、beforeEach 标题与微信授权逻辑
│   ├── assets/
│   │   ├── css/             # 全局与页面样式（styl）
│   │   │   ├── main.styl, common.styl, index.styl, home.styl, chat.styl
│   │   │   ├── login.styl, report.styl, single-common.styl
│   │   │   ├── theme-light.styl, theme-dark.styl
│   │   │   ├── tailwind.css, custom-scroll.styl
│   │   │   ├── markdown/vue.css
│   │   │   └── mobile/*.styl
│   │   ├── img/             # 组件内引用的图标、插图
│   │   └── iconfont/        # 图标字体
│   ├── components/          # 通用组件
│   │   ├── HeaderBar.vue    # 顶部导航（早筛技术、政策指引等单页链接）
│   │   ├── FooterBar.vue    # 页脚版权
│   │   ├── ChatInput.vue    # 对话输入（多行输入、文件/报告选择、发送）
│   │   ├── ChatPrompt.vue   # 用户消息展示（Markdown、附件/报告列表）
│   │   ├── ChatReply.vue    # AI 回复展示（思考过程 + 正文、Markdown、代码高亮）
│   │   ├── Welcome.vue      # 欢迎区（标题 + ChatInput + 插图）
│   │   ├── FileSelect.vue   # 选择/上传报告（弹窗 + 筛选 + 表格）
│   │   ├── FileList.vue     # 已选文件列表
│   │   ├── ReportList.vue   # 已选报告列表
│   │   ├── ReportTable.vue  # 报告表格（分页、状态、链接）
│   │   ├── LoginDialog.vue  # 登录弹窗（若使用）
│   │   ├── ItemList.vue
│   │   ├── SendMsg.vue       # 发送验证码
│   │   └── mobile/
│   │       └── FileSelect.vue  # 移动端文件/报告选择
│   ├── views/               # 页面
│   │   ├── Index.vue        # 首页（导航入口、登录按钮）
│   │   ├── Home.vue         # PC 主布局：侧边栏 + 主内容区
│   │   ├── Chat.vue         # PC 对话页（消息列表 + SSE 流式）
│   │   ├── Login.vue        # PC 登录（手机号+验证码/密码、微信扫码）
│   │   ├── LoginCallback.vue # 微信回调页
│   │   ├── Report.vue       # 我的报告（单项/综合、筛选、ReportTable）
│   │   ├── Single*.vue      # 单页：早筛技术、政策指引、星启协爱、公益机构、开放平台、关于我们
│   │   ├── 404.vue
│   │   └── mobile/
│   │       ├── Home.vue     # 移动端壳：Tabbar + router-view
│   │       ├── AddChat.vue   # 新建会话（欢迎 + ChatInput）
│   │       ├── Chat.vue     # 移动端对话详情
│   │       ├── ChatList.vue  # 历史会话列表
│   │       ├── Report.vue   # 我的报告
│   │       ├── Profile.vue   # 更多（设置、单页入口等）
│   │       ├── Login.vue    # 移动端登录
│   │       └── Single*.vue  # 移动端单页
│   ├── store/
│   │   ├── sharedata.js     # 全局：登录态、用户信息、主题、登录弹窗
│   │   ├── theme.js         # 主题（localStorage）
│   │   ├── cache.js         # checkSession、clientId
│   │   ├── system.js        # 系统级工具（如文件类型、大小格式化）
│   │   ├── sidebar.js
│   │   └── session.js
│   └── utils/
│       ├── http.js          # axios 封装（get/post、withCredentials、401 跳转登录）
│       ├── dialog.js        # 消息/确认弹窗
│       ├── validate.js
│       └── libs.js          # 设备判断、日期/字符串/数组工具、Markdown 处理
├── vue.config.js            # 别名、devServer 端口与 proxy、publicPath
├── tailwind.config.js
├── .env.development / .env.production  # API 地址、标题、版本等
└── package.json
```

---

## 4. 路由与页面结构

### 4.1 路由一览

- **`/`** → 重定向到 `/home`
- **`/home`** → 重定向到 `/chat`，布局为 `Home.vue`（侧边栏 + 主内容）
  - 子路由：`/chat`、`/chat/:id`、`/report`、`/single-page/early-screening-technology` 等单页
- **`/login`**：PC 登录
- **`/login/wechat/callback`**：微信登录回调
- **`/mobile/login`**：移动端登录
- **`/mobile`**：移动端壳 `mobile/Home.vue`，默认重定向到 `/mobile/chat/add`
  - 子路由：`/mobile/chat/add`、`/mobile/chat/list`、`/mobile/report`、`/mobile/profile`、`/mobile/single-page/*`
- **`/mobile/chat/:id`**：移动端对话详情（独立路由，无 Tabbar）
- **`/:all(.*)`**：404

### 4.2 PC 与移动端分离策略

- **设备判断**：`utils/libs.js` 中的 `isMobileV2()` 通过 UA 判断是否为移动设备。
- **Index.vue**：若为移动端则直接 `router.push("/mobile/index")`（注：实际路由为 `/mobile`，可核对为 `/mobile`）。
- **Chat.vue**：若为移动端则根据是否有 `chatId` 跳转到 `/mobile/chat/:id` 或 `/mobile/chat/add`。
- **登录后**：根据设备跳转 `/chat` 或 `/mobile`。
- **接口 401**：`http.js` 中根据设备跳转 `/login` 或 `/mobile/login`。

---

## 5. 核心页面与组件设计

### 5.1 首页（Index.vue）

- 顶部：Logo + 登录/注册按钮（未登录时）。
- 主区：标题（`VUE_APP_TITLE`）、slogan、导航卡片（早筛技术、政策指引、星启协爱等），点击跳转对应单页或功能。
- 底部：`FooterBar` 版权信息。
- 进入时 `checkSession`，成功则 `isLogin = true` 隐藏登录按钮。

### 5.2 PC 主布局（Home.vue）

- **左侧边栏**（可折叠）：
  - Logo、折叠按钮。
  - 菜单：我的报告、新建会话。
  - 历史会话列表：滚动加载，调用 `/api/chat/list`，点击项进入 `/chat/:id`。
  - 底部：已登录显示头像、昵称、退出；未登录显示「登录/注册」入口。
- **主内容区**：`<router-view>`，用于渲染 Chat、Report、单页等。
- 通过 `provide('refreshChatList')` 向子组件提供刷新会话列表方法；进入时 `checkSession` 并拉取会话列表。

### 5.3 对话页（Chat.vue）

- **结构**：顶部 `HeaderBar`，中间聊天区域，底部输入区（有会话时显示）或欢迎区底部 `FooterBar`。
- **无会话时**：显示 `Welcome`（欢迎文案 + `ChatInput`），无历史消息。
- **有会话时**：
  - 根据 `chatId` 请求 `/api/chat/detail` 拉取对话与消息列表。
  - 若有最后一条消息且 `completion` 为空，则自动调用 `readChatMessage` 拉取 SSE 流补全内容。
- **消息列表**：每条为「用户消息 + AI 回复」：`ChatPrompt`（用户）+ `ChatReply`（AI）。`ChatReply` 支持「思考中/已完成思考」、loading、`reasoning` 与 `completion` 的 Markdown 渲染与代码高亮。
- **发送流程**：
  1. `ChatInput` 触发 `send`，父组件调用 `createChatMessage`（`/api/chat/create_stream`）。
  2. 若当前无 `chatId`，则创建后跳转到 `/chat/:id` 并刷新会话列表。
  3. 若有 `chatId`，则追加 `message`，并调用 `readChatMessage(messageId)` 建立 SSE。
- **SSE**：使用 `@microsoft/fetch-event-source` 请求 `POST /api/chat/read_stream`，根据 `data.type`（reasoning/answer/end）更新当前条目的 `reasoning`、`completion`，并滚动到底部。
- **复制**：通过 Clipboard 绑定 `.copy-reply`、`.copy-code-btn`，复制成功/失败用 Element Message 提示。

### 5.4 输入组件（ChatInput.vue）

- 多行输入框、占位「输入您的问题...」。
- 可选：医疗指南、研究创新（标签）、从「我的报告」选择或上传文件（`FileSelect` / `mobile/FileSelect`），展示 `FileList`、`ReportList`，支持移除。
- 发送前组装：`prompt`、`chat_id`、`file_ids`、`report_ids` 等，通过 `emit('send', chatItem)` 交给父组件。
- 支持「停止生成」按钮（与后端停止流式接口配合）。

### 5.5 用户消息展示（ChatPrompt.vue）

- 用户输入内容经 `processPrompt` 等处理后可做简单 Markdown 或纯文本展示。
- 若有 `files`、`reports`，展示文件/报告卡片（图标、文件名、类型、大小等），使用 `store/system` 的 `GetFileIcon`、`GetFileType`、`FormatFileSize`。

### 5.6 AI 回复展示（ChatReply.vue）

- Props：`data`（含 `message_id`、`reasoning`、`completion`）、`thinking`、`loading`。
- 顶部状态：「思考中...」/「已完成思考」。
- 若有 `data.reasoning`，先展示思考内容（Markdown）。
- 再展示 `data.completion`（Markdown + 代码高亮），使用 `markdown-it`、`highlight.js`、emoji、mathjax 插件。
- 支持复制代码按钮（与 Clipboard 配合）。

### 5.7 登录页（Login.vue）

- **账号登录**：手机号、区号；验证码登录（配合 `SendMsg` 发验证码）或密码登录；协议勾选；立即登录/切换登录方式。
- **微信登录**：切换为「扫码」后请求 `/api/login/wechat` 获取 `appid`、`redirect_uri`、`state`，使用 `vue-next-wxlogin` 渲染二维码，样式引用 `wechatConfig.href`（如后端提供的 wxlogin.css）。
- 登录成功：写入 `store`（userInfo、isLogin），根据设备跳转 `/chat` 或 `/mobile`。

### 5.8 我的报告（Report.vue）

- Tab：单项报告、综合报告。
- 单项报告下有多类型筛选（问卷量表、面容数据、眼动数据等），筛选条件与 `ReportTable` 的 `type` 联动。
- `ReportTable`：请求 `/api/reports/list`，表格展示报告列表，支持报告文件链接、分页等。

### 5.9 单页（Single*）

- 共用顶部 `HeaderBar`，主内容为图文排版（如 SingleAbout 的「我是谁 / 从哪来 / 做什么」、联系邮箱等）。
- PC 与移动端各有一套 Single* 页面，路由分别为 `/single-page/*` 与 `/mobile/single-page/*`。

### 5.10 移动端

- **mobile/Home.vue**：底部 `van-tabbar` 四个入口：新建会话、历史会话、我的报告、更多；主区为 `router-view`。
- **mobile/AddChat.vue**：欢迎区 + `ChatInput`，发送后调用 `/api/chat/create_stream` 再跳转 `/mobile/chat/:id`。
- **mobile/Chat.vue**：与 PC 类似的对话详情与 SSE 流式展示。
- **mobile/ChatList.vue**：历史会话列表（来自 `/api/chat/list`）。
- **mobile/Report.vue**：我的报告移动版。
- **mobile/Profile.vue**：更多（主题、单页入口、关于我们等）。
- **mobile/Login.vue**：移动端登录表单/流程。

---

## 6. 状态管理与数据流

- **Pinia**
  - **sharedata**：`showLoginDialog`、`theme`、`isLogin`、`userInfo`（与 good-storage 持久化），提供 `setUserInfo`、`setIsLogin`、`setTheme`、`setShowLoginDialog`。
  - **theme**：与 localStorage 同步的 `theme`，用于设置 `data-theme`。
  - **cache**：非 store 模块，提供 `checkSession()`（请求 `/api/users/profile`）、`getClientId()`。
- **登录态**：进入 Home/Index 时调用 `checkSession()`，成功则更新 `store.isLogin` 和 `store.userInfo`，并拉取会话列表等。
- **接口约定**：统一使用 `utils/http.js` 的 `httpPost`/`httpGet`，baseURL 为空，依赖 vue.config 的 proxy 或生产环境同域；`withCredentials: true` 携带 Cookie/Session；401 时前端跳转登录页。

---

## 7. 样式与主题

- **全局**：`main.js` 中按顺序引入 `theme-dark.styl`、`theme-light.styl`、`common.styl`；`App.vue` 中定义 `--primary-color`、标题字号、省略号等；根节点 `data-theme` 由 theme store 或 sharedata 的 `theme` 控制。
- **主题**：通过 `document.documentElement.setAttribute("data-theme", theme)` 切换，样式文件中通过 `[data-theme="dark"]` 等选择器区分亮色/暗色。
- **布局与页面样式**：各页面通过 `<style lang="stylus" scoped>` 引用 `assets/css` 下对应 styl 文件（如 `home.styl`、`chat.styl`、`login.styl`、`report.styl`、`mobile/*.styl`）。
- **Markdown**：使用 `assets/css/markdown/vue.css` 及 highlight.js 的 a11y-dark 等主题，保证代码块与排版一致。

---

## 8. 与后端接口对接

- **基础**：开发环境通过 `vue.config.js` 的 `devServer.proxy` 将 `/api` 指向 `VUE_APP_API_HOST`；生产环境需保证前端与后端同域或配置 CORS，且 `axios.defaults.withCredentials = true` 保持不变。
- **主要接口**：
  - 登录：`POST /api/login/send-code`、`/api/login/mobile`、`/api/login/mobile-pwd`、`/api/login/wechat`（获取扫码参数）、回调页为 `/login/wechat/callback`。
  - 用户：`POST /api/users/profile`、`POST /api/logout`。
  - 对话：`POST /api/chat/list`、`/api/chat/detail`、`/api/chat/create_stream`、`POST /api/chat/read_stream`（SSE）、`/api/chat/stop_stream`、`/api/chat/update`、`/api/chat/delete`。
  - 文件：`POST /api/files/upload`、`/api/files/detail`。
  - 报告：`POST /api/reports/list`；管理端上传报告等为 `/api/admin/*`。
- **流式协议**：`read_stream` 返回 SSE，每条 `data` 为 JSON，包含 `type`（如 start、reasoning、answer、end、stop、error）和 `content`，前端按类型更新当前消息的 `reasoning`/`completion` 或结束/错误处理。

---

## 9. 开发与构建配置

- **环境变量**：`.env.development` / `.env.production` 中配置 `VUE_APP_API_HOST`、`VUE_APP_TITLE`、`VUE_APP_VERSION` 等；登录页或调试用可配置 `VUE_APP_USER`、`VUE_APP_PASS`（勿提交真实密码）。
- **运行**：`npm install` 后 `npm run dev`，默认访问 http://localhost:8888。
- **构建**：`npm run build`，输出到 `dist/`，`publicPath: "/"`，生产环境可改为子路径或 CDN 路径。
- **其它**：`lintOnSave: false`、`productionSourceMap: false`；路径别名 `@` -> `src`；ResizeObserver 在 `App.vue` 中做了 200ms 节流以减轻性能开销。

---

## 10. 小结

`start-love-web` 采用 **Vue 3 + Pinia + 双端 UI（Element Plus / Vant）** 的 SPA 架构，通过路由与设备判断实现 PC 与移动端两套入口与布局；对话页与 **SSE 流式接口** 紧密配合，实现「创建会话 → 发送消息 → 读流更新」的完整流程；登录、报告、单页等信息类页面与后端 REST 接口一一对应。扩展时可继续在 `views`、`components` 下按现有分层添加页面与组件，并在 `router.js` 与 `http.js` 中增加路由与接口封装。
