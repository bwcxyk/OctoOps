<template>
  <div class="app-header">
    <div class="header-left">
      <img src="/src/assets/logo.png" alt="logo" class="logo" />
      <span class="title">OctoOps</span>
    </div>
  </div>
</template>

<script setup>

// 可根据需要引入 props 或全局配置
</script>

<style scoped>
.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #4A90E2;
  color: #fff;
  padding: 0 32px;
  height: 56px;
  width: 100vw;
  min-width: 0;
  box-sizing: border-box;
  margin: 0;
  border-radius: 0;
}
.header-left {
  display: flex;
  align-items: center;
}
.logo {
  width: 65px;
  height: 65px;
  margin-right: 14px;
  background: transparent;
  border-radius: 8px;
  padding: 0;
  object-fit: contain;
  display: block;
}
.title {
  font-size: 1.25rem;
  font-weight: bold;
  letter-spacing: 2px;
}
.header-right {
  font-size: 1rem;
  opacity: 0.95;
  display: flex;
  align-items: center;
  gap: 24px;
}
</style> 