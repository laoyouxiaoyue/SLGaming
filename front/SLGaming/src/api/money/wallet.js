import http from "@/utils/http";

export const getwalletapi = () => {
  return http({
    url: "/user/wallet",
  });
};

/**
 * 创建充值订单
 * @param {Object} data - 请求参数
 * @param {number} data.amount - 充值金额（分/帅币）, 示例: 100
 * @param {string} [data.payType] - 支付方式：alipay_page(PC，默认) / alipay_wap(H5) / alipay_app(APP), 枚举值: alipay_page, alipay_wap, alipay_app, 示例: alipay_page
 * @param {string} [data.returnUrl] - 同步回跳地址（可选）, 示例: https://example.com/pay/return
 * @returns {Promise}
 * HTTP 状态码：200
 * code: integer <int32>, 响应码，0表示成功
 * msg: string, 响应消息
 * data: object (RechargeCreateData)
 *   orderNo: string, 充值单号, 示例: RCFBWT5BX7514W
 *   payUrl: string, 支付跳转URL（PC/WAP）, 示例: https://openapi-sandbox.dl.alipaydev.com/gateway.do?...
 *   payForm: string, 支付表单（APP/PC场景可选）
 *   expiresIn: integer <int64>, 订单有效期（秒）, 示例: 1800
 */
export const createrechargeorderapi = (data) => {
  return http({
    url: "/user/recharge",
    method: "POST",
    data,
  });
};

/**
 * 支付宝异步通知接口
 * @param {Object} [data] - 支付宝异步通知参数
 * @returns {Promise}
 * HTTP 状态码：200
 * code: integer <int32>, 响应码，0表示成功
 * msg: string, 响应消息
 */
export const alipaynotifyapi = (data) => {
  return http({
    url: "/user/recharge/alipay/notify",
    method: "POST",
    data,
  });
};

/**
 * 查询充值订单
 * @param {string} orderNo - 充值单号
 * @returns {Promise}
 * HTTP 状态码：200
 * code: integer <int32>, 响应码，0表示成功
 * msg: string, 响应消息
 * data: object (RechargeQueryData)
 *   orderNo: string, 充值单号
 *   status: integer, 订单状态：0=待支付,1=成功,2=失败,3=关闭
 *   amount: integer <int64>, 金额（分/帅币）
 */
export const queryrechargeorderapi = (orderNo) => {
  return http({
    url: "/user/recharge",
    method: "GET",
    params: {
      orderNo,
    },
  });
};
