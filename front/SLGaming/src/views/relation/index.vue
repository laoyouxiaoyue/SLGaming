<script setup>
import { RouterView, RouterLink } from "vue-router";
import { useInfoStore } from "@/stores/infoStore";
import { storeToRefs } from "pinia";

const infoStore = useInfoStore();
const { info } = storeToRefs(infoStore);
</script>

<template>
  <div class="relation-container">
    <div class="sidebar">
      <div class="profile-summary">
        <el-avatar :size="70" :src="info.avatarUrl" class="user-avatar">
          <img src="https://cube.elemecdn.com/e/fd/0fc7d20532fdaf769a25683617711png.png" />
        </el-avatar>
        <div class="user-name">{{ info.nickname || "用户" }}</div>
        <div class="stats-row">
          <div class="stat-item">
            <span class="count">{{ info.followingCount || 0 }}</span>
            <span class="label">关注</span>
          </div>
          <div class="divider"></div>
          <div class="stat-item">
            <span class="count">{{ info.fansCount || 0 }}</span>
            <span class="label">粉丝</span>
          </div>
        </div>
      </div>

      <div class="nav-menu">
        <RouterLink to="/relation/follow" class="nav-item">
          <sl-icon name="icon-user-follow" size="20" />
          <span>我的关注</span>
        </RouterLink>
        <RouterLink to="/relation/fans" class="nav-item">
          <sl-icon name="icon-user-fans" size="20" />
          <span>我的粉丝</span>
        </RouterLink>
      </div>
    </div>

    <div class="main-content">
      <RouterView />
    </div>
  </div>
</template>

<style scoped lang="scss">
.relation-container {
  display: flex;
  width: 1200px;
  margin: 20px auto;
  gap: 20px;
  min-height: 700px;

  .sidebar {
    width: 280px;
    background: #fff;
    border-radius: 12px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
    padding: 40px 0;
    display: flex;
    flex-direction: column;

    .profile-summary {
      display: flex;
      flex-direction: column;
      align-items: center;
      padding-bottom: 30px;
      margin-bottom: 20px;
      border-bottom: 1px solid #f5f7fa;

      .user-avatar {
        margin-bottom: 16px;
        border: 3px solid #f0f2f5;
        transition: transform 0.3s;
        &:hover {
          transform: rotate(5deg) scale(1.05);
        }
      }

      .user-name {
        font-size: 18px;
        font-weight: 700;
        color: #1a1a1a;
        margin-bottom: 20px;
      }

      .stats-row {
        display: flex;
        width: 80%;
        background: linear-gradient(135deg, #fff6f2, #fff0ea);
        border-radius: 10px;
        padding: 12px 0;

        .stat-item {
          flex: 1;
          display: flex;
          flex-direction: column;
          align-items: center;

          .count {
            font-size: 16px;
            font-weight: 700;
            color: #ff6b35;
          }

          .label {
            font-size: 12px;
            color: #666;
            margin-top: 2px;
          }
        }

        .divider {
          width: 1px;
          background-color: #ffd6c5;
          height: 24px;
          align-self: center;
        }
      }
    }

    .nav-menu {
      padding: 0 20px;

      .nav-item {
        display: flex;
        align-items: center;
        height: 56px;
        padding: 0 24px;
        margin-bottom: 10px;
        color: #555;
        text-decoration: none;
        font-size: 15px;
        font-weight: 500;
        border-radius: 10px;
        transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
        gap: 14px;

        &:hover {
          color: #ff6b35;
          background-color: #fff6f2;
        }

        &.router-link-active {
          color: #fff;
          background: linear-gradient(135deg, #ff9ca4, #ff6b35);
          box-shadow: 0 4px 12px rgba(255, 107, 53, 0.35);
        }
      }
    }
  }
}

.main-content {
  flex: 1;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
  padding: 32px;
  min-height: 700px;
}
</style>
