<template>
  <div style="padding: 24px;">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <el-button type="primary" @click="fetchCustomTasks">刷新</el-button>
    </div>
    <el-table :data="pagedCustomTasks" style="width: 100%" v-loading="loadingCustom">
      <el-table-column type="index" label="序号" width="60" />
      <el-table-column prop="name" label="任务名称" />
      <el-table-column prop="custom_type" label="类型" />
      <el-table-column prop="cron_expr" label="调度周期" />
      <el-table-column prop="status" label="状态">
        <template #default="scope">
          <el-switch v-model="scope.row.status" :active-value="1" :inactive-value="0" @change="toggleTask(scope.row)" />
        </template>
      </el-table-column>
      <el-table-column prop="last_run_time" label="上次执行">
        <template #default="scope">
          {{ formatDateTime(scope.row.last_run_time) }}
        </template>
      </el-table-column>
      <el-table-column prop="last_result" label="上次结果" />
    </el-table>
    <el-pagination
      v-model:current-page="customPage"
      :page-size="customPageSize"
      :total="customTasks.length"
      layout="prev, pager, next"
      style="margin-top: 12px; text-align: right;"
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'

const customTasks = ref([])
const loadingCustom = ref(false)

// 分页相关
const customPage = ref(1)
const customPageSize = 5

const pagedCustomTasks = computed(() => {
  const start = (customPage.value - 1) * customPageSize
  return customTasks.value.slice(start, start + customPageSize)
})

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

function toggleTask(row) {
  fetch(`/api/custom-tasks/${row.id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status: row.status })
  })
    .then(res => res.json())
    .then(() => {
      ElMessage.success(row.status === 1 ? '已启用' : '已禁用')
      fetchCustomTasks()
    })
    .catch(() => {
      ElMessage.error('操作失败')
      fetchCustomTasks()
    })
}

function formatDateTime(dateTimeStr) {
  if (!dateTimeStr) return ''
  const date = new Date(dateTimeStr)
  return date.toLocaleString().replaceAll('/', '-')
}

fetchCustomTasks()
</script> 