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

/**
 * 上传头像
 * Body 参数 multipart/form-data
 * @param {File} avatar - 头像文件 (必需)
 * @returns {Promise}
 * @returns {object} data - UploadAvatarData
 * @returns {string} [data.avatarUrl] - 头像URL (可选)
 * @example
 * // 返回示例
 * // {
 * //   avatarUrl: "/uploads/avatars/1700000000000000000.png"
 * // }
 */
export const upavatarUrlapi = (avatar) => {
  const formData = new FormData();
  formData.append("avatar", avatar);
  return http({
    url: "/user/avatar",
    method: "POST",
    data: formData,
  });
};
