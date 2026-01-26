import http from "@/utils/http";

/**
 * 获取陪玩列表
 * @param {Object} params - 请求参数
 * @param {string[]} [params.gameSkills] - 游戏技能筛选
 * @param {number} [params.minPrice] - 最低价格
 * @param {number} [params.maxPrice] - 最高价格
 * @param {number} [params.status] - 状态筛选：0=离线, 1=在线, 2=忙碌（默认1）
 * @param {boolean} [params.isVerified] - 是否只返回认证陪玩
 * @param {number} [params.page] - 页码（从1开始）
 * @param {number} [params.pageSize] - 每页数量
 */
export const getcompanionlist = (params) => {
  return http({
    url: "/user/companions",
    method: "GET",
    params,
  });
};
