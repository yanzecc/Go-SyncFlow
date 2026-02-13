import { createApp } from "vue";
import { createPinia } from "pinia";
import ElementPlus from "element-plus";
import zhCn from "element-plus/dist/locale/zh-cn.mjs";
import "element-plus/dist/index.css";
import "./styles/variables.css";
import "./styles/global.css";

import App from "./App.vue";
import router from "./router";
import { api } from "./api";

const app = createApp(App);

app.use(createPinia());
app.use(ElementPlus, { locale: zhCn });
app.use(router);

// 动态设置浏览器标题（异步，不阻塞渲染）
api.get("/settings/ui").then((res) => {
  if (res.data.success && res.data.data.browserTitle) {
    document.title = res.data.data.browserTitle;
  }
}).catch(() => {});

app.mount("#app");
