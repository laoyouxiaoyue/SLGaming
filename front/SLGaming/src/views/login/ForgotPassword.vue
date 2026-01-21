<script setup>
import { ref } from "vue";
import { ElMessage } from "element-plus";
import "element-plus/theme-chalk/el-message.css";
import { useRouter } from "vue-router";

import { codeAPI } from "@/api/user/code";
import { forgetpasswordapi } from "@/api/user/forget";
import LoginPanel from "./component/LoginPanel.vue";

const form = ref({
  phone: "",
  code: "",
  password: "",
  rePassword: "",
});

const validatePass = (rule, value, callback) => {
  if (value === "") {
    callback(new Error("请再次输入密码"));
  } else if (value !== form.value.password) {
    callback(new Error("两次输入密码不一致!"));
  } else {
    callback();
  }
};

const rules = {
  phone: [
    { required: true, message: "手机号不能为空", trigger: "blur" },
    { pattern: /^\d{11}$/, message: "请输入有效的11位手机号", trigger: "blur" },
  ],
  code: [
    { required: true, message: "验证码不能为空", trigger: "blur" },
    { pattern: /^\d{6}$/, message: "验证码必须是6位数字", trigger: "blur" },
  ],
  password: [
    { required: true, message: "新密码不能为空", trigger: "blur" },
    { min: 6, max: 14, message: "密码长度为6-14个字符", trigger: "blur" },
  ],
  rePassword: [{ required: true, validator: validatePass, trigger: "blur" }],
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
    await codeAPI({ phone, purpose: "resetpassword" });
    ElMessage({ type: "success", message: "验证码发送成功" });
    countdown.value = 60;
    const timer = setInterval(() => {
      countdown.value--;
      if (countdown.value <= 0) {
        clearInterval(timer);
      }
    }, 1000);
  } catch {
    // 拦截器已经处理了错误提示
  }
};

const doReset = () => {
  formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        const { phone, code, password } = form.value;
        await forgetpasswordapi({ phone, code, password });
        ElMessage({ type: "success", message: "密码重置成功，请重新登录" });
        router.push("/login");
      } catch {
        // ElMessage({ type: "error", message: "重置失败，请检查验证码或手机号" });
      }
    }
  });
};
</script>

<template>
  <LoginPanel>
    <div class="wrapper">
      <nav>
        <a class="active">重置登录密码</a>
      </nav>
      <div class="account-box">
        <div class="form">
          <el-form
            label-position="right"
            label-width="90px"
            ref="formRef"
            :model="form"
            :rules="rules"
            status-icon
          >
            <el-form-item prop="phone" label="手机号">
              <el-input v-model="form.phone" placeholder="请输入绑定的手机号" />
            </el-form-item>
            <el-form-item prop="code" label="验证码">
              <el-input v-model="form.code" style="width: 55%" placeholder="6位验证码" />
              <el-button
                :disabled="countdown > 0"
                @click="sendCode"
                style="width: 40%; margin-left: 5%"
              >
                {{ countdown > 0 ? `${countdown}s` : "获取验证码" }}
              </el-button>
            </el-form-item>
            <el-form-item prop="password" label="新密码">
              <el-input
                v-model="form.password"
                type="password"
                show-password
                placeholder="6-14位字符"
              />
            </el-form-item>
            <el-form-item prop="rePassword" label="确认密码">
              <el-input
                v-model="form.rePassword"
                type="password"
                show-password
                placeholder="请再次输入新密码"
              />
            </el-form-item>
            <el-button size="large" class="subBtn" @click="doReset">提交修改</el-button>
            <div class="login-links">
              <RouterLink to="/login">返回登录</RouterLink>
              <span class="divider">|</span>
              <RouterLink to="/register">立即注册</RouterLink>
            </div>
          </el-form>
        </div>
      </div>
    </div>
  </LoginPanel>
</template>

<style scoped lang="scss">
.wrapper {
  width: 400px;
  background: #fff;
  position: absolute;
  left: 44%;
  top: 150px;
  transform: translate3d(100px, 0, 0);
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.15);

  nav {
    font-size: 14px;
    height: 55px;
    margin-bottom: 20px;
    border-bottom: 1px solid #f5f5f5;
    display: flex;
    padding: 0 40px;
    align-items: center;

    a {
      flex: 1;
      line-height: 1;
      display: inline-block;
      font-size: 18px;
      position: relative;
      text-align: center;

      &.active {
        color: $xtxColor;
        border-bottom: 2px solid $xtxColor;
      }
    }
  }
}

.account-box {
  .form {
    padding: 0 20px 20px 20px;

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
        margin: 0 8px;
        &:hover {
          text-decoration: underline;
        }
      }
      .divider {
        color: #ccc;
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
