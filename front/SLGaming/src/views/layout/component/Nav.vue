<script setup>
import { useUserStore } from "@/stores/userStore";
import { useWalletStore } from "@/stores/walletStore";
import { useInfoStore } from "@/stores/infoStore";
import { useCompanionStore } from "@/stores/companionStore";
import { useRouter } from "vue-router";
import { storeToRefs } from "pinia";
import { onMounted, computed, watch } from "vue";

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

const handleLogout = async () => {
  // 1. 执行 Store 中的通用退出逻辑
  await userStore.logout();
  // 2. 跳转到登录页
  router.replace("/login");
};

onMounted(() => {
  if (userStore.userInfo?.accessToken) {
    infoStore.getUserDetail();
    walletStore.getWallet();
  }
});

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
  <nav class="app-topnav">
    <div class="container">
      <ul>
        <li class="rank-item">
          <a href="javascript:;" class="menu-item" @click="$router.push('/')">
            <sl-icon name="icon-guanjun" size="16" class="rank-icon" />排行榜
          </a>
        </li>
        <!-- 多模版渲染 区分登录状态和非登录状态 -->
        <!-- 适配思路: 登录时显示第一块 非登录时显示第二块  是否有token -->
        <template v-if="userStore.userInfo?.accessToken">
          <li class="user-info">
            <el-popover placement="bottom" trigger="hover" width="200" popper-class="user-popover">
              <template #reference>
                <a href="javascript:;" class="avatar-link">
                  <div
                    class="avatar-box"
                    v-if="info.avatarUrl"
                    :style="{ backgroundImage: `url(${info.avatarUrl})` }"
                  ></div>
                  <sl-icon name="icon-touxiang1" v-else size="32" color="#fff" />
                  <!-- 状态展示：小圆点 -->
                  <div v-if="statusClass" class="status-tag" :class="statusClass"></div>
                </a>
              </template>
              <div class="user-popover-content">
                <div class="nickname">{{ info.nickname }}</div>
                <div class="wallet">帅币:{{ walletInfo?.balance || "0.00" }}</div>

                <div class="divider"></div>
                <a href="javascript:;" class="menu-item" @click="$router.push('/account/setting')">
                  <sl-icon name="iconfont icon-touxiang" size="16" color="#fff" />个人中心
                </a>
                <a href="javascript:;" class="menu-item logout" @click="handleLogout">
                  <sl-icon name="icon-tuichu" size="16" color="#fff" />退出登录
                </a>
              </div>
            </el-popover>
          </li>
          <li>
            <a href="javascript:;" @click="$router.push('/')"><sl-icon name="icon-shouye" />首页</a>
          </li>
          <li>
            <a href="javascript:;" @click="$router.push('/account/order')"
              ><sl-icon name="icon-dingdan1" />我的订单</a
            >
          </li>
          <li>
            <a href="javascript:;" @click="$router.push('/scion/recharge')"
              ><sl-icon name="icon-chongzhi" />帅币充值</a
            >
          </li>
        </template>
        <template v-else>
          <li>
            <a href="javascript:;" @click="$router.push('/login')"
              ><sl-icon name="icon-qudenglu" />去登录</a
            >
          </li>
        </template>

        <li>
          <a href="javascript:;"><sl-icon name="icon-bangzhuzhongxin" />帮助中心</a>
        </li>
      </ul>
    </div>
  </nav>
</template>

<style scoped lang="scss">
.app-topnav {
  background-image:
    linear-gradient(rgba(130, 129, 129, 0.45), rgba(111, 110, 110, 0)),
    url("@/assets/images/home.png");
  background-repeat: no-repeat;
  background-position: 50% 24%;
  background-size: cover;
  // background-color: #333;

  ul {
    height: 183px;
    display: flex;
    justify-content: flex-end;
    align-items: center;

    li {
      margin-top: -100px;

      &.rank-item {
        margin-right: auto; // 核心：利用 flex 布局特性，自动占据剩余空间，将自己推向最左侧
        a {
          flex-direction: row; // 图标和文字并排
          gap: 10px; // 图标和文字间距
          font-size: 25px;

          i {
            // 重置之前强制的 margin-bottom
            margin-bottom: 0 !important;
          }
        }
      }

      &.user-info {
        .avatar-link {
          position: relative;
          width: 52px;
          height: 52px;
          border-radius: 50%;
          border: 2px solid #fff;
          background-color: rgba(255, 255, 255, 0.2);
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 0; // 清除默认 padding

          .avatar-box {
            width: 100%;
            height: 100%;
            border-radius: 50%;
            background-size: cover;
            background-position: center;
          }

          .status-tag {
            position: absolute;
            top: 37px;
            right: -2px;
            width: 12px;
            height: 12px;
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
        }
      }

      a {
        padding: 0 15px;
        color: #ffffff;
        display: flex;
        flex-direction: column;
        align-items: center;
        font-weight: 549;
        font-size: 16px;

        // 仅针对非头像链接的图标设置强制大小
        &:not(.avatar-link) i {
          font-size: 26px !important;
          margin-right: 0;
          margin-bottom: 1px;
        }

        &.avatar-link :deep(i) {
          font-weight: normal !important;
        }

        &:hover {
          color: $xtxColor;
        }
      }
    }
  }
}

.app-topnav .container {
  width: 100%;
  max-width: 2560px;
  padding: 0 100px;
  position: relative;
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
