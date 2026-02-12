import { defineStore } from "pinia";
import { ref } from "vue";
import { getInfoAPI, updateInfoAPI } from "@/api/user/info";

export const useInfoStore = defineStore(
  "info",
  () => {
    // 1. 定义管理用户详细信息的state
    /**
     * info存放内容形式:
     * {
     *   id: string,             // 用户唯一标识，例如 "1996080936390758400"
     *   uid: number,            // 用户短ID，例如 38830062
     *   nickname: string,       // 用户昵称，例如 "王者"
     *   phone: string,          // 手机号，例如 "13124917464"
     *   role: number,           // 角色，例如 2
     *   avatarUrl: string,      // 头像URL
     *   bio: string,            // 个人简介，例如 "老司机，懂就上车"
     *   balance: number,        // 余额
     *   frozenBalance: number,  // 冻结余额
     *   followerCount: number,  // 粉丝数
     *   followingCount: number  // 关注数
     * }
     */
    const info = ref({});

    // 2. 定义获取接口数据的action函数
    const getUserDetail = async () => {
      try {
        const res = await getInfoAPI();
        if (res.data) {
          info.value = res.data;

          // 适配头像地址: 如果是远程的具体IP地址且端口不对（默认80），转为相对路径走代理
          if (info.value.avatarUrl && info.value.avatarUrl.includes("http://120.26.29.242")) {
            info.value.avatarUrl = info.value.avatarUrl.replace("http://120.26.29.242", "");
          }

          // 检查图片是否有效，若 404 则清空，防止控制台报错
          if (info.value.avatarUrl) {
            const img = new Image();
            img.src = info.value.avatarUrl;
            img.onerror = () => {
              info.value.avatarUrl = "";
            };
          }
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
