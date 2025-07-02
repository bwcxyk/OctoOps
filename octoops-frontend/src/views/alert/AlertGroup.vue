<template>
  <el-card>
    <div style="margin-bottom: 16px;">
      <div style="margin-top: 8px;">
        <el-button type="primary" @click="showEditDialog()">新增</el-button>
      </div>
    </div>
    <el-table :data="groups" style="width: 100%" v-loading="loading" empty-text="暂无数据">
      <el-table-column type="index" label="序号" width="60" />
      <el-table-column prop="name" label="告警组名称" width="160" />
      <el-table-column prop="description" label="描述" width="200" />
      <el-table-column prop="status" label="启用" width="80">
        <template #default="scope">
          <el-switch v-model="scope.row.status" :active-value="1" :inactive-value="0" @change="toggleEnable(scope.row)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="250">
        <template #default="scope">
          <el-button size="small" @click="showEditDialog(scope.row)">编辑</el-button>
          <el-button size="small" @click="showMemberDialog(scope.row)">成员管理</el-button>
          <el-button size="small" type="danger" @click="deleteGroup(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-dialog v-model="editDialogVisible" :title="editForm.id ? '编辑告警组' : '新增告警组'" width="500px">
      <el-form :model="editForm" :rules="rules" ref="editFormRef" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="editForm.description" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="editForm.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveGroup">保存</el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="memberDialogVisible" title="成员管理" width="600px">
      <div style="margin-bottom: 12px;">
        <el-select v-model="addMemberType" placeholder="选择类型" style="width: 120px; margin-right: 8px;">
          <el-option label="邮件" value="email" />
          <el-option label="钉钉" value="dingtalk" />
          <el-option label="企业微信" value="wechat" />
          <el-option label="飞书" value="feishu" />
        </el-select>
        <el-select v-model="addMemberId" filterable placeholder="选择渠道" style="width: 260px; margin-right: 8px;">
          <el-option v-for="item in availableChannels" :key="item.id" :label="item.name || item.target" :value="item.id" />
          <template #empty>暂无数据</template>
        </el-select>
        <el-button type="primary" @click="addMember" :disabled="!addMemberType || !addMemberId">添加成员</el-button>
      </div>
      <el-table :data="members" style="width: 100%" empty-text="暂无数据">
        <el-table-column type="index" label="序号" width="60" />
        <el-table-column prop="channel_type" label="类型" width="100" >
          <template #default="scope">
            {{ channelTypeLabel(scope.row.channel_type) }}
          </template>
        </el-table-column>
        <el-table-column prop="channel_id" label="渠道" width="150">
          <template #default="scope">
            {{ getChannelName(scope.row) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="scope">
            <el-button size="small" type="danger" @click="deleteMember(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { getAlerts, createAlert, updateAlert, deleteAlert } from '../../api/alert'
import axios from 'axios'
import { ElMessage } from 'element-plus'

const groups = ref([])
const loading = ref(false)
const editDialogVisible = ref(false)
const editForm = ref({})
const memberDialogVisible = ref(false)
const currentGroup = ref(null)
const members = ref([])
const addMemberType = ref()
const addMemberId = ref(null)
const availableChannels = ref([])
const allChannels = ref([])
const editFormRef = ref()
const rules = {
  name: [
    { required: true, message: '请输入告警组名称', trigger: 'blur' }
  ]
}

function fetchGroups() {
  loading.value = true
  axios.get('/api/alert-groups').then(res => {
    groups.value = res.data
    loading.value = false
  }).catch(() => { loading.value = false })
}

function showEditDialog(row) {
  if (row) {
    editForm.value = { ...row }
  } else {
    editForm.value = { name: '', description: '', status: 1 }
  }
  editDialogVisible.value = true
}

function saveGroup() {
  editFormRef.value.validate(valid => {
    if (!valid) return
    if (editForm.value.id) {
      axios.put(`/api/alert-groups/${editForm.value.id}`, editForm.value).then(() => {
        ElMessage.success('更新成功')
        editDialogVisible.value = false
        fetchGroups()
      })
    } else {
      axios.post('/api/alert-groups', editForm.value).then(() => {
        ElMessage.success('创建成功')
        editDialogVisible.value = false
        fetchGroups()
      })
    }
  })
}

function deleteGroup(row) {
  axios.delete(`/api/alert-groups/${row.id}`).then(() => {
    ElMessage.success('删除成功')
    fetchGroups()
  })
}

function toggleEnable(row) {
  axios.put(`/api/alert-groups/${row.id}`, { ...row, status: row.status }).then(() => {
    ElMessage.success(row.status ? '已启用' : '已禁用')
    fetchGroups()
  })
}

function showMemberDialog(group) {
  currentGroup.value = group
  memberDialogVisible.value = true
  fetchMembers()
  fetchAllChannels()
}

function fetchMembers() {
  if (!currentGroup.value) return
  axios.get(`/api/alert-groups/${currentGroup.value.id}/members`).then(res => {
    members.value = res.data
  })
}

function fetchAllChannels() {
  getAlerts().then(res => {
    allChannels.value = res.data
    updateAvailableChannels()
  })
}

function updateAvailableChannels() {
  availableChannels.value = allChannels.value.filter(item => item.type === addMemberType.value)
}

function addMember() {
  if (!currentGroup.value || !addMemberType.value || !addMemberId.value) return
  axios.post(`/api/alert-groups/${currentGroup.value.id}/members`, {
    channel_type: addMemberType.value,
    channel_id: addMemberId.value
  }).then(() => {
    ElMessage.success('添加成功')
    addMemberId.value = null
    fetchMembers()
  })
}

function deleteMember(row) {
  axios.delete(`/api/alert-groups/${currentGroup.value.id}/members/${row.id}`).then(() => {
    ElMessage.success('删除成功')
    fetchMembers()
  })
}

function getChannelName(member) {
  const found = allChannels.value.find(item => item.id === member.channel_id && item.type === member.channel_type)
  return found ? (found.name || found.target) : member.channel_id
}

function channelTypeLabel(type) {
  if (type === 'email') return '邮件'
  if (type === 'dingtalk') return '钉钉'
  if (type === 'wechat') return '企业微信'
  if (type === 'feishu') return '飞书'
  return type
}

watch(addMemberType, () => {
  addMemberId.value = null;
  updateAvailableChannels();
});

onMounted(fetchGroups)
</script> 