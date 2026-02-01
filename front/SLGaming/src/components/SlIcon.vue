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
    default: "#fff",
  },
});
</script>

<template>
  <i class="iconfont" :class="name" :style="{ fontSize: size + 'px', color: color }"></i>
</template>

<style scoped>
.iconfont {
  display: inline-block;
  font-style: normal;
  vertical-align: middle;
}
</style>
