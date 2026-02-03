import http from "@/utils/http";

/**
 * 获取陪玩资料
 * @returns {Promise}
 * @returns {object} data - CompanionInfo
 * @returns {number} data.userId - 用户ID (int64)
 * @returns {string} data.gameSkill - 游戏技能（单个游戏名称）
 * @returns {number} data.pricePerHour - 每小时价格（帅币）
 * @returns {number} data.status - 状态：0=离线, 1=在线, 2=忙碌
 * @returns {number} data.rating - 评分（0-5分）
 * @returns {number} data.totalOrders - 总接单数
 * @returns {boolean} data.isVerified - 是否认证
 * @returns {string} [data.nickname] - 昵称
 * @returns {string} [data.avatarUrl] - 头像URL
 * @returns {string} [data.bio] - 个人简介
 */
export const getcompanionapi = () => {
  return http({
    url: "/user/companion/profile",
    method: "GET",
  });
};
/**
 * 更新陪玩资料
 * @param {string} [gameSkill] - 游戏技能
 * @param {number} [pricePerHour] - 每小时价格
 * @param {number} [status] - 状态：0=离线, 1=在线, 2=忙碌
 * @returns {Promise}
 */
export const updatecompanionapi = ({ gameSkill, pricePerHour, status } = {}) => {
  return http({
    url: "/user/companion/profile",
    method: "PUT",
    data: { gameSkill, pricePerHour, status },
  });
};

/**
 * 申请成为陪玩
 * @param {string} gameSkill - 游戏技能（单个游戏名称）
 * @param {number} pricePerHour - 每小时价格（帅币）
 * @param {string} [bio] - 个人简介（可选）
 * @returns {Promise}
 * @returns {object} data - UserInfo
 * @returns {number} [data.id] - 用户ID (int64)
 * @returns {number} [data.uid] - 用户UID (int64)
 * @returns {string} [data.nickname] - 昵称
 * @returns {string} [data.phone] - 手机号
 * @returns {number} [data.role] - 用户角色：1=老板, 2=陪玩, 3=管理员
 * @returns {string} [data.avatarUrl] - 头像URL
 * @returns {string} [data.bio] - 个人简介
 */
export const applycompanionapi = ({ gameSkill, pricePerHour, bio } = {}) => {
  return http({
    url: "/user/companion/apply",
    method: "POST",
    data: { gameSkill, pricePerHour, bio },
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
