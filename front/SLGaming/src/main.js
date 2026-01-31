import { createApp } from "vue";
import { createPinia } from "pinia";
import piniaPluginPersistedstate from "pinia-plugin-persistedstate";

import App from "./App.vue";
import router from "./router";
import Icon from "@/components/SlIcon.vue";
// 引入初始化样式文件
import "@/styles/common.scss";

const app = createApp(App);
app.component("SlIcon", Icon);
const pinia = createPinia();

pinia.use(piniaPluginPersistedstate);
app.use(pinia);
app.use(router);

app.mount("#app");
