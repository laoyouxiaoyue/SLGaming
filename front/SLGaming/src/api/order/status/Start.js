import http from "@/utils/http";

/**
 * 陪玩开始服务(仅陪玩角色可用)
 *
 * @param {object} data - 请求体
 * @param {number} data.orderId - 订单ID (int64) - 必需
 *
 * @returns {Promise}
 * @returns {object} response.data - OrderInfo
 * @returns {number} response.data.id - 订单ID
 * @returns {string} response.data.orderNo - 订单号
 * @returns {number} response.data.bossId - 老板ID
 * @returns {number} response.data.companionId - 陪玩ID
 * @returns {string} response.data.gameName - 游戏名称
 * @returns {number} response.data.durationHours - 时长（小时）
 * @returns {number} response.data.pricePerHour - 每小时价格（帅币）
 * @returns {number} response.data.totalAmount - 订单总价（帅币）
 * @returns {number} response.data.status - 状态：1=CREATED, 2=PAID, 3=ACCEPTED, 4=IN_SERVICE, 5=COMPLETED, 6=CANCELLED, 7=RATED
 * @returns {number} response.data.createdAt - 创建时间（时间戳）
 * @returns {number} response.data.paidAt - 支付时间（时间戳）
 * @returns {number} response.data.acceptedAt - 接单时间（时间戳）
 * @returns {number} response.data.startAt - 开始服务时间（时间戳）
 * @returns {number} response.data.completedAt - 完成时间（时间戳）
 * @returns {number} response.data.cancelledAt - 取消时间（时间戳）
 * @returns {number} response.data.rating - 评分（0-5）
 * @returns {string} response.data.comment - 评价内容
 * @returns {string} response.data.cancelReason - 取消原因
 */
export const startOrderServiceAPI = (data) => {
  return http({
    url: "/order/start",
    method: "POST",
    data,
  });
};
