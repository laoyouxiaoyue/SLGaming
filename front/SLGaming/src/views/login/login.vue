<script setup>
import { ref } from "vue";
import { ElMessage } from "element-plus";
import "element-plus/theme-chalk/el-message.css";
import { useRouter } from "vue-router";

import { useUserStore } from "@/stores/userStore";
import { codeAPI } from "@/api/user/code";
import LoginPanel from "./component/loginpanel.vue";
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
  } catch {
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
  <LoginPanel>
    <div class="wrapper">
      <nav>
        <a
          href="javascript:;"
          :class="{ active: loginType === 'account' }"
          @click="loginType = 'account'"
          >账户登录</a
        >
        <a href="javascript:;" :class="{ active: loginType === 'code' }" @click="loginType = 'code'"
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
            <el-form-item prop="phone" label="账    户">
              <el-input v-model="form.phone" />
            </el-form-item>
            <template v-if="loginType === 'account'">
              <el-form-item prop="password" label="密    码">
                <el-input v-model="form.password" type="password" show-password />
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
            <div class="login-links">
              <RouterLink to="/register">注册账号</RouterLink>
              <span class="divider">|</span>
              <RouterLink to="/forgot">忘记密码？</RouterLink>
            </div>
          </el-form>
        </div>
      </div>
    </div>
  </LoginPanel>
</template>

<style scoped lang="scss">
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
    .login-links {
      display: flex;
      justify-content: center;
      align-items: center;
      font-size: 14px;
      margin-top: 15px;
      margin-bottom: 6px;
      color: #999;
      a {
        color: #409eff;
        text-decoration: none;
        margin: 0 4px;
        &:hover {
          text-decoration: underline;
        }
      }
      .divider {
        color: #ccc;
        margin: 0 2px;
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
  width: 70%;
  display: block;
  margin: 20px auto;
  color: #fff;
}
</style>
