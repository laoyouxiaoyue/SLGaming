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
