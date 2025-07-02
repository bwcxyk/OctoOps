<template>
  <div style="padding: 24px;">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <el-button type="primary" @click="openDialog()">新增</el-button>
    </div>
    <el-table :data="configs" style="width: 100%" v-loading="loading" empty-text="暂无数据">
      <el-table-column type="index" label="序号" width="60" />
      <el-table-column prop="account_name" label="账号名" width="120" />
      <el-table-column prop="access_key" label="AccessKey" width="220" />
      <el-table-column prop="region_id" label="RegionId" width="120" />
      <el-table-column prop="security_group_id" label="安全组ID" width="200" />
      <el-table-column prop="port_list" label="端口列表" width="150" show-overflow-tooltip />
      <el-table-column prop="last_ip" label="最近授权IP" width="140" />
      <el-table-column prop="last_ip_updated_at" label="IP更新时间" width="160">
        <template #default="scope">
          <span>{{ formatDateTime(scope.row.last_ip_updated_at) }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="updated_at" label="配置更新时间" width="160">
        <template #default="scope">
          <span>{{ formatDateTime(scope.row.updated_at) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="scope">
          <el-button size="small" type="success" @click="handleSyncOne(scope.row)">同步</el-button>
          <el-button size="small" @click="openDialog(scope.row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(scope.row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-dialog v-model="dialogVisible" :title="editConfig.id ? '编辑配置' : '新增配置'" width="500px">
      <el-form :model="editConfig" label-width="120px" :rules="rules" ref="formRef">
        <el-form-item label="账号名" prop="account_name">
          <el-input v-model="editConfig.account_name" />
        </el-form-item>
        <el-form-item label="AccessKey" prop="access_key">
          <el-input v-model="editConfig.access_key" />
        </el-form-item>
        <el-form-item label="AccessSecret" prop="access_secret">
          <el-input v-model="editConfig.access_secret" type="password" show-password />
        </el-form-item>
        <el-form-item label="RegionId" prop="region_id">
          <el-input v-model="editConfig.region_id" />
        </el-form-item>
        <el-form-item label="安全组ID" prop="security_group_id">
          <el-input v-model="editConfig.security_group_id" />
        </el-form-item>
        <el-form-item label="端口列表" prop="port_list">
          <el-input v-model="editConfig.port_list" placeholder="如 22,80,443" />
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
import { getAliyunSGConfigs, createAliyunSGConfig, updateAliyunSGConfig, deleteAliyunSGConfig, syncAliyunSGConfigs, syncAliyunSGConfig } from '../api/aliyun'
import { ElMessage, ElMessageBox } from 'element-plus'

const configs = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const editConfig = ref({})
const formRef = ref()
const rules = {
  account_name: [{ required: true, message: '账号名必填', trigger: 'blur' }],
  access_key: [{ required: true, message: 'AccessKey必填', trigger: 'blur' }],
  access_secret: [{ required: true, message: 'AccessSecret必填', trigger: 'blur' }],
  region_id: [{ required: true, message: 'RegionId必填', trigger: 'blur' }],
  security_group_id: [{ required: true, message: '安全组ID必填', trigger: 'blur' }],
  port_list: [{ required: true, message: '端口列表必填', trigger: 'blur' }],
}

function fetchConfigs() {
  loading.value = true
  getAliyunSGConfigs().then(res => {
    configs.value = res.data
    loading.value = false
  }).catch(() => {
    loading.value = false
  })
}

function openDialog(row = null) {
  if (row) {
    editConfig.value = { ...row }
  } else {
    editConfig.value = {
      account_name: '', access_key: '', access_secret: '', region_id: '', security_group_id: '', port_list: ''
    }
  }
  dialogVisible.value = true
}

function handleSave() {
  formRef.value.validate(valid => {
    if (!valid) return
    if (editConfig.value.id) {
      updateAliyunSGConfig(editConfig.value.id, editConfig.value).then(() => {
        ElMessage.success('更新成功')
        dialogVisible.value = false
        fetchConfigs()
      })
    } else {
      createAliyunSGConfig(editConfig.value).then(() => {
        ElMessage.success('创建成功')
        dialogVisible.value = false
        fetchConfigs()
      })
    }
  })
}

function handleDelete(id) {
  ElMessageBox.confirm('确定要删除该配置吗？', '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(() => {
    deleteAliyunSGConfig(id).then(() => {
      ElMessage.success('删除成功')
      fetchConfigs()
    })
  })
}

function handleSync() {
  loading.value = true
  syncAliyunSGConfigs().then(() => {
    ElMessage.success('同步成功')
    fetchConfigs()
  }).catch(() => {
    ElMessage.error('同步失败')
    loading.value = false
  })
}

function handleSyncOne(row) {
  loading.value = true
  syncAliyunSGConfig(row.id).then(res => {
    loading.value = false
    ElMessage.success('同步成功')
    fetchConfigs()
  }).catch(err => {
    loading.value = false
    ElMessage.error('同步失败: ' + (err.response?.data?.error || err.message))
  })
}

function formatDateTime(dateTimeStr) {
  if (!dateTimeStr) return ''
  const date = new Date(dateTimeStr)
  return date.toLocaleString('zh-CN', { hour12: false }).replaceAll('/', '-')
}

function shortPortList(portList) {
  if (!portList) return ''
  return portList.length > 10 ? portList.slice(0, 10) + '...' : portList
}

onMounted(fetchConfigs)
</script> 