<template>
  <div style="padding: 24px;">
    <!-- 查询条件区域 -->
    <el-card class="search-card" shadow="never" style="margin-bottom: 12px;">
      <el-form :inline="true" class="search-form">
        <el-form-item label="名称">
          <el-input v-model="searchForm.name" placeholder="名称" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="AccessKey">
          <el-input v-model="searchForm.access_key" placeholder="AccessKey" clearable @keyup.enter="handleSearch" />
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
    <el-card class="operation-card" shadow="never" style="margin-bottom: 12px;">
      <el-button type="primary" @click="openDialog()">新增</el-button>
      <el-tooltip placement="right" effect="light">
        <template #content>
          <div style="max-width: 320px;">
            <div>RAM子账号需要以下权限才能正常操作安全组：</div>
            <pre style="background: #f6f8fa; border-radius: 4px; padding: 8px; margin: 8px 0 0 0; font-size: 13px;">
  ecs:AuthorizeSecurityGroup
  ecs:AuthorizeSecurityGroupEgress
  ecs:ModifySecurityGroupEgressRule
  ecs:ModifySecurityGroupRule
  ecs:RevokeSecurityGroup
  ecs:RevokeSecurityGroupEgress
  ecs:DescribeSecurityGroupAttribute
            </pre>
          </div>
        </template>
        <el-icon style="color: #909399; margin-left: 8px; cursor: pointer; vertical-align: middle;">
          <InfoFilled />
        </el-icon>
      </el-tooltip>
    </el-card>
    <el-table :data="configs" style="width: 100%" v-loading="loading" empty-text="暂无数据">
      <el-table-column type="index" label="序号" width="60" />
      <el-table-column prop="name" label="名称" width="120" />
      <el-table-column prop="access_key" label="AccessKey" width="220" />
      <el-table-column prop="security_group_id" label="安全组ID" width="200" />
      <el-table-column prop="port_list" label="端口列表" width="150" show-overflow-tooltip />
      <el-table-column prop="last_ip" label="上次授权IP" width="140" />
      <el-table-column prop="last_ip_updated_at" label="同步时间" width="160">
        <template #default="scope">
          <span>{{ formatDateTime(scope.row.last_ip_updated_at) }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="updated_at" label="更新时间" width="160">
        <template #default="scope">
          <span>{{ formatDateTime(scope.row.updated_at) }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="scope">
          <el-switch
            v-model="scope.row.status"
            :active-value="1"
            :inactive-value="0"
            @change="handleStatusChange(scope.row)"
          />
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
        <el-form-item label="名称" prop="name">
          <el-input v-model="editConfig.name" />
        </el-form-item>
        <el-form-item label="AccessKey" prop="access_key">
          <el-input v-model="editConfig.access_key" />
        </el-form-item>
        <el-form-item label="AccessSecret" prop="access_secret">
          <el-input v-model="editConfig.access_secret" type="password" show-password :key="inputKey" />
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
import { InfoFilled } from '@element-plus/icons-vue'

const configs = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const editConfig = ref({})
const formRef = ref()
const rules = {
  name: [{ required: true, message: '名称必填', trigger: 'blur' }],
  access_key: [{ required: true, message: 'AccessKey必填', trigger: 'blur' }],
  access_secret: [{ required: true, message: 'AccessSecret必填', trigger: 'blur' }],
  region_id: [{ required: true, message: 'RegionId必填', trigger: 'blur' }],
  security_group_id: [{ required: true, message: '安全组ID必填', trigger: 'blur' }],
  port_list: [{ required: true, message: '端口列表必填', trigger: 'blur' }],
}
const inputKey = ref(Date.now())
const searchForm = ref({ name: '', access_key: '', status: '' })

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
      name: '', access_key: '', access_secret: '', region_id: '', security_group_id: '', port_list: ''
    }
  }
  inputKey.value = Date.now()
  dialogVisible.value = true
}

function handleSave() {
  formRef.value.validate(valid => {
    if (!valid) return
    // 构造 payload，只包含需要的字段
    const payload = {
      name: editConfig.value.name,
      access_key: editConfig.value.access_key,
      access_secret: editConfig.value.access_secret,
      region_id: editConfig.value.region_id,
      security_group_id: editConfig.value.security_group_id,
      port_list: editConfig.value.port_list
    }
    if (editConfig.value.id) {
      updateAliyunSGConfig(editConfig.value.id, payload).then(() => {
        ElMessage.success('更新成功')
        dialogVisible.value = false
        fetchConfigs()
      })
    } else {
      createAliyunSGConfig(payload).then(() => {
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

function handleSearch() {
  loading.value = true
  const params = {}
  if (searchForm.value.name && searchForm.value.name.trim() !== '') params.name = searchForm.value.name.trim()
  if (searchForm.value.status !== '' && searchForm.value.status !== undefined) params.status = searchForm.value.status
  if (searchForm.value.access_key && searchForm.value.access_key.trim() !== '') params.access_key = searchForm.value.access_key.trim()
  getAliyunSGConfigs(params).then(res => {
    configs.value = res.data
    loading.value = false
  }).catch(() => {
    loading.value = false
  })
}

function handleReset() {
  searchForm.value = { name: '', access_key: '', status: '' }
  // 不自动查询
}

function handleStatusChange(row) {
  updateAliyunSGConfig(row.id, { status: row.status }).then(() => {
    ElMessage.success('状态更新成功')
    handleSearch()
  }).catch(() => {
    ElMessage.error('状态更新失败')
    // 回滚状态
    row.status = row.status === 1 ? 0 : 1
  })
}

onMounted(handleSearch)
</script> 