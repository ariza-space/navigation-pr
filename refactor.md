# 前端 Vue 化改造计划

## 目标

将当前基于 `index.html`、`static/css/app.css`、`static/js/app.js` 的单页原生前端，改造成基于 Vue 3 的前端应用，并引入 shadcn-vue 与 Tailwind CSS。后端 Go API 与 SQLite 存储逻辑保持不变，改造重点放在前端工程化、组件化、样式体系和静态资源构建方式。

## 当前状态

- Go 入口 `main.go` 通过 `//go:embed index.html static/*` 打包静态文件。
- HTTP 路由在 `internal/transport/http/handler.go` 中提供：
  - 页面：`/`、`/index.html`
  - 静态资源：`/static/*`
  - API：`/api/login`、`/api/session`、`/api/logout`、`/api/account`、`/api/settings`、`/api/sites`、`/api/categories`、`/api/category-stats`、`/api/stats`
- 当前 UI、状态管理、事件绑定、接口调用集中在 `static/js/app.js`。
- 当前样式集中在 `static/css/app.css`，主题通过 `body[data-theme]` 切换。

## 技术选型

- Vue 3：使用 Composition API 和 `<script setup>`。
- Vite：作为前端构建工具，输出静态资源供 Go embed。
- TypeScript：建议启用，便于约束 API 数据结构和表单模型。
- Tailwind CSS：替换大部分手写布局、间距、颜色和响应式样式。
- shadcn-vue：用于 Button、Dialog、DropdownMenu、Input、Textarea、Select、Tabs、Card、Badge、Form 等基础组件。
- lucide-vue-next：配合 shadcn-vue 使用图标，例如 Search、Settings、LogOut、Plus、X、ChevronDown。

## 建议目录结构

```text
.
├── frontend/
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── tailwind.config.ts
│   ├── postcss.config.js
│   ├── components.json
│   └── src/
│       ├── main.ts
│       ├── App.vue
│       ├── assets/
│       │   └── main.css
│       ├── components/
│       │   ├── AppShell.vue
│       │   ├── HeroSection.vue
│       │   ├── SiteGrid.vue
│       │   ├── SiteCard.vue
│       │   ├── CategoryTabs.vue
│       │   ├── ThemeSwitcher.vue
│       │   ├── UserMenu.vue
│       │   ├── SiteDialog.vue
│       │   ├── CategoryDialog.vue
│       │   ├── EmojiDialog.vue
│       │   ├── AccountDialog.vue
│       │   └── SettingsDialog.vue
│       ├── components/ui/
│       │   └── shadcn-vue 生成组件
│       ├── composables/
│       │   ├── useAuth.ts
│       │   ├── useSites.ts
│       │   ├── useSettings.ts
│       │   └── useTheme.ts
│       ├── lib/
│       │   ├── api.ts
│       │   └── utils.ts
│       └── types/
│           └── api.ts
├── web/
│   └── dist/
│       └── Vite 构建产物，供 Go embed
├── main.go
└── internal/
```

说明：

- `frontend/` 放完整前端源码，不直接由 Go 服务读取。
- `web/dist/` 放 `vite build` 的产物，并作为 Go embed 的目标。
- 旧的 `index.html`、`static/css/app.css`、`static/js/app.js` 在迁移完成后删除或仅保留到一个兼容分支，不再作为运行入口。

## 后端静态资源调整

将当前 embed：

```go
//go:embed index.html static/*
var staticFiles embed.FS
```

调整为：

```go
//go:embed web/dist/*
var staticFiles embed.FS
```

并修改 `serveIndex` 与静态文件服务：

- `/assets/*` 或 Vite 生成的静态资源应从 `web/dist` 下提供。
- `/`、`/index.html` 返回 `web/dist/index.html`。
- 如果以后需要前端路由，可以让未知非 `/api/*` 路径回退到 `index.html`；当前导航站没有前端路由，可先保持只支持 `/` 和 `/index.html`。

建议实现方式：

- 使用 `fs.Sub(staticFiles, "web/dist")` 得到 dist 子文件系统。
- `http.FileServerFS(distFS)` 服务 Vite 产物。
- API 路由继续优先注册在 `/api/*`。

## 构建脚本调整

新增或修改命令：

```bash
cd frontend && npm install
cd frontend && npm run dev
cd frontend && npm run build
go test ./...
./build.sh
```

`frontend/package.json` 建议脚本：

```json
{
  "scripts": {
    "dev": "vite --host 0.0.0.0",
    "build": "vue-tsc --noEmit && vite build",
    "preview": "vite preview --host 0.0.0.0",
    "lint": "eslint ."
  }
}
```

`vite.config.ts` 建议：

- `outDir: "../web/dist"`
- `emptyOutDir: true`
- 开发环境代理 `/api` 到 Go 服务，例如 `http://localhost:8080`

## shadcn-vue 与 Tailwind 初始化步骤

1. 在 `frontend/` 初始化 Vite Vue 项目。
2. 安装依赖：

```bash
npm install vue @vitejs/plugin-vue vite typescript vue-tsc
npm install -D tailwindcss postcss autoprefixer
npm install class-variance-authority clsx tailwind-merge lucide-vue-next radix-vue
```

3. 初始化 Tailwind：

```bash
npx tailwindcss init -p
```

4. 初始化 shadcn-vue：

```bash
npx shadcn-vue@latest init
```

5. 添加首批组件：

```bash
npx shadcn-vue@latest add button dialog dropdown-menu input textarea select tabs card badge form label scroll-area separator
```

## 组件拆分计划

### App.vue

职责：

- 初始化登录态、站点列表、分类、统计信息、页面设置。
- 维护当前分类、搜索关键字、弹窗开关状态。
- 组合页面主结构。

### AppShell.vue

职责：

- 页面背景、主内容宽度、顶部用户菜单、主题切换入口。
- 保持整体布局与响应式约束。

### HeroSection.vue

职责：

- 展示 badge、主标题、简介。
- 搜索框。
- 统计信息：站点数、分类数、覆盖率。

### CategoryTabs.vue

职责：

- 使用 shadcn-vue Tabs 展示 `全部 + categories`。
- 切换分类后触发站点重新加载或本地过滤。

### SiteGrid.vue / SiteCard.vue

职责：

- 网格展示站点卡片。
- 未登录时只展示访问入口。
- 登录后展示编辑、删除等管理操作。

### SiteDialog.vue

职责：

- 新增/编辑站点。
- 字段包括名称、分类、URL、图标、排序、光效颜色、描述。
- 使用 shadcn-vue Dialog、Input、Textarea、Select/Button。
- 表单提交调用 `POST /api/sites` 或 `PUT /api/sites/{id}`。

### CategoryDialog.vue

职责：

- 分类管理。
- 对接 `DELETE /api/categories/{name}`。
- 如后端后续支持分类排序或重命名，可在此扩展。

### EmojiDialog.vue

职责：

- 展示现有 emojiOptions。
- 点击后更新站点表单的 icon 字段。

### UserMenu.vue

职责：

- 登录态展示用户名。
- 未登录点击打开登录弹窗。
- 已登录展示设置、账号修改、退出登录。

### AccountDialog.vue

职责：

- 修改账号密码。
- 对接 `PUT /api/account`。

### SettingsDialog.vue

职责：

- 页面设置：title、badge、heroTitle、subtitle、defaultTheme。
- 对接 `GET/PUT /api/settings`。

### ThemeSwitcher.vue

职责：

- 使用本地 `localStorage` 保存主题覆盖。
- 与后端 settings.theme 合并：本地覆盖优先，否则使用全局默认主题。
- 将当前主题写入根元素 class 或 data 属性。

## API 封装计划

在 `frontend/src/lib/api.ts` 中统一封装请求：

- 默认携带 `credentials: "same-origin"`。
- 默认使用 `Content-Type: application/json`。
- `204` 返回 `null`。
- `401` 抛出带状态码的错误，由 `useAuth` 决定是否打开登录弹窗。
- 业务错误读取后端 `{ "error": "..." }`。

建议 API 函数：

```ts
getSession()
login(input)
logout()
updateAccount(input)
getSettings()
updateSettings(input)
listSites(params)
createSite(input)
updateSite(id, input)
deleteSite(id)
listCategories()
deleteCategory(name)
getStats()
getCategoryStats()
```

## 数据类型计划

在 `frontend/src/types/api.ts` 定义：

```ts
export interface Site {
  id: string
  name: string
  url: string
  category: string
  description: string
  icon: string
  glow: string
  sort: number
}

export interface AppSettings {
  siteTitle: string
  badge: string
  heroTitle: string
  subtitle: string
  theme: string
}

export interface UserSession {
  username: string
}

export interface Stats {
  siteCount: number
  categoryCount: number
  coverage?: string
}
```

实际字段名需以 Go `domain` 结构体 JSON tag 为准，迁移时先核对 `internal/domain`。

## 样式与主题迁移

1. 将 `static/css/app.css` 中的视觉 token 梳理为 Tailwind theme：
   - 背景色
   - 前景色
   - border
   - muted
   - primary
   - accent
   - destructive
2. shadcn-vue 默认使用 CSS variables，建议把主题变量放在 `frontend/src/assets/main.css`。
3. 当前 `dark`、`morning`、`forest`、`plum` 可映射为：
   - `html[data-theme="dark"]`
   - `html[data-theme="morning"]`
   - `html[data-theme="forest"]`
   - `html[data-theme="plum"]`
4. 卡片 glow 颜色可保留为内联 CSS 变量，例如 `style="{ '--glow': site.glow }"`。
5. 迁移过程中避免一次性完全复刻所有 CSS，优先保证：
   - 搜索、分类、卡片、弹窗、登录、设置等主流程可用。
   - 移动端布局不溢出。
   - shadcn-vue 组件状态一致。

## 迁移阶段

### 阶段 1：前端工程初始化

- 创建 `frontend/`。
- 初始化 Vue 3 + Vite + TypeScript。
- 配置 Tailwind CSS。
- 初始化 shadcn-vue。
- 配置 Vite dev proxy 到 Go `/api`。
- 确认 `npm run dev` 可打开空页面。

验收：

- `frontend` 可独立启动。
- Tailwind 样式生效。
- shadcn-vue Button/Dialog 等组件可正常渲染。

### 阶段 2：API 与类型层

- 添加 `types/api.ts`。
- 添加 `lib/api.ts`。
- 将现有 `requestJSON` 行为迁移到统一请求函数。
- 用简单页面验证：
  - `GET /api/settings`
  - `GET /api/sites`
  - `GET /api/categories`
  - `GET /api/stats`

验收：

- Vue 页面能展示真实站点、分类、统计、设置。
- 未登录访问公开接口不弹错误。

### 阶段 3：主页面组件迁移

- 实现 AppShell、HeroSection、CategoryTabs、SiteGrid、SiteCard。
- 实现搜索与分类过滤。
- 实现主题切换。
- 保持原有文案和交互行为。

验收：

- 首页核心浏览流程可用。
- 站点卡片点击可访问 URL。
- 搜索和分类切换结果正确。
- 主题切换可持久化到 localStorage。

### 阶段 4：登录与管理流程迁移

- 实现 LoginDialog 或登录遮罩。
- 实现 UserMenu。
- 实现新增/编辑/删除站点。
- 实现分类管理。
- 实现账号修改。
- 实现页面设置。

验收：

- 未登录只可浏览。
- 登录后显示新增站点、分类管理、设置、账号修改、退出登录。
- 401 时能提示登录。
- 所有管理操作成功后刷新对应数据。

### 阶段 5：Go embed 与生产构建接入

- Vite build 输出到 `web/dist`。
- 修改 `main.go` embed 路径。
- 修改 HTTP 静态文件服务读取 `web/dist`。
- 修改 `build.sh`：
  - 先执行 `cd frontend && npm ci && npm run build`。
  - 再执行 Go build。
- 修改 Dockerfile：
  - 使用 Node 阶段构建前端。
  - Go build 阶段复制 `web/dist`。

验收：

- `./build.sh` 可生成包含新前端的 `bin/navigation`。
- `docker build -t navigation .` 可成功。
- `go run . -port 8080 -data data` 能访问 Vue 生产页面。

### 阶段 6：清理旧前端

- 删除旧 `static/js/app.js`。
- 删除旧 `static/css/app.css`。
- 删除或替换根目录旧 `index.html`。
- 更新 README 中的开发命令。
- 检查 `.gitignore`：
  - 忽略 `frontend/node_modules/`
  - 忽略前端构建缓存
  - 根据团队策略决定是否提交 `web/dist`

验收：

- 仓库不再依赖旧原生 JS 页面。
- README 能指导本地开发、构建、Docker 运行。

## 测试与验证清单

- `go test ./...`
- `cd frontend && npm run build`
- `go run . -port 8080 -data data`
- 浏览器手动验证：
  - 首页加载
  - 搜索站点
  - 分类切换
  - 主题切换和刷新后保留
  - 登录
  - 新增站点
  - 编辑站点
  - 删除站点
  - 分类管理
  - 页面设置保存
  - 修改账号密码
  - 退出登录
- 移动端宽度验证：
  - 375px
  - 768px
  - 1280px

## 风险与注意事项

- shadcn-vue 组件依赖 Tailwind CSS variables，主题迁移时不要只迁移 class，需同步迁移 CSS 变量。
- 后端当前 cookie 登录依赖 same-origin，Vite dev server 下需要配置 `/api` proxy，否则跨端口请求会遇到 cookie 问题。
- 如果 `web/dist` 不提交到 Git，Go embed 在本地构建前必须先执行前端 build；`build.sh` 和 Dockerfile 需要确保顺序正确。
- 旧页面所有 DOM id 事件绑定迁移后会消失，测试时必须覆盖每个弹窗和表单。
- 站点 URL、分类名、设置项的校验仍应以后端 service 层为准，前端校验只做交互优化。

## 推荐执行顺序

1. 先搭建 `frontend/` 并跑通 Vite + Tailwind + shadcn-vue。
2. 再迁移只读首页，确保 API、主题、搜索、分类可用。
3. 然后迁移登录和管理弹窗。
4. 最后接入 Go embed、build.sh、Dockerfile，并删除旧静态页面。
