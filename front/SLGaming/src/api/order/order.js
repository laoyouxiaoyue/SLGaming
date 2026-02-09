import http from "@/utils/http";

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
 * 创建订单
 *
 * @param {object} data - 订单请求参数
 * @param {number} data.companionId - 陪玩ID (int64) - 必需
 * @param {string} data.gameName - 游戏名称 - 必需
 * @param {number} data.durationHours - 服务时长（小时） (int32) - 必需
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
export const createOrderAPI = (data) => {
  return http({
    url: "/order",
    method: "POST",
    data,
  });
};

/**
 * 陪玩接单(仅陪玩角色可用)
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
export const acceptOrderAPI = (data) => {
  return http({
    url: "/order/accept",
    method: "POST",
    data,
  });
};

/**
 * 取消订单（老板和陪玩都可以发起）
 *
 * @param {object} data - 请求体
 * @param {number} data.orderId - 订单ID (int64) - 必需
 * @param {string} [data.reason] - 取消原因 - 可选
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
export const cancelOrderAPI = (data) => {
  return http({
    url: "/order/cancel",
    method: "POST",
    data,
  });
};

/**
 * 完成订单(仅陪玩角色可用)
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
export const completeOrderAPI = (data) => {
  return http({
    url: "/order/complete",
    method: "POST",
    data,
  });
};
/**
 * 评价订单 (仅老板角色可用)
 *
 * @param {object} data - 请求体
 * @param {number} data.orderId - 订单ID (int64) - 必需
 * @param {number} data.rating - 评分（0-5） - 必需
 * @param {string} [data.comment] - 评价内容 - 可选
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
export const rateOrderAPI = (data) => {
  return http({
    url: "/order/rate",
    method: "POST",
    data,
  });
};

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
