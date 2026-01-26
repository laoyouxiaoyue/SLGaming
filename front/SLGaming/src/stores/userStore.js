import { defineStore } from "pinia";
import { ref } from "vue";
import { loginAPI, codeLoginAPI } from "@/api/user/login";

export const useUserStore = defineStore(
  "user",
  () => {
    // 1. 定义管理用户数据的state
    const userInfo = ref({});
    // 2. 定义获取接口数据的action函数
    const getUserInfo = async ({ phone, password }) => {
      const res = await loginAPI({ phone, password });
      userInfo.value = res.data;
    };
    const getUserInfoByCode = async ({ phone, code }) => {
      const res = await codeLoginAPI({ phone, code });
      userInfo.value = res.data;
    };
    const clearUserInfo = () => {
      userInfo.value = {};
    };
    // 3. 以对象的格式把state和action return
    return {
      userInfo,
      getUserInfo,
      getUserInfoByCode,
      clearUserInfo,
    };
  },
  {
    // Pinia 解决了组件通信和响应式问题。
    // LocalStorage 解决了持久化（防刷新、跨会话保存）问题
    // 通过 persist: true，你享受了 Pinia 的便利，同时自动获得了 LocalStorage 的持久化能力
    persist: true,
  },
);
