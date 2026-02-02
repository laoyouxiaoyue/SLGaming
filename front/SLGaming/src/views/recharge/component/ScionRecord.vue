<template>
  <div class="setting-info">
    <div class="panel-title">消费记录</div>
    <div class="setting-content">
      <el-empty description="暂无记录" />
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
