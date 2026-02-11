import http from "@/utils/http";

/**
 * 关注用户接口
 *
 * @param {object} data - 请求体参数
 * @param {number} data.userId - 目标用户ID (必须)
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
 * @param {number} params.userId - 目标用户ID (必须)
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
 * @returns {object} response.data - CheckFollowStatusData
 * @returns {boolean} response.data.isFollowing - 当前用户是否关注目标用户
 * @returns {boolean} response.data.isFollowed - 目标用户是否关注当前用户
 * @returns {boolean} response.data.isMutual - 是否互相关注
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
 * @returns {object} response.data - GetMyFollowersListData
 * @returns {Array<object>} response.data.users - 粉丝列表 (UserFollowInfo)
 * @returns {number} response.data.users[].userId - 用户ID
 * @returns {string} response.data.users[].nickname - 昵称
 * @returns {string} response.data.users[].avatarUrl - 头像URL
 * @returns {number} response.data.users[].role - 用户角色：1=老板,2=陪玩
 * @returns {boolean} response.data.users[].isVerified - 是否验证（仅陪玩）
 * @returns {number} response.data.users[].rating - 评分（仅陪玩）
 * @returns {number} response.data.users[].totalOrders - 总接单数（仅陪玩）
 * @returns {boolean} response.data.users[].isMutual - 是否互相关注
 * @returns {number} response.data.users[].followedAt - 关注时间戳
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
 * @returns {object} response.data - GetMyFollowingListData
 * @returns {Array<object>} response.data.users - 关注列表 (UserFollowInfo)
 * @returns {number} response.data.users[].userId - 用户ID
 * @returns {string} response.data.users[].nickname - 昵称
 * @returns {string} response.data.users[].avatarUrl - 头像URL
 * @returns {number} response.data.users[].role - 用户角色：1=老板,2=陪玩
 * @returns {boolean} response.data.users[].isVerified - 是否验证（仅陪玩）
 * @returns {number} response.data.users[].rating - 评分（仅陪玩）
 * @returns {number} response.data.users[].totalOrders - 总接单数（仅陪玩）
 * @returns {boolean} response.data.users[].isMutual - 是否互相关注
 * @returns {number} response.data.users[].followedAt - 关注时间戳
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
