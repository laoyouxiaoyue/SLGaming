<script setup>
import { ref, onUnmounted } from "vue";
import { ElMessage } from "element-plus";
import { codeAPI } from "@/api/user/code";
import { changePhoneAPI } from "@/api/user/info";
import { useInfoStore } from "@/stores/infoStore";

const props = defineProps({
  phone: {
    type: String,
    required: true,
  },
});

const infoStore = useInfoStore();

// 获取旧手机验证码逻辑
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
// 获取新手机验证码逻辑
const countdownNew = ref(0);
const timerNew = ref(null);
const startTimerNew = () => {
  countdownNew.value = 60;
  timerNew.value = setInterval(() => {
    if (countdownNew.value > 0) {
      countdownNew.value--;
    } else {
      clearInterval(timerNew.value);
    }
  }, 1000);
};

onUnmounted(() => {
  if (timer.value) clearInterval(timer.value);
  if (timerNew.value) clearInterval(timerNew.value);
});

const sendOldCode = async () => {
  if (!props.phone) {
    ElMessage.error("获取当前手机号失败");
    return;
  }
  try {
    await codeAPI({ phone: props.phone, purpose: "change_phone" });
    ElMessage.success("验证码已发送至原手机");
    startTimer();
  } catch (error) {
    console.error(error);
  }
};

const sendNewCode = async () => {
  if (!phoneForm.value.newPhone) {
    ElMessage.warning("请先输入新手机号");
    return;
  }
  if (!/^1[3-9]\d{9}$/.test(phoneForm.value.newPhone)) {
    ElMessage.warning("新手机号格式不正确");
    return;
  }
  try {
    await codeAPI({ phone: phoneForm.value.newPhone, purpose: "change_phone_new" });
    ElMessage.success("验证码已发送至新手机");
    startTimerNew();
  } catch (error) {
    console.error(error);
  }
};

// 修改手机号业务
const phoneFormRef = ref(null);
const phoneForm = ref({
  oldCode: "",
  newPhone: "",
  newCode: "",
});

const phoneRules = {
  oldCode: [{ required: true, message: "请输入原手机验证码", trigger: "blur" }],
  newPhone: [
    { required: true, message: "请输入新手机号", trigger: "blur" },
    { pattern: /^1[3-9]\d{9}$/, message: "手机号格式不正确", trigger: "blur" },
  ],
  newCode: [{ required: true, message: "请输入新手机验证码", trigger: "blur" }],
};

const handleUpdatePhone = async () => {
  if (!phoneFormRef.value) return;
  await phoneFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        await changePhoneAPI({
          oldPhone: props.phone,
          oldCode: phoneForm.value.oldCode,
          newPhone: phoneForm.value.newPhone,
          newCode: phoneForm.value.newCode,
        });
        ElMessage.success("手机号修改成功");
        phoneForm.value = { oldCode: "", newPhone: "", newCode: "" };
        infoStore.getUserDetail();
      } catch (error) {
        console.error(error);
      }
    }
  });
};
</script>

<template>
  <el-form
    ref="phoneFormRef"
    :model="phoneForm"
    :rules="phoneRules"
    label-width="100px"
    class="security-form"
  >
    <el-form-item label="当前手机">
      <el-input :value="phone" disabled />
    </el-form-item>

    <el-form-item label="原手机验证" prop="oldCode">
      <div class="code-input-group">
        <el-input v-model="phoneForm.oldCode" placeholder="原手机验证码" />
        <el-button :disabled="countdown > 0" @click="sendOldCode" class="code-btn">
          {{ countdown > 0 ? `${countdown}s后重发` : "获取验证码" }}
        </el-button>
      </div>
    </el-form-item>

    <el-form-item label="新手机号" prop="newPhone">
      <el-input v-model="phoneForm.newPhone" placeholder="请输入新手机号" />
    </el-form-item>

    <el-form-item label="新手机验证" prop="newCode">
      <div class="code-input-group">
        <el-input v-model="phoneForm.newCode" placeholder="新手机验证码" />
        <el-button :disabled="countdownNew > 0" @click="sendNewCode" class="code-btn">
          {{ countdownNew > 0 ? `${countdownNew}s后重发` : "获取验证码" }}
        </el-button>
      </div>
    </el-form-item>

    <el-form-item>
      <el-button type="primary" class="submit-btn" @click="handleUpdatePhone">
        确认修改手机号
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
