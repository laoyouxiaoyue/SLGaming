<script setup>
import { ref, onMounted } from "vue";
import { getFollowingsAPI, followUserAPI, unfollowUserAPI } from "@/api/relation";
import { ElMessage, ElMessageBox } from "element-plus";
import { Menu } from "@element-plus/icons-vue";

const followings = ref([]);
const loading = ref(false);
// eslint-disable-next-line no-unused-vars
const total = ref(0);
const params = ref({
  page: 1,
  pageSize: 20,
});

const fetchFollowings = async () => {
  loading.value = true;
  try {
    const res = await getFollowingsAPI(params.value);
    followings.value = (res.data.users || []).map((u) => ({ ...u, isFollowed: true }));
    total.value = res.data.total;
  } catch (error) {
    console.error("获取关注列表失败:", error);
  } finally {
    loading.value = false;
  }
};

const handleAction = async (user) => {
  if (user.isFollowed) {
    try {
      await ElMessageBox.confirm(`确定取消关注 ${user.nickname} 吗？`, "提示", {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
      });
      await unfollowUserAPI({ userId: user.userId });
      user.isFollowed = false;
      ElMessage.success("已取消关注");
    } catch (error) {
      if (error !== "cancel") {
        console.error("取消关注失败:", error);
      }
    }
  } else {
    try {
      await followUserAPI({ userId: user.userId });
      user.isFollowed = true;
      ElMessage.success("关注成功");
    } catch (error) {
      console.error("关注失败:", error);
    }
  }
};

onMounted(() => {
  fetchFollowings();
});
</script>

<template>
  <div class="setting-info">
    <div class="panel-title">我的关注</div>
    <div class="follow-list" v-loading="loading">
      <div v-if="followings.length === 0 && !loading" class="empty-container">
        <el-empty description="暂无关注" />
      </div>
      <div v-for="item in followings" :key="item.userId" class="follow-card">
        <div class="left-section">
          <el-avatar :size="60" :src="item.avatarUrl" />
        </div>
        <div class="right-section">
          <div class="nickname">{{ item.nickname }}</div>
          <div class="action-btn" @click="handleAction(item)">
            <el-icon class="icon"><Menu /></el-icon>
            <span>{{ !item.isFollowed ? "关注" : item.isMutual ? "已互关" : "已关注" }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.setting-info {
  padding: 0 10px;

  .panel-title {
    font-size: 20px;
    font-weight: 600;
    margin-bottom: 25px;
    color: #333;
    border-left: 4px solid #ff6b35;
    padding-left: 12px;
  }
}

.follow-list {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  padding: 16px;
}

.empty-container {
  grid-column: 1 / -1;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 300px;
}

@media screen and (max-width: 1200px) {
  .follow-list {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media screen and (max-width: 768px) {
  .follow-list {
    grid-template-columns: 1fr;
  }
}

.follow-card {
  display: flex;
  align-items: center;
  padding: 16px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
  gap: 16px;
}

.left-section {
  flex-shrink: 0;
}

.right-section {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 8px;
  flex: 1;
}

.nickname {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  width: fit-content;
  padding: 6px 16px;
  background-color: #f4f4f5;
  color: #909399;
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  user-select: none;
  transition: all 0.2s ease;
}

.action-btn:hover {
  background-color: #e9e9eb;
  color: #606266;
}

.icon {
  font-size: 14px;
}
</style>
