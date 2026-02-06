import http from "@/utils/http";

/**
 * 获取订单详情
 *
 * @param {object} params - 查询参数
 * @param {number} [params.id] - 订单ID (int64) - 可选
 * @param {string} [params.orderNo] - 订单号 - 可选
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
export const getOrderDetailAPI = (params) => {
  return http({
    url: "/order",
    method: "GET",
    params,
  });
};

/**
 * 获取订单列表
 *
 * @param {object} params - 查询参数
 * @param {'boss' | 'companion'} [params.role] - 角色：boss / companion（默认 boss） - 可选
 * @param {number} [params.status] - 状态筛选：1=CREATED, 2=PAID, 3=ACCEPTED, 4=IN_SERVICE, 5=COMPLETED, 6=CANCELLED, 7=RATED - 可选
 * @param {number} [params.page] - 页码（从1开始） - 可选
 * @param {number} [params.pageSize] - 每页数量 - 可选
 *
 * @returns {Promise}
 * @returns {object} response.data - GetOrderListData
 * @returns {Array<object>} response.data.orders - 订单列表
 * @returns {number} response.data.total - 总数
 * @returns {number} response.data.page - 当前页码
 * @returns {number} response.data.pageSize - 每页数量
 */
export const getOrderListAPI = (params) => {
  return http({
    url: "/orders",
    method: "GET",
    params,
  });
};
