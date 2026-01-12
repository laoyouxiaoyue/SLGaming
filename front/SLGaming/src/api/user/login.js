import http from "@/utils/http";

export const loginAPI = (data) => {
  return http({
    url: "/user/login",
    method: "POST",
    data,
  });
};

export const codeLoginAPI = (data) => {
  return http({
    url: "/user/login-by-code",
    method: "POST",
    data,
  });
};
