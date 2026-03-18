## AsdProject 前后端交互与使用说明

本说明文档基于：

- 后端：`asd_backend/asd_backend/ARCHITECTURE.md`
- 前端：`start-love-web/ARCHITECTURE.md`

对 **前后端交互方式** 以及 **本地/生产使用方式** 做一个统一概览，方便开发与部署。

---

### 1. 整体架构与角色分工

- **后端 (`asd_backend`)**
  - 技术栈：Go + Beego v2 + PostgreSQL + Redis + 腾讯云 COS + 阿里云百炼/RAG 服务。
  - 职责：
    - 用户认证（短信验证码、手机号密码、微信扫码）与 Session 管理
    - 对话（Chat）管理：会话列表、消息记录、流式聊天（SSE）
    - 文件上传与解析（通过 `markitdown` 抽取内容）、COS 下载链接
    - 诊断报告管理：列表、详情、RAG 相关数据源
    - 操作日志、中间件鉴权、配置加载、数据库初始化等

- **前端 (`start-love-web`)**
  - 技术栈：Vue 3 + Vue Router 4 + Pinia + Element Plus（PC）+ Vant 4（移动端）+ Stylus + Tailwind。
  - 职责：
    - 提供 PC 端与移动端两套 UI（路由与布局分离）
    - 实现登录、对话、报告、文件上传、信息单页等交互界面
    - 通过 Axios 与后端 `/api` 接口通信，SSE 流式消费对话输出
    - 管理用户登录态、主题、客户端 ID 等前端状态

前后端之间以 **HTTP/HTTPS + JSON** 为主、辅以 **SSE（Server-Sent Events）** 实现流式回复。

---

### 2. 网络与环境配置

#### 2.1 后端服务地址

后端默认监听 HTTP 端口来自：

- 启动参数：`-port`（在 `asd_backend/boot/init/init.go` 中解析）
- 或 Beego 配置 `httpport`（如未传入 `-port`）

常见场景：

- 本地开发：建议监听 `http://localhost:5678`（可自行指定）
- 生产环境：通过域名（示例：`https://asd.evotrek.cn`）对外暴露

数据库、Redis、COS、RAG 等连接参数配置在：

- `asd_backend/asd_backend/conf/config.yaml`

#### 2.2 前端环境变量与代理

前端运行环境：

- 开发模式下（`npm run dev`）：
  - 端口：`http://localhost:8888`
  - 通过 `vue.config.js` 里的 `devServer.proxy` 将 `/api` 与 `/static/upload/` 代理到后端：

    ```js
    proxy: {
      "/static/upload/": {
        target: process.env.VUE_APP_API_HOST,
        changeOrigin: true,
      },
      "/api": {
        target: process.env.VUE_APP_API_HOST,
        changeOrigin: true,
      },
    }
    ```

  - `.env.development` 中配置后端地址：

    ```ini
    VUE_APP_API_HOST=http://localhost:5678
    # 开发时建议改为本地后端地址
    ```

- 生产模式下（`npm run build`）：
  - 默认 `publicPath: "/"`，构建产物部署后，前端直接通过相对路径访问 `/api/...`，需与后端同域或正确配置 CORS。

Axios 全局配置（`start-love-web/src/utils/http.js`）：

- `axios.defaults.baseURL = ""`：依赖相对路径与反向代理
- `axios.defaults.withCredentials = true`：前端会携带 Cookie，用于后端 Session 鉴权

---

### 3. 认证与会话机制

#### 3.1 登录流程（后端接口）

后端在 `asd_backend/routers/api.go` 中定义登录相关路由：

- `POST /api/login/send-code`：发送短信验证码
- `POST /api/login/mobile`：手机号 + 验证码登录
- `POST /api/login/mobile-pwd`：手机号 + 密码登录
- `GET  /api/login/wechat`：微信扫码登录 URL（或参数）
- `POST /api/login/wechat`：返回前端所需微信登录参数（appid、state 等）
- `GET/POST /api/login/wechat/callback`：微信回调，拿 code 换 token 与用户信息

登录成功后，后端会在 Session 中写入：

- `conf.USER_ID`（键名 `userId`）—— 当前登录用户 ID

中间件 `CheckApi` 用于对 `/api/*` 路径做统一鉴权：

- `/api/login*`：放行
- `/api/admin*`：校验 `Authorization: Bearer xxx`
- 其它 `/api/*`：若 Session 中无 `USER_ID`，返回 `401 Unauthorized`

#### 3.2 前端登录使用方式

PC 端登录页：`start-love-web/src/views/Login.vue`，支持两种方式：

- **验证码登录**：
  - 点击发送验证码：通过 `SendMsg` 组件调用 `POST /api/login/send-code`
  - 登录请求：`POST /api/login/mobile`，Body 包含 `mobile` 与 `passcode`
- **密码登录**：
  - 登录请求：`POST /api/login/mobile-pwd`，Body 包含 `mobile` 与 `password`

微信扫码登录：

- 切换到二维码登录后，调用 `POST /api/login/wechat` 获取：
  - `appid`、`redirect_uri`、`state`、`href` 等参数
- 使用 `vue-next-wxlogin` 渲染二维码，用户扫码后浏览器跳转至 `/login/wechat/callback`，后端处理并在 Session 写入用户，最后重定向到 `/`。

登录成功后，前端会：

- 将用户信息写入 Pinia store（`useSharedStore`）与 `good-storage`：
  - `store.setUserInfo(res.data)`
  - `store.setIsLogin(true)`
- 根据设备类型跳转：
  - PC：`/chat`
  - 移动端：`/mobile`

#### 3.3 登录态校验与 401 处理

前端在 `http.js` 的响应拦截器中处理 `401`：

- 若返回 `401`，则：
  - PC：重定向到 `/login`
  - 移动端：重定向到 `/mobile/login`

进入主布局或首页时，会调用：

- `checkSession()`（`store/cache.js`），内部 `POST /api/users/profile`：
  - 若成功，表示已有有效 Session，前端设置 `store.isLogin = true` 并填充用户信息
  - 若失败，保持未登录状态（部分页面会引导到登录）

登出接口：

- `POST /api/logout`：清除 Session 中的 `USER_ID`
- 前端在 `Home.vue` 中调用该接口，成功后清空本地用户信息并跳转登录页

---

### 4. 对话（Chat）交互流程

#### 4.1 主要后端接口

位于 `asd_backend/routers/api.go`：

- `POST /api/chat/list`：获取当前用户的对话列表
- `POST /api/chat/detail`：获取指定对话详情（含消息列表）
- `POST /api/chat/update`：更新对话标题
- `POST /api/chat/delete`：删除对话
- `POST /api/chat/create_stream`：创建流式对话（生成一条新消息并异步调用大模型）
- `POST /api/chat/read_stream`：读取流式对话结果（SSE）
- `POST /api/chat/stop_stream`：停止当前流式生成

所有接口均基于 Session 的 `USER_ID` 做权限校验。

#### 4.2 前端调用方式

##### 4.2.1 PC 端 Chat 核心组件

- 页面：`start-love-web/src/views/Chat.vue`
- 输入组件：`components/ChatInput.vue`
- 显示组件：`components/ChatPrompt.vue`、`components/ChatReply.vue`

**创建/发送消息：**

1. 用户在 `ChatInput` 输入问题，可附加文件与报告：
   - 通过 `FileSelect` 选择已有报告或上传文件（下文详述）
2. `ChatInput` 在点击「发送」时，发出 `send` 事件，携带 `chatItem`（含 `prompt`、`file_ids`、`report_ids` 等）。
3. 父组件 `Chat.vue` 的 `sendMessage`：
   - 将当前路由参数中的 `chatId` 写入 `chatItem.chat_id`
   - 调用 `createChatMessage(chatItem)`：

      ```js
      httpPost("/api/chat/create_stream", chatItem)
        .then((res) => {
          // 若已有 chat_id：追加消息并启动读流
          // 若没有 chat_id：刷新列表并跳转新对话
        })
      ```

4. 若当前已有对话（`chatId` 非空）：
   - 直接在本地 `messages` 中追加返回的 `message`，调用 `readChatMessage(message_id)` 建立 SSE 连接。
5. 若是新建对话（`chatId` 为空）：
   - 调用 `refreshChatList()` 重新加载侧边栏对话列表
   - 跳转到 `/chat/{chat_id}`，由新页面加载对话详情并拉起流。

**流式读取（SSE）：**

- 在 `readChatMessage` 中使用 `fetchEventSource` 调用：

  - URL：`/api/chat/read_stream`
  - Method：`POST`
  - Body：`{ chat_id, message_id }`（JSON）
  - Headers：`Content-Type: application/json`，`Accept: text/event-stream`
  - `credentials: 'include'` 以携带 Cookie

- 服务器端返回 SSE 事件，`onmessage` 中解析 `event.data`，根据 `data.type` 更新当前消息：\n
  - `reasoning`：累加到思考区域（灰色小字）\n
  - `answer`：累加到最终回答内容\n
  - `end`：结束一次对话，清空缓冲

**停止对话：**

- 前端在 `ChatInput` 中暴露「停止生成」按钮，调用 `/api/chat/stop_stream`（Body：`chat_id`, `message_id`）：
  - 后端在 Redis 中打上 `stop` 标记，整理已生成的文本，并将最终内容写入数据库。
  - 前端收到 `type: "stop"` 事件后结束读流。

##### 4.2.2 移动端对话

- 移动端使用独立视图（`views/mobile/Chat.vue`、`views/mobile/AddChat.vue`），但对话接口与 PC 端完全一致：
  - 创建：`/api/chat/create_stream`
  - 读取：`/api/chat/read_stream`（同样使用 SSE）
  - 停止：`/api/chat/stop_stream`

区别仅在于 UI 与路由结构（Tabbar 与页面布局）。

---

### 5. 文件与报告交互

#### 5.1 文件上传

后端接口（`api_file.go`）：

- `POST /api/files/upload`：
  - 表单字段名：`file`（Content-Type: `multipart/form-data`）
  - 后端使用 `gfile.FileUpload` 保存文件，生成 `file_id` 等信息
  - 调用 `markitdown` 命令行工具从文件抽取 Markdown 形式的内容存入数据库
  - 返回文件详情（`file_id`、`file_name`、`content`、`download_url` 等）

前端调用方式（`components/FileSelect.vue`）：

- `el-upload` 配置 `http-request="afterRead"`，在 `afterRead` 中：

  ```js
  const formData = new FormData();
  formData.append("file", file.file, file.name);
  httpPost("/api/files/upload", formData)
    .then((res) => {
      // res.data 即为文件详情，回传给父组件以添加到当前对话的文件列表
    })
  ```

- 选择完成后通过 `uploadFile` 事件将文件信息传回 `ChatInput`，最终一起提交给 `/api/chat/create_stream`。

#### 5.2 报告列表与选择

后端接口（`api_report.go`）：

- `POST /api/reports/list`：
  - 请求体包含 `offset`, `limit`, `year`, `type` 等筛选参数
  - 当前 demo 中，将 `userId` 强制为 `1`（全局展示示例报告）
  - 返回 `list`（报告数组）与 `count`（总数）

前端使用场景：

- **我的报告页（`views/Report.vue` + `components/ReportTable.vue`）**
  - 根据报告类型（单项/综合）与筛选条件，调用 `/api/reports/list` 填充表格
  - 为 `report_file_url`、`original_file_url` 提供下载链接

- **对话中从报告中选择（`components/FileSelect.vue`）**
  - 打开弹窗，按年份筛选，分页调用 `/api/reports/list`
  - 用户选中报告后，通过 `selectReport(report)` 事件将报告对象传给 `ChatInput`
  - `ChatInput` 将报告 ID 列入 `report_ids`，提交给 `/api/chat/create_stream`，后端在构造 Prompt 时将报告内容拼入，从而实现结合报告的智能对话。

---

### 6. 开发与运行指南（前后端联调）

#### 6.1 后端启动

1. 安装依赖与 Bee 工具（可选）：
   - 参考 `asd_backend/asd_backend/README.md`：
     - `go install github.com/beego/bee/v2@latest`
     - `go mod tidy`
2. 初始化数据库（MySQL 或 PostgreSQL），执行 README 中给出的 SQL 语句（至少 PostgreSQL 部分）。
3. 配置 `conf/config.yaml`（如有需要，修改数据库、Redis、RAG 地址等）。注意避免提交真实密钥到仓库。
4. 在 `asd_backend/asd_backend` 目录运行：

   ```bash
   go run main.go -mode=api -port=5678
   ```

5. 确认后端已在 `http://localhost:5678` 提供 `/api` 接口。

#### 6.2 前端启动

1. 在 `start-love-web` 目录安装依赖：

   ```bash
   npm install
   ```

2. 配置 `.env.development`，将 `VUE_APP_API_HOST` 指向后端地址，如：

   ```ini
   VUE_APP_API_HOST=http://localhost:5678
   VUE_APP_TITLE="星启诶艾"
   VUE_APP_VERSION=v0.0.1
   ```

3. 启动前端开发服务器：

   ```bash
   npm run dev
   ```

4. 在浏览器访问：
   - PC：`http://localhost:8888`
   - 移动端：可使用移动设备访问同一地址，或使用浏览器的移动模式，路由自动切换到 `/mobile`。

---

### 7. 生产部署建议

1. 后端部署：
   - 将 `asd_backend` 构建为二进制或以容器形式部署到服务器。
   - 配置 `conf/config.yaml` 中的数据库、Redis、COS、RAG、外部 API Key 为生产值（通过环境变量或秘密管理工具）。
   - 在前置 Nginx/网关上代理 `/api`、`/static/upload` 到后端服务。

2. 前端构建与部署：
   - 在 `start-love-web` 中执行：

     ```bash
     npm run build
     ```

   - 将生成的 `dist/` 目录部署到 Web 服务器（可与后端域名相同，或通过 Nginx 做反向代理）。
   - 确保生产环境中 `/api` 的代理指向后端，且支持 Cookie（`withCredentials`）。

3. 域名与 HTTPS：
   - 推荐在网关层绑定域名并启用 HTTPS。
   - 如与后端同域，可简化 CORS 问题；如跨域，则需在后端或网关显式开启 CORS 并允许携带凭证。

---

### 8. 快速功能测试示例

**1）登录与会话：**

1. 打开前端 `http://localhost:8888`，点击「登录/注册」进入登录页。
2. 使用配置在 `.env.development` 中的测试账号（如 `VUE_APP_USER` / `VUE_APP_PASS`）进行短信或密码登录：
   - 验证码登录：先「发送验证码」→ 填入验证码 → 提交。
   - 密码登录：切换到密码登录 → 输入手机号和密码 → 提交。
3. 登录成功后应被跳转到 `/chat` 或 `/mobile`，左侧显示历史会话或空列表。

**2）新建对话（PC）：**

1. 在 `/chat` 页面，初次进入会看到欢迎界面 `Welcome`。
2. 在输入框输入问题，点击发送。
3. 页面会自动创建新对话并开始流式显示 AI 回复；侧边栏将出现新对话条目。

**3）从报告中提问：**

1. 在对话输入框点击「从我的报告中添加」（`FileSelect` 图标）。
2. 在弹窗中选择某一年份与报告记录，点击「选择」。
3. 发送问题时，后端会将报告内容拼接到 Prompt 中，AI 结合报告内容回答。

---

本文件旨在帮助你快速理解 **AsdProject 的前后端交互模式** 以及 **如何在本地/生产环境中正确使用这些接口**。更详细的架构与代码分层说明，请分别参考：`asd_backend/asd_backend/ARCHITECTURE.md` 与 `start-love-web/ARCHITECTURE.md`。

