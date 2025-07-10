<template>
  <div style="padding: 24px;">
    <el-button type="primary" @click="openDialog()" style="margin-bottom: 16px;">新增</el-button>
    <el-table :data="alerts" style="width: 100%" v-loading="loading" empty-text="暂无数据">
      <el-table-column type="index" label="序号" width="60" />
      <el-table-column prop="name" label="名称" width="140" />
      <el-table-column prop="target" label="邮箱" width="220" />
      <el-table-column prop="status" label="启用" width="80">
        <template #default="scope">
          <el-switch v-model="scope.row.status" :active-value="1" :inactive-value="0" />
        </template>
      </el-table-column>
      <el-table-column prop="updated_at" label="更新时间" width="160">
        <template #default="scope">
          <span>{{ formatDateTime(scope.row.updated_at) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="scope">
          <el-button size="small" @click="openDialog(scope.row)">编辑</el-button>
          <el-button size="small" type="primary" @click="handleTest(scope.row)">测试</el-button>
          <el-button size="small" type="danger" @click="handleDelete(scope.row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :total="total"
      layout="total, prev, pager, next, jumper"
      style="margin-top: 16px; text-align: right;"
    />
    <el-dialog v-model="dialogVisible" :title="editNotif.id ? '编辑邮件通知' : '新增邮件通知'" width="500px">
      <el-form :model="editNotif" label-width="100px" :rules="rules" ref="formRef">
        <el-form-item label="名称" prop="name">
          <el-input v-model="editNotif.name" />
        </el-form-item>
        <el-form-item label="邮箱" prop="target">
          <el-input v-model="editNotif.target" placeholder="请输入邮箱地址" />
        </el-form-item>
        <el-form-item label="启用" prop="status">
          <el-switch v-model="editNotif.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="告警模板" prop="template_id">
          <el-select v-model="editNotif.template_id" placeholder="请选择告警模板">
            <el-option v-for="tpl in templates" :key="tpl.id" :label="tpl.name" :value="tpl.id" />
            <template #empty>暂无数据</template>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getAlerts, createAlert, updateAlert, deleteAlert } from '../../api/alert'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'

const alerts = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const editNotif = ref({})
const formRef = ref()
const rules = {
  name: [{ required: true, message: '名称必填', trigger: 'blur' }],
  target: [{ required: true, message: '邮箱必填', trigger: 'blur' }],
}
const templates = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

function fetchAlerts() {
  loading.value = true
  getAlerts({ page: page.value, size: pageSize.value }).then(res => {
    const data = res.data.items || res.data
    alerts.value = data.filter(n => n.type === 'email')
    total.value = res.data.total || data.length || 0
    loading.value = false
  }).catch(() => {
    loading.value = false
  })
}

function fetchTemplates() {
  axios.get('/api/alert-templates').then(res => {
    templates.value = res.data.filter(t => t.type === 'email')
  })
}

function openDialog(row = null) {
  if (row) {
    editNotif.value = { ...row }
  } else {
    editNotif.value = { name: '', type: 'email', target: '', status: 1, template_id: null }
  }
  dialogVisible.value = true
  fetchTemplates()
}

function handleSave() {
  formRef.value.validate(valid => {
    if (!valid) return
    if (editNotif.value.id) {
      updateAlert(editNotif.value.id, editNotif.value).then(() => {
        ElMessage.success('更新成功')
        dialogVisible.value = false
        fetchAlerts()
      })
    } else {
      createAlert(editNotif.value).then(() => {
        ElMessage.success('创建成功')
        dialogVisible.value = false
        fetchAlerts()
      })
    }
  })
}

function handleDelete(id) {
  ElMessageBox.confirm('确定要删除该邮件通知吗？', '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(() => {
    deleteAlert(id).then(() => {
      ElMessage.success('删除成功')
      fetchAlerts()
    })
  })
}

function handleTest(row) {
  axios.post(`/api/alerts/${row.id}/test`).then(res => {
    ElMessage.success(res.data.message || '测试发送成功')
  }).catch(err => {
    ElMessage.error(err.response?.data?.error || '测试发送失败')
  })
}

function formatDateTime(dateTimeStr) {
  if (!dateTimeStr) return ''
  const date = new Date(dateTimeStr)
  return date.toLocaleString('zh-CN', { hour12: false }).replaceAll('/', '-')
}

onMounted(fetchAlerts)
</script> 