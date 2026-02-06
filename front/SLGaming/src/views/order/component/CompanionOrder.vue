<script setup>
import { ref, computed, onMounted, watch } from "vue";
import { useInfoStore } from "@/stores/infoStore";
import { getOrderListAPI } from "@/api/order/order";

const infoStore = useInfoStore();
const activeRole = ref("companion");
const activeStatus = ref("all");
const loading = ref(false);
const orders = ref([]);

const statusOptions = [
  { label: "全部", value: "all" },
  { label: "待支付", value: "1" },
  { label: "待接单", value: "2" },
  { label: "已接单", value: "3" },
  { label: "服务中", value: "4" },
  { label: "已完成", value: "5" },
  { label: "已取消", value: "6" },
  { label: "已评价", value: "7" },
];

// 从 Store 获取角色：1=老板, 2=陪玩
const userRole = computed(() => infoStore.info.role ?? 1);
const roleTabsVisible = computed(() => userRole.value === 2);
const queryStatus = computed(() =>
  activeStatus.value === "all" ? undefined : Number(activeStatus.value),
);

// 格式化时间
const formatTime = (timestamp) => {
  if (!timestamp) return "-";
  return new Date(timestamp).toLocaleString();
};

// 格式化时长 (分钟 -> 小时/分)
const formatDuration = (minutes) => {
  if (!minutes) return "-";
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return h > 0 ? `${h}小时${m > 0 ? ` ${m}分` : ""}` : `${m}分`;
};

// 获取状态文本
const getStatusText = (status) => {
  const option = statusOptions.find((opt) => opt.value == status);
  return option ? option.label : "未知状态";
};

// 获取状态对应 Tag 类型
const getStatusType = (status) => {
  // 1=CREATED, 2=PAID, 3=ACCEPTED, 4=IN_SERVICE, 5=COMPLETED, 6=CANCELLED, 7=RATED
  const map = {
    1: "info",
    2: "warning",
    3: "primary",
    4: "success",
    5: "success",
    6: "danger",
    7: "warning",
  };
  return map[status] || "";
};

// 按钮操作逻辑 (需对接具体API)
const handleAccept = (order) => console.log("Accept", order.id);
const handleStart = (order) => console.log("Start", order.id);
const handleComplete = (order) => console.log("Complete", order.id);

const loadOrders = async () => {
  loading.value = true;
  try {
    const res = await getOrderListAPI({
      role: activeRole.value,
      status: queryStatus.value,
      page: 1,
      pageSize: 10,
    });
    orders.value = res?.data?.orders || [];
  } finally {
    loading.value = false;
  }
};

const initData = async () => {
  // 如果 Store 里没数据，则请求一次
  if (!infoStore.info.id) {
    await infoStore.getUserDetail();
  }
  await loadOrders();
};

onMounted(() => {
  initData();
});

watch([activeStatus], () => {
  loadOrders();
});
</script>

<template>
  <div class="setting-info">
    <div class="setting-content">
      <!-- 订单状态筛选 -->
      <el-tabs v-model="activeStatus" class="status-tabs">
        <el-tab-pane
          v-for="item in statusOptions"
          :key="item.value"
          :label="item.label"
          :name="item.value"
        />
      </el-tabs>

      <el-skeleton v-if="loading" animated :rows="4" />
      <template v-else>
        <el-empty v-if="orders.length === 0" description="暂无订单" />
        <div v-else class="order-list">
          <el-card v-for="item in orders" :key="item.id" shadow="hover" class="order-card">
            <template #header>
              <div class="card-header">
                <span>订单号：{{ item.orderNo }}</span>
                <el-tag :type="getStatusType(item.status)">
                  {{ getStatusText(item.status) }}
                </el-tag>
              </div>
            </template>

            <el-descriptions :column="2" border size="small">
              <el-descriptions-item label="游戏名称">
                {{ item.gameName }}
              </el-descriptions-item>
              <el-descriptions-item label="时长">
                {{ formatDuration(item.duration) }}
              </el-descriptions-item>
              <el-descriptions-item label="单价">
                {{ item.pricePerHour }} 帅币/小时
              </el-descriptions-item>
              <el-descriptions-item label="总价">
                {{ item.totalAmount }} 帅币
              </el-descriptions-item>
              <el-descriptions-item label="创建时间" :span="2">
                {{ formatTime(item.createdAt) }}
              </el-descriptions-item>

              <!-- 动态状态字段 -->
              <el-descriptions-item v-if="item.paidAt" label="支付时间" :span="2">
                {{ formatTime(item.paidAt) }}
              </el-descriptions-item>
              <el-descriptions-item v-if="item.acceptedAt" label="接单时间" :span="2">
                {{ formatTime(item.acceptedAt) }}
              </el-descriptions-item>
              <el-descriptions-item v-if="item.startAt" label="开始服务" :span="2">
                {{ formatTime(item.startAt) }}
              </el-descriptions-item>
              <el-descriptions-item v-if="item.completedAt" label="完成时间" :span="2">
                {{ formatTime(item.completedAt) }}
              </el-descriptions-item>

              <el-descriptions-item v-if="item.status === 6" label="取消时间">
                {{ formatTime(item.cancelledAt) }}
              </el-descriptions-item>
              <el-descriptions-item v-if="item.status === 6" label="取消原因">
                {{ item.cancelReason }}
              </el-descriptions-item>

              <el-descriptions-item v-if="item.status === 7" label="评分">
                <el-rate v-model="item.rating" disabled show-score text-color="#ff9900" />
              </el-descriptions-item>
              <el-descriptions-item v-if="item.status === 7" label="评价" :span="2">
                {{ item.comment }}
              </el-descriptions-item>
            </el-descriptions>

            <div class="card-footer">
              <el-button
                v-if="item.status === 2"
                type="primary"
                size="small"
                @click="handleAccept(item)"
              >
                接单
              </el-button>
              <el-button
                v-if="item.status === 3"
                type="success"
                size="small"
                @click="handleStart(item)"
              >
                开始服务
              </el-button>
              <el-button
                v-if="item.status === 4"
                type="warning"
                size="small"
                @click="handleComplete(item)"
              >
                结束服务
              </el-button>
            </div>
          </el-card>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped lang="scss">
.setting-info {
  padding: 0 10px;

  .setting-content {
    padding-top: 20px;
  }

  .role-tabs,
  .status-tabs {
    margin-bottom: 12px;
  }

  .order-list {
    display: grid;
    gap: 12px;
  }

  .order-card {
    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
    .card-footer {
      margin-top: 10px;
      display: flex;
      justify-content: flex-end;
      gap: 10px;
    }
  }
}
</style>
