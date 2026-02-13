import http from "@/utils/http";

/**
 * 获取用户信息
 * @param {object} [params]
 * @param {number} [params.id] - 用户ID（可选，不传则获取当前用户）
 * @param {number} [params.uid] - 用户UID（可选）
 * @param {string} [params.phone] - 手机号（可选）
 * @returns {Promise}
 * @returns {object} data
 * @returns {string} data.avatarUrl - 头像URL
 * @returns {number} data.balance - 帅币可用余额
 * @returns {string} data.bio - 个人简介
 * @returns {number} data.followerCount - 粉丝数
 * @returns {number} data.followingCount - 关注数
 * @returns {number} data.frozenBalance - 冻结帅币余额（预留）
 * @returns {number} data.id - 用户ID (int64)
 * @returns {string} data.nickname - 昵称
 * @returns {string} data.phone - 手机号
 * @returns {number} data.role - 用户角色：1=老板, 2=陪玩, 3=管理员
 * @returns {number} data.uid - 用户唯一标识 (int64)
 */
export const getInfoAPI = ({ id, uid, phone } = {}) => {
  return http({
    url: "/user",
    method: "GET",
    params: { id, uid, phone },
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

/**
 * 修改手机号码
 * @param {string} oldPhone - 原手机号 (必需)
 * @param {string} oldCode - 原手机号验证码 (必需)
 * @param {string} newPhone - 新手机号 (必需)
 * @param {string} [newCode] - 新手机号验证码 (必须)
 * @returns {Promise}
 */
export const changePhoneAPI = ({ oldPhone, oldCode, newPhone, newCode } = {}) => {
  return http({
    url: "/user/change-phone",
    method: "PUT",
    data: { oldPhone, oldCode, newPhone, newCode },
  });
};

/**
 * 修改密码（通过手机验证）
 * @param {string} oldPhone - 原手机号 (必需)
 * @param {string} oldCode - 原手机号验证码 (必需)
 * @param {string} newPassword - 新密码 (必需)
 * @returns {Promise}
 */
export const changePasswordAPI = ({ oldPhone, oldCode, newPassword } = {}) => {
  return http({
    url: "/user/change-password",
    method: "PUT",
    data: { oldPhone, oldCode, newPassword },
  });
};
