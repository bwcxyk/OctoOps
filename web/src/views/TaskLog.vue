<template>
  <el-card>
    <div style="margin-bottom: 16px; display: flex; justify-content: space-between; align-items: center;">
    </div>

    <el-form :inline="true" class="search-form">
      <el-form-item label="作业ID">
        <el-input v-model="searchForm.job_id" placeholder="请输入作业ID" clearable />
      </el-form-item>
      <el-form-item label="作业类型">
        <el-select v-model="searchForm.task_type" placeholder="请选择类型" clearable style="width: 120px">
          <el-option label="全部" :value="''" />
          <el-option label="batch" value="batch" />
          <el-option label="stream" value="stream" />
          <el-option label="custom" value="custom" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="pagedLogs" style="width: 100%" v-loading="loading" empty-text="暂无数据">
      <el-table-column type="index" label="序号" width="60"/>
      <el-table-column prop="job_id" label="作业ID" width="200" />
      <el-table-column prop="job_name" label="作业名称" width="200">
        <template #default="scope">
          <span
            @click.stop="showDetail(scope.row)"
            style="color: #409EFF; cursor: pointer;"
            title="点击查看详情"
          >
            {{ scope.row.job_name }}
          </span>
        </template>
      </el-table-column>
      <el-table-column prop="task_type" label="作业类型" width="100">
        <template #default="scope">
          <span>{{ scope.row.task_type }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="提交时间" width="180">
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
        <el-descriptions-item label="作业ID">{{ detailLog.job_id }}</el-descriptions-item>
        <el-descriptions-item label="作业名称">{{ detailLog.job_name }}</el-descriptions-item>
        <el-descriptions-item label="返回内容">
          <el-input type="textarea" :rows="8" :model-value="detailLog.result" readonly />
        </el-descriptions-item>
        <el-descriptions-item label="提交时间">{{ formatDateTime(detailLog.created_at) }}</el-descriptions-item>
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
  job_id: '',
  task_type: ''
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
  fetchLogs({
    job_id: searchForm.value.job_id,
    task_type: searchForm.value.task_type
  })
}

function handleReset() {
  searchForm.value = { job_id: '', task_type: '' }
  currentPage.value = 1
  fetchLogs()
}

function handlePageChange(val) {
  currentPage.value = val
  fetchLogs({
    job_id: searchForm.value.job_id,
    task_type: searchForm.value.task_type
  })
}

function handleSizeChange(val) {
  pageSize.value = val
  currentPage.value = 1
  fetchLogs({
    job_id: searchForm.value.job_id,
    task_type: searchForm.value.task_type
  })
}

function showDetail(log) {
  detailLog.value = { ...log }
  detailDialogVisible.value = true
}

function formatDateTime(dateTimeStr) {
  if (!dateTimeStr) return ''
  const date = new Date(dateTimeStr)
  return date.toLocaleString().replaceAll('/', '-')
}

onMounted(() => fetchLogs())
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