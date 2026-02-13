import http from "@/utils/http";

/**
 * 获取接单排行榜列表
 * @param {object} params
 * @param {number} [params.page] - 页码（从1开始，默认1）
 * @param {number} [params.pageSize] - 每页数量（默认10）
 * @returns {Promise}
 * @returns {object} data
 * @returns {number} data.page
 * @returns {number} data.pageSize
 * @returns {number} data.total
 * @returns {Array} data.rankings
 * @returns {object} data.rankings[]
 * @returns {string} data.rankings[].avatarUrl - 头像URL
 * @returns {boolean} data.rankings[].isVerified - 是否认证
 * @returns {string} data.rankings[].nickname - 昵称
 * @returns {number} data.rankings[].rank - 排名（从1开始）
 * @returns {number} data.rankings[].rating - 评分（用于评分排名）
 * @returns {number} data.rankings[].totalOrders - 总接单数（用于接单数排名）
 * @returns {number} data.rankings[].userId - 用户ID
 */
export const getCompanionOrdersRankingAPI = ({ page = 1, pageSize = 10 } = {}) => {
  return http({
    url: "/user/companions/ranking/orders",
    method: "GET",
    params: {
      page,
      pageSize,
    },
  });
};

/**
 * 获取陪玩评分排行榜
 * @param {object} params
 * @param {number} [params.page] - 页码（从1开始，默认1）
 * @param {number} [params.pageSize] - 每页数量（默认10）
 * @returns {Promise}
 * @returns {object} data
 * @returns {number} data.page
 * @returns {number} data.pageSize
 * @returns {number} data.total
 * @returns {Array} data.rankings
 * @returns {object} data.rankings[]
 * @returns {string} data.rankings[].avatarUrl - 头像URL
 * @returns {boolean} data.rankings[].isVerified - 是否认证
 * @returns {string} data.rankings[].nickname - 昵称
 * @returns {number} data.rankings[].rank - 排名（从1开始）
 * @returns {number} data.rankings[].rating - 评分（用于评分排名）
 * @returns {number} data.rankings[].totalOrders - 总接单数（用于接单数排名）
 * @returns {number} data.rankings[].userId - 用户ID
 */
export const getCompanionRatingRankingAPI = ({ page = 1, pageSize = 10 } = {}) => {
  return http({
    url: "/user/companions/ranking/ratings",
    method: "GET",
    params: {
      page,
      pageSize,
    },
  });
};
