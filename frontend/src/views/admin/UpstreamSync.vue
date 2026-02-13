<template>
  <div class="upstream-page">
    <!-- Tab 切换：连接器 / 同步规则 / 属性映射 -->
    <el-tabs v-model="activeTab" type="border-card">
      <!-- ==================== 上游连接器 ==================== -->
      <el-tab-pane label="上游连接器" name="connectors">
        <div class="tab-toolbar">
          <span class="tab-desc">配置IM平台、LDAP/AD或数据库作为上游用户数据源，同步至本地</span>
          <el-button type="primary" @click="openConnDialog()">
            <el-icon><Plus /></el-icon> 新增连接器
          </el-button>
        </div>

        <el-table :data="connectors" v-loading="connLoading" stripe size="small">
          <el-table-column prop="name" label="名称" min-width="140">
            <template #default="{ row }"><span class="row-name">{{ row.name }}</span></template>
          </el-table-column>
          <el-table-column label="类型" width="140" align="center">
            <template #default="{ row }">
              <el-tag :type="platformTagType(row.type)" size="small" effect="light">
                {{ platformLabel(row.type) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="地址" min-width="160">
            <template #default="{ row }">
              <span v-if="isIMType(row.type)" class="conn-addr">{{ row.imAppId ? '已配置' : '未配置' }}</span>
              <span v-else class="conn-addr">{{ row.host }}{{ row.port ? ':' + row.port : '' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="SSO" width="70" align="center">
            <template #default="{ row }">
              <el-tag v-if="row.imEnableSso" type="success" size="small" effect="light">开</el-tag>
              <span v-else class="text-muted">-</span>
            </template>
          </el-table-column>
          <el-table-column label="健康" width="90" align="center">
            <template #default="{ row }">
              <span class="health-dot" :class="{ online: row.lastTestOk, offline: row.lastTestAt && !row.lastTestOk }"></span>
              {{ row.lastTestOk ? '正常' : (row.lastTestAt ? '异常' : '未测试') }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link size="small" @click="testConn(row)" :loading="testingId === row.id">测试</el-button>
              <el-button type="primary" link size="small" @click="openConnDialog(row)">编辑</el-button>
              <el-button type="danger" link size="small" @click="deleteConn(row)">删除</el-button>
            </template>
          </el-table-column>
          <template #empty>
            <el-empty description="暂无上游连接器" :image-size="100" />
          </template>
        </el-table>
      </el-tab-pane>

      <!-- ==================== 同步规则 ==================== -->
      <el-tab-pane label="同步规则" name="rules">
        <div class="tab-toolbar">
          <span class="tab-desc">定义从上游数据源同步用户到本地系统的规则</span>
          <el-button type="primary" @click="openRuleDialog()">
            <el-icon><Plus /></el-icon> 新增规则
          </el-button>
        </div>

        <el-table :data="rules" v-loading="ruleLoading" stripe size="small">
          <el-table-column prop="name" label="规则名称" min-width="160">
            <template #default="{ row }"><span class="row-name">{{ row.name }}</span></template>
          </el-table-column>
          <el-table-column label="连接器" width="150">
            <template #default="{ row }">
              <el-tag size="small" effect="light" :type="platformTagType(row.connector?.type)">
                {{ row.connector?.name || '-' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="触发方式" width="200">
            <template #default="{ row }">
              <div class="trigger-tags">
                <template v-if="row.enableSchedule">
                  <el-tag v-if="row.scheduleType === 'interval' || (!row.scheduleType && row.scheduleInterval)" type="primary" size="small" effect="light">
                    间隔 {{ row.scheduleInterval || 60 }}分钟
                  </el-tag>
                  <el-tag v-else type="primary" size="small" effect="light">
                    定点 {{ formatScheduleTimes(row.scheduleTime) }}
                  </el-tag>
                </template>
                <el-tag v-if="row.enableChangeDetect" type="warning" size="small" effect="light">
                  变更检测 {{ row.changeDetectInterval || 60 }}s
                </el-tag>
                <span v-if="!row.enableSchedule && !row.enableChangeDetect" class="text-muted">手动</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small" effect="light">
                {{ row.status === 1 ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="280" fixed="right">
            <template #default="{ row }">
              <el-button type="success" link size="small" @click="triggerRule(row)" :loading="triggeringId === row.id">
                <el-icon><Refresh /></el-icon> 同步
              </el-button>
              <el-button type="primary" link size="small" @click="openMappingDialog(row)">映射</el-button>
              <el-button type="primary" link size="small" @click="openRuleDialog(row)">编辑</el-button>
              <el-button type="danger" link size="small" @click="deleteRule(row)">删除</el-button>
            </template>
          </el-table-column>
          <template #empty>
            <el-empty description="暂无同步规则" :image-size="100" />
          </template>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- ==================== 连接器编辑弹窗 ==================== -->
    <el-dialog v-model="connDialogVisible" :title="connIsEdit ? '编辑上游连接器' : '新增上游连接器'" width="700px" destroy-on-close>
      <el-form :model="connForm" label-width="120px">
        <el-form-item label="连接器名称" required>
          <el-input v-model="connForm.name" placeholder="如：钉钉-总公司 / AD域用户源" />
        </el-form-item>
        <el-form-item label="连接器类型" required>
          <el-select v-model="connForm.type" class="full-width" :disabled="connIsEdit" @change="onConnTypeChange">
            <el-option-group label="IM 平台">
              <el-option label="钉钉 DingTalk" value="im_dingtalk" />
              <el-option label="企业微信 WeChatWork" value="im_wechatwork" />
              <el-option label="飞书 FeiShu" value="im_feishu" />
              <el-option label="WeLink" value="im_welink" />
            </el-option-group>
            <el-option-group label="目录服务">
              <el-option label="LDAP / Active Directory" value="ldap_ad" />
            </el-option-group>
            <el-option-group label="数据库">
              <el-option label="MySQL" value="db_mysql" />
              <el-option label="PostgreSQL" value="db_postgresql" />
              <el-option label="Oracle" value="db_oracle" />
              <el-option label="SQL Server" value="db_sqlserver" />
            </el-option-group>
          </el-select>
        </el-form-item>

        <!-- ============ IM 平台配置 ============ -->
        <template v-if="isIMType(connForm.type)">
          <el-divider content-position="left">应用凭证</el-divider>
          <el-form-item label="AppID / AppKey" required>
            <el-input v-model="connForm.imAppId" placeholder="应用的 AppKey / AppID" />
          </el-form-item>
          <el-form-item label="AppSecret" required>
            <el-input v-model="connForm.imAppSecret" type="password" show-password :placeholder="connIsEdit ? '留空不修改' : '应用的 AppSecret'" />
          </el-form-item>
          <el-form-item label="CorpID" v-if="connForm.type === 'im_dingtalk' || connForm.type === 'im_wechatwork'">
            <el-input v-model="connForm.imCorpId" placeholder="企业 CorpID" />
          </el-form-item>
          <el-form-item label="AgentID" v-if="connForm.type === 'im_dingtalk'">
            <el-input v-model="connForm.imAgentId" placeholder="应用的 AgentID（用于发送工作通知）" />
          </el-form-item>

          <el-divider content-position="left">同步配置</el-divider>
          <el-form-item label="用户匹配字段">
            <el-select v-model="connForm.imMatchField" class="full-width">
              <el-option label="手机号 (mobile)" value="mobile" />
              <el-option label="邮箱 (email)" value="email" />
              <el-option label="IM平台UserID (userid)" value="userid" />
              <el-option label="用户名 (username)" value="username" />
            </el-select>
            <div class="field-hint">用于匹配上游用户与本地用户的字段，修改后会同步到属性映射</div>
          </el-form-item>
          <el-form-item label="用户名生成">
            <el-select v-model="connForm.imUsernameRule" class="full-width">
              <el-option label="中文转拼音（重名加数字）" value="pinyin" />
              <el-option label="邮箱前缀" value="email_prefix" />
              <el-option label="手机号" value="mobile" />
              <el-option label="邮箱" value="email" />
              <el-option label="IM平台UserID" value="userid" />
            </el-select>
            <div class="field-hint">新用户创建时用户名的生成规则，修改后会同步到属性映射</div>
          </el-form-item>
          <el-form-item label="自动注册">
            <el-switch v-model="connForm.imAutoRegister" />
            <span class="field-hint">同步时自动创建本地用户</span>
          </el-form-item>
          <el-form-item label="默认角色">
            <el-select v-model="connForm.imDefaultRoleId" class="full-width" clearable placeholder="新用户默认角色">
              <el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id" />
            </el-select>
          </el-form-item>

          <el-divider content-position="left">SSO 免登</el-divider>
          <el-form-item label="启用SSO" v-if="connForm.type !== 'im_welink'">
            <el-switch v-model="connForm.imEnableSso" />
            <span class="field-hint">允许用户通过该IM平台免登进入系统</span>
          </el-form-item>
          <el-form-item label="SSO按钮文字" v-if="connForm.imEnableSso && connForm.type !== 'im_welink'">
            <el-input v-model="connForm.imSsoLabel" placeholder="如：钉钉登录" />
          </el-form-item>
          <el-form-item label="SSO优先级" v-if="connForm.imEnableSso && connForm.type !== 'im_welink'">
            <el-input-number v-model="connForm.imSsoPriority" :min="0" :max="100" />
            <span class="field-hint">数字越大越靠前</span>
          </el-form-item>
        </template>

        <!-- ============ LDAP/AD 配置 ============ -->
        <template v-if="connForm.type === 'ldap_ad'">
          <el-divider content-position="left">连接参数</el-divider>
          <el-form-item label="服务器地址" required>
            <div class="addr-row">
              <el-input v-model="connForm.host" placeholder="如：ldap.example.com" class="addr-host" />
              <el-input-number v-model="connForm.port" :min="1" :max="65535" class="addr-port" />
            </div>
          </el-form-item>
          <el-form-item label="LDAPS (TLS)">
            <el-switch v-model="connForm.useTls" @change="onTlsChange" />
            <span class="field-hint">启用SSL/TLS加密连接</span>
          </el-form-item>
          <el-form-item label="Base DN" required>
            <el-input v-model="connForm.baseDn" placeholder="dc=example,dc=com" />
          </el-form-item>
          <el-form-item label="Bind DN" required>
            <el-input v-model="connForm.bindDn" placeholder="cn=admin,dc=example,dc=com" />
          </el-form-item>
          <el-form-item label="密码" required>
            <el-input v-model="connForm.bindPassword" type="password" show-password :placeholder="connIsEdit ? '留空不修改' : '绑定密码'" />
          </el-form-item>
          <el-form-item label="用户过滤器">
            <el-input v-model="connForm.userFilter" placeholder="如：(objectClass=person)" />
            <div class="field-hint">LDAP搜索过滤器，留空则使用默认</div>
          </el-form-item>
          <el-form-item label="UPN后缀">
            <el-input v-model="connForm.upnSuffix" placeholder="@example.com" />
          </el-form-item>
        </template>

        <!-- ============ 数据库配置 ============ -->
        <template v-if="isDBType(connForm.type)">
          <el-divider content-position="left">连接参数</el-divider>
          <el-form-item label="服务器地址" required>
            <div class="addr-row">
              <el-input v-model="connForm.host" placeholder="如：192.168.1.100" class="addr-host" />
              <el-input-number v-model="connForm.port" :min="1" :max="65535" class="addr-port" />
            </div>
          </el-form-item>
          <el-form-item :label="connForm.type === 'db_oracle' ? '服务名' : '数据库名'" required>
            <el-input v-model="connForm.database" placeholder="数据库名称" />
          </el-form-item>
          <el-form-item label="用户名" required>
            <el-input v-model="connForm.dbUser" placeholder="数据库用户名" />
          </el-form-item>
          <el-form-item label="密码" required>
            <el-input v-model="connForm.dbPassword" type="password" show-password :placeholder="connIsEdit ? '留空不修改' : '数据库密码'" />
          </el-form-item>
          <el-form-item label="用户表名">
            <el-input v-model="connForm.userTable" placeholder="如：users（存储用户数据的表名）" />
          </el-form-item>
          <el-form-item label="字符集">
            <el-input v-model="connForm.charset" placeholder="utf8mb4" />
          </el-form-item>
        </template>

        <!-- ============ 通用字段 ============ -->
        <template v-if="!isIMType(connForm.type)">
          <el-divider content-position="left">同步配置</el-divider>
          <el-form-item label="自动注册">
            <el-switch v-model="connForm.imAutoRegister" />
            <span class="field-hint">同步时自动创建本地用户</span>
          </el-form-item>
          <el-form-item label="用户匹配字段">
            <el-select v-model="connForm.imMatchField" class="full-width">
              <el-option label="用户名" value="username" />
              <el-option label="手机号" value="mobile" />
              <el-option label="邮箱" value="email" />
            </el-select>
          </el-form-item>
          <el-form-item label="默认角色">
            <el-select v-model="connForm.imDefaultRoleId" class="full-width" clearable placeholder="新用户默认角色">
              <el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id" />
            </el-select>
          </el-form-item>
        </template>

        <el-form-item label="超时(秒)" v-if="!isIMType(connForm.type)">
          <el-input-number v-model="connForm.timeout" :min="1" :max="60" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="connDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveConn" :loading="connSaving">保存</el-button>
      </template>
    </el-dialog>

    <!-- ==================== 规则编辑弹窗 ==================== -->
    <el-dialog v-model="ruleDialogVisible" :title="ruleIsEdit ? '编辑同步规则' : '新增同步规则'" width="560px" destroy-on-close>
      <el-form :model="ruleForm" label-width="110px">
        <el-form-item label="规则名称" required>
          <el-input v-model="ruleForm.name" placeholder="如：钉钉全量同步 / LDAP用户导入" />
        </el-form-item>
        <el-form-item label="上游连接器" required>
          <el-select v-model="ruleForm.connectorId" class="full-width">
            <el-option v-for="c in connectors" :key="c.id" :label="c.name + ' (' + platformLabel(c.type) + ')'" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="同步范围" v-if="isIMType(selectedRuleConnectorType)">
          <el-select v-model="ruleForm.scopeType" class="full-width">
            <el-option label="全部部门" value="all" />
            <el-option label="指定部门" value="selected" />
          </el-select>
        </el-form-item>
        <el-form-item label="指定部门ID" v-if="ruleForm.scopeType === 'selected' && isIMType(selectedRuleConnectorType)">
          <el-input v-model="ruleForm.scopeDeptIds" placeholder="多个ID用逗号分隔" />
        </el-form-item>
        <el-divider content-position="left">定时同步</el-divider>
        <el-form-item label="定时同步">
          <el-switch v-model="ruleForm.enableSchedule" />
        </el-form-item>
        <template v-if="ruleForm.enableSchedule">
          <el-form-item label="调度模式">
            <el-radio-group v-model="ruleForm.scheduleType">
              <el-radio value="times">定点时间</el-radio>
              <el-radio value="interval">固定间隔</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="自动同步时间" v-if="ruleForm.scheduleType === 'times'">
            <div class="schedule-times-list">
              <div v-for="(t, idx) in ruleForm.scheduleTimes" :key="idx" class="schedule-time-row">
                <el-time-picker v-model="ruleForm.scheduleTimes[idx]" format="HH:mm" value-format="HH:mm" placeholder="选择时间" style="width: 140px;" />
                <el-button type="danger" size="small" @click="removeScheduleTime(idx)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button type="primary" size="small" @click="addScheduleTime">
                <el-icon><Plus /></el-icon>
              </el-button>
            </div>
          </el-form-item>
          <el-form-item label="同步间隔(分钟)" v-if="ruleForm.scheduleType === 'interval'">
            <el-input-number v-model="ruleForm.scheduleInterval" :min="5" :max="1440" :step="5" />
          </el-form-item>
        </template>

        <template v-if="isDBType(selectedRuleConnectorType)">
          <el-divider content-position="left">变更检测</el-divider>
          <el-form-item label="启用变更检测">
            <el-switch v-model="ruleForm.enableChangeDetect" />
            <span class="field-hint">当数据库中的数据发生变更时自动触发同步</span>
          </el-form-item>
          <template v-if="ruleForm.enableChangeDetect">
            <el-form-item label="检测间隔(秒)">
              <el-input-number v-model="ruleForm.changeDetectInterval" :min="10" :max="3600" :step="10" />
              <span class="field-hint">每隔多少秒检查一次数据库变更</span>
            </el-form-item>
            <el-form-item label="变更字段">
              <el-input v-model="ruleForm.changeDetectField" placeholder="如：updated_at / modify_time" />
              <div class="field-hint">用于判断记录是否更新的时间戳字段名（数据库表中的列名）</div>
            </el-form-item>
          </template>
        </template>

        <el-divider content-position="left">其他</el-divider>
        <el-form-item label="状态">
          <el-switch v-model="ruleForm.statusBool" active-text="启用" inactive-text="禁用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveRule" :loading="ruleSaving">保存</el-button>
      </template>
    </el-dialog>

    <!-- ==================== 属性映射弹窗 ==================== -->
    <el-dialog v-model="mappingDialogVisible" title="属性映射配置" width="860px" destroy-on-close>
      <div class="mapping-header">
        <div class="mapping-info">
          <span class="mapping-rule-name">规则：{{ mappingRuleName }}</span>
          <el-tag size="small" :type="platformTagType(mappingConnType)" effect="light" style="margin-left: 8px;">
            {{ platformLabel(mappingConnType) }}
          </el-tag>
        </div>
        <div class="mapping-actions">
          <el-button size="small" @click="addMappingRow">
            <el-icon><Plus /></el-icon> 添加映射
          </el-button>
          <el-button size="small" type="warning" @click="resetMappings">
            <el-icon><Refresh /></el-icon> 恢复默认
          </el-button>
        </div>
      </div>

      <el-alert type="info" :closable="false" style="margin-bottom: 12px;">
        <template #title>
          上游属性来源于{{ isIMType(mappingConnType) ? 'IM平台' : mappingConnType === 'ldap_ad' ? 'LDAP/AD目录' : '数据库表' }}，目标属性为本地用户/群组字段。映射类型支持"直接映射"和"转换"两种模式。
        </template>
      </el-alert>

      <el-table :data="mappings" v-loading="mappingLoading" size="small" stripe max-height="420">
        <el-table-column label="启用" width="60" align="center">
          <template #default="{ row }">
            <el-switch v-model="row.isEnabled" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="对象" width="100" align="center">
          <template #default="{ row }">
            <el-select v-model="row.objectType" size="small">
              <el-option label="用户" value="user" />
              <el-option label="群组" value="group" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="上游属性（源）" min-width="180">
          <template #default="{ row }">
            <el-select v-model="row.sourceAttribute" size="small" class="full-width" filterable allow-create placeholder="选择或输入">
              <el-option v-for="opt in getSourceOptions(row.objectType)" :key="opt.value" :label="opt.label" :value="opt.value" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column width="40" align="center">
          <template #default><span style="color: var(--color-text-quaternary); font-size: 16px;">&#8594;</span></template>
        </el-table-column>
        <el-table-column label="本地属性（目标）" min-width="180">
          <template #default="{ row }">
            <el-select v-model="row.targetAttribute" size="small" class="full-width" filterable allow-create placeholder="选择或输入">
              <el-option v-for="opt in getTargetOptions(row.objectType)" :key="opt.value" :label="opt.label" :value="opt.value" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="110" align="center">
          <template #default="{ row }">
            <el-select v-model="row.mappingType" size="small">
              <el-option label="直接映射" value="mapping" />
              <el-option label="转换" value="transform" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="转换规则" width="160">
          <template #default="{ row }">
            <el-select v-if="row.mappingType === 'transform'" v-model="row.transformRule" size="small" class="full-width" placeholder="选择转换规则" clearable>
              <el-option label="中文转拼音（重名加数字）" value="pinyin" />
              <el-option label="邮箱前缀" value="email_prefix" />
              <el-option label="手机号" value="mobile" />
              <el-option label="邮箱" value="email" />
              <el-option label="IM平台UserID" value="userid" />
            </el-select>
            <span v-else class="text-muted" style="font-size: 12px;">-</span>
          </template>
        </el-table-column>
        <el-table-column label="优先级" width="80" align="center">
          <template #default="{ row }">
            <el-input-number v-model="row.priority" size="small" :min="0" :max="100" controls-position="right" />
          </template>
        </el-table-column>
        <el-table-column label="" width="50" align="center">
          <template #default="{ $index }">
            <el-button type="danger" link size="small" @click="removeMappingRow($index)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无映射规则，点击【添加映射】或【恢复默认】" :image-size="60" />
        </template>
      </el-table>

      <template #footer>
        <el-button @click="mappingDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveMappings" :loading="mappingSaving">保存映射</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Plus, Refresh, Delete } from "@element-plus/icons-vue";
import type { Ref } from "vue";
import { syncApi, roleApi } from "../../api";

// ===== 类型判断工具 =====
const isIMType = (type: string) => type?.startsWith('im_');
const isDBType = (type: string) => type?.startsWith('db_');

const platformLabel = (type: string) => {
  const map: Record<string, string> = {
    im_dingtalk: '钉钉', im_wechatwork: '企业微信', im_feishu: '飞书', im_welink: 'WeLink',
    ldap_ad: 'LDAP AD', db_mysql: 'MySQL', db_postgresql: 'PostgreSQL', db_oracle: 'Oracle', db_sqlserver: 'SQL Server'
  };
  return map[type] || type;
};
const platformTagType = (type: string) => {
  if (type?.startsWith('im_')) return 'primary';
  if (type?.startsWith('db_')) return 'success';
  if (type === 'ldap_ad') return 'warning';
  return 'info';
};
const dbTypePorts: Record<string, number> = { db_mysql: 3306, db_postgresql: 5432, db_oracle: 1521, db_sqlserver: 1433 };

const formatScheduleTimes = (raw: string) => {
  if (!raw) return '-';
  try {
    const arr = JSON.parse(raw);
    if (Array.isArray(arr)) return arr.join(', ');
  } catch {}
  return raw;
};

// ===== Tab =====
const activeTab = ref('connectors');

// ===== 连接器 =====
const connectors = ref<any[]>([]);
const connLoading = ref(false);
const testingId = ref(0);
const connDialogVisible = ref(false);
const connIsEdit = ref(false);
const connEditingId = ref(0);
const connSaving = ref(false);
const roles = ref<any[]>([]);

const defaultConnForm = {
  name: '', type: 'im_dingtalk', direction: 'upstream',
  // IM 字段
  imAppId: '', imAppSecret: '', imCorpId: '', imAgentId: '',
  imMatchField: 'mobile', imUsernameRule: 'pinyin',
  imAutoRegister: true, imDefaultRoleId: 0,
  imEnableSso: false, imSsoLabel: '', imSsoPriority: 0,
  // LDAP 字段
  host: '', port: 389, useTls: false,
  baseDn: '', bindDn: '', bindPassword: '', userFilter: '', upnSuffix: '',
  // 数据库字段
  database: '', dbUser: '', dbPassword: '', userTable: '', charset: 'utf8mb4',
  // 通用
  timeout: 5
};
const connForm = ref({ ...defaultConnForm });

const loadConnectors = async () => {
  connLoading.value = true;
  try {
    const res = await syncApi.upstreamConnectors();
    connectors.value = (res as any).data?.data || [];
  } finally { connLoading.value = false; }
};

const loadRoles = async () => {
  try {
    const res = await roleApi.list();
    roles.value = (res as any).data?.data || [];
  } catch {}
};

const openConnDialog = (row?: any) => {
  connForm.value = { ...defaultConnForm };
  if (row) {
    connIsEdit.value = true;
    connEditingId.value = row.id;
    connForm.value = {
      name: row.name, type: row.type, direction: 'upstream',
      imAppId: row.imAppId || '', imAppSecret: '',
      imCorpId: row.imCorpId || '', imAgentId: row.imAgentId || '',
      imMatchField: row.imMatchField || 'mobile',
      imUsernameRule: row.imUsernameRule || 'pinyin',
      imAutoRegister: row.imAutoRegister ?? true,
      imDefaultRoleId: row.imDefaultRoleId || 0,
      imEnableSso: row.imEnableSso ?? false,
      imSsoLabel: row.imSsoLabel || '',
      imSsoPriority: row.imSsoPriority || 0,
      host: row.host || '', port: row.port || 389, useTls: row.useTls ?? false,
      baseDn: row.baseDn || '', bindDn: row.bindDn || '', bindPassword: '',
      userFilter: row.userFilter || '', upnSuffix: row.upnSuffix || '',
      database: row.database || '', dbUser: row.dbUser || '', dbPassword: '',
      userTable: row.userTable || '', charset: row.charset || 'utf8mb4',
      timeout: row.timeout || 5
    };
  } else {
    connIsEdit.value = false;
    connEditingId.value = 0;
  }
  connDialogVisible.value = true;
};

const onConnTypeChange = (val: string) => {
  if (val === 'ldap_ad') {
    connForm.value.port = connForm.value.useTls ? 636 : 389;
    connForm.value.imEnableSso = false;
  } else if (isDBType(val)) {
    connForm.value.port = dbTypePorts[val] || 3306;
    connForm.value.imEnableSso = false;
  } else {
    const labelMap: Record<string, string> = { im_dingtalk: '钉钉登录', im_wechatwork: '企微登录', im_feishu: '飞书登录' };
    connForm.value.imSsoLabel = labelMap[val] || '';
  }
};

const onTlsChange = (val: boolean) => {
  connForm.value.port = val ? 636 : 389;
};

const saveConn = async () => {
  if (!connForm.value.name) {
    ElMessage.warning('请填写连接器名称');
    return;
  }
  const t = connForm.value.type;
  if (isIMType(t) && !connForm.value.imAppId) {
    ElMessage.warning('请填写 AppID / AppKey');
    return;
  }
  if (t === 'ldap_ad' && (!connForm.value.host || !connForm.value.baseDn)) {
    ElMessage.warning('请填写服务器地址和 Base DN');
    return;
  }
  if (isDBType(t) && (!connForm.value.host || !connForm.value.database)) {
    ElMessage.warning('请填写服务器地址和数据库名');
    return;
  }
  connSaving.value = true;
  try {
    if (connIsEdit.value) {
      await syncApi.updateUpstreamConnector(connEditingId.value, connForm.value);
      // 同步匹配字段/用户名规则到关联的同步规则映射
      syncConnSettingsToMappings(connEditingId.value, connForm.value.imMatchField, connForm.value.imUsernameRule);
    } else {
      await syncApi.createUpstreamConnector(connForm.value);
    }
    ElMessage.success('保存成功');
    connDialogVisible.value = false;
    loadConnectors();
  } finally { connSaving.value = false; }
};

// 将连接器的匹配字段/用户名规则同步到关联映射
const syncConnSettingsToMappings = async (connId: number, matchField: string, usernameRule: string) => {
  // 找到该连接器关联的所有规则
  const relatedRules = rules.value.filter((r: any) => r.connectorId === connId);
  for (const rule of relatedRules) {
    try {
      const res = await syncApi.upstreamRuleMappings(rule.id);
      const currentMappings = ((res as any).data?.data || []);
      let changed = false;

      for (const m of currentMappings) {
        // 同步 username 转换规则
        if (m.targetAttribute === 'username' && m.mappingType === 'transform') {
          if (m.transformRule !== usernameRule) {
            m.transformRule = usernameRule;
            changed = true;
          }
        }
      }

      if (changed) {
        await syncApi.updateUpstreamRuleMappings(rule.id, currentMappings);
      }
    } catch {}
  }
};

const testConn = async (row: any) => {
  testingId.value = row.id;
  try {
    const res = await syncApi.testUpstreamConnector(row.id);
    const d = (res as any).data?.data;
    if (d?.ok) {
      ElMessage.success(d?.message || '连接成功');
    } else {
      ElMessage.error(d?.message || '连接失败');
    }
    loadConnectors();
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '连接失败');
    loadConnectors();
  } finally { testingId.value = 0; }
};

const deleteConn = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定删除连接器「${row.name}」？关联的同步规则也会失效。`, '确认删除', { type: 'warning' });
    await syncApi.deleteUpstreamConnector(row.id);
    ElMessage.success('删除成功');
    loadConnectors();
  } catch {}
};

// ===== 同步规则 =====
const rules = ref<any[]>([]);
const ruleLoading = ref(false);
const triggeringId = ref(0);
const ruleDialogVisible = ref(false);
const ruleIsEdit = ref(false);
const ruleEditingId = ref(0);
const ruleSaving = ref(false);

const defaultRuleForm = {
  name: '', connectorId: 0, direction: 'upstream',
  scopeType: 'all', scopeDeptIds: '',
  enableSchedule: false,
  scheduleType: 'times' as string,
  scheduleTimes: [] as string[],
  scheduleInterval: 60,
  enableChangeDetect: false,
  changeDetectInterval: 60,
  changeDetectField: 'updated_at',
  statusBool: true
};
const ruleForm = ref({ ...defaultRuleForm });

const addScheduleTime = () => {
  ruleForm.value.scheduleTimes.push('08:00');
};
const removeScheduleTime = (idx: number) => {
  ruleForm.value.scheduleTimes.splice(idx, 1);
};

const selectedRuleConnectorType = computed(() => {
  const c = connectors.value.find((x: any) => x.id === ruleForm.value.connectorId);
  return c?.type || '';
});

const loadRules = async () => {
  ruleLoading.value = true;
  try {
    const res = await syncApi.upstreamRules();
    rules.value = (res as any).data?.data || [];
  } finally { ruleLoading.value = false; }
};

const parseScheduleTimes = (raw: string): string[] => {
  if (!raw) return [];
  try {
    const arr = JSON.parse(raw);
    if (Array.isArray(arr)) return arr;
  } catch {}
  if (raw.includes(',')) return raw.split(',').map((s: string) => s.trim()).filter(Boolean);
  return [raw];
};

const openRuleDialog = (row?: any) => {
  ruleForm.value = { ...defaultRuleForm, scheduleTimes: [] };
  if (row) {
    ruleIsEdit.value = true;
    ruleEditingId.value = row.id;
    ruleForm.value = {
      name: row.name,
      connectorId: row.connectorId,
      direction: 'upstream',
      scopeType: row.scopeType || 'all',
      scopeDeptIds: row.scopeDeptIds || '',
      enableSchedule: row.enableSchedule ?? false,
      scheduleType: row.scheduleType || 'times',
      scheduleTimes: parseScheduleTimes(row.scheduleTime),
      scheduleInterval: row.scheduleInterval || 60,
      enableChangeDetect: row.enableChangeDetect ?? false,
      changeDetectInterval: row.changeDetectInterval || 60,
      changeDetectField: row.changeDetectField || 'updated_at',
      statusBool: row.status === 1
    };
  } else {
    ruleIsEdit.value = false;
    ruleEditingId.value = 0;
    if (connectors.value.length > 0) {
      ruleForm.value.connectorId = connectors.value[0].id;
    }
  }
  ruleDialogVisible.value = true;
};

const saveRule = async () => {
  if (!ruleForm.value.name || !ruleForm.value.connectorId) {
    ElMessage.warning('请填写必填项');
    return;
  }
  if (ruleForm.value.enableSchedule && ruleForm.value.scheduleType === 'times' && ruleForm.value.scheduleTimes.length === 0) {
    ElMessage.warning('请至少添加一个定时同步时间');
    return;
  }
  ruleSaving.value = true;
  try {
    const payload = {
      ...ruleForm.value,
      status: ruleForm.value.statusBool ? 1 : 0
    };
    if (ruleIsEdit.value) {
      await syncApi.updateUpstreamRule(ruleEditingId.value, payload);
    } else {
      await syncApi.createUpstreamRule(payload);
    }
    ElMessage.success('保存成功');
    ruleDialogVisible.value = false;
    loadRules();
  } finally { ruleSaving.value = false; }
};

const triggerRule = async (row: any) => {
  triggeringId.value = row.id;
  try {
    await syncApi.triggerUpstreamRule(row.id);
    ElMessage({ message: '上游同步已触发，正在后台执行...', type: 'success', duration: 5000 });
    // 定时刷新规则列表以获取最新同步状态
    const poll = setInterval(() => loadRules(), 3000);
    setTimeout(() => clearInterval(poll), 30000); // 最多轮询30秒
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '同步触发失败');
  } finally { triggeringId.value = 0; }
};

const deleteRule = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定删除规则「${row.name}」？`, '确认删除', { type: 'warning' });
    await syncApi.deleteUpstreamRule(row.id);
    ElMessage.success('删除成功');
    loadRules();
  } catch {}
};

// ===== 属性映射 =====
const mappingDialogVisible = ref(false);
const mappingLoading = ref(false);
const mappingSaving = ref(false);
const mappingRuleId = ref(0);
const mappingRuleName = ref('');
const mappingConnType = ref('');
const mappingConnId = ref(0);
const mappings = ref<any[]>([]);

// 上游源属性建议
// ===== 上游源属性选项（中文标签） =====
const imSourceOptions = [
  { value: 'name', label: '姓名 (name)' },
  { value: 'mobile', label: '手机号 (mobile)' },
  { value: 'email', label: '邮箱 (email)' },
  { value: 'userid', label: '用户ID (userid)' },
  { value: 'avatar', label: '头像 (avatar)' },
  { value: 'title', label: '职位 (title)' },
  { value: 'unionid', label: 'UnionID (unionid)' },
  { value: 'openid', label: 'OpenID (openid)' },
];
const ldapSourceOptions = [
  { value: 'sAMAccountName', label: '登录名 (sAMAccountName)' },
  { value: 'cn', label: '通用名 (cn)' },
  { value: 'displayName', label: '显示名 (displayName)' },
  { value: 'mail', label: '邮箱 (mail)' },
  { value: 'mobile', label: '手机号 (mobile)' },
  { value: 'telephoneNumber', label: '电话 (telephoneNumber)' },
  { value: 'title', label: '职位 (title)' },
  { value: 'department', label: '部门 (department)' },
  { value: 'userPrincipalName', label: 'UPN (userPrincipalName)' },
  { value: 'givenName', label: '名 (givenName)' },
  { value: 'sn', label: '姓 (sn)' },
];
const dbSourceOptions = [
  { value: 'username', label: '用户名 (username)' },
  { value: 'name', label: '姓名 (name)' },
  { value: 'display_name', label: '显示名 (display_name)' },
  { value: 'email', label: '邮箱 (email)' },
  { value: 'phone', label: '手机号 (phone)' },
  { value: 'mobile', label: '手机号 (mobile)' },
  { value: 'password', label: '密码 (password)' },
  { value: 'department', label: '部门 (department)' },
  { value: 'position', label: '职位 (position)' },
];
const imGroupSourceOptions = [
  { value: 'name', label: '部门名称 (name)' },
  { value: 'deptId', label: '部门ID (deptId)' },
  { value: 'parentId', label: '父部门ID (parentId)' },
];
const ldapGroupSourceOptions = [
  { value: 'ou', label: '组织单元 (ou)' },
  { value: 'cn', label: '通用名 (cn)' },
  { value: 'description', label: '描述 (description)' },
];

// ===== 本地目标属性选项（中文标签） =====
const localUserOptions = [
  { value: 'username', label: '用户名 (username)' },
  { value: 'nickname', label: '昵称 (nickname)' },
  { value: 'email', label: '邮箱 (email)' },
  { value: 'phone', label: '手机号 (phone)' },
  { value: 'avatar', label: '头像 (avatar)' },
  { value: 'position', label: '职位 (position)' },
  { value: 'department', label: '部门 (department)' },
  { value: 'im_user_id', label: 'IM用户ID (im_user_id)' },
  { value: 'password_hash', label: '密码哈希 (password_hash)' },
];
const localGroupOptions = [
  { value: 'name', label: '群组名称 (name)' },
  { value: 'description', label: '描述 (description)' },
  { value: 'remote_dept_id', label: '远程部门ID (remote_dept_id)' },
];

const getSourceOptions = (objectType: string) => {
  if (objectType === 'group') {
    return isIMType(mappingConnType.value) ? imGroupSourceOptions : ldapGroupSourceOptions;
  }
  if (isIMType(mappingConnType.value)) return imSourceOptions;
  if (mappingConnType.value === 'ldap_ad') return ldapSourceOptions;
  return dbSourceOptions;
};

const getTargetOptions = (objectType: string) => {
  return objectType === 'group' ? localGroupOptions : localUserOptions;
};

const openMappingDialog = async (rule: any) => {
  mappingRuleId.value = rule.id;
  mappingRuleName.value = rule.name;
  mappingConnType.value = rule.connector?.type || '';
  mappingConnId.value = rule.connectorId || rule.connector?.id || 0;
  mappingDialogVisible.value = true;
  await loadMappings();
};

const loadMappings = async () => {
  mappingLoading.value = true;
  try {
    const res = await syncApi.upstreamRuleMappings(mappingRuleId.value);
    mappings.value = ((res as any).data?.data || []).map((m: any) => ({
      ...m,
      isEnabled: m.isEnabled ?? true
    }));
  } finally { mappingLoading.value = false; }
};

const addMappingRow = () => {
  const maxPriority = mappings.value.reduce((max: number, m: any) => Math.max(max, m.priority || 0), 0);
  mappings.value.push({
    objectType: 'user',
    sourceAttribute: '',
    targetAttribute: '',
    mappingType: 'mapping',
    transformRule: '',
    priority: maxPriority + 1,
    isEnabled: true
  });
};

const removeMappingRow = (index: number) => {
  mappings.value.splice(index, 1);
};

const saveMappings = async () => {
  mappingSaving.value = true;
  try {
    await syncApi.updateUpstreamRuleMappings(mappingRuleId.value, mappings.value);

    // 从映射中提取 username 转换规则，同步回连接器
    const usernameMapping = mappings.value.find(
      (m: any) => m.targetAttribute === 'username' && m.mappingType === 'transform' && m.isEnabled
    );
    if (usernameMapping && mappingConnId.value) {
      const newRule = usernameMapping.transformRule || '';
      // 获取当前连接器信息
      const conn = connectors.value.find((c: any) => c.id === mappingConnId.value);
      if (conn && conn.imUsernameRule !== newRule) {
        await syncApi.updateUpstreamConnector(mappingConnId.value, { imUsernameRule: newRule });
        loadConnectors(); // 刷新连接器列表
      }
    }

    ElMessage.success('映射已保存');
    mappingDialogVisible.value = false;
  } finally { mappingSaving.value = false; }
};

const resetMappings = async () => {
  try {
    await ElMessageBox.confirm('将根据连接器类型恢复默认属性映射，当前自定义映射将被覆盖。确定？', '恢复默认映射', { type: 'warning' });
    mappingLoading.value = true;
    const res = await syncApi.resetUpstreamRuleMappings(mappingRuleId.value);
    mappings.value = ((res as any).data?.data || []).map((m: any) => ({
      ...m,
      isEnabled: m.isEnabled ?? true
    }));
    ElMessage.success('已恢复默认映射');
  } catch {} finally { mappingLoading.value = false; }
};

// ===== 初始化 =====
onMounted(() => {
  loadConnectors();
  loadRules();
  loadRoles();
});
</script>

<style scoped>
.upstream-page { display: flex; flex-direction: column; gap: var(--spacing-lg); }
.tab-toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.tab-desc { font-size: var(--font-size-sm); color: var(--color-text-tertiary); }
.row-name { font-weight: 500; color: var(--color-text-primary); }
.trigger-tags { display: flex; gap: 6px; flex-wrap: wrap; }
.text-muted { color: var(--color-text-tertiary); font-size: var(--font-size-sm); }
.full-width { width: 100%; }
.field-hint { margin-left: var(--spacing-sm); color: var(--color-text-tertiary); font-size: var(--font-size-xs); }
.conn-addr { font-family: monospace; font-size: var(--font-size-sm); color: var(--color-text-secondary); }
.addr-row { display: flex; gap: var(--spacing-sm); }
.addr-host { flex: 1; }
.addr-port { width: 120px; }
.health-dot {
  display: inline-block; width: 8px; height: 8px; border-radius: 50%;
  background: var(--color-text-quaternary); margin-right: 4px; vertical-align: middle;
}
.health-dot.online { background: var(--color-success); box-shadow: 0 0 0 3px rgba(82,196,26,0.15); }
.health-dot.offline { background: var(--color-error); box-shadow: 0 0 0 3px rgba(255,77,79,0.15); }
/* 多时间选择 */
.schedule-times-list { display: flex; flex-direction: column; gap: 8px; }
.schedule-time-row { display: flex; align-items: center; gap: 8px; }
/* 属性映射弹窗 */
.mapping-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.mapping-info { display: flex; align-items: center; }
.mapping-rule-name { font-weight: 500; font-size: 14px; color: var(--color-text-primary); }
.mapping-actions { display: flex; gap: 8px; }
</style>
