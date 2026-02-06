<script setup>
import { ref, computed, onMounted } from "vue";
import { useRoute } from "vue-router";
import { ElMessage } from "element-plus";
import { getCompanionPublicProfileAPI } from "@/api/companion/companion.js";
import { createOrderAPI } from "@/api/order/manage/create.js";

const route = useRoute();
const loading = ref(true);
const ordering = ref(false);
const companionInfo = ref(null);

const statusText = {
  0: "离线",
  1: "在线",
  2: "忙碌",
};

const orderForm = ref({
  durationHours: 1,
});

const totalAmount = computed(() => {
  if (!companionInfo.value) return 0;
  return companionInfo.value.pricePerHour * orderForm.value.durationHours;
});

const fetchCompanionInfo = async () => {
  try {
    const userId = route.params.id;
    if (!userId) {
      ElMessage.error("用户ID缺失");
      return;
    }
    const res = await getCompanionPublicProfileAPI({ userId });
    companionInfo.value = res.data;
  } catch (error) {
    console.error("获取陪玩信息失败:", error);
    ElMessage.error("获取陪玩信息失败");
  } finally {
    loading.value = false;
  }
};

const createOrder = async () => {
  if (!companionInfo.value) return;

  try {
    ordering.value = true;
    const data = {
      companionId: String(companionInfo.value.userId),
      gameName: companionInfo.value.gameSkill,
      durationHours: orderForm.value.durationHours,
    };
    const res = await createOrderAPI(data);
    ElMessage.success("订单创建成功");
    console.log("订单信息:", res.data);
    // 可以跳转到订单详情页或订单列表页
  } catch (error) {
    console.error("创建订单失败:", error);
  } finally {
    ordering.value = false;
  }
};

onMounted(() => {
  fetchCompanionInfo();
});
</script>
<template>
  <div class="companion-detail">
    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="companionInfo" class="content-wrapper">
      <div class="left-section">
        <div class="profile-header">
          <img :src="companionInfo.avatarUrl" :alt="companionInfo.nickname" class="avatar" />
          <div class="info">
            <h1>{{ companionInfo.nickname || "未设置昵称" }}</h1>
            <p class="game-skill">{{ companionInfo.gameSkill }}</p>
            <div class="status-rating">
              <span :class="['status', `status-${companionInfo.status}`]">
                {{ statusText[companionInfo.status] }}
              </span>
              <span class="rating">评分: {{ companionInfo.rating }}/5</span>
            </div>
            <p class="price">每小时价格: {{ companionInfo.pricePerHour }} 帅币</p>
            <p class="orders">总接单数: {{ companionInfo.totalOrders }}</p>
            <p v-if="companionInfo.isVerified" class="verified">✓ 已认证</p>
            <div class="bio-content">
              <h3>个人简介</h3>
              <p>{{ companionInfo.bio || "暂无简介" }}</p>
            </div>
          </div>
        </div>
      </div>

      <div class="right-section">
        <div class="order-section">
          <h3>下单服务</h3>
          <div class="order-panel">
            <el-form :model="orderForm" label-position="top">
              <el-form-item label="选择服务时长">
                <div class="duration-selector">
                  <el-input-number
                    v-model="orderForm.durationHours"
                    :min="1"
                    :max="24"
                    size="large"
                  />
                  <span class="unit">小时</span>
                </div>
              </el-form-item>
            </el-form>

            <div class="order-footer">
              <div class="price-summary">
                <span class="label">总计费用</span>
                <div class="amount-group">
                  <span class="number">{{ totalAmount }}</span>
                  <span class="currency">帅币</span>
                </div>
              </div>
              <el-button
                type="primary"
                size="large"
                class="submit-btn"
                @click="createOrder"
                :loading="ordering"
              >
                立即下单
              </el-button>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="error">获取陪玩信息失败</div>
  </div>
</template>

<style scoped>
.companion-detail {
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
  gap: 40px;
  margin-bottom: 24px;
}

/* 左侧个人资料区域 */
.left-section {
  flex: 2;
  min-width: 0;
}

/* 右侧下单服务区域 */
.right-section {
  width: 380px;
  flex-shrink: 0;
}

/* 卡片通用样式 */
.profile-header,
.order-section {
  background: #ffffff;
  border-radius: 16px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.04);
  padding: 32px;
  transition: all 0.3s ease;
  border: 1px solid #f0f2f5;
}

.profile-header:hover,
.order-section:hover {
  transform: translateY(-2px);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.08);
}

.profile-header {
  display: flex;
  align-items: flex-start;
  gap: 32px;
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
  margin: 0 0 12px 0;
  font-size: 32px;
  color: #303133;
  font-weight: 700;
  letter-spacing: -0.5px;
}

.game-skill {
  display: inline-block;
  font-size: 16px;
  color: #409eff;
  background: #ecf5ff;
  padding: 6px 16px;
  border-radius: 20px;
  margin: 0 0 20px 0;
  font-weight: 600;
}

.status-rating {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
  background: #f8f9fa;
  padding: 12px 20px;
  border-radius: 12px;
  width: fit-content;
}

.status {
  padding: 4px 12px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  display: flex;
  align-items: center;
}

.status::before {
  content: "";
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 8px;
  background: currentColor;
}

.status-0 {
  background: #f4f4f5;
  color: #909399;
}
.status-1 {
  background: #f0f9eb;
  color: #67c23a;
}
.status-2 {
  background: #fdf6ec;
  color: #e6a23c;
}

.rating {
  font-size: 16px;
  color: #606266;
  font-weight: 500;
  display: flex;
  align-items: center;
}

.rating::before {
  content: "★";
  color: #f39c12;
  margin-right: 6px;
  font-size: 18px;
}

.price,
.orders {
  font-size: 16px;
  color: #606266;
  margin: 8px 0;
  display: flex;
  align-items: center;
}

.price::before {
  content: "💰";
  margin-right: 10px;
  font-size: 18px;
}

.orders::before {
  content: "📦";
  margin-right: 10px;
  font-size: 18px;
}

.verified {
  color: #67c23a;
  font-weight: 600;
  margin-top: 16px;
  display: flex;
  align-items: center;
  font-size: 15px;
  background: #f0f9eb;
  padding: 8px 16px;
  border-radius: 8px;
  display: inline-block;
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

.order-section h3 {
  margin: 0 0 24px 0;
  font-size: 20px;
  color: #303133;
  font-weight: 700;
  border-left: 5px solid #409eff;
  padding-left: 16px;
}

.order-panel {
  width: 100%;
}

.duration-selector {
  display: flex;
  align-items: center;
  gap: 12px;
}

.duration-selector .unit {
  font-size: 16px;
  color: #606266;
}

.order-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px dashed #e4e7ed;
}

.price-summary {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.price-summary .label {
  font-size: 14px;
  color: #909399;
}

.amount-group {
  display: flex;
  align-items: baseline;
  gap: 4px;
  color: #f56c6c;
}

.amount-group .number {
  font-size: 32px;
  font-weight: 700;
  line-height: 1;
}

.amount-group .currency {
  font-size: 14px;
  font-weight: 500;
}

.submit-btn {
  padding: 12px 40px;
  font-size: 16px;
  border-radius: 8px;
  height: auto;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
  transition: all 0.3s;
}

.submit-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 16px rgba(64, 158, 255, 0.4);
}

/* 响应式布局 */
@media (max-width: 768px) {
  .content-wrapper {
    flex-direction: column;
  }

  .right-section {
    width: 100%;
  }
}
</style>
