import http from "@/utils/http";

// purpose:'register'/'login'
export const codeAPI = ({ phone, purpose }) => {
  return http({
    url: "/code/send",
    method: "POST",
    data: {
      phone,
      purpose,
    },
  });
};
