<script setup>
import { onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { CircleCheckFilled, WarningFilled } from "@element-plus/icons-vue";

const route = useRoute();
const router = useRouter();
const status = ref("waiting"); // waiting, success, fail

onMounted(() => {
  // 简单的参数检查，Alipay PC支付同步回调通常会带有 out_trade_no 等参数
  // 注意：前端的状态仅供展示，实际资金到账以服务端异步通知为准
  const { out_trade_no } = route.query;

  if (out_trade_no) {
    status.value = "success";
  } else {
    // 如果没有参数，可能是用户直接访问，或者支付取消等
    status.value = "fail";
  }
});

const goWallet = () => {
  router.replace("/account/wallet");
};
</script>

<template>
  <div class="pay-result-container">
    <div class="result-card">
      <div v-if="status === 'success'" class="status-icon-wrapper">
        <el-icon class="icon-success"><CircleCheckFilled /></el-icon>
      </div>
      <div v-else class="status-icon-wrapper">
        <el-icon class="icon-fail"><WarningFilled /></el-icon>
      </div>

      <h2 class="title">
        {{ status === "success" ? "支付提交成功" : "未检测到支付结果" }}
      </h2>
      <p class="desc" v-if="status === 'success'">
        您的充值订单已提交，系统正在确认金额到账情况。<br />
        通常会在 1-3 分钟内到账，请留意钱包余额变化。
      </p>
      <p class="desc" v-else>请确认您是否已完成支付，或尝试重新发起订单。</p>

      <div class="actions">
        <el-button type="primary" @click="goWallet">查看钱包</el-button>
        <el-button @click="router.push('/scion/recharge')">继续充值</el-button>
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
      color: #e6a23c;
    }
  }

  .title {
    font-size: 24px;
    font-weight: 600;
    color: #303133;
    margin-bottom: 15px;
  }

  .desc {
    font-size: 14px;
    color: #909399;
    line-height: 1.6;
    margin-bottom: 40px;
  }

  .actions {
    display: flex;
    gap: 20px;
  }
}
</style>
