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

/**
 * 更新陪玩上下线状态
 * @param {Object} data - 请求参数
 * @param {number} data.status - 状态：0=离线, 1=在线, 2=忙碌 (必需)
 * @returns {Promise}
 * data: object (CompanionInfo)
 *   userId: number <int64>, 用户ID
 *   gameSkill: string, 游戏技能
 *   pricePerHour: number <int64>, 每小时价格（帅币）
 *   status: number, 状态：0=离线, 1=在线, 2=忙碌
 *   rating: number <double>, 评分（0-5分）
 *   totalOrders: number <int64>, 总接单数
 *   isVerified: boolean, 是否认证
 *   nickname: string, 昵称（可选）
 *   avatarUrl: string, 头像URL（可选）
 *   bio: string, 个人简介（可选）
 */
export const updateCompanionStatusAPI = (data) => {
  return http({
    url: "/user/companion/status",
    method: "PUT",
    data,
  });
};
