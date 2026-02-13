基于钉钉同步搭建用户平台和LDAP服务（默认支持samba），用户密码支持同步至windows-AD
# Go-SyncFlow 统一身份同步与管理平台

基于 Go + Vue3 的企业级统一身份同步与管理平台，支持多上游（IM平台/LDAP/数据库）同步至本地用户，再下游同步至 LDAP/AD/数据库。

- 如有问题，请提交issue，不定期回复消息（杭州板砖ing，天天修机器！）

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
