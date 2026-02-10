<script setup>
import { useUserStore } from "@/stores/userStore";
import { useWalletStore } from "@/stores/walletStore";
import { useInfoStore } from "@/stores/infoStore";
import { storeToRefs } from "pinia";
import { onMounted } from "vue";
import InfoPopver from "./InfoPopver.vue";

const userStore = useUserStore();
const walletStore = useWalletStore();
const infoStore = useInfoStore();

// 使用 storeToRefs 获取响应式 state
const { info } = storeToRefs(infoStore);
const { walletInfo } = storeToRefs(walletStore);

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
        <li class="rank-item">
          <a href="javascript:;" class="menu-item" @click="$router.push('/')">
            <sl-icon name="icon-guanjun" size="16" class="rank-icon" />排行榜
          </a>
        </li>
        <!-- 多模版渲染 区分登录状态和非登录状态 -->
        <!-- 适配思路: 登录时显示第一块 非登录时显示第二块  是否有token -->
        <template v-if="userStore.userInfo?.accessToken">
          <li class="user-info">
            <InfoPopver />
          </li>
          <li>
            <a href="javascript:;" @click="$router.push('/')"><sl-icon name="icon-shouye" />首页</a>
          </li>
          <li>
            <a href="javascript:;" @click="$router.push('/order/boss')"
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
            <a href="javascript:;" @click="$router.push('/')"><sl-icon name="icon-shouye" />首页</a>
          </li>
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
</style>
