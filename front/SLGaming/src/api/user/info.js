import http from "@/utils/http";

// export const getInfoAPI = ({ id = "", uid = "", phone = "" } = {}) => {
//   return http({
//     url: "/user",
//     params: {
//       id,
//       uid,
//       phone,
//     },
//   });
// };

export const getInfoAPI = () => {
  return http({
    url: "/user",
  });
};
