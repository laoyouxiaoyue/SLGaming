<script setup>
import { RouterView, RouterLink } from "vue-router";
import { useInfoStore } from "@/stores/infoStore";
import { useWalletStore } from "@/stores/walletStore";
import { storeToRefs } from "pinia";

const infoStore = useInfoStore();
const walletStore = useWalletStore();
const { info } = storeToRefs(infoStore);
const { walletInfo } = storeToRefs(walletStore);
</script>

<template>
  <div class="recharge-container">
    <div class="sidebar">
      <div class="user-summary">
        <div class="user-info-row">
          <el-avatar :size="60" :src="info.avatarUrl" class="user-avatar">
            <img src="https://cube.elemecdn.com/e/fd/0fc7d20532fdaf769a25683617711png.png" />
          </el-avatar>
          <div class="user-name">{{ info.nickname || "用户" }}</div>
        </div>
        <div class="wallet-balance">
          <span class="label">帅币</span>
          <span class="amount">{{ walletInfo.balance || "0" }}</span>
        </div>
      </div>

      <div class="menu-list">
        <RouterLink to="/scion/recharge" class="menu-item">
          <sl-icon name="icon-chongzhi1" size="18" />
          <span>充值中心</span>
        </RouterLink>
        <RouterLink to="/scion/recond" class="menu-item">
          <sl-icon name="icon-dingdan" size="18" />
          <span>消费记录</span>
        </RouterLink>
      </div>
    </div>
    <div class="content-area">
      <RouterView />
    </div>
  </div>
</template>

<style scoped lang="scss">
.recharge-container {
  display: flex;
  width: 1200px;
  margin: 20px auto;
  gap: 20px;
  min-height: 600px;

  .sidebar {
    width: 260px;
    background: #fff;
    border-radius: 8px;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
    padding: 30px 0;
    display: flex;
    flex-direction: column;

    .user-summary {
      display: flex;
      flex-direction: column;
      align-items: center;
      padding-bottom: 25px;
      margin-bottom: 20px;
      border-bottom: 1px dashed #eee;

      .user-info-row {
        display: flex;
        align-items: center;
        width: 80%;
        margin-bottom: 20px;
        gap: 15px;

        .user-avatar {
          border: 2px solid #fff;
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
          flex-shrink: 0;
        }

        .user-name {
          font-size: 16px;
          font-weight: 600;
          color: #333;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
      }

      .wallet-balance {
        display: flex;
        flex-direction: row;
        align-items: center;
        justify-content: center;
        gap: 10px;
        background: linear-gradient(135deg, #fff6f2, #fff0ea);
        padding: 12px 10px;
        border-radius: 12px;
        width: 80%;

        .label {
          font-size: 16px;
          color: #666;
          margin-bottom: 0;
        }

        .amount {
          font-size: 18px;
          font-weight: bold;
          color: #0c0909;
        }
      }
    }

    .menu-list {
      display: flex;
      flex-direction: column;
      padding: 0 20px;

      .menu-item {
        display: flex;
        align-items: center;
        height: 54px;
        padding: 0 20px;
        margin-bottom: 8px;
        color: #555;
        text-decoration: none;
        font-size: 15px;
        border-radius: 8px;
        transition: all 0.3s;
        gap: 12px;

        &:hover {
          color: #ff6b35;
          background-color: #fff6f2;
        }

        &.router-link-active {
          color: #fff;
          background: linear-gradient(135deg, #ff9ca4, #ff6b35);
          box-shadow: 0 4px 12px rgba(255, 107, 53, 0.3);
        }
      }
    }
  }

  .content-area {
    flex: 1;
    background: #fff;
    border-radius: 8px;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
    padding: 30px;
    min-height: 600px;
  }
}
</style>
