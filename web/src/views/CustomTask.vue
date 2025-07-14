<template>
  <div style="padding: 24px;">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <el-button type="primary" @click="fetchCustomTasks">刷新</el-button>
    </div>
    <el-table :data="customTasks" style="width: 100%" v-loading="loading">
      <el-table-column type="index" label="序号" width="60" />
      <el-table-column prop="name" label="任务名称" width="180" />
      <el-table-column prop="type" label="类型" width="180" />
      <el-table-column prop="cron_expr" label="调度周期" width="180" />
      <el-table-column prop="status" label="状态" width="120" >
        <template #default="scope">
          <el-switch v-model="scope.row.status" :active-value="1" :inactive-value="0" @change="toggleTask(scope.row)" />
        </template>
      </el-table-column>
      <el-table-column prop="last_run_time" label="上次执行" width="180" >
        <template #default="scope">
          {{ formatDateTime(scope.row.last_run_time) }}
        </template>
      </el-table-column>
      <el-table-column prop="last_result" label="上次结果" width="250" show-overflow-tooltip />
    </el-table>
    <el-pagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :total="total"
      layout="total, prev, pager, next, jumper"
      style="margin-top: 16px; text-align: right;"
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'

const customTasks = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

function fetchCustomTasks() {
  loading.value = true
  axios.get('/api/custom-tasks', { params: { page: page.value, size: pageSize.value } }).then(res => {
    // 适配后端返回 { data: [...], total: n }
    customTasks.value = Array.isArray(res.data.data) ? res.data.data : []
    total.value = typeof res.data.total === 'number' ? res.data.total : 0
    loading.value = false
  }).catch(() => { loading.value = false })
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

function handlePageChange(val) {
  page.value = val
  fetchCustomTasks()
}

fetchCustomTasks()
</script> 