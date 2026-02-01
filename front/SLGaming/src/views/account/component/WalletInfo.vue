<script setup>
import { onMounted } from "vue";
import { useWalletStore } from "@/stores/walletStore";
import { storeToRefs } from "pinia";
import { Money, Lock } from "@element-plus/icons-vue";

const walletStore = useWalletStore();
const { walletInfo } = storeToRefs(walletStore);

onMounted(() => {
  if (Object.keys(walletInfo.value).length === 0) {
    walletStore.getWallet();
  }
});
</script>

<template>
  <div class="setting-info">
    <div class="panel-title">我的钱包</div>
    <div class="setting-content">
      <el-row :gutter="20">
        <el-col :span="12">
          <el-card shadow="never" class="wallet-card balance-card">
            <template #header>
              <div class="card-header">
                <div class="header-left">
                  <el-icon class="icon-orange"><Money /></el-icon>
                  <span>账户余额</span>
                </div>
              </div>
            </template>
            <div class="card-body">
              <span class="currency">¥</span>
              <span class="amount">{{ walletInfo.balance || 0 }}</span>
              <el-button
                type="primary"
                round
                class="action-btn"
                @click="$router.push('/scion/recharge')"
                >充值</el-button
              >
            </div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card shadow="never" class="wallet-card frozen-card">
            <template #header>
              <div class="card-header">
                <div class="header-left">
                  <el-icon><Lock /></el-icon>
                  <span>冻结金额</span>
                </div>
              </div>
            </template>
            <div class="card-body">
              <span class="currency">¥</span>
              <span class="amount">{{ walletInfo.frozenBalance || 0 }}</span>
              <div class="desc">交易中的资金暂时冻结</div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<style scoped lang="scss">
.setting-info {
  padding: 0 10px;
}

.panel-title {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 25px;
  color: #333;
  border-left: 4px solid #ff6b35;
  padding-left: 12px;
}

.setting-content {
  padding-top: 10px;
}

.wallet-card {
  height: 100%;
  border-radius: 12px;
  transition: all 0.3s;
  border: 1px solid #eee;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(255, 107, 53, 0.08);
    border-color: #ffccb8;
  }
}

.balance-card :deep(.el-card__header) {
  background: linear-gradient(to right, #fff6f2, #ffffff);
  border-bottom: 1px solid #f9f9f9;
}

.frozen-card :deep(.el-card__header) {
  background: linear-gradient(to right, #f8f8f8, #ffffff);
  border-bottom: 1px solid #f9f9f9;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  color: #333;
  font-weight: 500;

  .icon-orange {
    color: #ff6b35;
  }
}

.card-body {
  padding: 20px 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: flex-start;
  position: relative;
}

.currency {
  font-size: 24px;
  font-weight: 600;
  color: #333;
  margin-right: 4px;
}

.amount {
  font-size: 48px;
  font-weight: 700;
  color: #333;
  font-family: DINAlternate-Bold, "Helvetica Neue", Helvetica, sans-serif;
  line-height: 1;
  margin-bottom: 20px;
}

.desc {
  font-size: 12px;
  color: #909399;
}

.action-btn {
  position: absolute;
  right: 0;
  bottom: 10px;
  padding: 8px 24px;
  background: linear-gradient(135deg, #ff8e61, #ff6b35);
  border: none;

  &:hover {
    background: linear-gradient(135deg, #ff9ca4, #ff7a45);
    opacity: 0.9;
  }
}
</style>
