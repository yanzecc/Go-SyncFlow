# Go-SyncFlow 统一身份同步与管理平台 - API 接口文档

> 版本：v3.3 | 更新日期：2026-02-10 | 接口总数：170+

---

## 通用说明

### 基础 URL

- HTTP: `http://服务器IP:8080/api`
- HTTPS: `https://服务器IP:8443/api`

### 认证方式

本平台支持两种认证方式：

#### 方式一：JWT 令牌认证（Web 登录）

适用于前端页面和需要用户登录的场景：

```
Authorization: Bearer <token>
```

Token 通过 POST /api/auth/login 获取，接口路径为 `/api/*`。

#### 方式二：AppID/AppKey 认证（开放 API）

适用于第三方系统集成，无需登录账号密码：

```
X-App-ID: <your-app-id>
X-App-Key: <your-app-key>
```

密钥在管理后台「系统设置 → API 密钥」中创建，接口路径为 `/api/open/*`。

也支持通过 URL 参数传递：`?app_id=xxx&app_key=xxx`

### 统一响应格式

成功响应：

```json
{"success": true, "data": {}, "message": "操作成功"}
```

错误响应：

```json
{"success": false, "message": "错误描述"}
```

### HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 400 | 请求参数错误 |
| 401 | 未认证或 Token 过期 |
| 403 | 权限不足或 IP 被拒绝 |
| 404 | 资源不存在 |
| 429 | 请求过于频繁 |
| 500 | 服务器内部错误 |

### 限流策略

| 类型 | 限制 |
|------|------|
| 通用 API | 100 次/分钟/IP（已登录 200 次） |
| 登录接口 | 10 次/分钟/IP |
| 敏感操作 | 5 次/分钟/IP |

### API 使用限制（开放 API 安全策略）

通过 AppID/AppKey 认证的开放 API（`/api/open/*`）有以下安全限制：

1. **禁止管理 API 密钥**：无法通过 API 创建、删除、重置 API 密钥（开放 API 不暴露密钥管理接口）
2. **禁止删除日志**：无法通过 API 删除任何系统日志（仅提供查询接口）
3. **禁止操作管理员账户**：无法通过 API 查看、修改、删除 admin 用户信息，无法重置 admin 密码
4. **禁止分配超级管理员**：无法通过 API 给任何用户分配 `super_admin` 角色
5. **禁止修改超管权限**：无法通过 API 修改超级管理员角色的权限设置（开放 API 不暴露权限修改接口）

违反上述限制时，接口返回 HTTP 403：

```json
{"success": false, "message": "API 安全限制：不允许通过 API 操作管理员账户"}
```

---

## 一、公开接口（无需认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/auth/login | 用户登录 |
| POST | /api/auth/dingtalk | 钉钉免登 |
| POST | /api/auth/forgot-password/check | 忘记密码-检查用户 |
| POST | /api/auth/forgot-password/send-code | 忘记密码-发送验证码 |
| POST | /api/auth/forgot-password/reset | 忘记密码-重置密码 |
| GET | /api/settings/ui | 获取 UI 配置 |
| GET | /api/settings/dingtalk/status | 获取钉钉启用状态 |
| GET | /api/crypto/public-key | 获取 RSA 公钥 |

### 1.1 用户登录

POST /api/auth/login

请求体：
```json
{"username": "admin", "password": "RSA加密密码", "encrypted": true}
```

响应 data 字段：
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {"id": 1, "username": "admin", "roles": ["super_admin"], "permissions": [...]}
}
```

### 1.2 忘记密码

POST /api/auth/forgot-password/check
```json
{"username": "zhangsan"}
```
响应包含该用户可用的验证方式（受消息策略控制）。

POST /api/auth/forgot-password/send-code
```json
{"username": "zhangsan", "method": "sms"}
```

POST /api/auth/forgot-password/reset
```json
{"username": "zhangsan", "method": "sms", "code": "283746", "newPassword": "RSA加密新密码", "encrypted": true}
```

---

## 二、认证管理（需登录）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/auth/logout | 登出 |
| GET | /api/auth/info | 获取当前用户信息 |
| GET | /api/docs | 获取文档列表 |
| GET | /api/docs/:name | 下载文档 |
| PUT | /api/auth/password | 修改密码 |

---

## 三、个人中心

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/profile | 获取个人资料（含 allowedVerifyMethods） |
| PUT | /api/profile | 更新个人资料 |
| PUT | /api/profile/password | 修改密码（多方式） |
| POST | /api/profile/verify-code | 发送验证码 |

修改密码请求体：
```json
{"method": "password", "oldPassword": "旧密码", "newPassword": "新密码"}
```
或
```json
{"method": "sms", "code": "283746", "newPassword": "新密码"}
```

---

## 四、用户管理

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | /api/users | user:list | 用户列表 |
| POST | /api/users | user:create | 新增用户 |
| GET | /api/users/:id | user:list | 用户详情 |
| PUT | /api/users/:id | user:update | 更新用户 |
| DELETE | /api/users/:id | user:delete | 删除用户 |
| PUT | /api/users/:id/status | user:toggle_status | 启用/禁用 |
| PUT | /api/users/:id/reset-password | user:reset_password | 重置密码 |
| POST | /api/users/batch-reset-password | user:reset_password | 批量重置密码 |
| GET | /api/users/export | user:export | 导出用户（Excel格式） |

查询参数：page, pageSize, search, groupId, status

新增用户请求体：
```json
{"username": "zhangsan", "nickname": "张三", "password": "Abc@1234", "email": "z@a.com", "phone": "138xxx", "groupId": 1, "roleIds": [2]}
```

重置密码请求体：
```json
{"notifyChannels": ["sms", "dingtalk"]}
```

---

## 五、群组管理

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | /api/groups | user:list | 群组列表（树形） |
| POST | /api/groups | user:create_group | 新增群组 |
| PUT | /api/groups/:id | user:update | 更新群组 |
| DELETE | /api/groups/:id | user:delete | 删除群组 |

---

## 六、角色管理

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | /api/roles | role:list | 角色列表 |
| POST | /api/roles | role:create | 新增角色 |
| GET | /api/roles/:id | role:list | 角色详情 |
| PUT | /api/roles/:id | role:update | 更新角色（含 sidebarMode, landingPage） |
| DELETE | /api/roles/:id | role:delete | 删除角色 |
| GET | /api/roles/:id/permissions | role:list | 获取角色权限 |
| PUT | /api/roles/:id/permissions | role:permission | 更新角色权限 |
| GET | /api/roles/:id/auto-assign | role:list | 获取自动分配规则 |
| PUT | /api/roles/:id/auto-assign | role:update | 更新自动分配规则 |
| POST | /api/roles/auto-assign/apply | role:update | 执行自动分配 |

GET /api/permissions/tree - 获取权限树（任何已登录用户）

---

## 七、日志管理

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | /api/logs/system | log:login / log:operation | 系统日志（登录+操作合并） |
| GET | /api/logs/login | log:login | 登录日志（原始） |
| GET | /api/logs/operation | log:operation | 操作日志（原始） |
| GET | /api/logs/sync | log:operation | 同步日志 |

### 7.1 系统日志查询参数

| 参数 | 说明 |
|------|------|
| page | 页码，默认1 |
| size | 每页条数，默认20 |
| type | 日志类型：login / operation / 空=全部 |
| keyword | 用户名模糊搜索 |
| startDate | 开始日期（YYYY-MM-DD） |
| endDate | 结束日期（YYYY-MM-DD） |

### 7.2 同步日志查询参数

| 参数 | 说明 |
|------|------|
| page | 页码，默认1 |
| size | 每页条数，默认20 |
| direction | 同步方向：upstream / downstream / 空=全部 |
| event | 事件类型：full_sync / user_create / user_update 等 |
| status | 状态：success / partial / failed |
| startDate | 开始日期 |
| endDate | 结束日期 |

---

## 八、系统设置

| 方法 | 路径 | 说明 |
|------|------|------|
| PUT | /api/settings/ui | 更新 UI 配置 |
| GET | /api/settings/dingtalk | 获取钉钉配置 |
| PUT | /api/settings/dingtalk | 更新钉钉配置 |
| POST | /api/settings/dingtalk/test | 测试钉钉连接 |
| GET | /api/settings/ldap | 获取 LDAP 配置 |
| PUT | /api/settings/ldap | 更新 LDAP 配置 |
| POST | /api/settings/ldap/test | 测试 LDAP 服务 |
| GET | /api/settings/ldap/status | 获取 LDAP 状态 |
| GET | /api/settings/https | 获取 HTTPS 配置 |
| PUT | /api/settings/https | 更新 HTTPS 配置 |
| POST | /api/settings/https/cert | 上传 SSL 证书 |
| DELETE | /api/settings/https/cert | 删除 SSL 证书 |
| GET | /api/settings/crypto | 获取 RSA 加密配置 |
| PUT | /api/settings/crypto | 更新 RSA 加密配置 |

---

## 九、系统监控

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/system/status | 获取系统状态（CPU/内存/磁盘/网络/Go运行时） |

---

## 十、安全中心

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/security/dashboard | 安全仪表盘 |
| GET | /api/security/events | 安全事件列表 |
| PUT | /api/security/events/:id/resolve | 处理安全事件 |
| GET | /api/security/login-attempts | 登录尝试记录 |
| GET | /api/security/lockouts | 锁定记录 |
| POST | /api/security/lockouts/unlock-account | 解锁账号 |
| POST | /api/security/lockouts/unlock-ip | 解锁 IP |
| GET | /api/security/ip/blacklist | IP 黑名单 |
| POST | /api/security/ip/blacklist | 添加黑名单 |
| DELETE | /api/security/ip/blacklist/:id | 删除黑名单 |
| GET | /api/security/ip/whitelist | IP 白名单 |
| POST | /api/security/ip/whitelist | 添加白名单 |
| DELETE | /api/security/ip/whitelist/:id | 删除白名单 |
| GET | /api/security/ip/whitelist/mode | 白名单模式 |
| POST | /api/security/ip/check | 检查 IP 状态 |
| GET | /api/security/sessions | 所有会话 |
| GET | /api/security/sessions/my | 我的会话 |
| DELETE | /api/security/sessions/:id | 终止会话 |
| DELETE | /api/security/sessions/user/:userId | 终止用户全部会话 |
| GET | /api/security/config | 获取所有安全配置 |
| GET | /api/security/config/:key | 获取指定安全配置 |
| PUT | /api/security/config/:key | 更新安全配置 |

---

## 十一、通知渠道

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/security/alerts/channels | 渠道列表 |
| POST | /api/security/alerts/channels | 创建渠道 |
| PUT | /api/security/alerts/channels/:id | 更新渠道 |
| DELETE | /api/security/alerts/channels/:id | 删除渠道 |
| POST | /api/security/alerts/channels/:id/test | 测试渠道 |

channelType 可选值：sms_aliyun, sms_tencent, sms_huawei, sms_baidu, sms_ctyun, sms_ronghe, sms_cmcc, sms_wecom, sms_dingtalk, sms_feishu, sms_https, sms_custom, email, webhook, dingtalk_work

---

## 十二、告警规则

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/security/alerts/rules | 告警规则列表 |
| POST | /api/security/alerts/rules | 创建告警规则 |
| PUT | /api/security/alerts/rules/:id | 更新告警规则 |
| DELETE | /api/security/alerts/rules/:id | 删除告警规则 |
| GET | /api/security/alerts/logs | 告警日志 |

---

## 十三、消息模板

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/notify/templates | 模板列表 |
| GET | /api/notify/templates/:id | 模板详情 |
| POST | /api/notify/templates | 创建模板 |
| PUT | /api/notify/templates/:id | 更新模板 |
| DELETE | /api/notify/templates/:id | 删除模板 |

---

## 十四、消息策略

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/notify/policies | 所有策略 |
| GET | /api/notify/policies/scene | 按场景查询策略 |
| POST | /api/notify/policies | 创建/更新策略 |
| PUT | /api/notify/policies/batch | 批量更新策略 |
| POST | /api/notify/policies/group | 创建群组策略 |
| PUT | /api/notify/policies/group/:id | 更新群组策略 |
| DELETE | /api/notify/policies/group/:id | 删除群组策略 |

---

## 十五、连接器管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/connectors | 连接器列表 |
| POST | /api/connectors | 创建连接器 |
| GET | /api/connectors/:id | 连接器详情 |
| PUT | /api/connectors/:id | 更新连接器 |
| DELETE | /api/connectors/:id | 删除连接器 |
| POST | /api/connectors/:id/test | 测试连接 |
| GET | /api/connectors/:id/columns | 发现字段 |

type 可选值：ldap_ad, database
dbType 可选值：mysql, postgresql, oracle, sqlserver

---

## 十六、同步器管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/synchronizers | 同步器列表 |
| POST | /api/synchronizers | 创建同步器 |
| GET | /api/synchronizers/:id | 同步器详情 |
| PUT | /api/synchronizers/:id | 更新同步器 |
| DELETE | /api/synchronizers/:id | 删除同步器 |
| POST | /api/synchronizers/:id/trigger | 触发同步 |
| GET | /api/synchronizers/:id/logs | 同步日志 |
| GET | /api/synchronizers/:id/mappings | 属性映射列表 |
| POST | /api/synchronizers/:id/mappings | 创建映射 |
| PUT | /api/synchronizers/:id/mappings/:mid | 更新映射 |
| DELETE | /api/synchronizers/:id/mappings/:mid | 删除映射 |
| PUT | /api/synchronizers/:id/mappings-batch | 批量更新映射 |
| GET | /api/sync/events | 可订阅事件列表 |
| GET | /api/sync/source-fields | 本地可用字段 |
| GET | /api/sync/target-fields | 目标系统字段 |
| POST | /api/sync/trigger-all | 触发所有同步器 |

---

## 十七、钉钉管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/dingtalk/departments | 钉钉部门列表 |
| GET | /api/dingtalk/users | 钉钉用户列表 |
| POST | /api/dingtalk/sync | 触发钉钉同步 |
| GET | /api/dingtalk/sync/status | 同步状态 |
| GET | /api/dingtalk/settings | 同步设置 |
| PUT | /api/dingtalk/settings | 更新同步设置 |

---

## 十八、API 密钥管理

> 认证方式：JWT 令牌 | 权限要求：settings:system

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/apikeys | API 密钥列表 |
| POST | /api/apikeys | 创建 API 密钥 |
| GET | /api/apikeys/:id | 密钥详情 |
| PUT | /api/apikeys/:id | 更新密钥信息 |
| POST | /api/apikeys/:id/reset | 重置 AppKey |
| PUT | /api/apikeys/:id/toggle | 启用/禁用密钥 |
| DELETE | /api/apikeys/:id | 删除密钥 |

### 创建密钥

```
POST /api/apikeys
```

请求体：

```json
{
  "name": "ERP 系统集成",
  "description": "ERP 系统调用用户同步接口",
  "appId": "",
  "appKey": "",
  "ipWhitelist": ["10.0.0.0/8", "192.168.1.*"],
  "ipBlacklist": [],
  "rateLimit": 60,
  "expiresAt": "2027-12-31"
}
```

- `appId`、`appKey` 留空则自动生成
- `rateLimit` 单位为次/分钟，默认 60
- `expiresAt` 留空则永不过期

响应 data 字段（**仅创建时返回明文 AppKey**）：

```json
{
  "id": 1,
  "appId": "ak_06ee7418df3856da",
  "appKey": "f9394cc330b6e20c...完整密钥",
  "name": "ERP 系统集成",
  "message": "API密钥创建成功，请妥善保存 AppKey。"
}
```

### 重置 AppKey

```
POST /api/apikeys/:id/reset
```

请求体（可选）：

```json
{"appKey": "自定义新密钥，留空自动生成"}
```

---

## 十九、开放 API（AppID/AppKey 认证）

> 认证方式：X-App-ID + X-App-Key 请求头
> 基础路径：`/api/open`

### 调用示例

```bash
curl -H "X-App-ID: ak_06ee7418df3856da" \
     -H "X-App-Key: f9394cc330b6e20cf7..." \
     https://服务器:8443/api/open/users
```

### 接口列表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/open/users | 用户列表（分页） |
| POST | /api/open/users | 创建用户 |
| GET | /api/open/users/:id | 用户详情 |
| PUT | /api/open/users/:id | 更新用户 |
| DELETE | /api/open/users/:id | 删除用户 |
| PUT | /api/open/users/:id/status | 启用/禁用用户 |
| PUT | /api/open/users/:id/reset-password | 重置用户密码 |
| GET | /api/open/groups | 群组列表 |
| POST | /api/open/groups | 创建群组 |
| PUT | /api/open/groups/:id | 更新群组 |
| DELETE | /api/open/groups/:id | 删除群组 |
| GET | /api/open/roles | 角色列表 |
| GET | /api/open/roles/:id | 角色详情 |
| GET | /api/open/roles/:id/permissions | 角色权限 |
| GET | /api/open/logs/login | 登录日志 |
| GET | /api/open/logs/operation | 操作日志 |
| GET | /api/open/logs/sync | 同步日志 |
| POST | /api/open/dingtalk/sync | 触发钉钉同步 |
| GET | /api/open/dingtalk/sync/status | 钉钉同步状态 |
| GET | /api/open/system/status | 系统状态 |

### 错误码

| 状态码 | 说明 |
|--------|------|
| 401 | AppID 或 AppKey 无效 |
| 403 | 密钥已禁用 / 已过期 / IP 被拒绝 |

---

## 二十、上游同步管理

> 认证方式：JWT 令牌 | 权限要求：sync:*

### 上游连接器

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/sync/upstream/connectors | 上游连接器列表 |
| POST | /api/sync/upstream/connectors | 创建上游连接器 |
| PUT | /api/sync/upstream/connectors/:id | 更新上游连接器 |
| DELETE | /api/sync/upstream/connectors/:id | 删除上游连接器 |
| POST | /api/sync/upstream/connectors/:id/test | 测试连接 |

type 可选值：im_dingtalk, im_wechatwork, im_feishu, im_welink, ldap_ad, db_mysql, db_postgresql, db_sqlserver, db_oracle, db_sqlite

### 上游同步规则

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/sync/upstream/rules | 上游规则列表 |
| POST | /api/sync/upstream/rules | 创建上游规则 |
| PUT | /api/sync/upstream/rules/:id | 更新上游规则 |
| DELETE | /api/sync/upstream/rules/:id | 删除上游规则 |
| POST | /api/sync/upstream/rules/:id/trigger | 手动触发同步 |

### 上游属性映射

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/sync/upstream/rules/:id/mappings | 获取规则的属性映射列表 |
| PUT | /api/sync/upstream/rules/:id/mappings | 批量更新属性映射 |
| POST | /api/sync/upstream/rules/:id/mappings/reset | 恢复默认映射 |

映射字段说明：
- sourceAttribute：上游源属性
- targetAttribute：本地目标属性
- mappingType：mapping（直接映射）、transform（转换）
- transformRule：pinyin / email_prefix / mobile / email / userid

---

## 二十一、下游同步管理

> 认证方式：JWT 令牌 | 权限要求：sync:*

### 下游连接器

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/sync/downstream/connectors | 下游连接器列表 |
| POST | /api/sync/downstream/connectors | 创建下游连接器 |
| PUT | /api/sync/downstream/connectors/:id | 更新下游连接器 |
| DELETE | /api/sync/downstream/connectors/:id | 删除下游连接器 |
| POST | /api/sync/downstream/connectors/:id/test | 测试连接 |

### 下游同步规则

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/sync/downstream/rules | 下游规则列表 |
| POST | /api/sync/downstream/rules | 创建下游规则 |
| PUT | /api/sync/downstream/rules/:id | 更新下游规则 |
| DELETE | /api/sync/downstream/rules/:id | 删除下游规则 |
| POST | /api/sync/downstream/rules/:id/trigger | 手动触发同步 |

### 下游属性映射

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/sync/downstream/rules/:id/mappings | 获取规则的属性映射列表 |
| PUT | /api/sync/downstream/rules/:id/mappings | 批量更新属性映射 |

---

## 二十二、SSO 提供者

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/sso/providers | 获取已启用的 SSO 提供者列表 |
| POST | /api/auth/sso/:provider | SSO 免登（provider: dingtalk/feishu/wechatwork） |

---

## 二十三、日志设置

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/logs/settings | 获取日志保留配置 |
| PUT | /api/logs/settings | 更新日志保留配置 |

请求体：
```json
{
  "loginLogRetentionDays": 90,
  "operationLogRetentionDays": 90,
  "syncLogRetentionDays": 60,
  "securityEventRetentionDays": 180,
  "apiAccessLogRetentionDays": 30,
  "cleanupTime": "03:00"
}
```

---

## 二十四、API 调用日志

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/logs/api-access | API 调用日志列表 |
| GET | /api/logs/api-access/export | 导出 API 调用日志 CSV |

查询参数：page, pageSize, method, path, appId, statusCode, startTime, endTime

---

*本文档随系统版本更新，最后更新：2026-02-09。Go-SyncFlow v3.0*
