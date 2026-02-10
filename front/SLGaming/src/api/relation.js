import http from "@/utils/http";

/**
 * 关注用户接口
 *
 * @param {object} data - 请求体参数
 * @param {number} data.targetUserId - 目标用户ID (必须)
 *
 * @returns {Promise}
 * @returns {boolean} response.data.success - 操作是否成功
 */
export const followUserAPI = (data) => {
  return http({
    url: "/user/follow",
    method: "POST",
    data,
  });
};
/**
 * 取消关注接口
 *
 * @param {object} params - 查询参数
 * @param {number} params.targetUserId - 目标用户ID (必须)
 *
 * @returns {Promise}
 * @returns {boolean} response.data.success - 操作是否成功
 */
export const unfollowUserAPI = (params) => {
  return http({
    url: "/user/follow",
    method: "DELETE",
    params,
  });
};
/**
 * 检查关注状态接口
 *
 * @param {object} params - 查询参数
 * @param {number} params.targetUserId - 目标用户ID (必须)
 *
 * @returns {Promise}
 * @returns {boolean} response.data.following - 是否已关注
 */
export const checkFollowStatusAPI = (params) => {
  return http({
    url: "/user/follow/status",
    method: "GET",
    params,
  });
};
/**
 * 获取粉丝列表接口
 *
 * @param {object} params - 查询参数
 * @param {number} [params.page=1] - 页码
 * @param {number} [params.pageSize=20] - 每页数量
 *
 * @returns {Promise}
 * @returns {object} response.data - FollowListData
 * @returns {Array<object>} response.data.items - 粉丝列表
 * @returns {number} response.data.items[].userId - 用户ID
 * @returns {number} response.data.items[].uid - 用户UID
 * @returns {string} response.data.items[].nickname - 昵称
 * @returns {string} response.data.items[].avatarUrl - 头像URL
 * @returns {string} response.data.items[].bio - 个人简介
 * @returns {string} response.data.items[].followedAt - 关注时间
 * @returns {number} response.data.total - 总数
 * @returns {number} response.data.page - 当前页码
 * @returns {number} response.data.pageSize - 每页数量
 */
export const getFollowersAPI = (params) => {
  return http({
    url: "/user/followers",
    method: "GET",
    params,
  });
};
/**
 * 获取关注列表接口
 *
 * @param {object} params - 查询参数
 * @param {number} [params.page=1] - 页码
 * @param {number} [params.pageSize=20] - 每页数量
 *
 * @returns {Promise}
 * @returns {object} response.data - FollowListData
 * @returns {Array<object>} response.data.items - 关注列表
 * @returns {number} response.data.items[].userId - 用户ID
 * @returns {number} response.data.items[].uid - 用户UID
 * @returns {string} response.data.items[].nickname - 昵称
 * @returns {string} response.data.items[].avatarUrl - 头像URL
 * @returns {string} response.data.items[].bio - 个人简介
 * @returns {string} response.data.items[].followedAt - 关注时间
 * @returns {number} response.data.total - 总数
 * @returns {number} response.data.page - 当前页码
 * @returns {number} response.data.pageSize - 每页数量
 */
export const getFollowingsAPI = (params) => {
  return http({
    url: "/user/followings",
    method: "GET",
    params,
  });
};
