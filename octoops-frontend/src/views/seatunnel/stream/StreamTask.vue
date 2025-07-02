<template>
  <el-container>
    <el-main>
      <!-- 查询条件区域 -->
      <el-card class="search-card" shadow="never">
        <el-form :inline="true" class="search-form">
          <el-form-item label="作业ID">
            <el-input v-model="searchForm.jobid" placeholder="请输入作业ID" clearable />
          </el-form-item>
          <el-form-item label="作业名称">
            <el-input v-model="searchForm.name" placeholder="请输入作业名称" clearable />
          </el-form-item>
          <el-form-item label="作业状态">
            <el-select v-model="searchForm.job_status" placeholder="请选择作业状态" clearable style="width: 150px">
              <el-option label="全部" :value="''" />
              <el-option label="运行中" :value="'RUNNING'" />
              <el-option label="已完成" :value="'FINISHED'" />
              <el-option label="失败" :value="'FAILED'" />
              <el-option label="已取消" :value="'CANCEL'" />
              <el-option label="未知" :value="'UNKNOWN'" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <!-- 操作按钮区域 -->
      <el-card class="operation-card" shadow="never">
        <el-button type="primary" @click="goToNewTask">新增</el-button>
        <el-button type="success" @click="handleSyncJobStatus" :loading="syncing">同步作业状态</el-button>
      </el-card>

      <!-- 列表区域 -->
      <el-card class="table-card" shadow="never">
        <el-table :data="tasks" style="width: 100%" @row-click="null" v-loading="loading" empty-text="暂无数据">
          <el-table-column type="index" label="序号" width="60"/>
          <el-table-column prop="jobid" label="作业ID" width="300"/>
          <el-table-column prop="name" label="作业名称" width="200">
            <template #default="scope">
              <span
                @click.stop="showDetail(scope.row)"
                style="color: #409EFF; cursor: pointer;"
                title="点击查看详情"
              >
                {{ scope.row.name }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="description" label="描述" width="250" show-overflow-tooltip/>
          <el-table-column label="作业状态" width="150">
            <template #default="scope">
              <el-tag :type="jobStatusTagType(scope.row.job_status)">
                {{ jobStatusText(scope.row.job_status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="320" fixed="right">
            <template #default="scope">
              <el-button size="small" @click.stop="goToEditTask(scope.row.id)">编辑</el-button>
              <el-button size="small" type="primary" @click.stop="handleSubmitJob(scope.row)">提交作业</el-button>
              <el-button size="small" type="warning" @click.stop="handleStopJob(scope.row)">停止作业</el-button>
              <el-button size="small" type="danger" @click.stop="handleDelete(scope.row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </el-main>
    <el-dialog v-model="submitDialogVisible" title="提交作业" width="400px">
      <el-form>
        <el-form-item label="是否使用 SavePoint 启动">
          <el-switch v-model="isStartWithSavePoint" active-text="是" inactive-text="否" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="submitDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="doSubmitJob">确定</el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="stopDialogVisible" title="停止作业" width="400px">
      <el-form>
        <el-form-item label="是否使用 SavePoint 停止">
          <el-switch v-model="isStopWithSavePoint" active-text="是" inactive-text="否" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="stopDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="doStopJob">确定</el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="detailDialogVisible" title="任务详情" width="600px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="作业ID">{{ detailTask.jobid }}</el-descriptions-item>
        <el-descriptions-item label="作业名称">{{ detailTask.name }}</el-descriptions-item>
        <el-descriptions-item label="描述">{{ detailTask.description || '无' }}</el-descriptions-item>
        <el-descriptions-item label="作业状态">
          <el-tag :type="jobStatusTagType(detailTask.job_status)">
            {{ jobStatusText(detailTask.job_status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="类型">{{ detailTask.task_type }}</el-descriptions-item>
        <el-descriptions-item label="作业配置">
          <el-input type="textarea" :rows="10" v-model="detailTask.config" readonly />
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDateTime(detailTask.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatDateTime(detailTask.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </el-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getTasks, createTask, updateTask, deleteTask, submitJob, stopJob } from '../../../api/task'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'

const tasks = ref([])
const submitDialogVisible = ref(false)
const stopDialogVisible = ref(false)
const isStartWithSavePoint = ref(false)
const isStopWithSavePoint = ref(false)
const detailDialogVisible = ref(false)
const detailTask = ref({})
const loading = ref(false)
const searchForm = ref({
  name: '',
  job_status: '',
  jobid: ''
})
const rules = {
  name: [
    { required: true, message: '任务名称为必填', trigger: 'blur' }
  ],
  config: [
    { required: true, message: '作业配置为必填', trigger: 'blur' }
  ]
}
const syncing = ref(false)
const router = useRouter()
let currentTask = null

function fetchTasks() {
  loading.value = true
  getTasks('stream').then(res => {
    tasks.value = res.data
    loading.value = false
  }).catch(() => {
    loading.value = false
  })
}

function handleSearch() {
  loading.value = true
  getTasks('stream', searchForm.value).then(res => {
    tasks.value = res.data
    loading.value = false
  }).catch(() => {
    loading.value = false
  })
}

function handleReset() {
  searchForm.value = {
    name: '',
    job_status: '',
    jobid: ''
  }
}

function goToNewTask() {
  router.push({ name: 'StreamTaskNew' })
}

function goToEditTask(id) {
  router.push({ name: 'StreamTaskEdit', params: { id } })
}

function handleSave() {
  formRef.value.validate(valid => {
    if (!valid) return
    if (editTask.value.id) {
      updateTask(editTask.value.id, editTask.value).then(() => {
        ElMessage.success('更新成功')
        dialogVisible.value = false
        fetchTasks()
      })
    } else {
      createTask(editTask.value).then(() => {
        ElMessage.success('创建成功')
        dialogVisible.value = false
        fetchTasks()
      })
    }
  })
}

function handleDelete(id) {
  ElMessageBox.confirm('确定要删除该任务吗？', '提示', {
    type: 'warning'
  }).then(() => {
    deleteTask(id).then(() => {
      ElMessage.success('删除成功')
      fetchTasks()
    })
  })
}

function handleSubmitJob(task) {
  currentTask = task
  isStartWithSavePoint.value = false
  submitDialogVisible.value = true
}

function doSubmitJob() {
  if (!currentTask.id) {
    ElMessage.error('任务ID不能为空')
    return
  }
  submitJob(
    "",
    { params: { isStartWithSavePoint: isStartWithSavePoint.value, id: currentTask.id } }
  ).then(res => {
    ElMessage.success('提交作业成功')
    submitDialogVisible.value = false
  }).catch(() => {
    ElMessage.error('提交作业失败')
  })
}

function handleStopJob(task) {
  currentTask = task
  isStopWithSavePoint.value = false
  stopDialogVisible.value = true
}

function doStopJob() {
  if (!currentTask.id) {
    ElMessage.error('任务ID不能为空')
    return
  }
  stopJob(
    null,
    { params: { isStopWithSavePoint: isStopWithSavePoint.value, id: currentTask.id } }
  ).then(res => {
    ElMessage.success('停止作业成功')
    stopDialogVisible.value = false
  }).catch(() => {
    ElMessage.error('停止作业失败')
  })
}

function showDetail(task) {
  detailTask.value = { ...task }
  detailDialogVisible.value = true
}

function formatDateTime(dateTimeStr) {
  if (!dateTimeStr) return ''
  const date = new Date(dateTimeStr)
  return date.toLocaleString().replaceAll('/', '-')
}

function jobStatusTagType(status) {
  switch (status) {
    case 'RUNNING': return 'success'
    case 'FINISHED': return 'info'
    case 'FAILED': return 'danger'
    case 'CANCEL': return 'warning'
    default: return ''
  }
}

function jobStatusText(status) {
  switch (status) {
    case 'RUNNING': return '运行中'
    case 'FINISHED': return '已完成'
    case 'FAILED': return '失败'
    case 'CANCEL': return '已取消'
    case 'UNKNOWN': return '未知'
    default: return status || '未知'
  }
}

function handleSyncJobStatus() {
  syncing.value = true
  fetch('/api/sync-job-status', { method: 'POST' })
    .then(res => res.json())
    .then(() => {
      ElMessage.success('同步作业状态已触发')
      fetchTasks()
    })
    .catch(() => {
      ElMessage.error('同步失败')
    })
    .finally(() => {
      syncing.value = false
    })
}

onMounted(fetchTasks)
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

.el-form-item {
  margin-bottom: 16px;
}

.el-table {
  border-radius: 8px;
}
</style> 