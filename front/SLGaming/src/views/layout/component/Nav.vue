<script setup>
import { useUserStore } from "@/stores/userStore";
import { useWalletStore } from "@/stores/walletStore";
import { useInfoStore } from "@/stores/infoStore";
import { useRouter } from "vue-router";
import { storeToRefs } from "pinia";
import { onMounted } from "vue";
import { getlogoutAPI } from "@/api/user/logout";

const userStore = useUserStore();
const walletStore = useWalletStore();
const infoStore = useInfoStore();
const router = useRouter();

// 使用 storeToRefs 获取响应式 state
const { info } = storeToRefs(infoStore);
const { walletInfo } = storeToRefs(walletStore);

const confirm = async () => {
  // 退出登录业务逻辑实现
  // 1.清除用户信息 触发action
  await getlogoutAPI();
  userStore.clearUserInfo();
  infoStore.clearInfo();
  // 2.跳转到登录页
  router.push("/login");
};

onMounted(() => {
  if (userStore.userInfo?.accessToken) {
    infoStore.getUserDetail();
    walletStore.getWallet();
  }
});
</script>

<template>
  <nav class="app-topnav">
    <div class="container">
      <ul>
        <!-- 多模版渲染 区分登录状态和非登录状态 -->
        <!-- 适配思路: 登录时显示第一块 非登录时显示第二块  是否有token -->
        <template v-if="userStore.userInfo?.accessToken">
          <li class="user-info">
            <el-popover placement="bottom" trigger="hover" width="200" popper-class="user-popover">
              <template #reference>
                <a href="javascript:;" class="avatar-link">
                  <div class="avatar-box" v-if="info.avatarUrl">
                    <img :src="info.avatarUrl" />
                  </div>
                  <sl-icon name="icon-touxiang1" v-else size="64" color="#fff" />
                </a>
              </template>
              <div class="user-popover-content">
                <div class="nickname">{{ info.nickname }}</div>
                <div class="divider"></div>
                <!-- 钱包余额 -->
                <div class="wallet-info">
                  <div class="wallet-item">
                    <span class="label">账户余额</span>
                    <span class="value">¥{{ walletInfo.balance || "0.00" }}</span>
                  </div>
                  <div class="wallet-item">
                    <span class="label">冻结金额</span>
                    <span class="value">¥{{ walletInfo.frozenBalance || "0.00" }}</span>
                  </div>
                </div>
                <div class="divider"></div>
                <a href="javascript:;" class="menu-item" @click="$router.push('/account/setting')">
                  <sl-icon name="icon-touxiang" size="16" />个人中心
                </a>
                <a href="javascript:;" class="menu-item logout" @click="confirm">
                  <sl-icon name="icon-tuichu" size="16" />退出登录
                </a>
              </div>
            </el-popover>
          </li>
          <li>
            <a href="javascript:;" @click="$router.push('/')"><sl-icon name="icon-shouye" />首页</a>
          </li>
          <li>
            <a href="javascript:;" @click="$router.push('/account/order')"
              ><sl-icon name="icon-dingdan" />我的订单</a
            >
          </li>
          <li>
            <a href="javascript:;" @click="$router.push('/account/wallet')"
              ><sl-icon name="icon-chongzhi" />帅币充值</a
            >
          </li>
        </template>
        <template v-else>
          <li>
            <a href="javascript:;" @click="$router.push('/login')"
              ><sl-icon name="icon-touxiang" />去登录</a
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
      &.user-info {
        a {
          .avatar-box {
            width: 52px;
            height: 52px;
            border-radius: 50%;
            border: 2px solid #fff;
            background-color: rgba(255, 255, 255, 0.2);

            img {
              width: 100%;
              height: 100%;
              object-fit: cover;
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
  padding: 0 80px;
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
    font-size: 16px;
    font-weight: 600;
    color: #333;
    padding: 8px 16px;
    text-align: center;
  }

  .wallet-info {
    padding: 8px 16px;
    background-color: #f9f9f9;
    margin: 4px 0;

    .wallet-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 4px;
      font-size: 13px;

      &:last-child {
        margin-bottom: 0;
      }

      .label {
        color: #666;
      }

      .value {
        color: #ff6b35;
        font-weight: 600;
      }
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
