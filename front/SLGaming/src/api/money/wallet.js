import http from "@/utils/http";

export const getwalletapi = () => {
  return http({
    url: "/user/wallet",
  });
};
