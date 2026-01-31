<script setup>
import { RouterView, RouterLink } from "vue-router";
import { useInfoStore } from "@/stores/infoStore";
import { ref } from "vue";
const infoStore = useInfoStore();
const info = ref({});
info.value = infoStore.info;
</script>

<template>
  <div class="account-container">
    <div class="sidebar">
      <div class="menu-title">个人中心</div>
      <div class="menu-list">
        <RouterLink to="/account/setting" class="menu-item">
          <sl-icon name="icon-touxiang" size="16" />
          <span>我的信息</span>
        </RouterLink>
        <RouterLink to="/account/companion" class="menu-item" v-if="info.role === 2">
          <sl-icon name="icon-peiwandailian" size="16" />
          <span>陪玩设置</span>
        </RouterLink>
        <RouterLink to="/account/order" class="menu-item">
          <sl-icon name="icon-dingdan" size="16" />
          <span>我的订单</span>
        </RouterLink>
        <RouterLink to="/account/wallet" class="menu-item">
          <sl-icon name="icon-qianbao" size="16" />
          <span>我的钱包</span>
        </RouterLink>
      </div>
    </div>
    <div class="content-area">
      <RouterView />
    </div>
  </div>
</template>

<style scoped lang="scss">
.account-container {
  display: flex;
  width: 1200px;
  margin: 20px auto;
  gap: 20px;
  min-height: 600px;

  .sidebar {
    width: 220px;
    background: #fff;
    border-radius: 4px;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
    padding: 20px 0;

    .menu-title {
      font-size: 18px;
      padding: 0 24px 15px;
      border-bottom: 1px solid #f0f0f0;
      margin-bottom: 10px;
      color: #666565;
      text-align: center;
    }

    .menu-list {
      display: flex;
      flex-direction: column;
    }

    .menu-item {
      display: flex;
      align-items: center;
      justify-content: center;
      height: 50px;
      padding: 0 24px;
      color: #222;
      text-decoration: none;
      font-size: 14px;
      transition: all 0.3s;
      gap: 10px;

      &:hover {
        color: #409eff;
        background-color: #ecf5ff;
      }

      &.router-link-active {
        color: #409eff;
        background-color: #ecf5ff;
        border-right: 3px solid #409eff;
      }
    }
  }

  .content-area {
    flex: 1;
    background: #fff;
    border-radius: 4px;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
    padding: 24px;
    min-height: 600px;
  }
}
</style>
