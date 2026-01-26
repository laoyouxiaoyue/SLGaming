import http from "@/utils/http";

export const getgameskillapi = () => {
  return http({
    url: "/user/gameskills",
  });
};
