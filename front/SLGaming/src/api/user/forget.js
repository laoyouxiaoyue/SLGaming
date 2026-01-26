import http from "@/utils/http";

/**
 * 忘记密码-重置密码接口
 * @param {Object} params
 * @param {string} params.phone - 手机号，必需，示例：13800138000
 * @param {string} params.code - 验证码，必需，示例：123456
 * @param {string} params.password - 新密码，必需，示例：newpassword123
 * @returns {Promise} 请求结果 Promise
 *
 * 用法示例：
 *   forgetpasswordapi({ phone: '13800138000', code: '123456', password: 'newpassword123' })
 */
export const forgetpasswordapi = ({ phone, code, password }) => {
  return http({
    url: "/user/forgetPassword",
    method: "PUT",
    data: {
      phone,
      code,
      password,
    },
  });
};
