<script setup>
import { onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { CircleCheckFilled, WarningFilled, Loading } from "@element-plus/icons-vue";
import { queryrechargeorderapi } from "@/api/money/wallet";

const route = useRoute();
const router = useRouter();
const status = ref("checking"); // checking, success, fail, pending
const orderInfo = ref(null);

const checkOrderStatus = async (orderNo) => {
  try {
    const res = await queryrechargeorderapi(orderNo);
    if (res.code === 0 && res.data) {
      orderInfo.value = res.data;
      // 0=待支付, 1=成功, 2=失败, 3=关闭
      if (res.data.status === 1) {
        status.value = "success";
      } else if (res.data.status === 2 || res.data.status === 3) {
        status.value = "fail";
      } else {
        status.value = "pending";
      }
    } else {
      status.value = "fail";
    }
  } catch (error) {
    status.value = "fail";
    console.error(error);
  }
};

onMounted(() => {
  const { out_trade_no } = route.query;

  if (out_trade_no) {
    // 稍微延迟一下查询，等待后端可能的一步通知处理
    setTimeout(() => {
      checkOrderStatus(out_trade_no);
    }, 1000);
  } else {
    status.value = "fail";
  }
});

const goWallet = () => {
  router.replace("/account/wallet");
};

const retryQuery = () => {
  const { out_trade_no } = route.query;
  if (out_trade_no) {
    status.value = "checking";
    checkOrderStatus(out_trade_no);
  }
};
</script>

<template>
  <div class="pay-result-container">
    <div class="result-card">
      <!-- 加载中 -->
      <div v-if="status === 'checking'" class="status-icon-wrapper">
        <el-icon class="icon-loading is-loading"><Loading /></el-icon>
      </div>

      <!-- 成功 -->
      <div v-else-if="status === 'success'" class="status-icon-wrapper">
        <el-icon class="icon-success"><CircleCheckFilled /></el-icon>
      </div>

      <!-- 待处理/处理中 -->
      <div v-else-if="status === 'pending'" class="status-icon-wrapper">
        <el-icon class="icon-pending"><Loading /></el-icon>
      </div>

      <!-- 失败 -->
      <div v-else class="status-icon-wrapper">
        <el-icon class="icon-fail"><WarningFilled /></el-icon>
      </div>

      <h2 class="title" v-if="status === 'checking'">正在查询支付结果...</h2>
      <h2 class="title" v-else-if="status === 'success'">支付成功</h2>
      <h2 class="title" v-else-if="status === 'pending'">订单支付确认中</h2>
      <h2 class="title" v-else>未支付或支付失败</h2>

      <!-- 描述信息 -->
      <div class="desc-area" v-if="status === 'success'">
        <p class="amount" v-if="orderInfo">
          已成功充值 <span>{{ orderInfo.amount / 100 }}</span> 帅币
        </p>
        <p class="sub-desc">您现在可以前往查看钱包余额</p>
      </div>

      <div class="desc-area" v-if="status === 'pending'">
        <p class="sub-desc">系统暂未收到充值成功通知，如果您已完成付款，请稍后再次刷新查询。</p>
      </div>

      <div class="desc-area" v-if="status === 'fail'">
        <p class="sub-desc">我们未检测到该订单的成功支付记录，或订单已失效。</p>
      </div>

      <div class="actions">
        <el-button type="primary" @click="goWallet" v-if="status === 'success'">查看钱包</el-button>
        <el-button type="primary" @click="retryQuery" v-if="status === 'pending'"
          >刷新状态</el-button
        >
        <el-button @click="router.push('/scion/recharge')">{{
          status === "success" ? "继续充值" : "重新充值"
        }}</el-button>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.pay-result-container {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding-top: 50px;
  min-height: 600px;
}

.result-card {
  background: #fff;
  padding: 60px 100px;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.05);
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;

  .status-icon-wrapper {
    margin-bottom: 20px;

    .el-icon {
      font-size: 80px;
    }

    .icon-success {
      color: #67c23a;
    }

    .icon-fail {
      color: #f56c6c;
    }

    .icon-loading,
    .icon-pending {
      color: #409eff;
    }
  }

  .title {
    font-size: 24px;
    font-weight: 600;
    color: #303133;
    margin-bottom: 20px;
  }

  .desc-area {
    margin-bottom: 40px;

    .amount {
      font-size: 18px;
      color: #333;
      margin-bottom: 10px;

      span {
        font-size: 24px;
        color: #ff6b35;
        font-weight: bold;
        margin: 0 4px;
      }
    }

    .sub-desc {
      font-size: 14px;
      color: #909399;
      line-height: 1.6;
    }
  }

  .actions {
    display: flex;
    gap: 20px;
  }
}
</style>
