<script setup>
import { ref, computed, onMounted, watch } from "vue";
import { useInfoStore } from "@/stores/infoStore";
import {
  getOrderListAPI,
  cancelOrderAPI,
  acceptOrderAPI,
  startOrderServiceAPI,
} from "@/api/order/order";
import { ElMessageBox, ElMessage } from "element-plus";
import "element-plus/theme-chalk/el-message-box.css";
const infoStore = useInfoStore();
const activeRole = ref("companion");
const activeStatus = ref("all");
const loading = ref(false);
const orders = ref([]);

const statusOptions = [
  { label: "全部", value: "all" },
  { label: "待接单", value: "1" },
  { label: "已接单", value: "3" },
  { label: "服务中", value: "4" },
  { label: "已完成", value: "5" },
  { label: "已取消", value: "6" },
  { label: "已评价", value: "7" },
];

// 从 Store 获取角色：1=老板, 2=陪玩
const queryStatus = computed(() =>
  activeStatus.value === "all" ? undefined : Number(activeStatus.value),
);

// 格式化时间
const formatTime = (timestamp) => {
  if (!timestamp) return "-";
  // 传入的是秒级时间戳，需要乘以1000转换为毫秒
  return new Date(Number(timestamp) * 1000).toLocaleString();
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
    1: "warning",
    3: "primary",
    4: "success",
    5: "success",
    6: "danger",
    7: "warning",
  };
  return map[status] || "";
};

// 取消按钮操作逻辑
const handleCancel = async (order) => {
  try {
    const { value } = await ElMessageBox.prompt("请输入取消原因", "确认取消订单", {
      confirmButtonText: "确认取消",
      cancelButtonText: "暂不取消",
      inputPlaceholder: "请输入取消原因（选填）",
    });

    await cancelOrderAPI({
      orderId: order.id,
      reason: value || "陪玩者取消订单",
    });

    ElMessage.success("订单取消成功");
    // 刷新列表
    loadOrders();
  } catch (error) {
    if (error !== "cancel") {
      console.error("取消订单失败:", error);
    }
  }
};
//接单按钮操作逻辑
const handleAccept = async (order) => {
  try {
    await ElMessageBox.confirm(`确认接单吗？<br>一旦确认不可以取消了哦`, "接单确认", {
      confirmButtonText: "确认接单",
      cancelButtonText: "取消",
      type: "warning",
      dangerouslyUseHTMLString: true,
    });

    await acceptOrderAPI({ orderId: order.id });
    ElMessage.success("接单成功！");
    loadOrders();
  } catch (error) {
    if (error !== "cancel") {
      console.error("接单失败:", error);
    }
  }
};
const handleStart = async (order) => {
  try {
    await ElMessageBox.confirm(
      `确认开始服务吗？<br>请确保您已经做好准备开始陪玩服务`,
      "开始服务确认",
      {
        confirmButtonText: "立即开始",
        cancelButtonText: "稍后",
        type: "success",
        dangerouslyUseHTMLString: true,
      },
    );

    await startOrderServiceAPI({ orderId: order.id });
    ElMessage.success("服务已开始！");
    loadOrders();
  } catch (error) {
    if (error !== "cancel") {
      console.error("开始服务失败:", error);
    }
  }
};

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
      <el-tabs v-model="activeStatus" class="status-tabs" type="card">
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

            <el-descriptions :column="2" border size="default">
              <el-descriptions-item label="游戏名称">
                {{ item.gameName }}
              </el-descriptions-item>
              <el-descriptions-item label="时长">
                {{ item.durationHours }} 小时
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
                v-if="item.status === 1"
                type="primary"
                size="small"
                @click="handleAccept(item)"
              >
                接单
              </el-button>
              <el-button
                v-if="item.status === 1"
                type="danger"
                size="small"
                @click="handleCancel(item)"
              >
                取消订单
              </el-button>
              <el-button
                v-if="item.status === 3"
                type="success"
                size="small"
                @click="handleStart(item)"
              >
                开始服务
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
    padding-bottom: 20px;
  }

  .role-tabs,
  .status-tabs {
    margin-bottom: 1px;
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
