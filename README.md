基于钉钉同步搭建用户平台和LDAP服务（默认支持samba），用户密码支持同步至windows-AD
# Go-SyncFlow 统一身份同步与管理平台

基于 Go + Vue3 的企业级统一身份同步与管理平台，支持多上游（IM平台/LDAP/数据库）同步至本地用户，再下游同步至 LDAP/AD/数据库。

- 如有问题，请提交issue，不定期回复消息（杭州板砖ing，天天修机器！）
- 个人邮箱：auroraer888@163.com（每周六固定回复）

## 功能特性

- **上游同步**：支持钉钉、企业微信、飞书、WeLink 等 IM 平台，LDAP/AD，多种数据库（MySQL/PostgreSQL/SQL Server/Oracle/SQLite）
- **下游同步**：同步本地用户到 LDAP/AD、多种数据库
- **内置 LDAP 服务**：内嵌 LDAP 服务器，默认支持 Samba 属性（兼容群晖 NAS）
- **SSO 免登**：支持钉钉、飞书、企业微信免登
- **用户管理**：新增/编辑/删除/启禁用/重置密码，角色权限管理
- **通知系统**：邮件、短信（12家运营商）、Webhook、钉钉工作通知
- **日志管理**：登录日志、操作日志、同步日志、API调用日志
- **安全中心**：密码策略、登录策略、IP白名单、API Key认证
- **消息策略**：灵活配置消息类型到通道的映射
- **文档中心**：在线查看/下载系统文档
说明：

## 钉钉用户→本地用户目录=LDAP服务→AD服务器

- 1.上下游同步配置完成后，建议每10分钟钉钉同步1次，员工加入、离开钉钉时和部门信息发生变更，会自动同步至本地，同时触发事件同步，将本地用户和群组同步至windows-AD，这样你就有了一套基于钉钉组织架构的LDAP和AD域用户目录服务
- 2.钉钉架构的用户同步至本地时，系统默认生成复杂密码，密码生成时将发送至员工（配置消息策略即生效）
- 3.非钉钉架构里人员（外包或三方），人事专员可手工创建账号和群组，消息策略可设置专属密码通知渠道。（钉钉架构的人员密码生成发送至钉钉工作通知，外部人员将通过短信验证码接收密码）
- 4.系统设置-LDAP服务默认开启并支持samba，兼容群晖LDAP接入，提供ldaps能力（上传SSL证书即可手动开启）
- 5.事件触发指，当用户状态发生改变时，包括：用户创建，用户删除，用户禁用，用户信息更新，用户密码修改，用户密码生成，用户角色变更。
- 6.钉钉离职员工，同步至本地时，将改变账号状态为禁用，不会删除。未启用的本地账号，下游同步至AD时，可在同步规则中自定义设置删除账号或禁用账号。
说明：管理员也不能知道员工的密码。员工密码仅可登录数据库查看，管理员无法自定义设置密码，管理员重置密码时，密码必须通过通知渠道推送（建议普通员工不分配任何权限，将本服务链接添加进钉钉工作台，员工钉钉SSO免登后直接可修改密码或找回密码）。
  
## 技术栈

- 后端：Go 1.22+ / Gin / GORM / SQLite
- 前端：Vue3 / Vite / TypeScript / Element Plus / ECharts
- 内嵌：LDAP 服务器（gldap）、Samba 支持

## 目录结构

```
Go-SyncFlow/
├── backend/          # Go 后端
├── frontend/         # Vue3 前端
├── scripts/          # 部署脚本（start/stop/reset-admin/pack）
├── docs/             # 系统文档（MD + PDF）
├── tooling/          # 工具包（Go安装包等）
└── README.md         # 本文件
```

## 快速部署

```bash
tar -xzf go-syncflow-XXXXXX.tar.gz -C /opt/
cd /opt/Go-SyncFlow
chmod +x scripts/*.sh
./scripts/start.sh
```
登录后，先检查角色配置（当角色授权列表为空，登录页为个人修改密码页），HTTPS证书和传输加密

## 访问地址

- HTTP: `http://服务器IP:8080`（首次登录）
- HTTPS: `https://服务器IP:8443`（管理后台上传NGINX证书后可用）

## 默认账号

- 用户名：`admin`
- 密码：`Admin@2024`

## 常用命令

```bash
./scripts/start.sh          # 一键启动
./scripts/stop.sh           # 停止服务
./scripts/stop.sh && ./scripts/start.sh  # 重启服务
./scripts/reset-admin.sh    # 重置管理员密码
./scripts/pack.sh           # 打包项目
systemctl status go-syncflow  # 查看服务状态
journalctl -u go-syncflow -f  # 查看日志
```

## 后续计划
- 单点登录，支持CAS / SAML2.0 / OAuth2 / OIDC 接入方式
- 网络准入，支持Portal/radius/802.1X
- 物联网接入协议，内置MQTT服务，MQTT-Broker客户端，TCP服务和SNMP服务端
- 架构升级，使用MYSQL
- 登录认证支持双因素，OTP或SMS，OTP嵌入个人中心页面（员工钉钉免登打开后即可查看OTP验证码）
- 登录消息通知
## 说明

- LDAP 服务默认启用 Samba 属性支持，可直接对接群晖 NAS
- 通知渠道、同步连接器等需在管理界面中配置
- 数据库初始为空，首次启动自动初始化

## 协议

本项目采用 [Apache-2.0](LICENSE) 协议开源。
