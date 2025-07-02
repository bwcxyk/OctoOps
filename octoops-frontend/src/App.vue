<template>
  <Header />
  <div class="app-body">
    <el-container class="main-container">
      <el-aside width="220px" class="main-aside">
        <el-menu :default-active="activeMenu" @select="handleMenuSelect">
          <el-menu-item index="overview">
            <span>系统总览</span>
          </el-menu-item>
          <el-sub-menu index="aliyun">
            <template #title>
              <span>阿里云</span>
            </template>
            <el-menu-item index="ecs-security-group">
              <span>ECS安全组</span>
            </el-menu-item>
          </el-sub-menu>
          <el-sub-menu index="etl">
            <template #title>
              <span>ETL调度</span>
            </template>
            <el-menu-item index="batchtask">
              <span>离线数据集成</span>
            </el-menu-item>
            <el-menu-item index="streamtask">
              <span>实时数据集成</span>
            </el-menu-item>
          </el-sub-menu>
          <el-sub-menu index="task">
            <template #title>
              <span>任务管理</span>
            </template>
            <el-menu-item index="task-timer">
              <span>定时任务</span>
            </el-menu-item>
            <el-menu-item index="scheduler">
              <span>调度器</span>
            </el-menu-item>
          </el-sub-menu>
          <el-menu-item index="tasklog">
            <span>作业日志</span>
          </el-menu-item>
          <el-sub-menu index="notification">
            <template #title>
              <span>消息通知</span>
            </template>
            <el-menu-item index="alert-template">
              <span>告警模板</span>
            </el-menu-item>
            <el-menu-item index="alert-channel">
              <span>告警渠道</span>
            </el-menu-item>
            <el-menu-item index="alert-group">
              <span>告警组管理</span>
            </el-menu-item>
          </el-sub-menu>
        </el-menu>
      </el-aside>
      <el-container class="main-content">
        <el-main class="main-main">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import Header from './components/Header.vue'
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()
const activeMenu = ref('overview')

function handleMenuSelect(index) {
  if (index === 'overview') {
    router.push({ name: 'Overview' })
  } else if (index === 'batchtask') {
    router.push({ name: 'BatchTask' })
  } else if (index === 'streamtask') {
    router.push({ name: 'StreamTask' })
  } else if (index === 'scheduler') {
    router.push({ name: 'Scheduler' })
  } else if (index === 'task-timer') {
    router.push({ name: 'TaskTimer' })
  } else if (index === 'tasklog') {
    router.push({ name: 'TaskLog' })
  } else if (index === 'ecs-security-group') {
    router.push({ name: 'EcsSecurityGroup' })
  } else if (index === 'alert-channel') {
    router.push({ name: 'AlertChannel' })
  } else if (index === 'alert-group') {
    router.push({ name: 'AlertGroup' })
  } else if (index === 'alert-template') {
    router.push({ name: 'AlertTemplate' })
  }
}
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