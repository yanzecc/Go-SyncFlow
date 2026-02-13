<template>
  <div class="page-container">
    <el-card class="table-card">
      <template #header>
        <div class="filter-row">
          <el-select v-model="filterType" placeholder="日志类型" clearable style="width: 130px" @change="loadLogs">
            <el-option label="全部" value="" />
            <el-option label="登录日志" value="login" />
            <el-option label="操作日志" value="operation" />
          </el-select>
          <el-input v-model="keyword" placeholder="用户名" clearable style="width: 160px" @keyup.enter="loadLogs" />
          <el-date-picker v-model="dateRange" type="daterange" start-placeholder="开始" end-placeholder="结束"
            value-format="YYYY-MM-DD" style="width: 260px" />
          <el-button type="primary" @click="loadLogs">搜索</el-button>
        </div>
      </template>

      <el-table :data="logs" v-loading="loading" stripe size="small">
        <el-table-column label="时间" width="170">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column prop="username" label="用户" width="140" />
        <el-table-column label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.logType === 'login' ? 'primary' : 'warning'" size="small">
              {{ row.logType === 'login' ? '登录' : '操作' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="描述" min-width="300">
          <template #default="{ row }">
            <span>{{ row.summary }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
              {{ row.status === 'success' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="IP" width="140" />
      </el-table>

      <div class="pagination-row">
        <span class="total-text">共 {{ total }} 条</span>
        <el-pagination background layout="sizes, prev, pager, next" :total="total"
          v-model:current-page="page" v-model:page-size="pageSize"
          :page-sizes="[20, 50, 100]" @current-change="loadLogs" @size-change="loadLogs" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { api } from "../../api";

const loading = ref(false);
const logs = ref<any[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const filterType = ref("");
const keyword = ref("");
const dateRange = ref<string[]>([]);

const formatTime = (t: string) => t ? t.replace("T", " ").slice(0, 19) : "";

const loadLogs = async () => {
  loading.value = true;
  try {
    const params: any = { page: page.value, size: pageSize.value };
    if (filterType.value) params.type = filterType.value;
    if (keyword.value) params.keyword = keyword.value;
    if (dateRange.value?.length === 2) {
      params.startDate = dateRange.value[0];
      params.endDate = dateRange.value[1];
    }
    const res = await api.get("/logs/system", { params });
    if (res.data.success) {
      logs.value = res.data.data?.list || [];
      total.value = res.data.data?.total || 0;
    }
  } finally { loading.value = false; }
};

onMounted(loadLogs);
</script>

<style scoped>
.page-container { padding: 0; }
.filter-row { display: flex; gap: 10px; align-items: center; flex-wrap: wrap; }
.pagination-row { display: flex; justify-content: space-between; align-items: center; margin-top: 16px; }
.total-text { font-size: 13px; color: #909399; }
</style>
