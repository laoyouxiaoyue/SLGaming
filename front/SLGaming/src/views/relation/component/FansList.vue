<script setup>
import { ref, onMounted } from "vue";
import { getFollowersAPI, followUserAPI, unfollowUserAPI } from "@/api/relation";
import { ElMessage, ElMessageBox } from "element-plus";
import { Menu } from "@element-plus/icons-vue";

const followers = ref([]);
const loading = ref(false);
// eslint-disable-next-line no-unused-vars
const total = ref(0);
const params = ref({
  page: 1,
  pageSize: 20,
});

const fetchFollowers = async () => {
  loading.value = true;
  try {
    //   // 模拟数据 仅供测试
    //   await new Promise((resolve) => setTimeout(resolve, 500));
    //   const mockData = [
    //     {
    //       userId: 1,
    //       nickname: "测试粉丝1",
    //       avatarUrl: "https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png",
    //       isMutual: false,
    //     },
    //     {
    //       userId: 2,
    //       nickname: "互关大神",
    //       avatarUrl: "https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png",
    //       isMutual: true,
    //     },
    //     {
    //       userId: 3,
    //       nickname: "萌新小弟",
    //       avatarUrl: "https://cube.elemecdn.com/9/c2/f0ee8a3c7c9638a54940382568c9dpng.png",
    //       isMutual: false,
    //     },
    //   ];

    const res = await getFollowersAPI(params.value);
    followers.value = res.data.users || [];
    total.value = res.data.total;

    followers.value = mockData;
    total.value = mockData.length;
  } catch (error) {
    console.error("获取粉丝列表失败:", error);
  } finally {
    loading.value = false;
  }
};

const handleAction = async (user) => {
  if (user.isMutual) {
    try {
      await ElMessageBox.confirm(`确定取消关注 ${user.nickname}吗？`, "提示", {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
      });
      await unfollowUserAPI({ userId: user.userId });
      user.isMutual = false;
      ElMessage.success("已取消关注");
    } catch (error) {
      if (error !== "cancel") {
        console.error("取消关注失败:", error);
      }
    }
  } else {
    try {
      await followUserAPI({ userId: user.userId });
      user.isMutual = true;
      ElMessage.success("关注成功");
    } catch (error) {
      console.error("关注失败:", error);
    }
  }
};

onMounted(() => {
  fetchFollowers();
});
</script>

<template>
  <div class="fans-list" v-loading="loading">
    <div v-if="followers.length === 0 && !loading" class="empty-state">暂无粉丝</div>
    <div v-for="item in followers" :key="item.userId" class="follower-card">
      <div class="left-section">
        <el-avatar :size="60" :src="item.avatarUrl" />
      </div>
      <div class="right-section">
        <div class="nickname">{{ item.nickname }}</div>
        <div class="action-btn" :class="{ 'is-mutual': item.isMutual }" @click="handleAction(item)">
          <el-icon class="icon"><Menu /></el-icon>
          <span>{{ item.isMutual ? "已互关" : "回  关" }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.fans-list {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  padding: 16px;
}

@media screen and (max-width: 1200px) {
  .fans-list {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media screen and (max-width: 768px) {
  .fans-list {
    grid-template-columns: 1fr;
  }
}

.follower-card {
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
  background-color: #fff;
  color: #ef93a2;
  border: 1px solid #ef93a2;
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  user-select: none;
  transition: all 0.2s ease;
}

.action-btn:hover {
  background-color: #fff0f2;
}

.action-btn.is-mutual {
  background-color: #f4f4f5;
  color: #909399;
  border-color: transparent;
}

.action-btn.is-mutual:hover {
  background-color: #e9e9eb;
  color: #606266;
}

.icon {
  font-size: 14px;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: #909399;
}
</style>
