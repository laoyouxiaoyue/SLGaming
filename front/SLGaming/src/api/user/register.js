import http from "@/utils/http";

export const registerapi = ({ phone, code, password, nickname }) => {
  return http({
    url: "/user/register",
    method: "POST",
    data: {
      phone,
      code,
      password,
      nickname,
    },
  });
};
