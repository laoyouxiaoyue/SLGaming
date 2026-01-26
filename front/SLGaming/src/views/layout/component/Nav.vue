<script setup>
import { useUserStore } from "@/stores/userStore";
import { useRouter } from "vue-router";
import { getInfoAPI } from "@/api/user/info";
import { ref, onMounted } from "vue";
import { getlogoutAPI } from "@/api/user/logout";

const userStore = useUserStore();
const router = useRouter();
const confirm = async () => {
  // 退出登录业务逻辑实现
  // 1.清除用户信息 触发action
  await getlogoutAPI();
  userStore.clearUserInfo();
  // 2.跳转到登录页
  router.push("/login");
};

const info = ref({});
const getInfo = async () => {
  const res = await getInfoAPI();
  info.value = res.data;
};
onMounted(() => {
  if (userStore.userInfo?.accessToken) {
    getInfo();
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
          <li>
            <a href="javascript:;"><i class="iconfont icon-user"></i>{{ info.nickname }} </a>
          </li>
          <li>
            <el-popconfirm
              @confirm="confirm"
              title="确认退出吗?"
              confirm-button-text="确认"
              cancel-button-text="取消"
            >
              <template #reference>
                <a href="javascript:;">退出登录</a>
              </template>
            </el-popconfirm>
          </li>
          <li><a href="javascript:;">我的订单</a></li>
          <li><a href="javascript:;">会员中心</a></li>
        </template>
        <template v-else>
          <li><a href="javascript:;" @click="$router.push('/login')">请先登录</a></li>
          <li><a href="javascript:;">帮助中心</a></li>
          <li><a href="javascript:;">关于我们</a></li>
        </template>
      </ul>
    </div>
  </nav>
</template>

<style scoped lang="scss">
.app-topnav {
  background-image:
    linear-gradient(rgba(130, 129, 129, 0.3), rgba(111, 110, 110, 0)),
    url("@/assets/images/home.png");
  background-repeat: no-repeat;
  background-position: 50% 25%;
  background-size: cover;
  // background-color: #333;

  ul {
    height: 156px;
    display: flex;
    justify-content: flex-end;
    align-items: flex-start;

    li {
      a {
        margin-top: 30px;
        padding: 0 15px;
        color: #ffffff;
        line-height: 1;
        display: inline-block;
        font-weight: 400;
        font-size: 18px;

        i {
          font-size: 14px;
          margin-right: 2px;
        }

        &:hover {
          color: $xtxColor;
        }
      }

      ~ li {
        a {
          border-left: 2px solid #666;
        }
      }
    }
  }
}

.app-topnav .container {
  width: 100%;
  max-width: 2560px;
  padding: 0 66px;
  position: relative;
}
</style>
