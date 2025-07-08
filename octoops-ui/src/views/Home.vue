<template>
  <el-container style="height: 100vh">
    <el-aside width="220px" style="background: #f5f7fa">
      <el-menu :default-active="activeMenu" @select="handleMenuSelect">
        <el-menu-item index="overview">
          <span>系统总览</span>
        </el-menu-item>
        <el-menu-item index="batchtask">
          <span>离线数据集成</span>
        </el-menu-item>
        <el-menu-item index="streamtask">
          <span>实时数据集成</span>
        </el-menu-item>
        <el-menu-item index="scheduler">
          <span>调度器状态</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-main>
        <Overview v-if="activeMenu === 'overview'" />
        <BatchTask v-else-if="activeMenu === 'batchtask'" />
        <StreamTask v-else-if="activeMenu === 'streamtask'" />
        <Scheduler v-else-if="activeMenu === 'scheduler'" />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref } from 'vue'
import Overview from './Overview.vue'
import BatchTask from './BatchTask.vue'
import StreamTask from './StreamTask.vue'
import Scheduler from './Scheduler.vue'

const activeMenu = ref('overview')

function handleMenuSelect(index) {
  activeMenu.value = index
}
</script> 