<template>
  <div class="page-container">
    <el-card class="table-card">
      <template #header>
        <div class="filter-row">
          <el-select v-model="filters.direction" placeholder="方向" clearable style="width: 120px" size="small" @change="loadLogs">
            <el-option label="上游同步" value="upstream" />
            <el-option label="下游同步" value="downstream" />
          </el-select>
          <el-select v-model="filters.event" placeholder="事件" clearable style="width: 120px" size="small" @change="loadLogs">
            <el-option label="全量同步" value="full_sync" />
            <el-option label="用户创建" value="user_create" />
            <el-option label="用户更新" value="user_update" />
            <el-option label="用户删除" value="user_delete" />
            <el-option label="密码修改" value="password_change" />
            <el-option label="状态变更" value="user_status_change" />
          </el-select>
          <el-select v-model="filters.status" placeholder="状态" clearable style="width: 100px" size="small" @change="loadLogs">
            <el-option label="成功" value="success" />
            <el-option label="部分成功" value="partial" />
            <el-option label="失败" value="failed" />
          </el-select>
          <el-date-picker v-model="filters.dateRange" type="daterange" range-separator="-"
            start-placeholder="开始" end-placeholder="结束" value-format="YYYY-MM-DD"
            style="width: 220px" size="small" clearable @change="loadLogs" />
          <el-button type="primary" size="small" @click="loadLogs">搜索</el-button>
        </div>
      </template>

      <el-table :data="logs" v-loading="loading" border size="small" row-key="id" style="width: 100%;">
        <el-table-column type="expand">
          <template #default="{ row }">
            <div class="expand-content">
              <template v-if="row.detail">
                <div class="detail-raw">{{ row.detail }}</div>
              </template>
              <div v-else class="detail-empty">暂无详细信息</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="时间" width="155">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="方向" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.direction === 'upstream' ? 'primary' : 'success'" size="small">
              {{ row.direction === 'upstream' ? '上游同步' : '下游同步' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="triggerType" label="触发" width="70" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="row.triggerType === 'event' ? 'warning' : (row.triggerType === 'schedule' ? 'primary' : 'info')">
              {{ triggerTypeMap[row.triggerType] || row.triggerType }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="triggerEvent" label="事件" width="100">
          <template #default="{ row }">
            {{ eventMap[row.triggerEvent] || row.triggerEvent || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : (row.status === 'partial' ? 'warning' : 'danger')" size="small">
              {{ statusMap[row.status] || row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="概要" min-width="250" show-overflow-tooltip />
      </el-table>

      <div class="pagination-row">
        <span class="total-text">共 {{ pagination.total }} 条</span>
        <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.size"
          :total="pagination.total" :page-sizes="[20, 50, 100]"
          layout="sizes, prev, pager, next" @current-change="loadLogs" @size-change="loadLogs" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { logApi } from "../../api";

const loading = ref(false);
const logs = ref<any[]>([]);
const filters = reactive({ direction: "", event: "", status: "", dateRange: null as any });
const pagination = reactive({ page: 1, size: 20, total: 0 });

const triggerTypeMap: Record<string, string> = { event: '事件', schedule: '定时', manual: '手动' };
const eventMap: Record<string, string> = {
  password_change: '密码修改', full_sync: '全量同步', user_create: '用户创建',
  user_update: '用户更新', user_delete: '用户删除', user_status_change: '状态变更',
  role_change: '角色变更', dingtalk_sync: '钉钉同步'
};
const statusMap: Record<string, string> = { success: '成功', partial: '部分成功', failed: '失败' };

const formatTime = (time: string) => {
  if (!time) return '-';
  return new Date(time).toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' });
};

const loadLogs = async () => {
  loading.value = true;
  try {
    const params: any = {
      page: pagination.page, size: pagination.size,
      event: filters.event || undefined,
      status: filters.status || undefined,
      direction: filters.direction || undefined
    };
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.startDate = filters.dateRange[0];
      params.endDate = filters.dateRange[1];
    }
    const res = await logApi.syncLogs(params);
    const data = (res as any).data?.data;
    logs.value = data?.list || [];
    pagination.total = data?.total || 0;
  } finally { loading.value = false; }
};

onMounted(loadLogs);
</script>

<style scoped>
.page-container { display: flex; flex-direction: column; gap: 16px; }
.filter-row { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.table-card :deep(.el-card__header) { padding: 12px 16px; }
.table-card :deep(.el-card__body) { padding: 16px; }
.pagination-row { display: flex; justify-content: space-between; align-items: center; margin-top: 16px; }
.total-text { font-size: 14px; color: var(--color-text-tertiary); }
.expand-content { padding: 12px 20px; }
.detail-raw { background: var(--color-fill-secondary); border: 1px solid var(--color-border); border-radius: 4px; padding: 12px; font-size: 12px; line-height: 1.8; white-space: pre-wrap; word-break: break-all; max-height: 500px; overflow-y: auto; }
.detail-empty { color: var(--color-text-tertiary); font-size: 12px; }
</style>
