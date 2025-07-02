<template>
  <el-container>
    <el-main style="overflow:auto; min-height:0; padding-bottom: 16px;">
      <!-- 调度器状态卡片 -->
      <el-card class="status-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>调度器状态</span>
            <el-button type="primary" @click="refreshStatus">刷新</el-button>
          </div>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="活跃任务数量">
            <el-tag type="success">{{ scheduler.active_tasks_count || 0 }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="调度器状态">
            <el-tag type="success">运行中</el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 活跃任务列表 -->
      <el-card class="tasks-card" shadow="never" v-loading="loading">
        <template #header>
          <div class="card-header">
            <span>活跃任务列表</span>
          </div>
        </template>
        <el-table :data="pagedActiveTasks" style="width: 100%" empty-text="暂无活跃任务">
          <el-table-column type="index" label="序号" width="60" />
          <el-table-column prop="task_name" label="任务名称" width="200"/>
          <el-table-column prop="next_run" label="下次运行时间" width="200">
            <template #default="scope">
              {{ formatDateTime(scope.row.next_run) }}
            </template>
          </el-table-column>
          <el-table-column label="剩余时间" width="150">
            <template #default="scope">
              {{ getTimeRemaining(scope.row.next_run) }}
            </template>
          </el-table-column>
        </el-table>
        <el-pagination
          v-model:current-page="activePage"
          :page-size="activePageSize"
          :total="(scheduler.active_tasks || []).length"
          layout="prev, pager, next"
          style="margin-top: 12px; text-align: right;"
        />
      </el-card>

      <!-- 操作按钮 -->
      <el-card class="operation-card" shadow="never">
        <el-button type="primary" @click="reloadScheduler">重新加载调度器</el-button>
        <el-button type="warning" @click="stopScheduler">停止调度器</el-button>
        <el-button type="success" @click="startScheduler">启动调度器</el-button>
      </el-card>
    </el-main>
  </el-container>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { ElMessage } from 'element-plus'

const scheduler = ref({})
const loading = ref(false)
const customTasks = ref([])
const loadingCustom = ref(false)
let refreshTimer = null

// 分页相关
const customPage = ref(1)
const customPageSize = 5
const activePage = ref(1)
const activePageSize = 5

const pagedCustomTasks = computed(() => {
  const start = (customPage.value - 1) * customPageSize
  return customTasks.value.slice(start, start + customPageSize)
})

const pagedActiveTasks = computed(() => {
  const tasks = scheduler.value.active_tasks || []
  const start = (activePage.value - 1) * activePageSize
  return tasks.slice(start, start + activePageSize)
})

function fetchScheduler() {
  loading.value = true
  fetch('/api/scheduler/status')
    .then(res => res.json())
    .then(data => {
      scheduler.value = data
      loading.value = false
    })
    .catch(err => {
      ElMessage.error('获取调度器状态失败')
      console.error(err)
      loading.value = false
    })
}

function fetchCustomTasks() {
  loadingCustom.value = true
  fetch('/api/custom-tasks')
    .then(res => res.json())
    .then(data => {
      customTasks.value = data
      loadingCustom.value = false
    })
    .catch(() => { loadingCustom.value = false })
}

function refreshStatus() {
  fetchScheduler()
}

function reloadScheduler() {
  fetch('/api/scheduler/reload', {
    method: 'POST'
  })
    .then(res => res.json())
    .then(data => {
      ElMessage.success('调度器重新加载成功')
      fetchScheduler()
    })
    .catch(err => {
      ElMessage.error('重新加载调度器失败')
      console.error(err)
    })
}

function stopScheduler() {
  ElMessage.info('停止调度器功能待实现')
}

function startScheduler() {
  ElMessage.info('启动调度器功能待实现')
}

function formatDateTime(dateTimeStr) {
  if (!dateTimeStr) return ''
  const date = new Date(dateTimeStr)
  return date.toLocaleString().replaceAll('/', '-')
}

function getTimeRemaining(nextRun) {
  if (!nextRun) return ''
  const now = new Date()
  const next = new Date(nextRun)
  const diff = next - now
  
  if (diff <= 0) return '即将执行'
  
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
  const seconds = Math.floor((diff % (1000 * 60)) / 1000)
  
  if (hours > 0) {
    return `${hours}小时${minutes}分钟`
  } else if (minutes > 0) {
    return `${minutes}分钟${seconds}秒`
  } else {
    return `${seconds}秒`
  }
}

function toggleTask(row) {
  const url = row.status === 1
    ? `/api/custom-tasks/${row.id}/enable`
    : `/api/custom-tasks/${row.id}/disable`
  fetch(url, { method: 'POST' }).then(fetchCustomTasks)
}

onMounted(() => {
  fetchScheduler()
  fetchCustomTasks()
})

</script>

<style scoped>
.status-card {
  margin-bottom: 12px;
}
.tasks-card {
  margin-bottom: 12px;
}
.custom-tasks-card {
  margin-bottom: 12px;
}
.operation-card {
  margin-bottom: 0;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.empty-state {
  text-align: center;
  padding: 40px 0;
}
.el-card {
  border-radius: 8px;
}
</style>