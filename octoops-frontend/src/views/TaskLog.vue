<template>
  <el-container>
    <el-main class="main-scroll">
      <!-- 查询条件区域 -->
      <el-card class="search-card" shadow="never">
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
      </el-card>
      <!-- 列表区域 -->
      <el-card class="table-card" shadow="never">
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
        </el-table>
        <el-pagination
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="logs.length"
          layout="prev, pager, next"
          style="margin-top: 12px; text-align: right;"
        />
      </el-card>
      <el-dialog v-model="detailDialogVisible" title="日志详情" width="600px">
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
    </el-main>
  </el-container>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import axios from 'axios'

const logs = ref([])
const loading = ref(false)
const detailDialogVisible = ref(false)
const detailLog = ref({})
const searchForm = ref({
  job_id: '',
  task_type: ''
})
const currentPage = ref(1)
const pageSize = ref(10)
const pagedLogs = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return logs.value.slice(start, start + pageSize.value)
})

function fetchLogs(params = {}) {
  loading.value = true
  axios.get('/api/task-logs', { params }).then(res => {
    logs.value = res.data
    loading.value = false
  }).catch(() => {
    loading.value = false
  })
}

function handleSearch() {
  fetchLogs({
    job_id: searchForm.value.job_id,
    task_type: searchForm.value.task_type
  })
}

function handleReset() {
  searchForm.value = { job_id: '', task_type: '' }
  fetchLogs()
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
.operation-card {
  margin-bottom: 16px;
}
.table-card {
  margin-bottom: 16px;
}
.el-card {
  border-radius: 8px;
}
.el-table {
  border-radius: 8px;
}
.main-scroll {
  height: 100%;
  max-height: calc(100vh - 40px);
  overflow: auto;
}
</style> 