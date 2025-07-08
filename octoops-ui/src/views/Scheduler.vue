<template>
  <div style="padding: 24px;">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <el-button type="primary" @click="refreshStatus">刷新</el-button>
    </div>
    <el-card class="status-card" shadow="never" style="margin-bottom: 16px;">
      <div style="display: flex; align-items: center; gap: 40px; padding: 12px 0;">
        <div style="display: flex; align-items: center;">
          <span style="font-weight: bold; margin-right: 8px;">活跃任务数量</span>
          <el-tag type="success">{{ scheduler.active_tasks_count || 0 }}</el-tag>
        </div>
        <div style="display: flex; align-items: center;">
          <span style="font-weight: bold; margin-right: 8px;">调度器状态</span>
          <el-tag type="success">运行中</el-tag>
        </div>
      </div>
    </el-card>
    <el-card class="tasks-card" shadow="never">
      <template #header>
        <div class="card-header" style="display: flex; justify-content: space-between; align-items: center;">
          <span>活跃任务列表</span>
          <div style="display: flex; flex: 1; align-items: center; justify-content: space-between; margin-left: 24px;">
            <div style="display: flex; align-items: center; gap: 12px;">
              <el-select v-model="taskTypeFilter" placeholder="任务类型" clearable style="width: 120px">
                <el-option label="全部" value="" />
                <el-option label="ETL任务" value="etl" />
                <el-option label="自定义任务" value="custom" />
              </el-select>
            </div>
            <div style="display: flex; align-items: center; gap: 12px;">
              <el-button type="primary" @click="reloadScheduler">重新加载调度器</el-button>
              <el-button type="warning" @click="stopScheduler">停止调度器</el-button>
              <el-button type="success" @click="startScheduler">启动调度器</el-button>
            </div>
          </div>
        </div>
      </template>
      <el-table :data="filteredPagedActiveTasks" style="width: 100%" empty-text="暂无活跃任务">
        <el-table-column type="index" label="序号" width="60" />
        <el-table-column prop="task_name" label="任务名称" width="200"/>
        <el-table-column label="任务类型" width="120">
          <template #default="scope">
            <el-tag :type="scope.row.task_type === 'etl' ? 'success' : 'info'">
              {{ scope.row.task_type === 'etl' ? 'ETL任务' : '自定义任务' }}
            </el-tag>
          </template>
        </el-table-column>
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
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { ElMessage } from 'element-plus'

const scheduler = ref({})
const loading = ref(false)
let refreshTimer = null

// 分页相关
const activePage = ref(1)
const activePageSize = 10

const taskTypeFilter = ref("")

const filteredActiveTasks = computed(() => {
  const tasks = scheduler.value.active_tasks || []
  if (!taskTypeFilter.value) return tasks
  return tasks.filter(t => t.task_type === taskTypeFilter.value)
})

const filteredPagedActiveTasks = computed(() => {
  const tasks = filteredActiveTasks.value
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
  fetch('/api/scheduler/stop', {
    method: 'POST'
  })
    .then(res => res.json())
    .then(data => {
      ElMessage.success('调度器已停止')
      fetchScheduler()
    })
    .catch(err => {
      ElMessage.error('停止调度器失败')
      console.error(err)
    })
}

function startScheduler() {
  fetch('/api/scheduler/start', {
    method: 'POST'
  })
    .then(res => res.json())
    .then(data => {
      ElMessage.success('调度器已启动')
      fetchScheduler()
    })
    .catch(err => {
      ElMessage.error('启动调度器失败')
      console.error(err)
    })
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

onMounted(() => {
  fetchScheduler()
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
