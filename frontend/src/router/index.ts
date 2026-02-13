import { createRouter, createWebHistory } from "vue-router";
import { useUserStore } from "../store/user";

const routes = [
  {
    path: "/login",
    name: "Login",
    component: () => import("../views/Login.vue"),
    meta: { public: true }
  },
  {
    path: "/",
    redirect: "/admin"
  },
  {
    path: "/admin",
    component: () => import("../layouts/MainLayout.vue"),
    children: [
      {
        path: "",
        name: "AdminHome",
        component: () => import("../views/admin/AdminHome.vue"),
        meta: { permission: "settings:system" }
      },
      // 个人中心
      {
        path: "profile",
        name: "Profile",
        component: () => import("../views/admin/Profile.vue"),
      },
      // 用户源 - 本地用户
      {
        path: "users/local",
        name: "LocalUsers",
        component: () => import("../views/admin/Users.vue"),
        meta: { permission: "user:list" }
      },
      // 连接器
      {
        path: "users/connectors",
        name: "Connectors",
        component: () => import("../views/admin/Connectors.vue"),
        meta: { permission: "settings:system" }
      },
      // 同步器
      {
        path: "users/synchronizers",
        name: "Synchronizers",
        component: () => import("../views/admin/Synchronizers.vue"),
        meta: { permission: "settings:system" }
      },
      // 角色管理
      {
        path: "roles",
        name: "Roles",
        component: () => import("../views/admin/Roles.vue"),
        meta: { permission: "role:list" }
      },
      // 日志管理
      {
        path: "logs/system",
        name: "SystemLogs",
        component: () => import("../views/admin/SystemLogs.vue"),
        meta: { permission: "log:login" }
      },
      {
        path: "logs/login",
        redirect: "/admin/logs/system"
      },
      {
        path: "logs/operation",
        redirect: "/admin/logs/system"
      },
      {
        path: "logs/sync",
        name: "SyncLogs",
        component: () => import("../views/admin/SyncLogs.vue"),
        meta: { permission: "log:operation" }
      },
      // 通知管理
      {
        path: "notify/channels",
        name: "NotifyChannels",
        component: () => import("../views/admin/NotifyChannels.vue"),
        meta: { permission: "settings:system" }
      },
      {
        path: "notify/templates",
        name: "MessageTemplates",
        component: () => import("../views/admin/MessageTemplates.vue"),
        meta: { permission: "settings:system" }
      },
      {
        path: "notify/rules",
        name: "NotifyRules",
        component: () => import("../views/admin/NotifyRules.vue"),
        meta: { permission: "settings:system" }
      },
      // 系统设置
      {
        path: "settings",
        name: "Settings",
        component: () => import("../views/admin/Settings.vue"),
        meta: { permission: "settings:system" }
      },
      // LDAP 服务配置
      {
        path: "settings/ldap",
        name: "LDAPSettings",
        component: () => import("../views/admin/LDAPSettings.vue"),
        meta: { permission: "settings:system" }
      },
      // 钉钉配置（移到设置下）
      {
        path: "settings/dingtalk",
        name: "DataSourceDingtalk",
        component: () => import("../views/admin/DataSourceDingtalk.vue"),
        meta: { permission: "settings:system" }
      },
      // 安全中心
      {
        path: "security",
        name: "Security",
        component: () => import("../views/admin/Security.vue"),
        meta: { permission: "settings:system" }
      },
      // API 密钥管理
      {
        path: "apikeys",
        name: "APIKeys",
        component: () => import("../views/admin/APIKeys.vue"),
        meta: { permission: "settings:system" }
      },
      // API文档
      {
        path: "api-docs",
        name: "ApiDocs",
        component: () => import("../views/admin/ApiDocs.vue"),
        meta: { permission: "settings:system" }
      },
      // ========== 同步管理 ==========
      // 上游同步
      {
        path: "sync/upstream",
        name: "UpstreamSync",
        component: () => import("../views/admin/UpstreamSync.vue"),
        meta: { permission: "settings:system" }
      },
      // 下游同步
      {
        path: "sync/downstream",
        name: "DownstreamSync",
        component: () => import("../views/admin/DownstreamSync.vue"),
        meta: { permission: "settings:system" }
      },
      // ========== 日志管理扩展 ==========
      // API 调用日志
      {
        path: "logs/api",
        name: "ApiAccessLogs",
        component: () => import("../views/admin/ApiAccessLogs.vue"),
        meta: { permission: "settings:system" }
      },
      // 日志设置
      {
        path: "logs/settings",
        name: "LogSettings",
        component: () => import("../views/admin/LogSettings.vue"),
        meta: { permission: "settings:system" }
      }
    ]
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

// 根据用户权限找到第一个可访问的页面
function getFirstAccessibleRoute(userStore: ReturnType<typeof useUserStore>): string | null {
  // 按优先级排列的页面-权限映射
  const routePermMap = [
    { path: '/admin/users/local', permission: 'user:list' },
    { path: '/admin/sync/upstream', permission: 'settings:system' },
    { path: '/admin/sync/downstream', permission: 'settings:system' },
    { path: '/admin/roles', permission: 'role:list' },
    { path: '/admin/logs/system', permission: 'log:login' },
    { path: '/admin/logs/api', permission: 'settings:system' },
    { path: '/admin/notify/channels', permission: 'settings:system' },
    { path: '/admin/settings', permission: 'settings:system' },
    { path: '/admin/security', permission: 'settings:system' },
  ];
  for (const item of routePermMap) {
    if (userStore.hasPermission(item.permission)) {
      return item.path;
    }
  }
  return null;
}

router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore();

  if (to.meta.public) {
    next();
    return;
  }

  if (!userStore.token) {
    next("/login");
    return;
  }

  if (!userStore.user) {
    try {
      await userStore.fetchUserInfo();
      if (!userStore.user) {
        userStore.clearAuth();
        next("/login");
        return;
      }
    } catch (e) {
      userStore.clearAuth();
      next("/login");
      return;
    }
  }

  // 检查路由权限
  const requiredPermission = to.meta.permission as string | undefined;
  if (requiredPermission && !userStore.hasPermission(requiredPermission)) {
    // 无权访问该页面 → 按优先级选择重定向目标：
    // 1. 后端返回的 landingPage（已根据合并权限验证过）
    // 2. 前端根据合并权限动态计算的首个可访问页面
    // 3. 个人中心（兜底）
    const landingPage = userStore.layoutConfig?.landingPage;
    if (landingPage && landingPage !== to.path) {
      next(landingPage);
    } else {
      const target = getFirstAccessibleRoute(userStore);
      if (target && target !== to.path) {
        next(target);
      } else {
        next({ name: 'Profile' });
      }
    }
    return;
  }

  next();
});

export default router;
