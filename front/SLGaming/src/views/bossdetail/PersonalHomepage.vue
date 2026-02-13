<script setup>
import { ref, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { getInfoAPI } from "@/api/user/info.js";
import { checkFollowStatusAPI, followUserAPI, unfollowUserAPI } from "@/api/relation";
import { useUserStore } from "@/stores/userStore";
import { useInfoStore } from "@/stores/infoStore";

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();
const InfoStore = useInfoStore();
const loading = ref(true);
const userInfo = ref(null);
const isFollowed = ref(false);

const toggleFollow = async () => {
  if (!userStore.userInfo?.accessToken) {
    ElMessage.warning("请先登录后操作");
    router.push("/login");
    return;
  }

  const userId = route.params.id;
  try {
    if (isFollowed.value) {
      await unfollowUserAPI({ userId: userId });
      isFollowed.value = false;
      ElMessage.info("已取消关注");
    } else {
      await followUserAPI({ userId: userId });
      isFollowed.value = true;
      ElMessage.success("关注成功");
    }
    InfoStore.getUserDetail();
  } catch (error) {
    console.error("关注操作失败:", error);
  }
};

const fetchUserInfo = async () => {
  try {
    const userId = route.params.id;
    if (!userId) {
      ElMessage.error("用户ID缺失");
      return;
    }
    const res = await getInfoAPI({ id: userId });
    userInfo.value = res.data;
  } finally {
    loading.value = false;
  }
};

const fetchFollowStatus = async () => {
  if (!userStore.userInfo?.accessToken) return;

  try {
    const userId = route.params.id;
    if (!userId) return;

    const res = await checkFollowStatusAPI({ targetUserId: userId });
    isFollowed.value = res.data.isFollowing;
  } catch (error) {
    console.error("获取关注状态失败:", error);
  }
};

onMounted(() => {
  fetchUserInfo();
  fetchFollowStatus();
});
</script>

<template>
  <div class="user-detail">
    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="userInfo" class="content-wrapper">
      <div class="left-section">
        <div class="profile-header">
          <img :src="userInfo.avatarUrl" :alt="userInfo.nickname" class="avatar" />
          <div class="info">
            <div class="name-header">
              <h1>{{ userInfo.nickname || "未设置昵称" }}</h1>
              <!-- 只有当查看的不是自己的时候，才显示关注按钮 -->
              <div
                v-if="InfoStore.info?.id != route.params.id"
                class="follow-btn"
                :class="{ 'is-followed': isFollowed }"
                @click="toggleFollow"
              >
                {{ isFollowed ? "已关注" : "+ 关注" }}
              </div>
            </div>

            <p class="user-id">UID: {{ userInfo.uid || userInfo.id }}</p>

            <div class="stats">
              <span class="stat-item">
                <span class="count">{{ userInfo.followingCount || 0 }}</span>
                <span class="label">关注</span>
              </span>
              <span class="divider"></span>
              <span class="stat-item">
                <span class="count">{{ userInfo.followerCount || 0 }}</span>
                <span class="label">粉丝</span>
              </span>
            </div>

            <div class="bio-content">
              <h3>个人简介</h3>
              <p>{{ userInfo.bio || "暂无简介" }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="error">获取用户信息失败</div>
  </div>
</template>

<style scoped>
.user-detail {
  max-width: 1200px;
  margin: 40px auto;
  padding: 0 20px;
  font-family:
    -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
}

.loading,
.error {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
  font-size: 18px;
  color: #909399;
}

/* 左右布局容器 */
.content-wrapper {
  display: flex;
  justify-content: center; /* 居中显示，因为移除了右侧 */
  margin-bottom: 24px;
}

/* 左侧个人资料区域 */
.left-section {
  width: 100%;
  max-width: 800px; /* 限制最大宽度，使其在没有右侧栏时看起来更协调 */
}

/* 卡片通用样式 */
.profile-header {
  background: #ffffff;
  border-radius: 16px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.04);
  padding: 32px;
  transition: all 0.3s ease;
  border: 1px solid #f0f2f5;
  display: flex;
  align-items: flex-start;
  gap: 32px;
}

.profile-header:hover {
  transform: translateY(-2px);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.08);
}

.avatar {
  width: 140px;
  height: 140px;
  border-radius: 50%;
  object-fit: cover;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  flex-shrink: 0;
}

.info {
  flex: 1;
}

.info h1 {
  margin: 0;
  font-size: 32px;
  color: #303133;
  font-weight: 700;
  letter-spacing: -0.5px;
}

.name-header {
  display: flex;
  align-items: center;
  gap: 40px;
  margin-bottom: 12px;
}

.follow-btn {
  padding: 6px 18px;
  border-radius: 20px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  background: #ef93a2;
  color: #ffffff;
  border: 1px solid #ef93a2;
  user-select: none;
  display: flex;
  align-items: center;
  justify-content: center;
}

.follow-btn:hover {
  opacity: 0.9;
  transform: scale(1.02);
}

.follow-btn:active {
  transform: scale(0.98);
}

.follow-btn.is-followed {
  background: transparent;
  color: #ef93a2;
  border: 1px solid #ef93a2;
}

.user-id {
  font-size: 15px;
  color: #909399;
  margin: 5px 0 20px 0;
}

.stats {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}

.stat-item {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.count {
  font-size: 18px;
  font-weight: 700;
  color: #303133;
}

.label {
  font-size: 14px;
  color: #909399;
}

.divider {
  width: 1px;
  height: 14px;
  background-color: #dcdfe6;
}

.bio-content {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px dashed #e4e7ed;
}

.bio-content h3 {
  margin: 0 0 12px 0;
  font-size: 18px;
  color: #303133;
  font-weight: 700;
}

.bio-content p {
  line-height: 1.6;
  color: #606266;
  font-size: 15px;
  white-space: pre-line;
}

/* 响应式布局 */
@media (max-width: 768px) {
  .profile-header {
    flex-direction: column;
    align-items: center;
    text-align: center;
  }

  .name-header {
    flex-direction: column;
    gap: 16px;
    justify-content: center;
  }
}
</style>
