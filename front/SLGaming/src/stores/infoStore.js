import { defineStore } from "pinia";
import { ref } from "vue";
import { getInfoAPI, updateInfoAPI } from "@/api/user/info";

export const useInfoStore = defineStore(
  "info",
  () => {
    // 1. 定义管理用户详细信息的state
    const info = ref({});

    // 2. 定义获取接口数据的action函数
    const getUserDetail = async () => {
      try {
        const res = await getInfoAPI();
        if (res.data) {
          info.value = res.data;
        }
      } catch (error) {
        console.error("Failed to fetch user info", error);
      }
    };

    // 3. 更新用户信息的action
    const updateUserInfo = async (data) => {
      await updateInfoAPI(data);
      // 更新成功后重新获取最新信息或直接更新本地state
      await getUserDetail();
    };

    // 4. 清除信息
    const clearInfo = () => {
      info.value = {};
    };

    return {
      info,
      getUserDetail,
      updateUserInfo,
      clearInfo,
    };
  },
  {
    persist: true, // 也可选择持久化，避免刷新后store丢失
  },
);
