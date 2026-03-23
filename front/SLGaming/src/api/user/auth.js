import http from "@/utils/http";
import { useUserStore } from "@/stores/userStore";

// 刷新token
export const refreshTokenAPI = () => {
  const userStore = useUserStore();
  const refreshToken = userStore.userInfo.refreshToken;

  return http({
    url: "/user/refresh-token",
    method: "POST",
    data: {
      refreshToken,
    },
  });
};
