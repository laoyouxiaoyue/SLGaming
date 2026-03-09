<script setup>
import { ref, computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { getCompanionPublicProfileAPI } from "@/api/companion/companion.js";
import { createOrderAPI } from "@/api/order/order";
import { useWalletStore } from "@/stores/walletStore";
import { checkFollowStatusAPI, followUserAPI, unfollowUserAPI } from "@/api/relation";
import { useUserStore } from "@/stores/userStore";
import { useInfoStore } from "@/stores/infoStore";

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();
const loading = ref(true);
const ordering = ref(false);
const companionInfo = ref(null);
const walletStore = useWalletStore();
const isFollowed = ref(false);
const InfoStore = useInfoStore();

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

const createOrder = async () => {
  if (!companionInfo.value) return;
  if (InfoStore.info?.id === route.params.id) {
    ElMessage.error("自己不可以给自己下单哦");
    return;
  }

  const balance = Number(walletStore.walletInfo.balance || 0);
  const price = totalAmount.value;

  if (balance < price) {
    try {
      await ElMessageBox.confirm(
        `当前余额不足 (余额: ${balance} 帅币, 需支付: ${price} 帅币)，是否前往充值？`,
        "余额不足",
        {
          confirmButtonText: "去充值",
          cancelButtonText: "取消",
          type: "warning",
        },
      );
      router.push("/scion/recharge");
      return;
    } catch {
      return;
    }
  }

  // 2. 余额充足，二次确认
  try {
    await ElMessageBox.confirm(
      `当前余额: ${balance} 帅币\n本次支付: ${price} 帅币\n确认支付吗？`,
      "确认支付",
      {
        confirmButtonText: "确定支付",
        cancelButtonText: "取消",
        type: "success",
      },
    );
  } catch {
    return; // 用户取消支付
  }

  try {
    ordering.value = true;
    const data = {
      companionId: companionInfo.value.userId,
      gameName: companionInfo.value.gameSkill,
      durationHours: orderForm.value.durationHours,
    };
    await createOrderAPI(data);
    ElMessage.success({
      message: "支付成功！",
      duration: 1500,
    });
    // 可以跳转到订单详情页或订单列表页
    setTimeout(async () => {
      router.replace("/order/boss");
    }, 1200);
  } finally {
    ordering.value = false;
  }
};

onMounted(() => {
  fetchCompanionInfo();
  fetchFollowStatus();
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
            <div class="name-header">
              <h1>{{ companionInfo.nickname || "未设置昵称" }}</h1>
              <div
                v-if="InfoStore.info?.id != route.params.id"
                class="follow-btn"
                :class="{ 'is-followed': isFollowed }"
                @click="toggleFollow"
              >
                {{ isFollowed ? "已关注" : "+ 关注" }}
              </div>
            </div>
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
