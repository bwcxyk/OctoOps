<template>
  <el-container>
    <el-main class="main-scroll">
      <!-- 查询条件区域 -->
      <el-card class="search-card" shadow="never">
        <el-form :inline="true" class="search-form">
          <el-form-item label="作业名称">
            <el-input v-model="searchForm.name" placeholder="请输入作业名称" clearable />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="searchForm.status" placeholder="请选择状态" clearable style="width: 120px">
              <el-option label="启用" :value="1" />
              <el-option label="禁用" :value="0" />
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
      </el-card>

      <!-- 列表区域 -->
      <el-card class="table-card" shadow="never">
        <el-table :data="tasks" style="width: 100%" @row-click="null" v-loading="loading" empty-text="暂无数据">
          <el-table-column type="index" label="序号" width="60"/>
          <el-table-column prop="name" label="作业名称" width="230" >
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
          <el-table-column prop="description" label="描述" width="220" show-overflow-tooltip/>
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-switch
                v-model="scope.row.status"
                :active-value="1"
                :inactive-value="0"
                @change="handleStatusChange(scope.row)"
              />
            </template>
          </el-table-column>
          <el-table-column prop="cron_expr" label="Cron表达式" width="180" />
          <el-table-column label="下次执行时间" width="180">
            <template #default="scope">
              {{ scope.row.next_run_time ? formatDateTime(scope.row.next_run_time) : '未设置' }}
            </template>
          </el-table-column>
          <el-table-column prop="last_run_time" label="最后运行时间" width="180">
            <template #default="scope">
              {{ scope.row.last_run_time ? formatDateTime(scope.row.last_run_time) : '未运行' }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="scope">
              <el-space>
                <el-button size="small" @click.stop="goToEditTask(scope.row.id)">编辑</el-button>
                <el-button size="small" type="primary" @click.stop="handleManualExecute(scope.row)">手动执行</el-button>
                <el-button size="small" type="danger" @click.stop="handleDelete(scope.row.id)">删除</el-button>
              </el-space>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </el-main>
    <el-dialog v-model="dialogVisible" :title="editTask.id ? '编辑作业' : '新建作业'" width="80%">
      <el-form :model="editTask" :rules="rules" ref="formRef" label-width="200px" label-position="top">
        <el-form-item label="作业名称" prop="name">
          <el-input v-model="editTask.name"/>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="editTask.description" type="textarea" :rows="3" placeholder="请输入任务描述"/>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch
            v-model="editTask.status"
            :active-value="1"
            :inactive-value="0"
          />
        </el-form-item>
        <el-form-item label="Cron表达式" prop="cron_expr">
          <el-input v-model="editTask.cron_expr" placeholder="秒 分 时 日 月 周" />
        </el-form-item>
        <el-form-item label="配置风格" prop="config_format">
          <el-select v-model="editTask.config_format" placeholder="请选择配置风格">
            <el-option label="JSON" value="json" />
            <el-option label="HOCON" value="hocon" />
          </el-select>
        </el-form-item>
        <el-form-item label="作业配置" prop="config">
          <el-input type="textarea" v-model="editTask.config" :rows="8" placeholder="请输入完整作业配置（json或hocon）"/>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="detailDialogVisible" title="任务详情" width="600px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="作业名称">{{ detailTask.name }}</el-descriptions-item>
        <el-descriptions-item label="描述">{{ detailTask.description || '无' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="detailTask.status === 1 ? 'success' : 'info'">
            {{ detailTask.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="类型">{{ detailTask.task_type }}</el-descriptions-item>
        <el-descriptions-item label="Cron表达式">{{ detailTask.cron_expr || '无' }}</el-descriptions-item>
        <el-descriptions-item label="下次执行时间">
          {{ detailTask.next_run_time ? formatDateTime(detailTask.next_run_time) : '未设置' }}
        </el-descriptions-item>
        <el-descriptions-item label="最后运行时间">
          {{ detailTask.last_run_time ? formatDateTime(detailTask.last_run_time) : '未运行' }}
        </el-descriptions-item>
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
import { getTasks, createTask, updateTask, deleteTask, submitJob } from '../../../api/task'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CronExpressionParser } from 'cron-parser'
import { useRouter } from 'vue-router'

const tasks = ref([])
const dialogVisible = ref(false)
const editTask = ref({})
const detailDialogVisible = ref(false)
const detailTask = ref({})
const formRef = ref()
const loading = ref(false)
const searchForm = ref({
  name: '',
  status: ''
})
const rules = {
  name: [
    { required: true, message: '任务名称为必填', trigger: 'blur' }
  ],
  config: [
    { required: true, message: '作业配置为必填', trigger: 'blur' }
  ],
  cron_expr: [
    {
      validator: (rule, value, callback) => {
        const expr = value.trim()
        try {
          CronExpressionParser.parse(expr)
          callback()
        } catch (e) {
          callback(new Error('无效的Cron表达式'))
        }
      },
      trigger: 'blur'
    }
  ]
}
const router = useRouter()

function fetchTasks() {
  loading.value = true
  getTasks('batch').then(res => {
    tasks.value = res.data
    loading.value = false
  }).catch(() => {
    loading.value = false
  })
}

function handleSearch() {
  loading.value = true
  getTasks('batch', searchForm.value).then(res => {
    tasks.value = res.data
    loading.value = false
  }).catch(() => {
    loading.value = false
  })
}

function handleReset() {
  searchForm.value = {
    name: '',
    status: ''
  }
}

function openDialog(task = null) {
  if (task) {
    editTask.value = { ...task }
  } else {
    editTask.value = {
      name: '',
      description: '',
      cron_expr: '',
      config: '',
      config_format: 'json',
      task_type: 'batch',
      status: 1
    }
  }
  dialogVisible.value = true
}

function handleSave() {
  formRef.value.validate(valid => {
    if (!valid) return
    // 只取业务字段
    const payload = {
      id: editTask.value.id,
      name: editTask.value.name,
      description: editTask.value.description,
      cron_expr: editTask.value.cron_expr,
      config: editTask.value.config,
      config_format: editTask.value.config_format,
      task_type: editTask.value.task_type,
      status: editTask.value.status
      // 其他需要的字段
    }
    if (editTask.value.id) {
      updateTask(editTask.value.id, payload).then(() => {
        ElMessage.success('更新成功')
        dialogVisible.value = false
        fetchTasks()
      })
    } else {
      createTask(payload).then(() => {
        ElMessage.success('创建成功')
        dialogVisible.value = false
        fetchTasks()
      })
    }
  })
}

function handleDelete(id) {
  ElMessageBox.confirm('确定要删除该任务吗？', '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(() => {
    deleteTask(id).then(() => {
      ElMessage.success('删除成功')
      fetchTasks()
    })
  })
}

function showDetail(task) {
  detailTask.value = { ...task }
  detailDialogVisible.value = true
}

function formatDateTime(dateTimeStr) {
  if (!dateTimeStr) return ''
  const date = new Date(dateTimeStr)
  return date.toLocaleString('zh-CN', { hour12: false }).replaceAll('/', '-')
}

function handleStatusChange(row) {
  if (row.status === 1 && (!row.cron_expr || !row.cron_expr.trim())) {
    ElMessage.error('离线任务启用时必须填写 Cron 表达式')
    row.status = 0
    return
  }
  updateTask(row.id, { status: row.status }).then(() => {
    ElMessage.success('状态更新成功')
    fetchTasks()
  })
}

function handleManualExecute(task) {
  ElMessageBox.confirm(`确定要手动执行任务"${task.name}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(() => {
    // 调用提交作业API，和实时任务保持一致
    submitJob(
      '',
      { params: { id: task.id, isStartWithSavePoint: false } }
    )
    .then(res => {
      if (res.data && res.data.error) {
        ElMessage.error(`执行失败: ${res.data.error}`)
      } else {
        ElMessage.success('手动执行成功')
        // 刷新任务列表以更新最后运行时间
        fetchTasks()
      }
    })
    .catch(err => {
      // 显示后端返回的具体错误信息
      const errorMsg = err.response?.data?.error || err.message || '手动执行失败'
      ElMessage.error(errorMsg)
      console.error(err)
    })
  })
}

function goToNewTask() {
  router.push({ name: 'BatchTaskNew' })
}

function goToEditTask(id) {
  router.push({ name: 'BatchTaskEdit', params: { id } })
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

.main-scroll {
  height: 100%;
  max-height: calc(100vh - 40px);
  overflow: auto;
}
</style> 