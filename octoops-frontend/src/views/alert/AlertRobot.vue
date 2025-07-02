<template>
  <div style="padding: 24px;">
    <el-button type="primary" @click="openDialog()" style="margin-bottom: 16px;">新增</el-button>
    <el-table :data="alerts" style="width: 100%" v-loading="loading" empty-text="暂无数据">
      <el-table-column type="index" label="序号" width="60" />
      <el-table-column prop="name" label="名称" width="140" />
      <el-table-column prop="type" label="平台" width="120">
        <template #default="scope">
          <span>{{ platformLabel(scope.row.type) }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="target" label="Webhook地址" width="260" show-overflow-tooltip />
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
    <el-dialog v-model="dialogVisible" :title="editNotif.id ? '编辑机器人通知' : '新增机器人通知'" width="500px">
      <el-form :model="editNotif" label-width="100px" :rules="rules" ref="formRef">
        <el-form-item label="机器人名称" prop="name">
          <el-input v-model="editNotif.name" />
        </el-form-item>
        <el-form-item label="机器人平台" prop="type">
          <el-select v-model="editNotif.type" placeholder="请选择平台">
            <el-option label="钉钉" value="dingtalk" />
            <el-option label="企业微信" value="wechat" />
            <el-option label="飞书" value="feishu" />
          </el-select>
        </el-form-item>
        <el-form-item label="Webhook" prop="target">
          <el-input v-model="editNotif.target" type="textarea" :rows="3" placeholder="请输入机器人Webhook地址" />
        </el-form-item>
        <el-form-item v-if="editNotif.type === 'dingtalk'" label="加签密钥" prop="dingtalk_secret">
          <el-input v-model="editNotif.dingtalk_secret" placeholder="请输入钉钉加签密钥" />
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
  type: [{ required: true, message: '平台必选', trigger: 'change' }],
  target: [{ required: true, message: 'Webhook地址必填', trigger: 'blur' }],
}
const templates = ref([])

function fetchAlerts() {
  loading.value = true
  getAlerts().then(res => {
    alerts.value = res.data.filter(n => n.type === 'robot')
    loading.value = false
  }).catch(() => {
    loading.value = false
  })
}

function openDialog(row = null) {
  if (row) {
    editNotif.value = { ...row }
  } else {
    editNotif.value = { name: '', type: '', target: '', status: 1, dingtalk_secret: '', template_id: null }
  }
  dialogVisible.value = true
  fetchTemplates()
}

function fetchTemplates() {
  axios.get('/api/alert-templates').then(res => {
    templates.value = res.data.filter(t => t.type === 'dingtalk' || t.type === 'wechat' || t.type === 'feishu')
  })
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
  ElMessageBox.confirm('确定要删除该机器人通知吗？', '提示', {
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

function formatDateTime(dateTimeStr) {
  if (!dateTimeStr) return ''
  const date = new Date(dateTimeStr)
  return date.toLocaleString('zh-CN', { hour12: false }).replaceAll('/', '-')
}

function platformLabel(val) {
  if (val === 'dingtalk') return '钉钉'
  if (val === 'feishu') return '飞书'
  if (val === 'wechat') return '企业微信'
  return val
}

function handleTest(row) {
  axios.post(`/api/alerts/${row.id}/test`).then(res => {
    ElMessage.success(res.data.message || '测试发送成功')
  }).catch(err => {
    ElMessage.error(err.response?.data?.error || '测试发送失败')
  })
}

onMounted(fetchAlerts)
</script> 