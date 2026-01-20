import http from "@/utils/http";

/**
 * 用户注册接口
 * @param {Object} params
 * @param {string} params.phone - 手机号
 * @param {string} params.code - 验证码
 * @param {string} params.password - 密码
 * @param {string} params.nickname - 昵称
 * @returns {Promise} 请求结果 Promise
 *
 * 用法示例：
 *   registerapi({ phone: '13800138000', code: '123456', password: 'abc123', nickname: '张三' })
 */
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
