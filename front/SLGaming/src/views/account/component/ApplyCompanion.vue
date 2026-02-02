<script setup>
import { ref } from "vue";
import { ElMessage } from "element-plus";
import { applycompanionapi } from "@/api/companion/companion";
import router from "@/router";

defineOptions({ name: "ApplyCompanion" });

const formRef = ref(null);
const loading = ref(false);
const form = ref({
  gameSkill: "",
  pricePerHour: null,
  bio: "",
});

const rules = {
  gameSkill: [{ required: true, message: "请输入游戏技能", trigger: "blur" }],
  pricePerHour: [{ required: true, message: "请输入每小时价格", trigger: "blur" }],
};

const onSubmit = async () => {
  if (!formRef.value || loading.value) return;
  await formRef.value.validate(async (valid) => {
    if (!valid) return;
    loading.value = true;
    try {
      await applycompanionapi({
        gameSkill: form.value.gameSkill,
        pricePerHour: Number(form.value.pricePerHour),
        bio: form.value.bio || undefined,
      });
      ElMessage.success({
        message: "您已申请成功，获得成为陪玩资格",
        duration: 4000,
      });
      router.push("/account/companion");
    } catch (error) {
      console.error(error);
      ElMessage.error("提交失败，请稍后再试");
    } finally {
      loading.value = false;
    }
  });
};

const onReset = () => {
  form.value = {
    gameSkill: "",
    pricePerHour: null,
    bio: "",
  };
  formRef.value?.clearValidate();
};
</script>

<template>
  <div class="setting-info">
    <div class="panel-title">申请成为陪玩</div>
    <div class="setting-content">
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
        label-position="left"
        class="user-form"
      >
        <el-form-item label="游戏技能" prop="gameSkill">
          <el-input v-model="form.gameSkill" placeholder="如：王者荣耀" />
        </el-form-item>

        <el-form-item label="每小时价格" prop="pricePerHour">
          <el-input-number
            v-model="form.pricePerHour"
            :min="1"
            :step="10"
            controls-position="right"
          />
          <span class="price-unit">帅币/小时</span>
        </el-form-item>

        <el-form-item label="个人简介">
          <el-input
            v-model="form.bio"
            type="textarea"
            :rows="4"
            placeholder="专业陪玩，技术过硬，带你上分！"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" class="save-btn" :loading="loading" @click="onSubmit">
            提交申请
          </el-button>
          <el-button class="reset-btn" :disabled="loading" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<style scoped lang="scss">
.setting-info {
  padding: 0 10px;

  .panel-title {
    font-size: 20px;
    font-weight: 600;
    margin-bottom: 25px;
    color: #333;
    border-left: 4px solid #ff6b35;
    padding-left: 12px;
  }

  .setting-content {
    .user-form {
      max-width: 700px;

      .price-unit {
        margin-left: 10px;
        color: #666;
      }

      .save-btn {
        width: 140px;
        background: linear-gradient(135deg, #ff8e61, #ff6b35);
        border: none;
        font-weight: 500;

        &:hover {
          background: linear-gradient(135deg, #ff9ca4, #ff7a45);
          opacity: 0.9;
        }
      }

      .reset-btn {
        margin-left: 10px;
      }
    }
  }
}
</style>
