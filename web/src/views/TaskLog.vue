<template>
  <el-card>
    <div style="margin-bottom: 16px; display: flex; justify-content: space-between; align-items: center;">
    </div>

    <el-form :inline="true" class="search-form">
      <el-form-item label="任务">
        <el-input v-model="searchForm.task_name" placeholder="请输入任务名称" clearable style="width: 240px" />
      </el-form-item>
      <el-form-item label="运行状态">
        <el-select v-model="searchForm.status" placeholder="请选择状态" clearable style="width: 120px">
          <el-option label="全部" :value="''" />
          <el-option label="成功" value="success" />
          <el-option label="失败" value="failed" />
        </el-select>
      </el-form-item>
      <el-form-item label="创建时间">
        <el-date-picker
          v-model="searchForm.timeRange"
          type="datetimerange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          format="YYYY-MM-DD HH:mm:ss"
          value-format="YYYY-MM-DD HH:mm:ss"
          clearable
          style="width: 360px"
        />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="pagedLogs" style="width: 100%" v-loading="loading" empty-text="暂无数据">
      <el-table-column prop="id" label="ID" width="60"/>
      <el-table-column prop="task_name" label="任务名称" width="200">
        <template #default="scope">
          <span
            @click.stop="showDetail(scope.row)"
            style="color: #409EFF; cursor: pointer;"
            title="点击查看详情"
          >
            {{ scope.row.task_name }}
          </span>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="运行状态" width="100">
        <template #default="scope">
          <el-tag :type="getStatusType(scope.row.status)">{{ getStatusText(scope.row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="scope">
          {{ formatDateTime(scope.row.created_at) }}
        </template>
      </el-table-column>
      <!-- 操作列已移除 -->
    </el-table>
    <el-pagination
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :total="total"
      layout="total, prev, pager, next, jumper"
      @current-change="handlePageChange"
      @size-change="handleSizeChange"
      style="margin-top: 16px; text-align: right;"
    />
    <el-dialog v-model="detailDialogVisible" title="日志详情" width="600px" :close-on-click-modal="false">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="任务名称">{{ detailLog.task_name }}</el-descriptions-item>
        <el-descriptions-item label="返回内容">
          <el-input type="textarea" :rows="8" :model-value="detailLog.result" readonly />
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDateTime(detailLog.created_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import axios from 'axios'

const logs = ref([])
const total = ref(0)
const loading = ref(false)
const detailDialogVisible = ref(false)
const detailLog = ref({})
const searchForm = ref({
  task_name: '',
  status: '',
  timeRange: []
})
const currentPage = ref(1)
const pageSize = ref(10)
const pagedLogs = computed(() => logs.value) // 后端分页，直接用 logs

function fetchLogs(params = {}) {
  loading.value = true
  axios.get('/api/task-logs', { params: { ...params, page: currentPage.value, page_size: pageSize.value } }).then(res => {
    logs.value = res.data.data
    total.value = res.data.total
    loading.value = false
  }).catch(() => {
    loading.value = false
  })
}

function handleSearch() {
  currentPage.value = 1
  const params = {
    task_name: searchForm.value.task_name,
    status: searchForm.value.status
  }
  if (searchForm.value.timeRange && searchForm.value.timeRange.length === 2) {
    params.start_time = searchForm.value.timeRange[0]
    params.end_time = searchForm.value.timeRange[1]
  }
  fetchLogs(params)
}

function handleReset() {
  searchForm.value = { task_name: '', status: '', timeRange: [] }
  currentPage.value = 1
  fetchLogs()
}

function handlePageChange(val) {
  currentPage.value = val
  const params = {
    task_name: searchForm.value.task_name,
    status: searchForm.value.status
  }
  if (searchForm.value.timeRange && searchForm.value.timeRange.length === 2) {
    params.start_time = searchForm.value.timeRange[0]
    params.end_time = searchForm.value.timeRange[1]
  }
  fetchLogs(params)
}

function handleSizeChange(val) {
  pageSize.value = val
  currentPage.value = 1
  const params = {
    task_name: searchForm.value.task_name,
    status: searchForm.value.status
  }
  if (searchForm.value.timeRange && searchForm.value.timeRange.length === 2) {
    params.start_time = searchForm.value.timeRange[0]
    params.end_time = searchForm.value.timeRange[1]
  }
  fetchLogs(params)
}

function showDetail(log) {
  detailLog.value = { ...log }
  detailDialogVisible.value = true
}

function getStatusType(status) {
  const statusMap = {
    'success': 'success',
    'failed': 'danger'
  }
  return statusMap[status] || 'info'
}

function getStatusText(status) {
  const statusMap = {
    'success': '成功',
    'failed': '失败'
  }
  return statusMap[status] || status
}

function formatDateTime(dateTimeStr) {
  if (!dateTimeStr) return ''
  const date = new Date(dateTimeStr)
  return date.toLocaleString().replaceAll('/', '-')
}

onMounted(() => {
  fetchLogs()
})
</script>

<style scoped>
.search-card {
  margin-bottom: 16px;
}
.search-form {
  margin-bottom: 0;
}
.el-card {
  border-radius: 8px;
}
.el-table {
  border-radius: 8px;
}
</style> 