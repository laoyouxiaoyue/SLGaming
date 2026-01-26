<template>
  <div class="login-container">
    <section class="login-section">
      <slot />
    </section>
  </div>
</template>

<style scoped lang="scss">
.login-section {
  background:
    // linear-gradient(rgba(255, 255, 255, 0.25), rgba(255, 255, 255, 0.25)),
    url("@/assets/images/login-bg.png") no-repeat center / cover;
  height: 934px;
  position: relative;
}
</style>
