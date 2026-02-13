<template>
  <div class="api-logs-page">
    <!-- 统计概览 -->
    <div class="stat-row">
      <div class="stat-card">
        <span class="stat-value">{{ stats.totalToday || 0 }}</span>
        <span class="stat-label">今日调用</span>
      </div>
      <div class="stat-card stat-success">
        <span class="stat-value">{{ stats.successRate || '0' }}%</span>
        <span class="stat-label">成功率</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ stats.avgDuration || 0 }}ms</span>
        <span class="stat-label">平均耗时</span>
      </div>
      <div class="stat-card stat-error">
        <span class="stat-value">{{ stats.errorCount || 0 }}</span>
        <span class="stat-label">今日异常</span>
      </div>
    </div>

    <!-- 日志列表 -->
    <el-card>
      <template #header>
        <div class="card-header-row">
          <span class="card-title">API 调用日志</span>
          <div class="header-actions">
          </div>
        </div>
      </template>

      <!-- 筛选区 -->
      <div class="filter-bar">
        <el-select v-model="filters.authType" placeholder="认证类型" clearable style="width:130px" @change="loadLogs">
          <el-option label="API Key" value="apikey" />
          <el-option label="JWT" value="jwt" />
        </el-select>
        <el-select v-model="filters.method" placeholder="请求方法" clearable style="width:110px" @change="loadLogs">
          <el-option label="GET" value="GET" />
          <el-option label="POST" value="POST" />
          <el-option label="PUT" value="PUT" />
          <el-option label="DELETE" value="DELETE" />
        </el-select>
        <el-input v-model="filters.path" placeholder="请求路径" clearable style="width:200px" @clear="loadLogs" @keyup.enter="loadLogs" />
        <el-input v-model="filters.ip" placeholder="来源IP" clearable style="width:150px" @clear="loadLogs" @keyup.enter="loadLogs" />
        <el-select v-model="filters.statusGroup" placeholder="状态" clearable style="width:110px" @change="loadLogs">
          <el-option label="成功(2xx)" value="2xx" />
          <el-option label="客户端错误(4xx)" value="4xx" />
          <el-option label="服务端错误(5xx)" value="5xx" />
        </el-select>
        <el-date-picker v-model="filters.dateRange" type="daterange" range-separator="-" start-placeholder="开始" end-placeholder="结束" value-format="YYYY-MM-DD" style="width:240px" @change="loadLogs" />
        <el-button type="primary" @click="loadLogs">搜索</el-button>
      </div>

      <el-table :data="logs" v-loading="loading" stripe size="small" row-key="id">
        <el-table-column type="expand">
          <template #default="{ row }">
            <div class="log-expand">
              <div class="expand-section" v-if="row.query">
                <strong>Query: </strong><code>{{ row.query }}</code>
              </div>
              <div class="expand-section" v-if="row.requestBody">
                <strong>Request Body: </strong>
                <pre class="code-block">{{ row.requestBody }}</pre>
              </div>
              <div class="expand-section" v-if="row.errorMessage">
                <strong>Error: </strong><span class="text-error">{{ row.errorMessage }}</span>
              </div>
              <div class="expand-section">
                <strong>User-Agent: </strong><span class="text-muted">{{ row.userAgent || '-' }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="时间" width="155">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="认证" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.authType === 'apikey' ? 'warning' : 'primary'" size="small" effect="light">
              {{ row.authType === 'apikey' ? 'APIKey' : 'JWT' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="调用者" width="120">
          <template #default="{ row }">
            {{ row.appId || row.username || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="方法" width="70" align="center">
          <template #default="{ row }">
            <el-tag :type="methodTagType(row.method)" size="small" effect="plain">{{ row.method }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路径" min-width="200" show-overflow-tooltip />
        <el-table-column label="状态" width="70" align="center">
          <template #default="{ row }">
            <el-tag :type="row.statusCode < 400 ? 'success' : (row.statusCode < 500 ? 'warning' : 'danger')" size="small" effect="light">
              {{ row.statusCode }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="耗时" width="80" align="center">
          <template #default="{ row }">
            <span :class="{ 'text-error': row.duration > 1000 }">{{ row.duration }}ms</span>
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="来源IP" width="130" />
        <el-table-column label="响应大小" width="90" align="center">
          <template #default="{ row }">{{ formatSize(row.responseSize) }}</template>
        </el-table-column>
      </el-table>

      <div class="pagination-bar">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="loadLogs"
          @current-change="loadLogs"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { Download } from "@element-plus/icons-vue";
import { logManagementApi } from "../../api";

const logs = ref<any[]>([]);
const loading = ref(false);
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const exporting = ref(false);
const stats = ref<any>({});

const filters = reactive({
  authType: '', method: '', path: '', ip: '',
  statusGroup: '', dateRange: null as string[] | null
});

const methodTagType = (m: string) => {
  const map: Record<string, string> = { GET: '', POST: 'success', PUT: 'warning', DELETE: 'danger' };
  return map[m] || 'info';
};

const formatTime = (t: string) => t ? new Date(t).toLocaleString('zh-CN') : '-';
const formatSize = (bytes: number) => {
  if (!bytes) return '-';
  if (bytes < 1024) return bytes + 'B';
  return (bytes / 1024).toFixed(1) + 'KB';
};

const loadLogs = async () => {
  loading.value = true;
  try {
    const params: any = { page: page.value, size: pageSize.value };
    if (filters.authType) params.authType = filters.authType;
    if (filters.method) params.method = filters.method;
    if (filters.path) params.path = filters.path;
    if (filters.ip) params.ip = filters.ip;
    if (filters.statusGroup) params.statusGroup = filters.statusGroup;
    if (filters.dateRange?.length === 2) {
      params.startDate = filters.dateRange[0];
      params.endDate = filters.dateRange[1];
    }
    const res = await logManagementApi.apiLogs(params);
    const data = (res as any).data?.data;
    logs.value = data?.list || [];
    total.value = data?.total || 0;
  } finally { loading.value = false; }
};

const loadStats = async () => {
  try {
    const res = await logManagementApi.apiLogStats();
    stats.value = (res as any).data?.data || {};
  } catch {}
};

const exportLogs = async () => {
  exporting.value = true;
  try {
    const res = await logManagementApi.exportApiLogs(filters);
    const blob = new Blob([res.data], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `api-logs-${new Date().toISOString().slice(0, 10)}.csv`;
    a.click();
    URL.revokeObjectURL(url);
    ElMessage.success('导出成功');
  } catch {
    ElMessage.error('导出失败');
  } finally { exporting.value = false; }
};

onMounted(() => { loadLogs(); loadStats(); });
</script>

<style scoped>
.api-logs-page { display: flex; flex-direction: column; gap: var(--spacing-lg); }
.stat-row { display: flex; gap: var(--spacing-lg); }
.stat-card {
  flex: 1; background: var(--color-bg-container); border: 1px solid var(--color-border-secondary);
  border-radius: var(--radius-xl); padding: var(--spacing-xl);
  display: flex; flex-direction: column; align-items: center; gap: var(--spacing-xs);
  transition: box-shadow 0.2s;
}
.stat-card:hover { box-shadow: var(--shadow-md); }
.stat-value { font-size: var(--font-size-2xl); font-weight: 700; color: var(--color-text-primary); }
.stat-label { font-size: var(--font-size-sm); color: var(--color-text-tertiary); }
.stat-success .stat-value { color: var(--color-success); }
.stat-error .stat-value { color: var(--color-error); }

.card-header-row { display: flex; justify-content: space-between; align-items: center; }
.card-title { font-size: var(--font-size-lg); font-weight: 600; color: var(--color-text-primary); }

.filter-bar { display: flex; flex-wrap: wrap; gap: var(--spacing-sm); margin-bottom: var(--spacing-lg); }

.pagination-bar { margin-top: var(--spacing-md); display: flex; justify-content: flex-end; }

.log-expand { padding: var(--spacing-md) var(--spacing-xl); }
.expand-section { margin-bottom: var(--spacing-sm); font-size: var(--font-size-sm); }
.code-block {
  background: var(--color-fill-secondary); padding: var(--spacing-sm);
  border-radius: var(--radius-sm); font-size: var(--font-size-xs);
  max-height: 200px; overflow: auto; white-space: pre-wrap; word-break: break-all;
}
.text-muted { color: var(--color-text-tertiary); }
.text-error { color: var(--color-error); }
</style>
