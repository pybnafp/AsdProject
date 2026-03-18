## Asd 后端服务架构说明（asd_backend）

### 1. 项目概览

`asd_backend` 是 AsdProject 的后端服务，使用 **Go 语言 + Beego v2** 框架开发，主要提供：

- **用户登录与会话管理**（短信验证码登录、手机号密码登录、微信扫码登录）
- **对话（Chat）管理**：创建对话、消息记录、流式对话（SSE）
- **文件管理**：文件上传、内容提取（通过外部 `markitdown` 工具）、COS 存储与下载链接
- **诊断报告（Report）管理**：报告上传、列表查询、详情获取
- **操作日志与权限校验**：中间件统一处理登录态与操作日志

整体采用典型的 **分层架构**：

- **路由层 (`routers`)**：定义 URL 与控制器方法映射
- **控制器层 (`app/controllers`)**：解析请求、参数校验、调用服务层，返回 JSON 或重定向
- **服务层 (`app/services`)**：封装业务逻辑，操作模型层和外部服务
- **模型层 (`app/models`)**：基于 Beego ORM 的数据库实体及 CRUD 方法
- **DTO/VO 层 (`app/dto`, `app/vo`)**：请求与响应的数据结构
- **中间件层 (`app/middleware`)**：登录校验、数据过滤、操作日志、模板函数
- **基础设施与工具层 (`boot`, `conf`, `utils`, `constant`)**：配置加载、数据库初始化、Session 设置、Redis、COS、短信、JWT 等

### 2. 技术栈与依赖

- **Web 框架**：`github.com/beego/beego/v2`
- **ORM**：Beego ORM（当前使用 PostgreSQL 驱动）
- **配置管理**：`spf13/viper` + YAML 配置文件
- **数据库**：PostgreSQL（`conf.CONFIG.Postgres`），同时预留 MySQL 配置
- **缓存与队列**：Redis（流式对话结果存储）
- **对象存储**：腾讯云 COS（文件与报告上传/下载）
- **外部 AI 服务**：阿里云百炼（对话大模型与用量统计）、RAG 服务（`apiconfig.rag_urls`）
- **其它**：短信服务、JWT 工具、自定义 `markitdown` 命令行工具用于文件内容抽取

### 3. 目录结构（后端部分）

主要关注 `asd_backend/asd_backend` 目录：

- `main.go`：程序入口
- `boot/`
  - `boot/config/config.go`：使用 Viper 加载 YAML 配置到全局 `conf.CONFIG`
  - `boot/init/init.go`：解析命令行参数 `-mode`、`-port`，设置运行模式与监听端口
  - `boot/postgres/postgres.go`：注册 PostgreSQL 数据库、配置连接池、自动建表（`orm.RunSyncdb`）
  - `boot/session/session.go`：配置 Session 过期时间与 Cookie 生命周期
- `conf/`
  - `conf/config.go`：定义全局配置结构体 `Config`（Mysql/Postgres/Redis/附件/系统/腾讯云/API 配置）
  - `conf/common.go`：全局变量 `CONFIG`、用户 Session Key（`USER_ID`）与 `RunMode`
  - `conf/config.yaml`：YAML 配置文件（数据库、Redis、文件上传、系统参数、腾讯云与百炼配置等）
- `routers/`
  - `routers/router.go`：初始化数据权限过滤、API 登录校验中间件、自定义模板函数
  - `routers/api.go`：定义所有 `/api/*` 路由与对应控制器方法
- `app/controllers/`
  - `api_user.go`：用户登录相关接口（短信验证码登录、手机号密码登录、微信扫码登录、用户信息、注销等）
  - `api_chat.go`：对话与消息接口（列表、详情、创建/更新/删除对话、流式对话创建/读取/停止）
  - `api_file.go`：文件上传、查询、删除接口
  - `api_report.go`：诊断报告列表接口
  - `api_admin.go` / `backend.go`：管理端与后台相关接口（如报告导入、RAG 状态等）
- `app/services/`
  - `chat.go`：对话增删改查、详情与用量统计 `UpdateUsageStats`
  - `chat_message.go`：消息增删改查、最近消息获取、消息详情与历史消息构造
  - `file.go`：文件列表、详情、添加、删除、状态变更、批量获取下载链接与详情
  - `report.go`：报告列表、详情、新增报告、批量查询报告
  - 其他：`bailian.go`（调用百炼大模型）、`rag.go`（与 RAG 服务交互）、`redis.go/redis_writer.go/utils.go` 等
- `app/models/`
  - `user.go`、`chat.go`、`chat_message.go`、`file.go`、`report.go`、`message_usage_stats.go` 等 ORM 实体
- `app/dto/`
  - `user.go`、`chat.go`、`chat_message.go`、`file.go`、`report.go`、`usage_stats.go` 等请求与内部传输结构
- `app/vo/`
  - `user.go`、`chat.go`、`chat_message.go`、`file.go`、`report.go`、`rag.go`、`oper_log.go` 等对外返回视图对象
- `app/middleware/`
  - `api.go`：登录与管理端鉴权中间件
  - `datafilter.go`：数据权限过滤器，将 `userId` 注入到上下文
  - `operlog.go`：操作日志中间件，记录修改类操作（add/update/delete 等）
  - `template.go`：自定义模板函数注入（主要用于页面模板场景）
- `utils/`
  - `utils/redis.go`：Redis 客户端封装
  - `utils/tencent/*`：COS 客户端与下载 URL 生成
  - `utils/sms/*`：短信发送与验证码校验
  - `utils/common/*`：JWT、UUID、通用工具、JSON & DataFilter 辅助方法
  - `utils/gfile/*`：封装 Beego 文件上传逻辑
  - `utils/gmd5/gmd5.go`、`utils/gconv/gconv.go`、`utils/gstr/gstr.go`、`utils/gregex/gregex.go` 等通用工具
- `constant/`
  - `constant.go`：配置环境变量名、配置文件默认路径等

### 4. 启动流程（程序生命周期）

以 `main.go` 为入口，程序整体启动流程如下：

1. **导入与初始化阶段（Go 包级 `init` 调用）**
   - `main.go` 中通过匿名导入 `_ "asd/boot/config"`, `_ "asd/boot/init"`, `_ "asd/boot/postgres"`, `_ "asd/boot/session"`, `_ "asd/routers"`，触发各包的 `init()`：
     - `boot/config`：使用 Viper 读取 `conf/config.yaml`（或环境变量 `CONFIG` 指定的路径），反序列化到全局 `conf.CONFIG`，并监听文件变化进行热更新。
     - `boot/init`：解析命令行参数：
       - `-mode`：运行模式（`api`, `job`, `console`），写入 `conf.RunMode`
       - `-port`：HTTP 监听端口，默认读取 `beego.AppConfig` 中 `httpport`，然后赋值给 `beego.BConfig.Listen.HTTPPort`
     - `boot/postgres`：
       - 注册 PostgreSQL 驱动与默认数据库（使用 `conf.CONFIG.Postgres.Default`）
       - 配置连接池（例如 `SetConnMaxLifetime(2 * time.Hour)`）
       - `orm.RunSyncdb` 自动建表（基于 `app/models` 中注册的模型）
       - Debug 模式下开启 SQL 日志
     - `boot/session`：配置 Session 过期时间与 Cookie 生命周期（默认 7200 秒）
     - `routers/router.go`：注册全局中间件（数据过滤、API 登录验证、模板函数初始化）
     - `routers/api.go`：注册所有 `/api/*` 路由与对应控制器方法

2. **主函数执行**
   - 在 `main()` 中：
     - 配置全局日志输出到 `global.log` 文件，并根据 `RunMode` 设置日志级别（`dev`：Debug，其它：Info）
     - 打印当前运行模式与端口：`logs.Info("run at mode=%v, port=%v", conf.RunMode, beego.BConfig.Listen.HTTPPort)`
     - 调用 `beego.Run()` 启动 HTTP 服务器并开始监听请求

### 5. 请求处理整体流程

以一个典型的 API 请求（例如 `POST /api/chat/create_stream`）为例，整体处理链路如下：

1. **请求进入 Beego 路由与中间件**
   - 全局过滤器：
     - `AddDataFitler()`：对所有路径 `/*` 生效，将 Session 中的 `userId` 注入到上下文数据中，供后续使用。
     - `CheckApi()`：对 `/api/*` 路径生效：
       - `/api/login*`：登录相关接口放行，不做登录校验
       - `/api/admin*`：检查 `Authorization` 头中是否存在特定 Bearer Token，用于简单的管理端鉴权
       - 其它 `/api/*`：使用 Session 中的 `conf.USER_ID` 判断是否登录，未登录返回 `401`
     - `OperLog()`：在 `AfterExec` 阶段执行，对非 GET、且路径包含 `/update`、`/delete`、`/add` 等关键字的请求记录操作日志

2. **路由匹配**
   - 例如：`beego.Router("/api/chat/create_stream", &controllers.ChatApiController{}, "post:CreateStream")`
   - Beego 将请求分发到对应控制器方法（如 `ChatApiController.CreateStream`）

3. **控制器处理**
   - 控制器统一继承 `BaseController`：
     - 负责 JSON 解析：`c.ParseJSON(&req)`
     - 参数校验错误通过 `c.ErrorJson(status, msg)` 返回
     - 提供 `c.GetUserId()` 等辅助方法，从中间件注入的上下文/Session 中获取当前用户
   - 业务入口示例：
     - 用户接口：`UserApiController` 处理登录、登出、个人信息等
     - 对话接口：`ChatApiController` 负责对话列表、详情、流式对话创建/读取/停止
     - 文件接口：`FileApiController` 处理文件上传、详情、删除
     - 报告接口：`ReportApiController` 提供报告列表与查询

4. **服务层调用与数据库/外部服务访问**
   - 控制器调用服务层（如 `services.Chat`, `services.ChatMessage`, `services.FileService`, `services.Report`），服务层中：
     - 通过 Beego ORM (`orm.NewOrm()`) 操作数据库模型
     - 调用工具类处理 COS、短信、JWT、Redis 等
     - 调用百炼大模型与 RAG 服务，处理流式对话并将结果写入 Redis/数据库

5. **响应构建与返回**
   - 大多数接口以统一的 JSON 结构返回（`common.JsonResult` 或 `common.JsonRes`），包含：
     - `Code`：业务状态码（0 表示成功）
     - `Msg`：提示信息
     - `Data`：数据主体
     - `Count`：分页场景下的总数
   - 流式接口（如 `ReadStream`）通过 **Server-Sent Events (SSE)** 的方式不断向客户端推送 `data: ...\n\n` 事件。

### 6. 关键业务流程示例

#### 6.1 用户登录流程

- **短信验证码登录**：`POST /api/login/send-code` + `POST /api/login/mobile`
  - `SendSmsCode`：
    - 解析请求（手机号）
    - 校验手机号格式（`sms.IsValidMobile`）
    - 生成 6 位随机验证码并通过短信服务发送
  - `MobileLogin`：
    - 校验验证码（`sms.VerifySmsCode`）
    - 根据手机号查找/创建用户（`models.User`）
    - 将用户 ID 写入 Session（键为 `conf.USER_ID`）
    - 返回用户昵称与头像信息

- **手机号密码登录**：`POST /api/login/mobile-pwd`
  - 如用户不存在则自动注册并设置默认密码（MD5 存储）
  - 校验输入密码的 MD5 是否匹配
  - 登录成功后同样写入 Session

- **微信扫码登录**：`GET/POST /api/login/wechat*`
  - `WechatLoginUrl` / `WechatLoginParams`：
    - 从配置中读取 `WechatAppID`、`WechatRedirectURI`
    - 生成随机 `state`，存入 Session（`WECHAT_LOGIN_STATE`）
    - 构造微信授权 URL 或返回前端所需参数
  - `WechatLoginCallback`：
    - 验证 `code`、`state` 一致性防止 CSRF
    - 调用微信 OAuth 接口换取 `access_token` 和 `openid`
    - 再根据 `access_token + openid` 获取用户信息
    - 在本地用户表中查找或创建用户，写入 Session 后重定向到前端页面

#### 6.2 对话与流式聊天流程

- **创建/更新/删除对话**
  - `List`：分页查询当前用户的对话列表（按创建时间倒序），返回简要信息 `ChatVo`
  - `Update`：根据 `ChatID` 更新标题，包含权限校验（只能操作自己的对话）
  - `Delete`：软删除对话（`mark=0`），同时调用 `ChatMessage.DeleteByChatID` 批量删除该对话的消息

- **流式对话创建：`POST /api/chat/create_stream`**
  - 步骤简要：
    1. 控制器解析 `StreamChatReq`，校验参数。
    2. 如果 `ChatID` 为空，自动创建新对话（标题默认使用截断后的 `Prompt`）。
    3. 校验对话是否属于当前用户。
    4. 生成 `MessageID` 并插入消息表（仅含用户输入 Prompt）。
    5. 生成 Redis 键名（`GenerateChatRedisKeys`），检查是否已有流式结果。
    6. 根据最近若干条历史消息，构造百炼模型的对话上下文（`Messages`）。
    7. 如关联了文件或报告，则拉取其内容并拼接到 `Prompt` 中。
    8. 将完整 Prompt 写回消息表的 `RawPrompt` 字段。
    9. 启动后台协程：
       - 创建 RedisResponseWriter，将百炼的流式输出实时写入 Redis 列表
       - 请求完成后更新消息的 `Completion` 与 `Reasoning` 字段
       - 同步更新使用统计（`UpdateUsageStats`），包括 `total_tokens`、`cost` 等信息
  - 接口返回 `chat_id`、`message_id` 以及当前消息详情。

- **流式结果读取：`POST /api/chat/read_stream`**
  - 使用 Server-Sent Events (SSE)：
    1. 设置响应头 `Content-Type: text/event-stream` 等
    2. 先发送一个 `start` 事件（包含 `chat_id`、`message_id`）
    3. 循环轮询 Redis 列表，将新的 JSON 片段（`type` + `content`）不断写入 SSE 流
    4. 当检测到 `type=="end"` 或 `type=="stop"` 时结束循环
    5. 若 Redis 中没有新数据但数据库已有完整 `Completion` 内容，则直接一次性将推理过程和回答内容写出并结束

- **停止流式对话：`POST /api/chat/stop_stream`**
  - 校验对话与消息归属。
  - 在 Redis 中写入停止标记，并追加一条 `type=="stop"` 的事件。
  - 从 Redis 中汇总已生成的 `reasoning` 和 `answer` 内容，更新消息表的 `Completion` 与 `Reasoning` 字段。

#### 6.3 文件上传与内容提取

- **上传文件：`POST /api/files/upload`**
  - 通过 `gfile.FileUpload` 把表单文件保存到服务器（并返回路径、大小、类型等信息）。
  - 生成 `FileAddReq` 填充基础元数据。
  - 使用外部命令 `markitdown <TmpFilePath>` 将文件内容转换为 Markdown 文本并保存到 `Content` 字段（若失败则仅存元数据）。
  - 调用 `FileService.Add` 将记录插入数据库。
  - 返回文件详情（包含下载 URL 与内容）。

#### 6.4 报告管理

- **报告列表：`GET/POST /api/reports/list`**
  - 按用户 ID、报告类型、年份等筛选，返回 `ReportVo` 列表。
  - 为原始报告文件与解析后的报告文件生成可访问的下载 URL。
  - 后续对话中可将报告内容联动到大模型 Prompt 中，用于诊断场景的 RAG。

### 7. 配置与环境说明（概念性）

- 配置文件位于 `conf/config.yaml`，通过 `boot/config` 加载。
- 主要配置项（不包含真实密钥，只说明作用）：
  - `mysql` / `postgres`：数据库连接信息（地址、端口、库名、账户、密码）。
  - `redis`：Redis 主机、端口、密码、DB 索引。
  - `attachment`：文件上传基础路径、大小限制、允许后缀等。
  - `systemconfig`：版本号、调试开关、图片域名等。
  - `tencentconfig`：腾讯云 COS 与微信登录所需的 ID/Secret。
  - `apiconfig`：
    - `file_cache_path`：文件缓存目录
    - `jwt_exp`：JWT 过期时间
    - `use_rag`：是否启用 RAG
    - `rag_urls`：RAG 服务地址列表
    - 百炼、OpenRouter 等外部 API 所需的 Key/URL

> 建议在生产环境中使用环境变量或独立的配置文件管理敏感信息，不在代码仓库中提交真实密钥。

### 8. 本项目的扩展与二次开发建议

- **新增业务模块**：
  - 按现有分层模式新增：`dto`（请求）、`vo`（响应）、`models`（实体）、`services`（业务）、`controllers`（控制器）、`routers/api.go` 中注册路由。
  - 如果涉及登录态或权限控制，可以直接复用 `app/middleware` 中的中间件逻辑。

- **接入新的外部模型或服务**：
  - 建议参考 `services.BailianService` 与 `services.Rag` 的封装方式，新建独立的 service 与配置节。
  - 流式场景可以复用当前 Redis + SSE 模式，将模型输出统一写入 Redis，再由 `ReadStream` 等接口消费。

- **前后端联调**：
  - 所有对外 REST 接口都以 `/api/*` 开头，统一使用 JSON 请求与响应。
  - 对话流式接口基于 SSE，前端只需要监听 `message` 事件并解析 JSON 载荷中的 `type` 与 `content` 字段即可。

以上即为 AsdProject 中 `asd_backend` 后端服务的整体架构、代码结构与主要执行流程说明，后续可以根据实际需要继续补充具体模型字段与接口文档。

