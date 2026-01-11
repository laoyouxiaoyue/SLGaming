import http from "@/utils/http";

export const getlogoutAPI = () => {
  return http({
    url: "/user/logout",
  });
};
