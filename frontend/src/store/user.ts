import { defineStore } from "pinia";
import { ref } from "vue";
import { authApi } from "../api";

export const useUserStore = defineStore("user", () => {
  const token = ref(localStorage.getItem("token") || "");
  const user = ref<any>(null);
  const permissions = ref<string[]>([]);
  const layoutConfig = ref<{ sidebarMode: string; landingPage: string }>({ sidebarMode: 'auto', landingPage: '' });

  const login = async (username: string, password: string) => {
    const res = await authApi.login({ username, password });
    if (res.data.success) {
      token.value = res.data.data.token;
      user.value = res.data.data.user;
      localStorage.setItem("token", token.value);
      await fetchUserInfo();
    }
    return res.data;
  };

  const logout = async () => {
    try {
      await authApi.logout();
    } catch (e) {
      // 忽略登出错误
    }
    clearAuth();
  };

  const clearAuth = () => {
    token.value = "";
    user.value = null;
    permissions.value = [];
    layoutConfig.value = { sidebarMode: 'auto', landingPage: '' };
    localStorage.removeItem("token");
  };

  const fetchUserInfo = async () => {
    if (!token.value) return;
    try {
      const res = await authApi.getInfo();
      if (res.data.success) {
        user.value = res.data.data;
        permissions.value = res.data.data.permissions || [];
        if (res.data.data.layoutConfig) {
          layoutConfig.value = res.data.data.layoutConfig;
        }
      }
    } catch (e) {
      // ignore
    }
  };

  const hasPermission = (code: string) => {
    return permissions.value.includes(code);
  };

  const setToken = (newToken: string) => {
    token.value = newToken;
    localStorage.setItem("token", newToken);
  };

  const setUser = (newUser: any) => {
    user.value = newUser;
  };

  return {
    token,
    user,
    permissions,
    layoutConfig,
    login,
    logout,
    clearAuth,
    fetchUserInfo,
    hasPermission,
    setToken,
    setUser
  };
});
