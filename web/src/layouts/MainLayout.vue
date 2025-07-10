<template>
  <Header />
  <div class="app-body">
    <el-container class="main-container">
      <el-aside width="220px" class="main-aside">
        <el-menu :default-active="activeMenu" @select="handleMenuSelect">
          <template v-for="menu in userStore.menus" :key="menu.code">
            <el-sub-menu v-if="menu.children && menu.children.length" :index="menu.code">
              <template #title>
                <span>{{ menu.name }}</span>
              </template>
              <el-menu-item v-for="child in menu.children" :index="child.path || child.code" :key="child.code">
                <span>{{ child.name }}</span>
              </el-menu-item>
            </el-sub-menu>
            <el-menu-item v-else :index="menu.path || menu.code" :key="menu.code">
              <span>{{ menu.name }}</span>
            </el-menu-item>
          </template>
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
import Header from '@/components/Header.vue'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/store/user'

const router = useRouter()
const userStore = useUserStore()
const activeMenu = ref('overview')

function handleMenuSelect(index) {
  // 路由 name 和 path 映射
  const nameMap = {
    'overview': 'Overview',
    'batchtask': 'BatchTask',
    'streamtask': 'StreamTask',
    'scheduler': 'Scheduler',
    'task-timer': 'CustomTask',
    'tasklog': 'TaskLog',
    'ecs-security-group': 'EcsSecurityGroup',
    'alert-channel': 'AlertChannel',
    'alert-group': 'AlertGroup',
    'alert-template': 'AlertTemplate',
    'rbac-user': 'UserManage',
    'rbac-role': 'RoleManage',
    'rbac-permission': 'PermissionManage',
  }
  if (nameMap[index]) {
    router.push({ name: nameMap[index] })
  } else {
    router.push({ path: index })
  }
}
</script>

<style scoped>
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
</style> 