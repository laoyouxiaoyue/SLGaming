import http from "@/utils/http";

export const getInfoAPI = () => {
  return http({
    url: "/user",
  });
};
