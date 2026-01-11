import http from "@/utils/http";

export const loginAPI = (data) => {
  return http({
    url: "/user/login",
    method: "POST",
    data,
  });
};
