import http from "@/utils/http";

export const getInfoAPI = () => {
  return http({
    url: "/user",
    method: "GET",
  });
};

/**
 * 更新用户信息
 * @param {string} [nickname] - 昵称（可选）
 * @param {string} [password] - 密码（可选）
 * @param {string} [phone] - 手机号（可选）
 * @param {number} [role] - 用户角色（可选，1=老板, 2=陪玩, 3=管理员）
 * @param {string} [avatarUrl] - 头像URL（可选）
 * @param {string} [bio] - 个人简介（可选）
 * @returns {Promise}
 */
export const updateInfoAPI = ({ nickname, password, phone, role, avatarUrl, bio } = {}) => {
  return http({
    url: "/user",
    method: "PUT",
    data: { nickname, password, phone, role, avatarUrl, bio },
  });
};
