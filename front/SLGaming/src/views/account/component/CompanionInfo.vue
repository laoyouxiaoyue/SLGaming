<script setup>
import { ref, onMounted, watch } from "vue";
import { useCompanionStore } from "@/stores/companionStore";
import { storeToRefs } from "pinia";
import { ElMessage } from "element-plus";

const companionStore = useCompanionStore();
const { companionInfo } = storeToRefs(companionStore);

const formRef = ref(null);
const form = ref({
  gameSkill: "",
  pricePerHour: 0,
  status: 0,
  rating: 0,
  totalOrders: 0,
  isVerified: false,
});

// 监听 store 数据变化，同步到表单
watch(
  () => companionInfo.value,
  (newVal) => {
    form.value = {
      gameSkill: newVal.gameSkill || "",
      pricePerHour: newVal.pricePerHour || 0,
      status: newVal.status ?? 0,
      rating: newVal.rating || 0,
      totalOrders: newVal.totalOrders || 0,
      isVerified: newVal.isVerified || false,
    };
  },
  { immediate: true, deep: true },
);

const rules = {
  gameSkill: [{ required: true, message: "请输入游戏技能", trigger: "blur" }],
  pricePerHour: [{ required: true, message: "请输入每小时价格", trigger: "blur" }],
  status: [{ required: true, message: "请选择状态", trigger: "change" }],
};

const onSave = async () => {
  if (!formRef.value) return;
  await formRef.value.validate(async (valid) => {
    if (valid) {
      await companionStore.updateCompanionDetail({
        gameSkill: form.value.gameSkill,
        pricePerHour: Number(form.value.pricePerHour),
        status: form.value.status,
      });
      ElMessage.success("保存成功");
    }
  });
};

onMounted(() => {
  companionStore.getCompanionDetail();
});
</script>

<template>
  <div class="setting-info">
    <!-- 标题栏 -->
    <div class="panel-title">陪玩设置</div>

    <!-- 表单内容 -->
    <div class="setting-content">
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
        label-position="left"
        class="user-form"
      >
        <!-- 只读信息区域 -->
        <div class="info-section">
          <el-descriptions title="基础数据" :column="2" border>
            <el-descriptions-item label="评分">
              <el-rate
                v-model="form.rating"
                disabled
                show-score
                text-color="#ff9900"
                score-template="{value}"
              />
            </el-descriptions-item>
            <el-descriptions-item label="总接单数">
              {{ form.totalOrders }} 单
            </el-descriptions-item>
            <el-descriptions-item label="认证状态">
              <el-tag :type="form.isVerified ? 'success' : 'info'">
                {{ form.isVerified ? "已认证" : "未认证" }}
              </el-tag>
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <el-divider content-position="left">服务设置</el-divider>

        <!-- 可编辑区域 -->
        <el-form-item label="游戏技能" prop="gameSkill">
          <el-input v-model="form.gameSkill" placeholder="请输入擅长的游戏，如：王者荣耀" />
        </el-form-item>

        <el-form-item label="服务价格" prop="pricePerHour">
          <el-input-number
            v-model="form.pricePerHour"
            :min="0"
            :step="10"
            controls-position="right"
          />
          <span style="margin-left: 10px">帅币/小时</span>
        </el-form-item>

        <el-form-item label="当前状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio-button :value="0">离线</el-radio-button>
            <el-radio-button :value="1">在线</el-radio-button>
            <el-radio-button :value="2">忙碌</el-radio-button>
          </el-radio-group>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="onSave" class="save-btn">保存设置</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<style scoped lang="scss">
.setting-info {
  height: 100%;
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
      max-width: 800px;

      .info-section {
        margin-bottom: 30px;
        :deep(.el-descriptions__title) {
          font-size: 15px;
          font-weight: 600;
          color: #333;
        }
      }

      .save-btn {
        width: 140px;
        margin-top: 20px;
        background: linear-gradient(135deg, #ff8e61, #ff6b35);
        border: none;
        font-weight: 500;

        &:hover {
          background: linear-gradient(135deg, #ff9ca4, #ff7a45);
          opacity: 0.9;
        }
      }
    }
  }
}
</style>
