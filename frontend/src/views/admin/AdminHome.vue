<template>
  <div class="admin-home">
    <!-- 欢迎区域 -->
    <div class="welcome-section">
      <div class="welcome-text">
        <h1>欢迎回来，{{ userStore.user?.nickname || userStore.user?.username }}</h1>
        <p>{{ greeting }}，祝您工作愉快！</p>
      </div>
      <div class="quick-stats">
        <div class="stat-item">
          <span class="stat-value">{{ systemStatus.runtime?.uptime || '-' }}</span>
          <span class="stat-label">运行时长</span>
        </div>
        <div class="stat-item">
          <span class="stat-value">{{ formatTime(new Date()) }}</span>
          <span class="stat-label">当前时间</span>
        </div>
      </div>
    </div>

    <!-- 系统状态卡片 -->
    <div class="status-cards">
      <div class="status-card" :class="getCpuClass(systemStatus.cpu?.usage || 0)">
        <div class="card-header">
          <el-icon class="card-icon"><Cpu /></el-icon>
          <span class="card-title">CPU使用率</span>
        </div>
        <div class="card-body">
          <div class="progress-ring">
            <svg viewBox="0 0 100 100">
              <circle class="bg" cx="50" cy="50" r="42" />
              <circle 
                class="progress" 
                cx="50" cy="50" r="42" 
                :stroke-dasharray="`${(systemStatus.cpu?.usage || 0) * 2.64} 264`"
              />
            </svg>
            <span class="ring-value">{{ (systemStatus.cpu?.usage || 0).toFixed(1) }}%</span>
          </div>
          <div class="card-detail">
            <span>核心数: {{ systemStatus.cpu?.cores || 0 }}</span>
          </div>
        </div>
      </div>

      <div class="status-card" :class="getMemoryClass(systemStatus.memory?.percent || 0)">
        <div class="card-header">
          <el-icon class="card-icon"><Coin /></el-icon>
          <span class="card-title">内存使用</span>
        </div>
        <div class="card-body">
          <div class="progress-ring">
            <svg viewBox="0 0 100 100">
              <circle class="bg" cx="50" cy="50" r="42" />
              <circle 
                class="progress" 
                cx="50" cy="50" r="42" 
                :stroke-dasharray="`${(systemStatus.memory?.percent || 0) * 2.64} 264`"
              />
            </svg>
            <span class="ring-value">{{ (systemStatus.memory?.percent || 0).toFixed(1) }}%</span>
          </div>
          <div class="card-detail">
            <span>{{ (systemStatus.memory?.used || 0).toFixed(2) }}GB / {{ (systemStatus.memory?.total || 0).toFixed(2) }}GB</span>
          </div>
        </div>
      </div>

      <div class="status-card" :class="getDiskClass(systemStatus.disk?.percent || 0)">
        <div class="card-header">
          <el-icon class="card-icon"><FolderOpened /></el-icon>
          <span class="card-title">磁盘使用</span>
        </div>
        <div class="card-body">
          <div class="progress-ring">
            <svg viewBox="0 0 100 100">
              <circle class="bg" cx="50" cy="50" r="42" />
              <circle 
                class="progress" 
                cx="50" cy="50" r="42" 
                :stroke-dasharray="`${(systemStatus.disk?.percent || 0) * 2.64} 264`"
              />
            </svg>
            <span class="ring-value">{{ (systemStatus.disk?.percent || 0).toFixed(1) }}%</span>
          </div>
          <div class="card-detail">
            <span>{{ (systemStatus.disk?.used || 0).toFixed(1) }}GB / {{ (systemStatus.disk?.total || 0).toFixed(1) }}GB</span>
          </div>
        </div>
      </div>

      <div class="status-card success">
        <div class="card-header">
          <el-icon class="card-icon"><Connection /></el-icon>
          <span class="card-title">网络流量</span>
        </div>
        <div class="card-body">
          <div class="network-stats">
            <div class="net-item">
              <el-icon class="up"><Top /></el-icon>
              <span class="net-value">{{ (systemStatus.network?.outRate || 0).toFixed(2) }} MB/s</span>
            </div>
            <div class="net-item">
              <el-icon class="down"><Bottom /></el-icon>
              <span class="net-value">{{ (systemStatus.network?.inRate || 0).toFixed(2) }} MB/s</span>
            </div>
          </div>
          <div class="card-detail">
            <span>实时网络速率</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 服务状态 & 快捷操作 -->
    <div class="section-row">
      <div class="section-card services-card">
        <div class="section-header">
          <h3><el-icon><SetUp /></el-icon> 服务状态</h3>
          <el-button type="primary" link @click="loadSystemStatus">
            <el-icon><Refresh /></el-icon> 刷新
          </el-button>
        </div>
        <div class="services-list">
          <div class="service-item" :class="{ online: systemStatus.database?.status === 'ok' }">
            <div class="service-icon">
              <el-icon><DataLine /></el-icon>
            </div>
            <div class="service-info">
              <span class="service-name">数据库服务</span>
              <span class="service-status">{{ systemStatus.database?.type || 'SQLite' }} 运行正常</span>
            </div>
            <el-tag :type="systemStatus.database?.status === 'ok' ? 'success' : 'danger'" size="small">
              {{ systemStatus.database?.status === 'ok' ? '在线' : '离线' }}
            </el-tag>
          </div>
          <div class="service-item online">
            <div class="service-icon">
              <el-icon><Monitor /></el-icon>
            </div>
            <div class="service-info">
              <span class="service-name">应用服务</span>
              <span class="service-status">Go {{ systemStatus.runtime?.goVersion || '' }} | {{ systemStatus.runtime?.goroutines || 0 }} goroutines</span>
            </div>
            <el-tag type="success" size="small">在线</el-tag>
          </div>
        </div>
      </div>

      <div class="section-card quick-actions-card">
        <div class="section-header">
          <h3><el-icon><Grid /></el-icon> 快捷操作</h3>
        </div>
        <div class="quick-actions">
          <div class="action-item" @click="router.push('/admin/users/local')">
            <el-icon><User /></el-icon>
            <span>本地用户</span>
          </div>
          <div class="action-item" @click="router.push('/admin/users/dingtalk')">
            <el-icon><UserFilled /></el-icon>
            <span>钉钉用户</span>
          </div>
          <div class="action-item" @click="router.push('/admin/roles')">
            <el-icon><Stamp /></el-icon>
            <span>角色管理</span>
          </div>
          <div class="action-item" @click="router.push('/admin/settings')">
            <el-icon><Setting /></el-icon>
            <span>系统设置</span>
          </div>
          <div class="action-item" @click="router.push('/admin/security')">
            <el-icon><Lock /></el-icon>
            <span>安全中心</span>
          </div>
          <div class="action-item" @click="loadSystemStatus">
            <el-icon><Refresh /></el-icon>
            <span>刷新状态</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 系统信息 -->
    <div class="section-card system-info-card">
      <div class="section-header">
        <h3><el-icon><InfoFilled /></el-icon> 系统信息</h3>
      </div>
      <div class="info-grid">
        <div class="info-item">
          <span class="info-label">Go版本</span>
          <span class="info-value">{{ systemStatus.runtime?.goVersion || '-' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">Goroutines</span>
          <span class="info-value">{{ systemStatus.runtime?.goroutines || 0 }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">启动时间</span>
          <span class="info-value">{{ systemStatus.runtime?.startTime || '-' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">数据库类型</span>
          <span class="info-value">{{ systemStatus.database?.type || 'SQLite' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">运行时长</span>
          <span class="info-value">{{ systemStatus.runtime?.uptime || '-' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">最后刷新</span>
          <span class="info-value">{{ lastRefresh }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { 
  Cpu, Coin, FolderOpened, Connection, Top, Bottom, SetUp, Refresh,
  DataLine, Lock, Grid, User, UserFilled, Setting, Monitor, Stamp,
  InfoFilled
} from "@element-plus/icons-vue";
import { useUserStore } from "../../store/user";
import { systemApi } from "../../api";

const router = useRouter();
const userStore = useUserStore();

const systemStatus = ref<any>({
  cpu: { usage: 0, cores: 0 },
  memory: { used: 0, total: 0, percent: 0 },
  disk: { used: 0, total: 0, percent: 0 },
  network: { inRate: 0, outRate: 0 },
  database: { status: 'ok', type: 'sqlite' },
  runtime: { uptime: '-', goVersion: '', goroutines: 0, startTime: '' }
});

const lastRefresh = ref('-');
let refreshTimer: number | null = null;

const greeting = computed(() => {
  const hour = new Date().getHours();
  if (hour < 6) return '夜深了';
  if (hour < 9) return '早上好';
  if (hour < 12) return '上午好';
  if (hour < 14) return '中午好';
  if (hour < 18) return '下午好';
  if (hour < 22) return '晚上好';
  return '夜深了';
});

const formatTime = (date: Date) => {
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
};

const getCpuClass = (value: number) => {
  if (value >= 90) return 'danger';
  if (value >= 70) return 'warning';
  return 'success';
};

const getMemoryClass = (value: number) => {
  if (value >= 90) return 'danger';
  if (value >= 70) return 'warning';
  return 'success';
};

const getDiskClass = (value: number) => {
  if (value >= 90) return 'danger';
  if (value >= 80) return 'warning';
  return 'success';
};

const loadSystemStatus = async () => {
  try {
    const res = await systemApi.status();
    if (res.data.success) {
      systemStatus.value = res.data.data;
      lastRefresh.value = new Date().toLocaleTimeString('zh-CN');
    }
  } catch (e) {
    // ignore
  }
};

onMounted(() => {
  loadSystemStatus();
  refreshTimer = window.setInterval(loadSystemStatus, 10000);
});

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer);
});
</script>

<style scoped>
.admin-home {
  max-width: 1400px;
  margin: 0 auto;
}

.welcome-section {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: var(--radius-xl);
  padding: 28px 32px;
  color: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-2xl);
}

.welcome-text h1 {
  font-size: 24px;
  font-weight: 600;
  margin: 0 0 8px 0;
}

.welcome-text p {
  margin: 0;
  opacity: 0.9;
  font-size: 15px;
}

.quick-stats {
  display: flex;
  gap: 32px;
}

.stat-item {
  text-align: center;
}

.stat-value {
  display: block;
  font-size: 24px;
  font-weight: 600;
}

.stat-label {
  font-size: 13px;
  opacity: 0.8;
}

.status-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.status-card {
  background: var(--color-bg-container);
  border-radius: var(--radius-xl);
  padding: 20px;
  box-shadow: var(--shadow-sm);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
}

.card-icon {
  font-size: 22px;
  color: var(--color-primary);
}

.status-card.warning .card-icon { color: var(--color-warning); }
.status-card.danger .card-icon { color: var(--color-error); }
.status-card.success .card-icon { color: var(--color-success); }

.card-title {
  font-size: 14px;
  color: var(--color-text-tertiary);
}

.card-body {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.progress-ring {
  width: 100px;
  height: 100px;
  position: relative;
}

.progress-ring svg {
  transform: rotate(-90deg);
}

.progress-ring circle {
  fill: none;
  stroke-width: 8;
  stroke-linecap: round;
}

.progress-ring .bg {
  stroke: var(--color-border-secondary);
}

.progress-ring .progress {
  stroke: var(--color-primary);
  transition: stroke-dasharray 0.5s ease;
}

.status-card.warning .progress-ring .progress { stroke: var(--color-warning); }
.status-card.danger .progress-ring .progress { stroke: var(--color-error); }
.status-card.success .progress-ring .progress { stroke: var(--color-success); }

.ring-value {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.card-detail {
  margin-top: 12px;
  font-size: 13px;
  color: var(--color-text-tertiary);
}

.network-stats {
  display: flex;
  gap: 24px;
  margin: 16px 0;
}

.net-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.net-item .up { color: var(--color-success); }
.net-item .down { color: var(--color-primary); }

.net-value {
  font-size: 16px;
  font-weight: 500;
  color: var(--color-text-primary);
}

.section-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 24px;
}

.section-card {
  background: var(--color-bg-container);
  border-radius: var(--radius-xl);
  padding: 20px;
  box-shadow: var(--shadow-sm);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--color-text-primary);
}

.services-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.service-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: var(--radius-lg);
  background: var(--color-fill-secondary);
  border: 1px solid var(--color-border-secondary);
}

.service-item.online {
  border-color: var(--color-success-border);
  background: var(--color-success-bg);
}

.service-icon {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-lg);
  background: var(--color-primary-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: var(--color-primary);
}

.service-item.online .service-icon {
  background: #d9f7be;
  color: var(--color-success);
}

.service-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.service-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text-primary);
}

.service-status {
  font-size: 12px;
  color: var(--color-text-tertiary);
}

.quick-actions {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.action-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px;
  border-radius: var(--radius-lg);
  background: var(--color-fill-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.action-item:hover {
  background: var(--color-primary-bg);
  color: var(--color-primary);
}

.action-item .el-icon {
  font-size: 24px;
}

.action-item span {
  font-size: 13px;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px;
  background: var(--color-fill-secondary);
  border-radius: var(--radius-lg);
}

.info-label {
  font-size: 12px;
  color: var(--color-text-tertiary);
}

.info-value {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text-primary);
}

@media (max-width: 1200px) {
  .status-cards {
    grid-template-columns: repeat(2, 1fr);
  }
  .section-row {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .welcome-section {
    flex-direction: column;
    gap: 20px;
    text-align: center;
  }
  .status-cards {
    grid-template-columns: 1fr;
  }
  .quick-actions {
    grid-template-columns: repeat(2, 1fr);
  }
  .info-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
