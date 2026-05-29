# navigation

一个轻量级个人导航/书签站。后端使用 Go 1.22 提供 HTTP API，数据存储在 SQLite 中，前端使用 Vue 3、Vite、TypeScript、Tailwind CSS 组织源码，并通过 Go embed 打包 `web/dist` 构建产物进二进制。

## 功能

- 单用户登录保护，首次启动生成随机初始密码。
- 管理导航站点：新增、编辑、删除、搜索。
- 按分类浏览站点，并支持删除分类。
- 提供站点数量、分类数量、分类统计等接口。
- 支持修改账号密码、站点标题、首页文案和主题。
- 首次启动时可从 `data/sites.json` 导入旧数据到 SQLite。
- 支持本地运行、二进制构建和 Docker 部署。

## 技术栈

- Go 1.22
- SQLite，驱动为 `github.com/mattn/go-sqlite3`
- Vue 3、Vite、TypeScript
- Tailwind CSS、shadcn-vue 风格基础组件、lucide-vue-next 图标

## 项目结构

```text
.
├── main.go                         # 程序入口，初始化配置、存储、服务和 HTTP 路由
├── frontend/                       # Vue 3 前端源码
├── web/dist/                       # Vite 构建产物，供 Go embed 使用
├── internal/config                 # 命令行参数和运行配置
├── internal/domain                 # 领域模型
├── internal/service                # 站点、分类、账号、会话和设置业务逻辑
├── internal/storage                # SQLite 存储实现
├── internal/transport/http         # HTTP 路由、处理器和响应封装
├── data/                           # 运行时数据目录
├── build.sh                        # 本地构建脚本
└── Dockerfile                      # 容器镜像构建文件
```

## 本地运行

首次开发前安装前端依赖：

```bash
cd frontend && npm install
```

开发时建议分别启动后端 API 和 Vite 前端：

```bash
go run . -port 8080 -data data
```

```bash
cd frontend && npm run dev
```

Vite 会把 `/api` 代理到 `http://localhost:8080`，因此开发页面仍然可以使用同源 Cookie 登录。

如需直接运行 Go 服务内嵌的生产前端，先构建前端产物：

```bash
cd frontend && npm run build
go run . -port 8080 -data data
```

启动后访问：

```text
http://localhost:8080
```
首次启动会自动生成随机初始密码，并打印到服务日志中。请登录后立即修改初始密码。

重置账号密码：

```bash
bin/navigation -data data -reset-auth
```

`-reset-auth` 会把账号密码重置为随机密码并打印到日志中，然后立即退出，不启动 HTTP 服务。

参数说明：

- `-port`：HTTP 服务端口，默认 `8080`。
- `-data`：数据目录，默认 `data`。SQLite 数据库会写入 `${data}/sites.db`。
- `-reset-auth`：重置账号密码为随机密码后退出。
- `-secure-cookie`：把登录 Cookie 标记为 `Secure`，用于 HTTPS 生产部署。

注意：项目使用 `go-sqlite3`，需要启用 CGO。本地环境必须有可用的 C 编译工具链。

## 构建

```bash
./build.sh
```

`build.sh` 会先执行前端构建，再编译 Go 二进制。首次执行时如果缺少 `frontend/node_modules`，会自动运行 `npm ci`。

默认输出到：

```text
bin/navigation
```

也可以指定输出路径：

```bash
OUTPUT=/tmp/navigation ./build.sh
```

## Docker

构建镜像：

```bash
docker build -t navigation .
```

运行容器：

```bash
docker run -p 8080:8080 -v "$PWD/data:/app/data" navigation
```

容器内数据目录为 `/app/data`，建议挂载到宿主机，避免容器删除后丢数据。

## API

公开读取接口无需登录；站点写入、分类修改、页面设置保存、账号修改和退出登录需要登录。登录成功后服务端会写入 `navigation_session` Cookie，会话有效期为 24 小时。

### 认证与账号

- `POST /api/login`：登录。
- `GET /api/session`：检查当前会话。
- `POST /api/logout`：退出登录。
- `PUT /api/account`：修改账号；`newPassword` 为空时只修改用户名。

登录请求：

```json
{
  "username": "admin",
  "password": "admin"
}
```

账号更新请求：

```json
{
  "username": "admin",
  "currentPassword": "admin",
  "newPassword": "new-password"
}
```

### 页面设置

- `GET /api/settings`：获取首页显示设置。
- `PUT /api/settings`：更新首页显示设置。

设置 JSON 字段：

```json
{
  "siteTitle": "导航站",
  "badge": "DEV PORTAL / 个人导航站",
  "subtitle": "聚合了常用网站",
  "heroTitle": "常用站点导航",
  "theme": "dark"
}
```

`theme` 当前支持 `dark`、`morning`、`forest`、`plum`；传入其他值会回落为 `dark`。

### 站点

- `GET /api/sites`：获取站点列表。
- `GET /api/sites?category=工具&q=go`：按分类和关键词过滤站点。
- `POST /api/sites`：创建站点。
- `PUT /api/sites/{id}`：更新站点。
- `DELETE /api/sites/{id}`：删除站点。

站点 JSON 字段：

```json
{
  "id": "site_xxx",
  "name": "OpenAI",
  "url": "https://openai.com",
  "category": "AI",
  "icon": "🔗",
  "description": "AI 平台",
  "glow": "rgba(96,165,250,.45)",
  "sort": 1,
  "createdAt": "2026-05-26T10:00:00+08:00",
  "updatedAt": "2026-05-26T10:00:00+08:00"
}
```

创建和更新时必须提供：

- `name`
- `url`，必须以 `http://` 或 `https://` 开头
- `category`

创建时如果缺少 `icon`、`glow` 或有效 `sort`，服务端会自动补默认图标、默认光效和下一个排序值。更新时保留原 ID、创建时间和原排序位置，除非请求里传入有效 `sort`。

### 分类

- `GET /api/categories`：获取分类列表，结果包含 `全部`。
- `PUT /api/categories/{name}`：重命名分类，请求体为 `{"name":"新分类名"}`；响应包含 `renamedSites`。
- `DELETE /api/categories/{name}`：删除分类，并将该分类下站点的分类清空；响应包含 `uncategorizedSites`。
- `GET /api/category-stats`：获取每个分类的站点数量。

### 统计

- `GET /api/stats`：获取站点数量、分类数量和覆盖率。

## 数据说明

- 默认数据库文件：`data/sites.db`。
- SQLite 启用 WAL，因此运行时可能生成 `data/sites.db-wal` 和 `data/sites.db-shm`。
- 旧版导入文件：`data/sites.json`。
- 首次启动时，如果 `sites` 表为空且存在旧版 JSON 数据，存储层会导入到 SQLite。
- 账号保存在 `users` 表，页面设置保存在 `settings` 表。
- 不要提交生成的数据库文件、WAL/SHM 文件或本地构建产物。

## 测试

```bash
go test ./...
```

## 开发约定

- HTTP 层只负责请求解析和响应输出。
- 业务校验放在 `internal/service`。
- SQL 和持久化细节放在 `internal/storage`。
- 修改 Go 代码后运行 `go fmt ./...` 和 `go test ./...`。
