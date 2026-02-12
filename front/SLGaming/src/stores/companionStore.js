import { defineStore } from "pinia";
import { ref } from "vue";
import {
  getcompanionapi,
  updatecompanionapi,
  updateCompanionStatusAPI,
} from "@/api/companion/companion";

export const useCompanionStore = defineStore(
  "companion",
  () => {
    // 1. 定义 state
    /**
     * companionInfo存放内容形式:
     * {
     *   gameSkill: string,      // 游戏技能名称，例如 "三角洲行动"
     *   pricePerHour: number,   // 每小时价格，例如 50
     *   status: number,         // 状态（1:在线, 0:离线等），例如 1
     *   rating: number,         // 评分，例如 5
     *   totalOrders: number,    // 总订单数，例如 1
     *   isVerified: boolean,    // 是否认证，例如 false
     *   userId: string,         // 用户ID，例如 "1996080936390758400"
     *   nickname: string,       // 昵称，例如 "王者"
     *   avatarUrl: string,      // 头像链接
     *   bio: string             // 简介，例如 "老司机，懂就上车"
     * }
     */
    const companionInfo = ref({
      gameSkill: "",
      pricePerHour: 0,
      status: 0,
      rating: 0,
      totalOrders: 0,
      isVerified: false,
      userId: 0,
      nickname: "",
      avatarUrl: "",
      bio: "",
    });

    // 2. 定义 actions
    // 获取陪玩信息
    const getCompanionDetail = async () => {
      try {
        const res = await getcompanionapi();
        if (res.code === 0 && res.data) {
          companionInfo.value = {
            ...companionInfo.value,
            ...res.data,
          };
        }
        return res;
      } catch (error) {
        console.error("获取陪玩信息失败:", error);
        throw error;
      }
    };

    // 更新陪玩信息
    const updateCompanionDetail = async (data) => {
      try {
        const res = await updatecompanionapi(data);
        if (res.code === 0) {
          // 更新成功后重新获取最新数据
          await getCompanionDetail();
        }
        return res;
      } catch (error) {
        console.error("更新陪玩信息失败:", error);
        throw error;
      }
    };

    // 更新陪玩在线状态
    const updateStatus = async (status) => {
      try {
        const res = await updateCompanionStatusAPI({ status });
        if (res.code === 0) {
          // 更新成功后同步本地状态
          companionInfo.value.status = status;
        }
        return res;
      } catch (error) {
        console.error("更新状态失败:", error);
        throw error;
      }
    };

    // 3. 清除数据 (用于退出登录)
    const clearCompanionInfo = () => {
      companionInfo.value = {
        gameSkill: "",
        pricePerHour: 0,
        status: 0,
        rating: 0,
        totalOrders: 0,
        isVerified: false,
        userId: 0,
        nickname: "",
        avatarUrl: "",
        bio: "",
      };
    };

    return {
      companionInfo,
      getCompanionDetail,
      updateCompanionDetail,
      updateStatus,
      clearCompanionInfo,
    };
  },
  {
    persist: true, // 开启持久化
  },
);
