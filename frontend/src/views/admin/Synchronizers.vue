<template>
  <div class="sync-page">
    <!-- 同步器列表 -->
    <el-card v-if="!editingSyncId">
      <template #header>
        <div class="card-header-row">
          <span class="card-title">同步器管理</span>
          <el-button type="primary" @click="openCreateDialog">
            <el-icon><Plus /></el-icon> 新增同步器
          </el-button>
        </div>
      </template>

      <el-table :data="list" v-loading="loading" stripe size="small">
        <el-table-column prop="name" label="名称" min-width="150">
          <template #default="{ row }">
            <span class="sync-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="连接器" width="150">
          <template #default="{ row }">
            <el-tag size="small" effect="light" :type="row.connector?.type === 'ldap_ad' ? 'primary' : 'success'">
              {{ row.connector?.name || '-' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="触发方式" min-width="160">
          <template #default="{ row }">
            <div class="trigger-tags">
              <el-tag v-if="row.enableEvent" type="warning" size="small" effect="light">
                <span class="trigger-label">⚡ 事件</span>
              </el-tag>
              <el-tag v-if="row.enableSchedule" type="primary" size="small" effect="light">
                <span class="trigger-label">🕐 {{ row.scheduleTime }}</span>
              </el-tag>
              <span v-if="!row.enableEvent && !row.enableSchedule" class="text-muted">手动</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="最后同步" width="180">
          <template #default="{ row }">
            <template v-if="row.lastSyncAt">
              <div class="last-sync" :class="row.lastSyncStatus === 'success' ? 'sync-ok' : 'sync-fail'">
                <span class="sync-time">{{ formatTime(row.lastSyncAt) }}</span>
                <el-tag :type="row.lastSyncStatus === 'success' ? 'success' : 'danger'" size="small" effect="light">
                  {{ row.lastSyncStatus === 'success' ? '成功' : '失败' }}
                </el-tag>
              </div>
            </template>
            <span v-else class="text-muted">从未同步</span>
          </template>
        </el-table-column>
        <el-table-column prop="syncCount" label="累计" width="70" align="center" />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small" effect="light">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="enterDetail(row.id)">
              <el-icon><Edit /></el-icon> 编辑规则
            </el-button>
            <el-button type="success" link size="small" @click="triggerSync(row)" :loading="triggeringId === row.id">
              <el-icon><Refresh /></el-icon> 立即同步
            </el-button>
            <el-dropdown trigger="click" @command="(cmd: string) => handleMore(cmd, row)">
              <el-button type="info" link size="small">更多</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="delete" class="text-error">删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无同步器，点击上方按钮新增" :image-size="120" />
        </template>
      </el-table>
    </el-card>

    <!-- 同步器详情 -->
    <div v-if="editingSyncId" class="sync-detail">
      <div class="detail-header">
        <el-button @click="exitDetail" :icon="ArrowLeft">返回</el-button>
        <span class="detail-title">{{ detailData.name }}</span>
        <el-tag :type="detailData.connector?.type === 'ldap_ad' ? 'primary' : 'success'" size="small" effect="light">
          {{ detailData.connector?.type === 'ldap_ad' ? 'LDAP AD' : 'MySQL' }}
        </el-tag>
        <div class="detail-header-actions">
          <el-button type="primary" @click="saveDetail" :loading="savingDetail">保存配置</el-button>
          <el-button @click="triggerSync(detailData)" :loading="triggeringId === editingSyncId">
            <el-icon><Refresh /></el-icon> 立即同步
          </el-button>
        </div>
      </div>

      <el-tabs v-model="activeTab" type="border-card">
        <!-- 基本信息 -->
        <el-tab-pane label="基本信息" name="basic">
          <div class="detail-sections">
            <!-- 基本配置区 -->
            <div class="section-card">
              <div class="section-title">基本配置</div>
              <el-form :model="detailForm" label-width="110px" class="detail-form">
                <el-form-item label="同步器名称">
                  <el-input v-model="detailForm.name" />
                </el-form-item>
                <el-form-item label="连接器">
                  <el-select v-model="detailForm.connectorId" class="full-width">
                    <el-option v-for="c in connectors" :key="c.id" :label="c.name + ' (' + (c.type === 'ldap_ad' ? 'AD' : 'MySQL') + ')'" :value="c.id" />
                  </el-select>
                </el-form-item>
                <el-form-item label="目标容器" v-if="selectedConnectorType === 'ldap_ad'">
                  <el-input v-model="detailForm.targetContainer" placeholder="cn=users,dc=example,dc=com" />
                  <div class="field-hint">AD 中用户存放的 OU 或容器 DN</div>
                </el-form-item>
                <el-form-item label="状态">
                  <el-switch v-model="detailForm.statusBool" active-text="启用" inactive-text="禁用" />
                </el-form-item>
              </el-form>
            </div>

            <!-- 触发方式区 -->
            <div class="section-card">
              <div class="section-title">触发方式</div>
              <div class="trigger-config-grid">
                <div class="trigger-option" :class="{ active: detailForm.enableSchedule }">
                  <div class="trigger-option-header">
                    <el-switch v-model="detailForm.enableSchedule" />
                    <span class="trigger-option-label">🕐 定时同步</span>
                  </div>
                  <div v-if="detailForm.enableSchedule" class="trigger-option-body">
                    <el-form label-width="80px">
                      <el-form-item label="执行时间">
                        <el-time-picker v-model="scheduleTimeObj" format="HH:mm" value-format="HH:mm" placeholder="选择时间" style="width: 140px" />
                      </el-form-item>
                    </el-form>
                  </div>
                  <div v-else class="trigger-option-desc">每天在指定时间自动执行同步</div>
                </div>
                <div class="trigger-option" :class="{ active: detailForm.enableEvent }">
                  <div class="trigger-option-header">
                    <el-switch v-model="detailForm.enableEvent" />
                    <span class="trigger-option-label">⚡ 事件驱动</span>
                  </div>
                  <div v-if="!detailForm.enableEvent" class="trigger-option-desc">用户变更时自动触发同步</div>
                  <div v-else class="trigger-option-desc active-desc">已启用，可在「同步事件」标签页配置具体事件</div>
                </div>
              </div>
            </div>

            <!-- AD 策略区（仅 AD 类型显示） -->
            <div class="section-card" v-if="selectedConnectorType === 'ldap_ad'">
              <div class="section-title">AD 策略</div>
              <el-form :model="detailForm" label-width="110px" class="detail-form">
                <el-form-item label="禁止用户改密">
                  <el-switch v-model="detailForm.preventPwdChange" />
                  <div class="field-hint">同步时设置 AD 用户"不能更改密码"，确保员工只能通过本系统修改密码</div>
                </el-form-item>
              </el-form>
            </div>
          </div>
        </el-tab-pane>

        <!-- 同步事件 -->
        <el-tab-pane label="同步事件" name="events">
          <el-alert type="info" :closable="false" show-icon class="tab-alert">
            勾选的事件发生时，将自动触发同步，将最新数据推送到目标系统。
          </el-alert>
          <el-table :data="allEvents">
            <el-table-column label="事件" min-width="200">
              <template #default="{ row }">
                <el-checkbox v-model="row.enabled" @change="saveEvents">{{ row.label }}</el-checkbox>
              </template>
            </el-table-column>
            <el-table-column label="标识" width="180">
              <template #default="{ row }">
                <code class="event-code">{{ row.key }}</code>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 属性映射 -->
        <el-tab-pane label="属性映射" name="mappings">
          <el-tabs v-model="mappingTab" type="card">
            <el-tab-pane label="用户属性映射" name="user" />
            <el-tab-pane label="组织属性映射" name="group" />
            <el-tab-pane label="角色属性映射" name="role" />
          </el-tabs>

          <div class="mapping-toolbar">
            <el-button type="primary" size="small" @click="addMapping">
              <el-icon><Plus /></el-icon> 新增映射
            </el-button>
          </div>

          <el-table :data="filteredMappings" size="small">
            <el-table-column label="本地属性" min-width="180">
              <template #default="{ row }">
                <el-select v-model="row.sourceAttribute" size="small" filterable class="full-width">
                  <el-option v-for="f in sourceFieldsForTab" :key="f.key" :label="f.label" :value="f.key" />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column label="" width="50" align="center">
              <template #default>
                <span class="mapping-arrow">&rarr;</span>
              </template>
            </el-table-column>
            <el-table-column label="目标属性" min-width="180">
              <template #default="{ row }">
                <el-select v-model="row.targetAttribute" size="small" filterable allow-create class="full-width">
                  <el-option v-for="f in targetFieldsForTab" :key="f.key" :label="f.label" :value="f.key" />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column label="映射方式" width="130">
              <template #default="{ row }">
                <el-select v-model="row.mappingType" size="small" class="full-width">
                  <el-option label="直接映射" value="mapping" />
                  <el-option label="转换" value="transform" />
                  <el-option label="常量" value="constant" />
                  <el-option label="表达式" value="expression" />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column label="转换规则" min-width="160">
              <template #default="{ row }">
                <el-input v-if="row.mappingType !== 'mapping'" v-model="row.transformRule" size="small" placeholder="如: append:@domain.com" />
                <span v-else class="text-muted-light">-</span>
              </template>
            </el-table-column>
            <el-table-column label="启用" width="60" align="center">
              <template #default="{ row }">
                <el-switch v-model="row.isEnabled" size="small" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="60" align="center">
              <template #default="{ $index }">
                <el-button type="danger" link size="small" @click="removeMapping($index)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="mapping-footer">
            <el-button type="primary" @click="saveMappings" :loading="savingMappings">保存映射</el-button>
          </div>
        </el-tab-pane>

        <!-- 同步日志 -->
        <el-tab-pane label="同步日志" name="logs">
          <el-table :data="syncLogs" v-loading="logsLoading" size="small" row-key="id">
            <el-table-column type="expand">
              <template #default="{ row }">
                <div class="log-expand">
                  <div v-if="row.detail" class="log-detail-box">{{ row.detail }}</div>
                  <div v-else class="text-muted">暂无详细信息</div>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="时间" width="155">
              <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
            </el-table-column>
            <el-table-column prop="triggerType" label="触发" width="70" align="center">
              <template #default="{ row }">
                <el-tag size="small" effect="light" :type="row.triggerType === 'event' ? 'warning' : (row.triggerType === 'schedule' ? 'primary' : 'info')">
                  {{ triggerTypeMap[row.triggerType as string] || row.triggerType }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="triggerEvent" label="事件" width="110">
              <template #default="{ row }">
                {{ eventMap[row.triggerEvent as string] || row.triggerEvent || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="username" label="用户" width="100" show-overflow-tooltip />
            <el-table-column prop="status" label="状态" width="80" align="center">
              <template #default="{ row }">
                <el-tag :type="row.status === 'success' ? 'success' : (row.status === 'partial' ? 'warning' : 'danger')" size="small" effect="light">
                  {{ statusMap[row.status as string] || row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="message" label="概要" min-width="180" show-overflow-tooltip />
            <el-table-column label="详情" width="70" align="center">
              <template #default="{ row }">
                <el-tag v-if="row.detail" type="danger" size="small" effect="light">展开</el-tag>
                <span v-else class="text-muted-light">-</span>
              </template>
            </el-table-column>
            <el-table-column label="耗时" width="70" align="center">
              <template #default="{ row }">{{ row.duration ? row.duration + 'ms' : '-' }}</template>
            </el-table-column>
          </el-table>
          <div class="log-pagination">
            <el-pagination
              v-model:current-page="logPage"
              :page-size="20"
              :total="logTotal"
              layout="prev, pager, next"
              @current-change="loadLogs"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 新增同步器弹窗 -->
    <el-dialog v-model="createDialogVisible" title="新增同步器" width="500px" destroy-on-close>
      <el-form :model="createForm" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="createForm.name" placeholder="如：钉钉-AD同步" />
        </el-form-item>
        <el-form-item label="连接器" required>
          <el-select v-model="createForm.connectorId" class="full-width">
            <el-option v-for="c in connectors" :key="c.id" :label="c.name + ' (' + (c.type === 'ldap_ad' ? 'AD' : 'MySQL') + ')'" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="启用事件触发">
          <el-switch v-model="createForm.enableEvent" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createSync" :loading="creating">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Plus, Delete, ArrowLeft, Edit, Refresh } from "@element-plus/icons-vue";
import { connectorApi, synchronizerApi } from "../../api";

const triggerTypeMap: Record<string, string> = { event: '事件', schedule: '定时', manual: '手动' };
const eventMap: Record<string, string> = {
  password_change: '密码修改', full_sync: '全量同步', user_create: '用户创建',
  user_update: '用户更新', user_delete: '用户删除', user_status_change: '状态变更', role_change: '角色变更'
};
const statusMap: Record<string, string> = { success: '成功', partial: '部分', failed: '失败' };

// ===== 列表 =====
const list = ref<any[]>([]);
const loading = ref(false);
const connectors = ref<any[]>([]);
const triggeringId = ref(0);

const loadList = async () => {
  loading.value = true;
  try {
    const [syncRes, connRes] = await Promise.all([synchronizerApi.list(), connectorApi.list()]);
    list.value = (syncRes as any).data?.data || [];
    connectors.value = (connRes as any).data?.data || [];
  } finally { loading.value = false; }
};

// ===== 新增 =====
const createDialogVisible = ref(false);
const creating = ref(false);
const createForm = reactive({ name: "", connectorId: 0 as number, enableEvent: true });

const openCreateDialog = () => {
  createForm.name = "";
  createForm.connectorId = connectors.value[0]?.id || 0;
  createForm.enableEvent = true;
  createDialogVisible.value = true;
};

const createSync = async () => {
  if (!createForm.name || !createForm.connectorId) { ElMessage.warning("请填写必填项"); return; }
  creating.value = true;
  try {
    const res = await synchronizerApi.create({ ...createForm, events: ["password_change","user_update","user_create","user_delete","role_change"], syncUsers: true });
    ElMessage.success("创建成功");
    createDialogVisible.value = false;
    loadList();
    enterDetail((res as any).data?.data?.id);
  } finally { creating.value = false; }
};

const handleMore = (cmd: string, row: any) => { if (cmd === 'delete') deleteSync(row); };

const deleteSync = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定删除同步器「${row.name}」？相关映射和日志也会被删除。`, "确认删除", { type: "warning" });
    await synchronizerApi.delete(row.id);
    ElMessage.success("删除成功");
    loadList();
  } catch {}
};

const triggerSync = async (row: any) => {
  triggeringId.value = row.id;
  try {
    await synchronizerApi.trigger(row.id);
    ElMessage.success("同步已触发");
    setTimeout(() => { loadList(); loadLogs(); }, 3000);
  } finally { triggeringId.value = 0; }
};

// ===== 详情 =====
const editingSyncId = ref(0);
const activeTab = ref("basic");
const detailData = ref<any>({});
const detailForm = reactive({ name: "", connectorId: 0, targetContainer: "", enableSchedule: false, enableEvent: true, statusBool: true, preventPwdChange: true });
const scheduleTimeObj = ref("");
const savingDetail = ref(false);

const selectedConnectorType = computed(() => { const c = connectors.value.find((x: any) => x.id === detailForm.connectorId); return c?.type || ""; });

const enterDetail = async (id: number) => {
  editingSyncId.value = id;
  activeTab.value = "basic";
  await loadDetail();
  await loadAllEvents();
  await loadMappingMeta();
  await loadMappings();
  await loadLogs();
};

const exitDetail = () => { editingSyncId.value = 0; loadList(); };

const loadDetail = async () => {
  const res = await synchronizerApi.get(editingSyncId.value);
  const data = (res as any).data?.data;
  detailData.value = data.synchronizer;
  allMappings.value = data.mappings || [];
  detailForm.name = detailData.value.name;
  detailForm.connectorId = detailData.value.connectorId;
  detailForm.targetContainer = detailData.value.targetContainer || "";
  detailForm.enableSchedule = detailData.value.enableSchedule;
  detailForm.enableEvent = detailData.value.enableEvent;
  detailForm.preventPwdChange = detailData.value.preventPwdChange ?? true;
  detailForm.statusBool = detailData.value.status === 1;
  scheduleTimeObj.value = detailData.value.scheduleTime || "";
};

const saveDetail = async () => {
  savingDetail.value = true;
  try {
    await synchronizerApi.update(editingSyncId.value, {
      name: detailForm.name, connectorId: detailForm.connectorId, targetContainer: detailForm.targetContainer,
      enableSchedule: detailForm.enableSchedule, scheduleTime: scheduleTimeObj.value || "",
      enableEvent: detailForm.enableEvent, preventPwdChange: detailForm.preventPwdChange,
      status: detailForm.statusBool ? 1 : 0
    });
    ElMessage.success("保存成功");
    loadDetail();
  } finally { savingDetail.value = false; }
};

// ===== 事件 =====
const allEvents = ref<any[]>([]);
const loadAllEvents = async () => {
  const res = await synchronizerApi.events();
  const eventDefs = (res as any).data?.data || [];
  let selected: string[] = [];
  try { selected = JSON.parse(detailData.value.events || "[]"); } catch { selected = []; }
  allEvents.value = eventDefs.map((e: any) => ({ ...e, enabled: selected.includes(e.key) }));
};
const saveEvents = async () => {
  const enabled = allEvents.value.filter((e: any) => e.enabled).map((e: any) => e.key);
  await synchronizerApi.update(editingSyncId.value, { events: enabled });
  ElMessage.success("事件配置已保存");
};

// ===== 映射 =====
const mappingTab = ref("user");
const allMappings = ref<any[]>([]);
const sourceFields = ref<Record<string, any[]>>({ user: [], group: [], role: [] });
const targetFields = ref<Record<string, any[]>>({ user: [], group: [], role: [] });
const savingMappings = ref(false);
const filteredMappings = computed(() => allMappings.value.filter((m: any) => m.objectType === mappingTab.value));
const sourceFieldsForTab = computed(() => sourceFields.value[mappingTab.value] || []);
const targetFieldsForTab = computed(() => targetFields.value[mappingTab.value] || []);

const loadMappingMeta = async () => {
  for (const t of ["user", "group", "role"]) {
    try { const sRes = await synchronizerApi.sourceFields(t); sourceFields.value[t] = (sRes as any).data?.data || []; } catch { sourceFields.value[t] = []; }
  }
};
watch(() => detailForm.connectorId, async (newId) => {
  if (!newId) return;
  for (const t of ["user", "group", "role"]) {
    try { const tRes = await synchronizerApi.targetFields(newId, t); targetFields.value[t] = (tRes as any).data?.data || []; } catch { targetFields.value[t] = []; }
  }
}, { immediate: true });

const loadMappings = async () => { const res = await synchronizerApi.getMappings(editingSyncId.value); allMappings.value = (res as any).data?.data || []; };
const addMapping = () => { allMappings.value.push({ synchronizerId: editingSyncId.value, objectType: mappingTab.value, sourceAttribute: "", targetAttribute: "", mappingType: "mapping", transformRule: "", priority: filteredMappings.value.length + 1, isEnabled: true }); };
const removeMapping = (filteredIndex: number) => { const item = filteredMappings.value[filteredIndex]; const realIndex = allMappings.value.indexOf(item); if (realIndex !== -1) allMappings.value.splice(realIndex, 1); };
const saveMappings = async () => {
  savingMappings.value = true;
  try { let p = 0; for (const m of allMappings.value) m.priority = p++; await synchronizerApi.batchUpdateMappings(editingSyncId.value, allMappings.value); ElMessage.success("映射已保存"); loadMappings(); } finally { savingMappings.value = false; }
};

// ===== 日志 =====
const syncLogs = ref<any[]>([]);
const logsLoading = ref(false);
const logPage = ref(1);
const logTotal = ref(0);
const loadLogs = async () => {
  if (!editingSyncId.value) return;
  logsLoading.value = true;
  try { const res = await synchronizerApi.logs(editingSyncId.value, { page: logPage.value, size: 20 }); const data = (res as any).data?.data; syncLogs.value = data?.list || []; logTotal.value = data?.total || 0; } finally { logsLoading.value = false; }
};

const formatTime = (t: string) => { if (!t) return "-"; return new Date(t).toLocaleString("zh-CN"); };
onMounted(loadList);
</script>

<style scoped>
.sync-page { display: flex; flex-direction: column; gap: var(--spacing-lg); }

/* 卡片头 */
.card-header-row { display: flex; justify-content: space-between; align-items: center; }
.card-title { font-size: var(--font-size-lg); font-weight: 600; color: var(--color-text-primary); }

/* 列表 */
.sync-name { font-weight: 500; color: var(--color-text-primary); }
.trigger-tags { display: flex; gap: 6px; flex-wrap: wrap; align-items: center; }
.trigger-label { display: inline-flex; align-items: center; gap: 2px; }

/* 最后同步 */
.last-sync { display: flex; flex-direction: column; gap: 4px; }
.sync-time { font-size: var(--font-size-sm); color: var(--color-text-tertiary); }

/* 详情 */
.sync-detail { min-height: 400px; }
.detail-header { display: flex; align-items: center; gap: var(--spacing-md); margin-bottom: var(--spacing-lg); flex-wrap: wrap; }
.detail-title { font-size: var(--font-size-xl); font-weight: 600; color: var(--color-text-primary); }
.detail-header-actions { margin-left: auto; display: flex; gap: var(--spacing-sm); }
.detail-form { max-width: 560px; }

/* 分区卡片 */
.detail-sections { display: flex; flex-direction: column; gap: var(--spacing-lg); }
.section-card {
  background: var(--color-bg-container);
  border: 1px solid var(--color-border-secondary);
  border-radius: var(--radius-lg);
  padding: var(--spacing-xl);
}
.section-title {
  font-size: var(--font-size-base);
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-lg);
  padding-bottom: var(--spacing-sm);
  border-bottom: 1px solid var(--color-border-secondary);
}

/* 触发方式配置网格 */
.trigger-config-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-md);
}
.trigger-option {
  border: 1px solid var(--color-border-secondary);
  border-radius: var(--radius-md);
  padding: var(--spacing-lg);
  transition: all 0.2s;
}
.trigger-option.active {
  border-color: var(--color-primary-border);
  background: var(--color-primary-bg);
}
.trigger-option-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
}
.trigger-option-label {
  font-weight: 500;
  color: var(--color-text-primary);
  font-size: var(--font-size-base);
}
.trigger-option-body {
  margin-top: var(--spacing-md);
  padding-top: var(--spacing-md);
  border-top: 1px dashed var(--color-border-secondary);
}
.trigger-option-desc {
  font-size: var(--font-size-sm);
  color: var(--color-text-tertiary);
  line-height: 1.5;
}
.trigger-option-desc.active-desc {
  color: var(--color-primary);
}

/* 通用 */
.text-muted { color: var(--color-text-tertiary); font-size: var(--font-size-sm); }
.text-muted-light { color: var(--color-text-quaternary); font-size: var(--font-size-xs); }
.full-width { width: 100%; }
.field-hint { font-size: var(--font-size-xs); color: var(--color-text-tertiary); margin-top: var(--spacing-xs); }
.tab-alert { margin-bottom: var(--spacing-lg); }

/* 映射 */
.mapping-toolbar { display: flex; justify-content: flex-end; margin: var(--spacing-md) 0; }
.mapping-arrow { font-size: var(--font-size-lg); color: var(--color-primary); font-weight: 700; }
.mapping-footer { margin-top: var(--spacing-lg); display: flex; justify-content: flex-end; }

/* 事件标识 */
.event-code { background: var(--color-fill-secondary); padding: 2px 6px; border-radius: var(--radius-sm); font-size: var(--font-size-xs); }

/* 日志 */
.log-expand { padding: var(--spacing-md) var(--spacing-xl); }
.log-detail-box {
  background: var(--color-error-bg);
  border: 1px solid var(--color-error-border);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  font-size: var(--font-size-xs);
  line-height: 1.8;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 500px;
  overflow-y: auto;
  color: var(--color-error);
}
.log-pagination { margin-top: var(--spacing-md); display: flex; justify-content: flex-end; }
</style>
