<script setup>
import { ref } from "vue";
import { useInfoStore } from "@/stores/infoStore";
import { storeToRefs } from "pinia";
import PhoneSetting from "./component/PhoneSetting.vue";
import PasswordSetting from "./component/PasswordSetting.vue";

defineOptions({ name: "SecuritySetting" });

const infoStore = useInfoStore();
const { info } = storeToRefs(infoStore);

const activeTab = ref("phone");
</script>

<template>
  <div class="setting-info">
    <div class="panel-title">安全设置</div>
    <div class="setting-content">
      <el-tabs v-model="activeTab" class="security-tabs">
        <el-tab-pane label="修改手机号" name="phone">
          <PhoneSetting :phone="info.phone" />
        </el-tab-pane>

        <el-tab-pane label="修改密码" name="password">
          <PasswordSetting :phone="info.phone" />
        </el-tab-pane>
      </el-tabs>
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
    .security-tabs {
      :deep(.el-tabs__item.is-active) {
        color: #ff6b35;
      }
      :deep(.el-tabs__active-bar) {
        background-color: #ff6b35;
      }
    }
  }
}
</style>
