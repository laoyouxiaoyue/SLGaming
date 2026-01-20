import http from "@/utils/http";

/**
 * 发送验证码接口
 * @param {Object} params
 * @param {string} params.phone - 手机号
 * @param {'register'|'login'|'resetpassword'} params.purpose - 发送目的：注册、登录、重置密码
 * @returns {Promise} 请求结果 Promise
 *
 * 用法示例：
 *   codeAPI({ phone: '13800138000', purpose: 'login' })
 */
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
