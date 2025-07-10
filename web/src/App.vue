<template>
  <router-view />
</template>

<script setup>
</script>

<style>
html, body, #app {
  height: 100%;
  margin: 0;
  padding: 0;
  box-sizing: border-box;
  overflow: hidden;
}
.app-body {
  height: calc(100vh - 56px);
  width: 100vw;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.main-container {
  height: 100%;
  width: 100%;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
}
.main-aside {
  height: 100%;
  min-width: 0;
  overflow: auto;
}
.main-content {
  height: 100%;
  width: 100%;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.main-main {
  height: 100%;
  width: 100%;
  min-width: 0;
  min-height: 0;
  overflow: auto;
  padding: 0;
}
body {
  min-height: 100vh;
  background: #f5f7fa;
}
</style> 