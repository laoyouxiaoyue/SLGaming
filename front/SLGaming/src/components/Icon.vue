<script setup>
defineProps({
  name: {
    type: String,
    required: true,
  },
  size: {
    type: [Number, String],
    default: 16,
  },
  color: {
    type: String,
    default: "inherit",
  },
});
</script>

<template>
  <i class="iconfont" :class="name" :style="{ fontSize: size + 'px', color: color }"></i>
</template>

<style scoped>
@import url("//at.alicdn.com/t/c/font_5120505_9p9k7k8trm.css");

.iconfont {
  display: inline-block;
  font-style: normal;
  vertical-align: middle;
}
</style>
