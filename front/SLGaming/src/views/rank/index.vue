<script setup>
import { ref, onMounted } from "vue";
import { ElMessage } from "element-plus";
import { getCompanionOrdersRankingAPI, getCompanionRatingRankingAPI } from "@/api/rank";

const loading = ref(false);
const activeTab = ref("orders"); // orders: 接单榜, rating: 评分榜

const ordersRankings = ref([]);
const ratingRankings = ref([]);

const page = ref(1);
const pageSize = ref(10);
const total = ref(0);

const fetchOrdersRank = async () => {
  loading.value = true;
  try {
    const res = await getCompanionOrdersRankingAPI({ page: page.value, pageSize: pageSize.value });
    ordersRankings.value = res.data?.rankings || [];
    total.value = res.data?.total || 0;
  } catch (error) {
    console.error("获取接单排行榜失败", error);
    ElMessage.error("获取接单排行榜失败");
  } finally {
    loading.value = false;
  }
};

const fetchRatingRank = async () => {
  loading.value = true;
  try {
    const res = await getCompanionRatingRankingAPI({ page: page.value, pageSize: pageSize.value });
    ratingRankings.value = res.data?.rankings || [];
    total.value = res.data?.total || 0;
  } catch (error) {
    console.error("获取评分排行榜失败", error);
    ElMessage.error("获取评分排行榜失败");
  } finally {
    loading.value = false;
  }
};

const handleTabChange = (name) => {
  activeTab.value = name;
  page.value = 1;
  if (name === "orders") {
    fetchOrdersRank();
  } else {
    fetchRatingRank();
  }
};

const handlePageChange = (val) => {
  page.value = val;
  if (activeTab.value === "orders") {
    fetchOrdersRank();
  } else {
    fetchRatingRank();
  }
};

onMounted(() => {
  handleTabChange("orders");
});
</script>

<template>
  <div class="rank-page">
    <h1 class="page-title">
      <span class="page-title-text">陪玩排行榜</span>
      <span class="page-title-badge">RANK</span>
    </h1>

    <el-card class="rank-card" shadow="never">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="接单榜" name="orders" />
        <el-tab-pane label="评分榜" name="rating" />
      </el-tabs>

      <div v-loading="loading" class="rank-content">
        <template v-if="activeTab === 'orders'">
          <div v-if="ordersRankings.length === 0" class="empty">
            <el-empty description="暂无接单排行榜数据" />
          </div>
          <ul v-else class="rank-list">
            <li v-for="item in ordersRankings" :key="item.userId" class="rank-item">
              <div class="rank-num" :class="[`top-${item.rank}`]">
                {{ item.rank }}
              </div>
              <el-avatar :size="48" :src="item.avatarUrl" class="avatar" />
              <div class="info">
                <div class="name-line">
                  <span class="nickname">{{ item.nickname || "未设置昵称" }}</span>
                  <span v-if="item.isVerified" class="verified">已认证</span>
                </div>
                <div class="meta">
                  <span>总接单数：{{ item.totalOrders || 0 }}</span>
                  <span class="dot" />
                  <span>评分：{{ item.rating ?? "-" }}</span>
                </div>
              </div>
            </li>
          </ul>
        </template>

        <template v-else>
          <div v-if="ratingRankings.length === 0" class="empty">
            <el-empty description="暂无评分排行榜数据" />
          </div>
          <ul v-else class="rank-list">
            <li v-for="item in ratingRankings" :key="item.userId" class="rank-item">
              <div class="rank-num" :class="[`top-${item.rank}`]">
                {{ item.rank }}
              </div>
              <el-avatar :size="48" :src="item.avatarUrl" class="avatar" />
              <div class="info">
                <div class="name-line">
                  <span class="nickname">{{ item.nickname || "未设置昵称" }}</span>
                  <span v-if="item.isVerified" class="verified">已认证</span>
                </div>
                <div class="meta">
                  <span>评分：{{ item.rating ?? "-" }}</span>
                  <span class="dot" />
                  <span>总接单数：{{ item.totalOrders || 0 }}</span>
                </div>
              </div>
            </li>
          </ul>
        </template>

        <div v-if="total > 0" class="pagination-wrapper">
          <el-pagination
            background
            layout="prev, pager, next"
            :total="total"
            :page-size="pageSize"
            :current-page="page"
            @current-change="handlePageChange"
          />
        </div>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.rank-page {
  max-width: 1000px;
  margin: 40px auto;
  padding: 0 20px;
}

.page-title {
  font-size: 28px;
  font-weight: 800;
  margin-bottom: 22px;
  color: #303133;
  display: flex;
  align-items: center;
  gap: 10px;
  position: relative;
}

.page-title::after {
  content: "";
  position: absolute;
  left: 0;
  bottom: -6px;
  width: 90px;
  height: 4px;
  border-radius: 999px;
  background: linear-gradient(90deg, #ff9a9e, #fad0c4);
}

.page-title-text {
  letter-spacing: 1px;
}

.page-title-badge {
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 999px;
  background: linear-gradient(135deg, #ff9a9e, #fecfef);
  color: #ffffff;
  font-weight: 600;
  box-shadow: 0 2px 6px rgba(255, 154, 158, 0.45);
}

.rank-card {
  border-radius: 16px;
}

.rank-content {
  padding-top: 12px;
}

.rank-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.rank-item {
  display: flex;
  align-items: center;
  padding: 14px 8px;
  border-bottom: 1px solid #f0f2f5;
}

.rank-item:last-child {
  border-bottom: none;
}

.rank-num {
  width: 32px;
  text-align: center;
  font-size: 18px;
  font-weight: 600;
  margin-right: 12px;
  color: #606266;
}

.rank-num.top-1 {
  color: #f56c6c;
}

.rank-num.top-2 {
  color: #e6a23c;
}

.rank-num.top-3 {
  color: #409eff;
}

.avatar {
  margin-right: 12px;
}

.info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.name-line {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.nickname {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.verified {
  font-size: 12px;
  color: #67c23a;
  padding: 2px 6px;
  border-radius: 10px;
  background-color: #f0f9eb;
}

.meta {
  font-size: 13px;
  color: #909399;
  display: flex;
  align-items: center;
  gap: 6px;
}

.dot {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background-color: #dcdfe6;
}

.empty {
  padding: 40px 0;
}

.pagination-wrapper {
  margin-top: 16px;
  display: flex;
  justify-content: center;
}

@media (max-width: 768px) {
  .rank-page {
    margin: 20px auto;
  }

  .rank-item {
    padding: 10px 4px;
  }
}
</style>
