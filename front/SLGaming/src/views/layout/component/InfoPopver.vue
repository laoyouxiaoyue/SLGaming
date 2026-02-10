<script setup>
import { useUserStore } from "@/stores/userStore";
import { useWalletStore } from "@/stores/walletStore";
import { useInfoStore } from "@/stores/infoStore";
import { useCompanionStore } from "@/stores/companionStore";
import { useRouter } from "vue-router";
import { storeToRefs } from "pinia";
import { computed, watch } from "vue";

const userStore = useUserStore();
const walletStore = useWalletStore();
const infoStore = useInfoStore();
const companionStore = useCompanionStore();
const router = useRouter();

// 使用 storeToRefs 获取响应式 state
const { info } = storeToRefs(infoStore);
const { walletInfo } = storeToRefs(walletStore);
const { companionInfo } = storeToRefs(companionStore);

// 用户状态样式映射
const statusClass = computed(() => {
  // 仅陪玩角色展示状态 (role: 1=老板, 2=陪玩)
  if (info.value.role !== 2) return "";
  const map = {
    0: "offline",
    1: "online",
    2: "busy",
  };
  // 优先使用 companionStore 中的状态，因为它是专门管理陪玩业务的
  return map[companionInfo.value.status] ?? map[info.value.status] ?? "offline";
});

const statusText = computed(() => {
  const map = {
    0: "离线",
    1: "在线",
    2: "忙碌",
  };
  return map[companionInfo.value.status] ?? "离线";
});

const handleLogout = async () => {
  // 1. 执行 Store 中的通用退出逻辑
  await userStore.logout();
  // 2. 跳转到登录页
  router.replace("/login");
};

// 监听角色变化，如果是陪玩则获取陪玩详情信息
watch(
  () => info.value.role,
  (newRole) => {
    if (newRole === 2) {
      companionStore.getCompanionDetail();
    }
  },
  { immediate: true },
);
</script>

<template>
  <el-popover placement="bottom" trigger="hover" width="200" popper-class="user-popover">
    <template #reference>
      <a href="javascript:;" class="avatar-link">
        <div
          class="avatar-box"
          v-if="info.avatarUrl"
          :style="{ backgroundImage: `url(${info.avatarUrl})` }"
        ></div>
        <sl-icon name="icon-touxiang1" v-else size="60" color="#fff" />
        <!-- 状态展示：小圆点 -->
        <div v-if="statusClass" class="status-tag" :class="statusClass"></div>
      </a>
    </template>
    <div class="user-popover-content">
      <div class="nickname">{{ info.nickname }}</div>
      <div class="wallet">帅币:{{ walletInfo?.balance || "0.00" }}</div>

      <div class="divider"></div>
      <!-- 陪玩专属信息：游戏技能与状态 -->
      <div
        v-if="info.role === 2"
        class="companion-info"
        @click="$router.push('/account/companion')"
      >
        <div class="info-row">
          <span class="value">{{ companionInfo.gameSkill || "" }}</span>
        </div>
        <div class="info-row">
          <span class="status-text" :class="statusClass">{{ statusText }}</span>
        </div>
      </div>

      <!-- 用户数据统计卡片 -->
      <div class="user-stats-card">
        <!-- 数据统计区 -->
        <div class="stats-grid">
          <div class="stat-item" @click="$router.push('/account/follows')">
            <div class="stat-num">{{ info.followingCount || 0 }}</div>
            <div class="stat-label">关注</div>
          </div>
          <div class="stat-divider"></div>
          <div class="stat-item" @click="$router.push('/account/fans')">
            <div class="stat-num">{{ info.followerCount || 0 }}</div>
            <div class="stat-label">粉丝</div>
          </div>
        </div>

        <!-- 底部渐变线 -->
        <div class="bottom-gradient"></div>
      </div>

      <div class="divider" v-if="info.role === 2"></div>
      <a href="javascript:;" class="menu-item" @click="$router.push('/account/setting')">
        <sl-icon name="iconfont icon-touxiang" size="16" color="#fff" />个人中心
      </a>
      <a href="javascript:;" class="menu-item logout" @click="handleLogout">
        <sl-icon name="icon-tuichu" size="16" color="#fff" />退出登录
      </a>
    </div>
  </el-popover>
</template>

<style scoped lang="scss">
.avatar-link {
  position: relative;
  width: 46px;
  height: 46px;
  border-radius: 50%;
  border: 2px solid #fff;
  background-color: rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0; // 清除默认 padding
  margin-right: 30px;

  .avatar-box {
    width: 100%;
    height: 100%;
    border-radius: 50%;
    background-size: cover;
    background-position: center;
  }

  .status-tag {
    position: absolute;
    top: 31px;
    right: -2px;
    width: 13px;
    height: 13px;
    border-radius: 50%;
    border: 1.5px solid #fff;
    box-shadow: 0 0 4px rgba(0, 0, 0, 0.2);
    z-index: 1;

    &.online {
      background-color: #52c41a; // 绿色
    }
    &.busy {
      background-color: #f5222d; // 红色
    }
    &.offline {
      background-color: #bfbfbf; // 灰色
    }
  }

  :deep(i) {
    font-weight: normal !important;
  }
}

:global(.user-popover) {
  padding: 0 !important;
  border-radius: 8px !important;
  overflow: hidden;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1) !important;
}

.user-popover-content {
  display: flex;
  flex-direction: column;
  padding: 12px 0;

  .nickname {
    font-size: 23px;
    font-weight: 600;
    color: #333;
    padding: 8px 16px;
    padding-bottom: 4px;
    text-align: center;
  }

  .wallet {
    margin-top: 2px;
    font-size: 12px;
    color: #292828;
    font-weight: 400;
    text-align: center;
    padding-bottom: 8px;
  }

  .companion-info {
    padding: 4px 16px;
    margin-bottom: 4px;
    display: flex;
    flex-direction: row;
    justify-content: center;
    gap: 13px;
    cursor: pointer;

    .info-row {
      display: flex;
      justify-content: center;
      align-items: center;
      gap: 6px;
      font-size: 12px;
      background-color: #f5f5f5;
      padding: 4px 12px;
      border-radius: 12px;

      .value {
        color: #333;
        font-weight: 500;
      }

      .status-text {
        font-weight: 600;
        &.online {
          color: #52c41a;
        }
        &.busy {
          color: #f5222d;
        }
        &.offline {
          color: #bfbfbf;
        }
      }
    }
  }

  .user-stats-card {
    margin: 8px 16px;
    background: #fdfdfd;
    border-radius: 8px;
    border: 1px solid #f0f0f0;
    overflow: hidden;
    transition: all 0.3s ease;

    &:hover {
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
      background: #fff;
    }

    .stats-grid {
      display: flex;
      align-items: center;
      padding: 12px 0;

      .stat-item {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 4px;
        cursor: pointer;
        transition: transform 0.2s;

        &:hover {
          transform: translateY(-1px);
          .stat-num {
            color: #409eff;
          }
        }

        .stat-num {
          font-size: 16px;
          font-weight: 700;
          color: #333;
          line-height: 1;
        }

        .stat-label {
          font-size: 11px;
          color: #909399;
        }
      }

      .stat-divider {
        width: 1px;
        height: 20px;
        background: #eee;
      }
    }

    .bottom-gradient {
      height: 2px;
      background: linear-gradient(90deg, #409eff 0%, #36cfc9 100%);
      opacity: 0.8;
    }
  }

  .divider {
    height: 1px;
    background-color: #eee;
    margin: 4px 0;
  }

  .menu-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 16px;
    font-size: 14px;
    color: #333;
    text-decoration: none;
    transition: background-color 0.2s;

    &:hover {
      background-color: #f5f5f5;
      color: $xtxColor;
    }
  }
}
</style>
