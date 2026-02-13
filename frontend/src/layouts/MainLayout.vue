<template>
  <div class="admin-layout" :class="{ 'sidebar-collapsed': isCollapsed, 'no-sidebar': !hasAdminMenus }">
    <!-- 侧边栏（仅管理员可见） -->
    <aside class="sidebar" v-if="hasAdminMenus" :style="{ width: isCollapsed ? '64px' : '220px' }">
      <!-- Logo区域 -->
      <div class="sidebar-header">
        <div class="logo">
          <span class="logo-icon">🛡️</span>
          <span class="logo-text" v-show="!isCollapsed">{{ uiConfig.browserTitle || '用户管理' }}</span>
        </div>
      </div>

      <!-- 菜单搜索 -->
      <div class="menu-search" v-show="!isCollapsed">
        <el-input
          v-model="menuSearch"
          placeholder="搜索菜单..."
          :prefix-icon="Search"
          size="small"
          clearable
        />
      </div>

      <!-- 菜单列表 -->
      <nav class="menu-nav">
        <template v-for="menu in filteredMenus" :key="menu.id">
          <!-- 一级菜单 -->
          <div
            class="menu-item level-1"
            :class="{ 
              active: isMenuActive(menu), 
              expanded: isExpanded(menu.id),
              'has-children': menu.children?.length 
            }"
            @click="handleMenuClick(menu)"
          >
            <el-icon class="menu-icon"><component :is="menu.icon" /></el-icon>
            <span class="menu-title" v-show="!isCollapsed">{{ menu.title }}</span>
            <el-icon class="expand-icon" v-if="menu.children?.length && !isCollapsed">
              <ArrowDown v-if="isExpanded(menu.id)" />
              <ArrowRight v-else />
            </el-icon>
          </div>

          <!-- 二级菜单 -->
          <div 
            class="submenu" 
            v-if="menu.children?.length && !isCollapsed"
            v-show="isExpanded(menu.id)"
          >
            <div
              v-for="child in menu.children"
              :key="child.id"
              class="menu-item level-2"
              :class="{ active: route.path === child.path }"
              @click.stop="navigateTo(child.path)"
            >
              <span class="menu-title">{{ child.title }}</span>
            </div>
          </div>
        </template>
      </nav>

      <!-- 底部折叠按钮 -->
      <div class="sidebar-footer">
        <div class="collapse-btn" @click="toggleCollapse">
          <el-icon><Fold v-if="!isCollapsed" /><Expand v-else /></el-icon>
          <span v-show="!isCollapsed">收起菜单</span>
        </div>
      </div>
    </aside>

    <!-- 主内容区 -->
    <div class="main-container">
      <!-- 顶部导航（仅管理员显示） -->
      <header class="topbar" v-if="hasAdminMenus">
        <div class="topbar-left">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/admin' }">管理后台</el-breadcrumb-item>
            <el-breadcrumb-item v-for="crumb in breadcrumbs" :key="crumb.path">
              {{ crumb.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <div class="topbar-right">
          <!-- 全局搜索 -->
          <div class="global-search">
            <el-input
              v-model="globalSearch"
              placeholder="搜索..."
              :prefix-icon="Search"
              size="default"
              style="width: 200px"
              clearable
            />
          </div>

          <!-- 通知中心 -->
          <el-badge :value="notificationCount" :hidden="notificationCount === 0" class="notification-badge">
            <el-button :icon="Bell" circle />
          </el-badge>

          <!-- 用户菜单 -->
          <el-dropdown @command="handleCommand" trigger="click">
            <div class="user-dropdown">
              <el-avatar :size="36" class="user-avatar">
                {{ userStore.user?.nickname?.charAt(0) || 'U' }}
              </el-avatar>
              <div class="user-info" v-if="!isCollapsed">
                <span class="user-name">{{ userStore.user?.nickname || userStore.user?.username }}</span>
              </div>
              <el-icon><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <el-icon><User /></el-icon> 个人中心
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon> 退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <!-- 内容区域 -->
      <main class="content-area">
        <router-view v-slot="{ Component }">
          <keep-alive :max="10">
            <component :is="Component" :key="$route.path" />
          </keep-alive>
        </router-view>
      </main>
    </div>

    <!-- 修改密码对话框 -->
    <el-dialog v-model="passwordDialogVisible" title="修改密码" width="420px" :close-on-click-modal="false">
      <el-form :model="passwordForm" label-width="80px">
        <el-form-item label="原密码">
          <el-input v-model="passwordForm.oldPassword" type="password" show-password placeholder="请输入原密码" />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="passwordForm.newPassword" type="password" show-password placeholder="请输入新密码" />
        </el-form-item>
        <el-form-item label="确认密码">
          <el-input v-model="passwordForm.confirmPassword" type="password" show-password placeholder="请再次输入新密码" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="changePassword">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { 
  Search, ArrowDown, ArrowRight, Fold, Expand, Bell, Monitor, 
  User, Key, SwitchButton, HomeFilled, UserFilled, Setting, 
  Lock, Tickets, List, DataAnalysis, Connection, Refresh, Sort
} from "@element-plus/icons-vue";
import { useUserStore } from "../store/user";
import { authApi, settingsApi } from "../api";

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();

// 状态
const isCollapsed = ref(localStorage.getItem('sidebarCollapsed') === 'true');
const menuSearch = ref('');
const globalSearch = ref('');
const notificationCount = ref(0);
const uiConfig = ref({ browserTitle: '', loginTitle: '' });
const passwordDialogVisible = ref(false);
const passwordForm = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' });

// 菜单数据
const menuConfig = [
  {
    id: 'home',
    title: '系统首页',
    icon: HomeFilled,
    path: '/admin',
    level: 1,
    order: 0,
    permission: ['settings:system']
  },
  {
    id: 'user-source',
    title: '用户管理',
    icon: UserFilled,
    level: 1,
    order: 1,
    permission: ['user:list'],
    children: [
      { id: 'local-users', title: '用户', path: '/admin/users/local', level: 2, permission: 'user:list' }
    ]
  },
  {
    id: 'sync-mgmt',
    title: '同步管理',
    icon: Refresh,
    level: 1,
    order: 2,
    permission: ['settings:system'],
    children: [
      { id: 'upstream-sync', title: '上游同步', path: '/admin/sync/upstream', level: 2, permission: 'settings:system' },
      { id: 'downstream-sync', title: '下游同步', path: '/admin/sync/downstream', level: 2, permission: 'settings:system' }
    ]
  },
  {
    id: 'roles',
    title: '角色管理',
    icon: List,
    path: '/admin/roles',
    level: 1,
    order: 3,
    permission: ['role:list']
  },
  {
    id: 'logs',
    title: '日志管理',
    icon: Tickets,
    level: 1,
    order: 4,
    permission: ['log:login', 'log:operation', 'settings:system'],
    children: [
      { id: 'system-logs', title: '系统日志', path: '/admin/logs/system', level: 2, permission: 'log:login' },
      { id: 'sync-logs', title: '同步日志', path: '/admin/logs/sync', level: 2, permission: 'log:operation' },
      { id: 'api-logs', title: 'API调用日志', path: '/admin/logs/api', level: 2, permission: 'settings:system' },
      { id: 'log-settings', title: '日志设置', path: '/admin/logs/settings', level: 2, permission: 'settings:system' }
    ]
  },
  {
    id: 'notify',
    title: '通知管理',
    icon: Bell,
    level: 1,
    order: 5,
    permission: ['settings:system'],
    children: [
      { id: 'channels', title: '通知渠道', path: '/admin/notify/channels', level: 2, permission: 'settings:system' },
      { id: 'msg-templates', title: '消息模板', path: '/admin/notify/templates', level: 2, permission: 'settings:system' },
      { id: 'alert-rules', title: '规则与策略', path: '/admin/notify/rules', level: 2, permission: 'settings:system' }
    ]
  },
  {
    id: 'settings',
    title: '系统设置',
    icon: Setting,
    level: 1,
    order: 6,
    permission: ['settings:system'],
    children: [
      { id: 'settings-ui', title: '界面与证书', path: '/admin/settings', level: 2, permission: 'settings:system' },
      { id: 'settings-ldap', title: 'LDAP 服务', path: '/admin/settings/ldap', level: 2, permission: 'settings:system' },
      { id: 'settings-apikeys', title: 'API 密钥', path: '/admin/apikeys', level: 2, permission: 'settings:system' }
    ]
  },
  {
    id: 'security',
    title: '安全中心',
    icon: Lock,
    path: '/admin/security',
    level: 1,
    order: 7,
    permission: ['settings:system']
  }
];

// 展开状态独立管理
const expandedMenus = ref<Set<string>>(new Set(['user-source']));

// 恢复菜单展开状态
const restoreMenuState = () => {
  const saved = localStorage.getItem('menuExpandState');
  if (saved) {
    try {
      const arr = JSON.parse(saved);
      if (Array.isArray(arr)) {
        expandedMenus.value = new Set(arr);
      }
    } catch (e) {
      // ignore
    }
  }
};

// 保存菜单展开状态
const saveMenuState = () => {
  localStorage.setItem('menuExpandState', JSON.stringify([...expandedMenus.value]));
};

// 判断菜单是否展开
const isExpanded = (menuId: string) => expandedMenus.value.has(menuId);

// 切换菜单展开状态
const toggleExpand = (menuId: string) => {
  if (expandedMenus.value.has(menuId)) {
    expandedMenus.value.delete(menuId);
  } else {
    expandedMenus.value.add(menuId);
  }
  saveMenuState();
};

// 当前用户是否显示侧边栏
const hasAdminMenus = computed(() => {
  const mode = userStore.layoutConfig?.sidebarMode || 'auto';
  if (mode === 'hidden') return false;
  if (mode === 'visible') return true;
  // auto模式：有多个可见顶级菜单时显示，只有1个或更少时隐藏
  const visibleMenus = filteredMenus.value.filter(m => !!(m as any).permission);
  return visibleMenus.length > 1;
});

// 过滤后的菜单（只计算一次权限）
const filteredMenus = computed(() => {
  return menuConfig.map(menu => {
    // 权限检查
    if (menu.permission) {
      const perms = Array.isArray(menu.permission) ? menu.permission : [menu.permission];
      if (!perms.some(p => userStore.hasPermission(p))) return null;
    }
    
    // 过滤子菜单权限
    let children = menu.children;
    if (children) {
      children = children.filter(child => {
        if (child.permission && !userStore.hasPermission(child.permission)) return false;
        return true;
      });
    }
    
    return { ...menu, children };
  }).filter(Boolean) as typeof menuConfig;
});

// 面包屑
const breadcrumbs = computed(() => {
  const path = route.path;
  const crumbs: { path: string; title: string }[] = [];
  
  for (const menu of menuConfig) {
    if (menu.path === path) {
      crumbs.push({ path: menu.path, title: menu.title });
      break;
    }
    if (menu.children) {
      const child = menu.children.find(c => c.path === path);
      if (child) {
        crumbs.push({ path: '', title: menu.title });
        crumbs.push({ path: child.path || '', title: child.title });
        break;
      }
    }
  }
  
  return crumbs;
});

// 判断菜单是否激活
const isMenuActive = (menu: any) => {
  if (menu.path === route.path) return true;
  if (menu.children?.some((c: any) => c.path === route.path)) return true;
  return false;
};

// 菜单点击处理
const handleMenuClick = (menu: any) => {
  if (menu.children?.length) {
    toggleExpand(menu.id);
  } else if (menu.path) {
    router.push(menu.path);
  }
};

// 导航到指定路径
const navigateTo = (path: string) => {
  if (path) router.push(path);
};

// 折叠/展开侧边栏
const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value;
  localStorage.setItem('sidebarCollapsed', String(isCollapsed.value));
};

// 键盘导航
const handleKeydown = (e: KeyboardEvent) => {
  // 简单的键盘导航支持
  if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
    e.preventDefault();
  }
};

// 用户命令处理
const handleCommand = async (command: string) => {
  if (command === 'logout') {
    await userStore.logout();
    router.push('/login');
  } else if (command === 'password') {
    router.push('/admin/profile');
  } else if (command === 'profile') {
    router.push('/admin/profile');
  }
};

// 修改密码
const changePassword = async () => {
  if (!passwordForm.oldPassword || !passwordForm.newPassword) {
    ElMessage.warning('请填写完整');
    return;
  }
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    ElMessage.warning('两次密码不一致');
    return;
  }
  try {
    const res = await authApi.changePassword({
      oldPassword: passwordForm.oldPassword,
      newPassword: passwordForm.newPassword
    });
    if (res.data.success) {
      ElMessage.success('密码修改成功');
      passwordDialogVisible.value = false;
    }
  } catch (e) {
    // handled by interceptor
  }
};

// 监听路由变化，自动展开对应菜单
watch(() => route.path, (newPath) => {
  menuConfig.forEach(menu => {
    if (menu.children?.some(c => c.path === newPath)) {
      expandedMenus.value.add(menu.id);
    }
  });
}, { immediate: true });

onMounted(async () => {
  restoreMenuState();
  
  try {
    const res = await settingsApi.getUI();
    if (res.data.success) {
      uiConfig.value = res.data.data;
      if (uiConfig.value.browserTitle) {
        document.title = uiConfig.value.browserTitle;
      }
    }
  } catch (e) {
    // ignore
  }
});
</script>

<style scoped>
.admin-layout {
  display: flex;
  min-height: 100vh;
  background: var(--color-bg-layout);
}

/* 侧边栏 */
.sidebar {
  background: var(--color-bg-container);
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.04);
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  z-index: 100;
  overflow: hidden;
}

.sidebar-header {
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border-secondary);
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
  color: var(--color-primary);
}

.logo-icon { font-size: 24px; }
.logo-text { font-size: var(--font-size-lg); white-space: nowrap; }

.menu-search {
  padding: var(--spacing-md) var(--spacing-lg);
}

/* 菜单导航 */
.menu-nav {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-sm) 0;
  outline: none;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 11px var(--spacing-lg);
  cursor: pointer;
  transition: all 0.2s;
  color: var(--color-text-secondary);
  position: relative;
}

.menu-item:hover {
  background: var(--color-bg-layout);
}

.menu-item.active {
  color: var(--color-primary);
  background: var(--color-primary-bg);
}

.menu-item.level-1 {
  font-size: var(--font-size-base);
  font-weight: 500;
}

.menu-item.level-2 {
  font-size: var(--font-size-base);
  padding-left: 52px;
  color: var(--color-text-tertiary);
}

.menu-item.level-2.active {
  color: var(--color-primary);
  background: var(--color-primary-bg);
}

.menu-item.level-2.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: var(--color-primary);
  border-radius: 0 2px 2px 0;
}

.menu-icon {
  font-size: 18px;
  margin-right: var(--spacing-md);
  flex-shrink: 0;
}

.menu-title {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.expand-icon {
  font-size: 12px;
  transition: transform 0.2s;
}

.submenu {
  overflow: hidden;
  transition: all 0.3s ease;
}

/* 底部折叠按钮 */
.sidebar-footer {
  padding: var(--spacing-md);
  border-top: 1px solid var(--color-border-secondary);
}

.collapse-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  padding: 10px;
  cursor: pointer;
  color: var(--color-text-tertiary);
  border-radius: var(--radius-md);
  transition: all 0.2s;
}

.collapse-btn:hover {
  background: var(--color-bg-layout);
  color: var(--color-primary);
}

/* 主内容区 */
.main-container {
  flex: 1;
  margin-left: 220px;
  transition: margin-left 0.3s ease;
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  max-width: calc(100vw - 220px);
  overflow-x: hidden;
}

.sidebar-collapsed .main-container { margin-left: 64px; max-width: calc(100vw - 64px); }
.no-sidebar .main-container { margin-left: 0; max-width: 100vw; }

/* 顶部导航 */
.topbar {
  height: 56px;
  background: var(--color-bg-container);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--spacing-2xl);
  position: sticky;
  top: 0;
  z-index: 50;
}

.topbar-left { display: flex; align-items: center; }

.topbar-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
}

.notification-badge { cursor: pointer; }

.user-dropdown {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  padding: 6px 10px;
  border-radius: var(--radius-lg);
  transition: background 0.2s;
}

.user-dropdown:hover {
  background: var(--color-bg-layout);
}

.user-avatar {
  background: linear-gradient(135deg, var(--color-primary), #36cfc9);
  color: #fff;
  font-weight: 500;
}

.user-info {
  display: flex;
  flex-direction: column;
  line-height: 1.3;
}

.user-name {
  font-size: var(--font-size-base);
  font-weight: 500;
  color: var(--color-text-primary);
}

.user-role {
  font-size: var(--font-size-xs);
  color: var(--color-text-tertiary);
}

/* 内容区域 */
.content-area {
  flex: 1;
  padding: var(--spacing-lg) var(--spacing-xl);
  overflow-y: auto;
  overflow-x: hidden;
  min-width: 0;
}

.no-sidebar .content-area { padding: 0; }

/* 页面切换动画 */
/* 页面切换动画已移除以提升响应速度 */

/* 折叠状态样式 */
.sidebar-collapsed .sidebar { width: 64px; }

.sidebar-collapsed .menu-item.level-1 {
  justify-content: center;
  padding: 14px;
}

.sidebar-collapsed .menu-icon {
  margin-right: 0;
  font-size: 20px;
}

.sidebar-collapsed .collapse-btn { justify-content: center; }

/* 响应式 */
@media (max-width: 480px) {
  .sidebar { transform: translateX(-100%); }
  .main-container { margin-left: 0 !important; max-width: 100vw !important; }
  .topbar-right .global-search { display: none; }
}
</style>
