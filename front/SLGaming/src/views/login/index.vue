<script setup>
import { ref } from "vue";
import { ElMessage } from "element-plus";
import "element-plus/theme-chalk/el-message.css";
import { useRouter } from "vue-router";

import { useUserStore } from "@/stores/userStore";
import { codeAPI } from "@/api/user/code";
const userStore = useUserStore();

const loginType = ref("account"); // 'account' or 'code'

const form = ref({
  phone: "13800138000",
  password: "password123",
  code: "",
});

const rules = {
  phone: [
    { required: true, message: "电话不能为空", trigger: "blur" },
    { pattern: /^\d{11}$/, message: "电话号码必须是11位数字", trigger: "blur" },
  ],
  password: [
    { required: true, message: "密码不能为空", trigger: "blur" },
    { min: 6, max: 14, message: "密码长度为6-14个字符", trigger: "blur" },
  ],
  code: [
    { required: true, message: "验证码不能为空", trigger: "blur" },
    { pattern: /^\d{6}$/, message: "验证码必须是6位数字", trigger: "blur" },
  ],
};

const formRef = ref(null);
const router = useRouter();
const countdown = ref(0);

const sendCode = async () => {
  if (countdown.value > 0) return;
  const { phone } = form.value;
  if (!phone || !/^\d{11}$/.test(phone)) {
    ElMessage({ type: "error", message: "请输入正确的手机号" });
    return;
  }
  try {
    await codeAPI({ phone, purpose: "login" });
    ElMessage({ type: "success", message: "验证码发送成功" });
    countdown.value = 60;
    const timer = setInterval(() => {
      countdown.value--;
      if (countdown.value <= 0) {
        clearInterval(timer);
      }
    }, 1000);
  } catch (error) {
    ElMessage({ type: "error", message: "验证码发送失败" });
  }
};
const doLogin = () => {
  const { phone, password, code } = form.value;
  // 调用实例方法
  formRef.value.validate(async (valid) => {
    // valid: 所有表单都通过校验  才为true
    // 以valid做为判断条件 如果通过校验才执行登录逻辑
    if (valid) {
      if (loginType.value === "account") {
        await userStore.getUserInfo({ phone, password });
      } else {
        await userStore.getUserInfoByCode({ phone, code });
      }
      // console.log(userStore.userInfo);
      // // 1. 提示用户
      ElMessage({ type: "success", message: "登录成功" });
      // // 2. 跳转首页
      router.replace({ path: "/" });
    }
  });
};
</script>

<template>
  <div>
    <header class="login-header">
      <div class="container m-top-20">
        <h1 class="logo">
          <RouterLink to="/">SLGaming</RouterLink>
        </h1>
        <RouterLink class="entry" to="/">
          进入网站首页
          <i class="iconfont icon-angle-right"></i>
          <i class="iconfont icon-angle-right"></i>
        </RouterLink>
      </div>
    </header>
    <section class="login-section">
      <div class="wrapper">
        <nav>
          <a
            href="javascript:;"
            :class="{ active: loginType === 'account' }"
            @click="loginType = 'account'"
            >账户登录</a
          >
          <a
            href="javascript:;"
            :class="{ active: loginType === 'code' }"
            @click="loginType = 'code'"
            >验证码登录</a
          >
        </nav>
        <div class="account-box">
          <div class="form">
            <el-form
              label-position="right"
              label-width="80px"
              ref="formRef"
              :model="form"
              :rules="rules"
              status-icon
            >
              <el-form-item prop="phone" label="账户">
                <el-input v-model="form.phone" />
              </el-form-item>
              <template v-if="loginType === 'account'">
                <el-form-item prop="password" label="密码">
                  <el-input v-model="form.password" />
                </el-form-item>
              </template>
              <template v-else>
                <el-form-item prop="code" label="验证码">
                  <el-input v-model="form.code" style="width: 60%" />
                  <el-button
                    :disabled="countdown > 0"
                    @click="sendCode"
                    style="width: 35%; margin-left: 5%"
                  >
                    {{ countdown > 0 ? `${countdown}s` : "发送验证码" }}
                  </el-button>
                </el-form-item>
              </template>
              <el-button size="large" class="subBtn" @click="doLogin">点击登录</el-button>
            </el-form>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped lang="scss">
.login-header {
  background: #fff;
  border-bottom: 1px solid #e4e4e4;

  .container {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
  }

  .logo {
    width: 200px;

    a {
      display: block;
      height: 132px;
      width: 100%;
      text-indent: -9999px;
      background: url("@/assets/images/logo.png") no-repeat center 18px / contain;
    }
  }

  .sub {
    flex: 1;
    font-size: 24px;
    font-weight: normal;
    margin-bottom: 38px;
    margin-left: 20px;
    color: #666;
  }

  .entry {
    width: 120px;
    margin-bottom: 38px;
    font-size: 16px;

    i {
      font-size: 14px;
      color: $xtxColor;
      letter-spacing: -5px;
    }
  }
}

.login-section {
  background: linear-gradient(rgba(255, 255, 255, 0.25), rgba(255, 255, 255, 0.25)),
    url("@/assets/images/login-bg.png") no-repeat center / cover;
  height: 783px;
  position: relative;

  .wrapper {
    // height: 300px;
    width: 380px;
    background: #fff;
    position: absolute;
    left: 44%;
    top: 200px;
    transform: translate3d(100px, 0, 0);
    box-shadow: 0 0 10px rgba(0, 0, 0, 0.15);

    nav {
      font-size: 14px;
      height: 55px;
      margin-bottom: 20px;
      border-bottom: 1px solid #f5f5f5;
      display: flex;
      padding: 0 40px;
      text-align: right;
      align-items: center;

      a {
        flex: 1;
        line-height: 1;
        display: inline-block;
        font-size: 18px;
        position: relative;
        text-align: center;
        cursor: pointer;

        &.active {
          color: $xtxColor;
          border-bottom: 2px solid $xtxColor;
        }
      }
    }
  }
}

.login-footer {
  padding: 30px 0 50px;
  background: #fff;

  p {
    text-align: center;
    color: #999;
    padding-top: 20px;

    a {
      line-height: 1;
      padding: 0 10px;
      color: #999;
      display: inline-block;

      ~ a {
        border-left: 1px solid #ccc;
      }
    }
  }
}

.account-box {
  .toggle {
    padding: 15px 40px;
    text-align: right;

    a {
      color: $xtxColor;

      i {
        font-size: 14px;
      }
    }
  }

  .form {
    padding: 0 20px 20px 20px;

    &-item {
      margin-bottom: 28px;

      .input {
        position: relative;
        height: 36px;

        > i {
          width: 34px;
          height: 34px;
          background: #cfcdcd;
          color: #fff;
          position: absolute;
          left: 1px;
          top: 1px;
          text-align: center;
          line-height: 34px;
          font-size: 18px;
        }

        input {
          padding-left: 44px;
          border: 1px solid #cfcdcd;
          height: 36px;
          line-height: 36px;
          width: 100%;

          &.error {
            border-color: $priceColor;
          }

          &.active,
          &:focus {
            border-color: $xtxColor;
          }
        }

        .code {
          position: absolute;
          right: 1px;
          top: 1px;
          text-align: center;
          line-height: 34px;
          font-size: 14px;
          background: #f5f5f5;
          color: #666;
          width: 90px;
          height: 34px;
          cursor: pointer;
        }
      }

      > .error {
        position: absolute;
        font-size: 12px;
        line-height: 28px;
        color: $priceColor;

        i {
          font-size: 14px;
          margin-right: 2px;
        }
      }
    }

    .agree {
      a {
        color: #069;
      }
    }

    .btn {
      display: block;
      width: 100%;
      height: 40px;
      color: #fff;
      text-align: center;
      line-height: 40px;
      background: $xtxColor;

      &.disabled {
        background: #cfcdcd;
      }
    }
  }

  .action {
    padding: 20px 40px;
    display: flex;
    justify-content: space-between;
    align-items: center;

    .url {
      a {
        color: #999;
        margin-left: 10px;
      }
    }
  }
}

.subBtn {
  background: $xtxColor;
  width: 100%;
  color: #fff;
}
</style>
