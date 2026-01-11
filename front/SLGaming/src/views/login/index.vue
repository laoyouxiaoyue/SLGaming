<script setup>
import { onMounted, ref } from "vue";
import { useUserStore } from "@/stores/userStore";
const data = ref({
  phone: "1111",
  password: "222",
});

const userStore = useUserStore();
onMounted(() => {
  userStore.getUserInfo(data);
  console.log("ssdsad");
  console.log(userStore.userInfo);
});
</script>

<template>登录页面</template>
