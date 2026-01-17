import { defineConfig, globalIgnores } from "eslint/config";
import globals from "globals";
import js from "@eslint/js";
import pluginVue from "eslint-plugin-vue";

export default defineConfig([
  {
    name: "app/files-to-lint",
    files: ["**/*.{vue,js,mjs,jsx}"],
  },

  globalIgnores(["**/dist/**", "**/dist-ssr/**", "**/coverage/**"]),

  {
    languageOptions: {
      globals: {
        ...globals.browser,
      },
    },
  },

  js.configs.recommended,
  ...pluginVue.configs["flat/essential"],

  // 覆盖规则：允许特定组件使用单词名
  {
    files: ["src/views/**/index.vue", "src/views/layout/component/Nav.vue"],
    rules: {
      "vue/multi-word-component-names": "off",
    },
  },
]);
