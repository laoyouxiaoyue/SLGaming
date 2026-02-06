<script setup>
import { ref, computed, watch } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useInfoStore } from "@/stores/infoStore";

const router = useRouter();
const route = useRoute();
const infoStore = useInfoStore();

// 若角色为2 (陪玩)，显示"陪玩订单" tab
const isCompanion = computed(() => infoStore.info.role === 2);

// 根据路由路径决定 active tab
const activeTab = ref(route.path.includes("companion") ? "companion" : "boss");

// 监听 tab 切换跳转路由
const handleTabClick = (tab) => {
  router.push(`/order/${tab.paneName}`);
};
</script>

<template>
  <div class="order-layout">
    <el-tabs v-model="activeTab" class="order-tabs" @tab-click="handleTabClick">
      <el-tab-pane label="我的订单" name="boss"> <RouterView /> </el-tab-pane>
      <el-tab-pane v-if="isCompanion" label="陪玩订单" name="companion">
        <RouterView />
      </el-tab-pane>
    </el-tabs>
    <div class="order-content"></div>
  </div>
</template>

<style scoped lang="scss">
.order-layout {
  padding: 20px;
  width: 1200px;
  margin: 20px auto;
  border-radius: 8px;

  .order-tabs {
    background: #fff;
    padding: 0 20px;
    margin-bottom: 20px;
    border-radius: 8px;
    border: 1px solid #eeeded;
  }
}
</style>
