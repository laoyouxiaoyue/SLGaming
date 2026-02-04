<template>
  <div class="companion-detail">
    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="companionInfo">
      <div class="profile-header">
        <img
          :src="companionInfo.avatarUrl || defaultAvatar"
          :alt="companionInfo.nickname"
          class="avatar"
        />
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
        </div>
      </div>

      <div class="bio-section">
        <h3>个人简介</h3>
        <p>{{ companionInfo.bio || "暂无简介" }}</p>
      </div>

      <div class="order-section">
        <h3>下单服务</h3>
        <el-form :model="orderForm" label-width="120px">
          <el-form-item label="服务时长（小时）">
            <el-input-number v-model="orderForm.durationHours" :min="1" :max="24" label="小时" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="createOrder" :loading="ordering">
              下单 (总价: {{ totalAmount }} 帅币)
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
    <div v-else class="error">获取陪玩信息失败</div>
  </div>
</template>

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
const defaultAvatar = "https://api.dicebear.com/7.x/avataaars/svg?seed=default";

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
      companionId: companionInfo.value.userId,
      gameName: companionInfo.value.gameSkill,
      durationMinutes: orderForm.value.durationHours * 60,
    };
    const res = await createOrderAPI(data);
    ElMessage.success("订单创建成功");
    console.log("订单信息:", res.data);
    // 可以跳转到订单详情页或订单列表页
  } catch (error) {
    console.error("创建订单失败:", error);
    ElMessage.error("创建订单失败");
  } finally {
    ordering.value = false;
  }
};

onMounted(() => {
  fetchCompanionInfo();
});
</script>

<style scoped>
.companion-detail {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.loading,
.error {
  text-align: center;
  padding: 50px;
  font-size: 18px;
}

.profile-header {
  display: flex;
  align-items: center;
  margin-bottom: 30px;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.avatar {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  margin-right: 20px;
}

.info h1 {
  margin: 0 0 10px 0;
  font-size: 24px;
}

.game-skill {
  font-size: 18px;
  color: #666;
  margin: 5px 0;
}

.status-rating {
  display: flex;
  gap: 15px;
  margin: 10px 0;
}

.status {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 14px;
  font-weight: bold;
}

.status-0 {
  background: #ccc;
  color: #333;
}
.status-1 {
  background: #67c23a;
  color: white;
}
.status-2 {
  background: #e6a23c;
  color: white;
}

.rating {
  font-size: 16px;
  color: #f39c12;
}

.price,
.orders {
  font-size: 16px;
  margin: 5px 0;
}

.verified {
  color: #67c23a;
  font-weight: bold;
}

.bio-section,
.order-section {
  margin-bottom: 30px;
}

.bio-section h3,
.order-section h3 {
  margin-bottom: 15px;
  font-size: 20px;
}

.order-section .el-form {
  max-width: 500px;
}
</style>
