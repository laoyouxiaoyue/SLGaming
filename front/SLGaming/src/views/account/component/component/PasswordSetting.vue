<script setup>
import { ref, onUnmounted } from "vue";
import { ElMessage } from "element-plus";
import { codeAPI } from "@/api/user/code";
import { changePasswordAPI } from "@/api/user/info";
import { useUserStore } from "@/stores/userStore";
import { useRouter } from "vue-router";

const props = defineProps({
  phone: {
    type: String,
    required: true,
  },
});

const userStore = useUserStore();
const router = useRouter();

// 验证码逻辑
const countdown = ref(0);
const timer = ref(null);
const startTimer = () => {
  countdown.value = 60;
  timer.value = setInterval(() => {
    if (countdown.value > 0) {
      countdown.value--;
    } else {
      clearInterval(timer.value);
    }
  }, 1000);
};

onUnmounted(() => {
  if (timer.value) clearInterval(timer.value);
});

const sendOldCode = async () => {
  if (!props.phone) {
    ElMessage.error("获取当前手机号失败");
    return;
  }
  try {
    await codeAPI({ phone: props.phone, purpose: "change_password" });
    ElMessage.success("验证码已发送至原手机");
    startTimer();
  } catch (error) {
    console.error(error);
  }
};

// 修改密码业务
const pwdFormRef = ref(null);
const pwdForm = ref({
  oldCode: "",
  newPassword: "",
  confirmPassword: "",
});

const pwdRules = {
  oldCode: [{ required: true, message: "请输入验证码", trigger: "blur" }],
  newPassword: [
    { required: true, message: "请输入新密码", trigger: "blur" },
    { min: 6, max: 20, message: "密码长度为 6-20 位", trigger: "blur" },
  ],
  confirmPassword: [
    { required: true, message: "请确认新密码", trigger: "blur" },
    {
      validator: (rule, value, callback) => {
        if (value !== pwdForm.value.newPassword) {
          callback(new Error("两次输入的密码不一致"));
        } else {
          callback();
        }
      },
      trigger: "blur",
    },
  ],
};

const handleUpdatePassword = async () => {
  if (!pwdFormRef.value) return;
  await pwdFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        await changePasswordAPI({
          oldPhone: props.phone,
          oldCode: pwdForm.value.oldCode,
          newPassword: pwdForm.value.newPassword,
        });
        ElMessage.success({
          message: "密码修改成功，请重新登录",
          duration: 2000,
        });
        pwdForm.value = { oldCode: "", newPassword: "", confirmPassword: "" };

        // 修改成功后跳转登录页逻辑
        setTimeout(async () => {
          await userStore.logout();
          router.replace("/login");
        }, 2000);
      } catch (error) {
        console.error(error);
      }
    }
  });
};
</script>

<template>
  <el-form
    ref="pwdFormRef"
    :model="pwdForm"
    :rules="pwdRules"
    label-width="100px"
    class="security-form"
  >
    <el-form-item label="当前手机">
      <el-input :value="phone" disabled />
    </el-form-item>

    <el-form-item label="验证码" prop="oldCode">
      <div class="code-input-group">
        <el-input v-model="pwdForm.oldCode" placeholder="原手机验证码" />
        <el-button :disabled="countdown > 0" @click="sendOldCode" class="code-btn">
          {{ countdown > 0 ? `${countdown}s后重发` : "获取验证码" }}
        </el-button>
      </div>
    </el-form-item>

    <el-form-item label="新密码" prop="newPassword">
      <el-input
        v-model="pwdForm.newPassword"
        type="password"
        show-password
        placeholder="请输入新密码"
      />
    </el-form-item>

    <el-form-item label="确认密码" prop="confirmPassword">
      <el-input
        v-model="pwdForm.confirmPassword"
        type="password"
        show-password
        placeholder="请再次填写新密码"
      />
    </el-form-item>

    <el-form-item>
      <el-button type="primary" class="submit-btn" @click="handleUpdatePassword">
        确认修改密码
      </el-button>
    </el-form-item>
  </el-form>
</template>

<style scoped lang="scss">
.security-form {
  max-width: 500px;
  margin-top: 20px;

  .code-input-group {
    display: flex;
    gap: 12px;
    width: 100%;

    .code-btn {
      width: 120px;
      flex-shrink: 0;
    }
  }

  .submit-btn {
    width: 100%;
    background: linear-gradient(135deg, #ff8e61, #ff6b35);
    border: none;
    margin-top: 10px;

    &:hover {
      background: linear-gradient(135deg, #ff9ca4, #ff7a45);
      opacity: 0.9;
    }
  }
}
</style>
