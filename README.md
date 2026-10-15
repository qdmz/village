# 乡村振兴村居日常工作管理系统

面向村“两委”、驻村干部和网格员的村居日常工作管理系统：党建引领、网格治理、户籍人口、民生保障、任务督办、村务公开、产业与集体经济一体化管理，并提供对外的村级门户首页与数据驾驶舱。

- 后端：Go 1.22+（Gin + GORM），JWT 认证，bcrypt 密码
- 前端：Vue 3 + Vite + Element Plus + ECharts（hash 路由），SheetJS 处理 Excel
- 数据库：SQLite（纯 Go 驱动 glebarez/sqlite，`CGO_ENABLED=0`，零配置）或 MySQL / MariaDB
- 部署：单个二进制文件，内嵌前端，同时提供 `/api` 接口

当前版本：见 `VERSION`（v0.1.1）。

在线演示：<https://village-demo-qdmz.style.dev>（临时演示环境，账号请联系作者）

## 功能模块

| 模块 | 说明 |
| --- | --- |
| 门户首页 | 横幅、工作快讯、通知公告/党务村务财务公开、村规民约、振兴故事、致富带头人、红黑榜、治理积分榜（无需登录） |
| 工作台 | 关键指标、人口结构、特殊人群、事件状态、任务完成率、资金收支图表、我的待办、最新公告 |
| 数据驾驶舱 | 一屏大屏总览，适合会议室/大屏展示 |
| 村情与网格 | 村 → 村民小组 → 网格三级组织树、村情信息、网格员分配 |
| 户籍与人口台账 | 家庭户、村民（身份证脱敏）、特殊人群标签、家庭户详情（成员+走访记录）、Excel/CSV 导入导出 |
| 网格巡查 | 巡查日期/时间/路线/情况/问题、现场照片（手机拍照上传） |
| 事件上报 | 上报 → 受理 → 处理中 → 办结 / 退回 → 重新上报，全流程时间线 |
| 矛盾纠纷调解 | 受理、调解员、调解状态与结果 |
| 任务督办 | 交办给个人，优先级、截止日期、进度更新、逾期预警 |
| 工作日志 | 工作日志、民情日记、走访记录（关联家庭户） |
| 党建引领 | 党组织架构图、党员花名册、三会一课/主题党日/党建活动 |
| 民生保障 | 低保、特困供养、优抚、残疾人、留守老人/儿童/妇女、临时救助 |
| 乡村治理 | 治理积分（爱党爱国/邻里和睦/人居环境/勤劳致富/移风易俗/遵纪守法）及排行、红黑榜、村规民约 |
| 乡村振兴 | 振兴概览、产业项目、资金台账（收支汇总）、集体资产、致富带头人 |
| 通知公告 | 通知公告、党务公开、村务公开、财务公开 |
| 门户管理 | 维护门户首页的横幅、快讯、振兴故事 |
| 系统管理 | 用户与角色、重置密码、操作日志、系统设置、JSON 备份/恢复、主题色切换 |

### 角色与权限

| 角色 | 权限 |
| --- | --- |
| 管理员 admin | 全部功能，含用户管理、操作日志、系统设置、备份恢复 |
| 村干部 cadre | 全部业务功能（含资金、资产、项目、门户管理），不含系统管理 |
| 网格员 grid | 仅能查看/维护**所负责网格**的户籍、人口、巡查、事件、民生、纠纷、积分；只能看到交办给自己的任务和本人日志；不可访问资金/资产/项目 |

权限在后端强制校验，前端菜单同步隐藏。

## 快速开始（零配置）

```bash
# 1. 构建（需要 Node 18+ 与 Go 1.22+）
scripts/build.sh linux amd64

# 2. 运行（不带配置文件时使用 SQLite，数据存放在 ./data）
./dist/village-mgmt-linux-amd64
```

首次启动会自动建表并写入演示数据，并在日志里打印 `admin` 的随机初始密码（也可在 `config.yaml` 的 `seed.admin_password` 指定）。浏览器访问 `http://127.0.0.1:8095`。

> 演示数据中的其他账号（村干部、网格员）使用随机密码，请用管理员在“用户管理”中重置。

## 配置文件

复制 `server/config.example.yaml` 为 `config.yaml`，用 `-c` 指定：

```yaml
server:
  addr: "127.0.0.1:8095"
  static_dir: ""            # 留空使用内嵌前端
  upload_dir: "./data/uploads"
  max_upload_mb: 5
database:
  driver: "mysql"           # sqlite | mysql
  dsn: "village:密码@tcp(127.0.0.1:3306)/village?charset=utf8mb4&parseTime=True&loc=Local"
jwt:
  secret: ""                # 留空自动生成并保存到数据库
  expire_hours: 72
seed:
  demo: true
  admin_password: ""
```

MySQL/MariaDB 需先建库：

```sql
CREATE DATABASE village DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'village'@'127.0.0.1' IDENTIFIED BY '强密码';
GRANT ALL ON village.* TO 'village'@'127.0.0.1';
```

## 生产部署（systemd）

```bash
sudo useradd -r -s /usr/sbin/nologin village
sudo mkdir -p /opt/village-mgmt/data && sudo chown -R village: /opt/village-mgmt
sudo cp dist/village-mgmt-linux-amd64 /opt/village-mgmt/village-mgmt
sudo cp server/config.example.yaml /opt/village-mgmt/config.yaml   # 按需修改
sudo cp deploy/village-mgmt.service /etc/systemd/system/
sudo systemctl daemon-reload && sudo systemctl enable --now village-mgmt
```

如需通过 Nginx 暴露端口或绑定域名，参考 `deploy/nginx-village.conf`。

## 开发

```bash
cd server && go run .            # 后端 127.0.0.1:8095
cd web && npm install && npm run dev   # 前端开发服务器，/api 自动代理到后端
```

### 目录结构

```
server/
  main.go                 入口
  internal/config         配置
  internal/db             数据库初始化（SQLite / MySQL）
  internal/model          数据模型
  internal/middleware     JWT 鉴权、角色控制、操作日志
  internal/handler        接口；crud.go 为通用增删改查/筛选/分页/导入导出
  internal/seed           首次启动演示数据
  internal/router         路由
  webui/                  内嵌前端产物
web/
  src/config/modules.js   各模块字段/筛选/表单配置（新增模块主要改这里）
  src/components/CrudPage.vue  通用列表页（表格/移动端卡片/表单/详情/导入导出）
  src/views/              各业务页面
```

### 新增一个业务模块

1. `server/internal/model` 增加模型并加入 `All()`；
2. `server/internal/handler` 声明一个 `Resource[T]`（搜索字段、筛选、权限、导出/导入列），在 `router.go` 中 `Register`；
3. `web/src/config/modules.js` 增加字段配置，在 `router.js` 用 `crud()` 注册路由，并在 `layout/SideMenu.vue` 增加菜单。

## 主要接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | /api/auth/login | 登录，返回 JWT |
| GET | /api/public/portal | 门户首页数据（公开） |
| GET | /api/dashboard | 工作台统计 |
| GET | /api/screen | 驾驶舱数据 |
| GET/POST/PUT/DELETE | /api/{households,villagers,patrols,events,tasks,worklogs,notices,projects,funds,assets,meetings,welfare,disputes,honors,points,leaders,party/members,party/orgs,portal,...} | 通用增删改查，支持 `kw`、筛选字段、`page`、`size`、`start`、`end` |
| GET | /api/{资源}/export | 导出 CSV（前端转换为 xlsx） |
| POST | /api/{资源}/import | 导入 CSV（前端可将 xlsx 转换后上传） |
| POST | /api/events/:id/flow | 事件流转：accept / process / finish / reject / resubmit |
| POST | /api/tasks/:id/progress | 任务进度更新 |
| GET | /api/system/backup | 全量备份（管理员） |
| POST | /api/system/restore | 从备份恢复（管理员） |

## 安全说明

- 密码 bcrypt 加密；登录失败 5 次锁定 5 分钟；JWT 默认 72 小时过期。
- 所有写操作记录操作日志（模块、动作、路径、IP、状态码）。
- 身份证号在列表与导出中脱敏；图片上传限制类型与大小并校验文件内容。
- 演示数据均为虚构，不对应真实人员。

## 许可证

仅供学习与演示使用。
