import { fileURLToPath, URL } from "node:url";

import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import vueDevTools from "vite-plugin-vue-devtools";

// elementPlus按需导入
import AutoImport from "unplugin-auto-import/vite";
import Components from "unplugin-vue-components/vite";
import { ElementPlusResolver } from "unplugin-vue-components/resolvers";
// import ElementPlus from "unplugin-element-plus/vite";
// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
    AutoImport({
      resolvers: [ElementPlusResolver()],
    }),
    Components({
      resolvers: [ElementPlusResolver({ importStyle: "sass" })],
    }),
  ],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        // 2. 自动导入定制化样式文件进行样式覆盖
        additionalData: `
          @use "@/styles/element/index.scss" as *;
          @use "@/styles/var.scss" as *;
        `,
      },
    },
  },
  server: {
    proxy: {
      // 匹配以 /api 开头的请求
      "/api": {
        // target: "http://120.26.29.194:8888", // 后端接口地址，包含 /api 前缀
        target: "http://192.168.0.108:8888", // 后端接口地址，包含 /api 前缀
        changeOrigin: true, // 开启跨域
        // rewrite: (path) => path.replace(/^\/api/, '') // 如果后端接口没有 /api 前缀，需要重写路径
      },
    },
  },
});
