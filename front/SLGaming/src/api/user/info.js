import http from "@/utils/http";

/**
 * 获取用户信息
 * @returns {Promise}
 * @returns {object} data
 * @returns {number} data.id - 用户ID (int64)
 * @returns {number} data.uid - 用户UID (int64)
 * @returns {string} data.nickname - 昵称
 * @returns {string} data.phone - 手机号
 * @returns {number} data.role - 用户角色：1=老板, 2=陪玩, 3=管理员
 * @returns {string} data.avatarUrl - 头像URL
 * @returns {string} data.bio - 个人简介
 */
export const getInfoAPI = () => {
  return http({
    url: "/user",
    method: "GET",
  });
};

/**
 * 更新用户信息
 *
 * @param {string} [nickname] - 昵称（可选）
 * @param {string} [password] - 密码（可选）
 * @param {string} [phone] - 手机号（可选）
 * @param {number} [role] - 用户角色（可选，1=老板, 2=陪玩, 3=管理员）
 * @param {string} [avatarUrl] - 头像URL（可选）
 * @param {string} [bio] - 个人简介（可选）
 * @returns {Promise}
 */
export const updateInfoAPI = ({ id, nickname, password, phone, role, avatarUrl, bio } = {}) => {
  return http({
    url: "/user",
    method: "PUT",
    data: { id, nickname, password, phone, role, avatarUrl, bio },
  });
};
