<template>
  <div class="setting-info">
    <div class="panel-title">我的订单</div>
    <div class="setting-content">
      <el-empty description="暂无订单" />
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
    padding-top: 20px;
  }
}
</style>
