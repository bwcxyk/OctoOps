<template>
  <el-card>
    <div style="margin-bottom: 16px; display: flex; align-items: center;">
      <el-select v-model="filterType" placeholder="全部类型" style="width: 140px; margin-right: 12px;">
        <el-option label="全部" value="" />
        <el-option label="邮件" value="email" />
        <el-option label="钉钉" value="dingtalk" />
        <el-option label="企业微信" value="wechat" />
        <el-option label="飞书" value="feishu" />
      </el-select>
      <el-button type="primary" @click="openDialog()">新增</el-button>
    </div>
    <el-table :data="filteredChannels" style="width: 100%" v-loading="loading" empty-text="暂无数据">
      <el-table-column type="index" label="序号" width="60" />
      <el-table-column prop="name" label="名称" width="140" />
      <el-table-column prop="type" label="类型" width="100">
        <template #default="scope">
          <span>{{ platformLabel(scope.row.type) }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="target" label="目标" width="220" show-overflow-tooltip />
      <el-table-column prop="status" label="启用" width="80">
        <template #default="scope">
          <el-switch v-model="scope.row.status" :active-value="1" :inactive-value="0" @change="toggleStatus(scope.row)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220">
        <template #default="scope">
          <el-button size="small" @click="openDialog(scope.row)">编辑</el-button>
          <el-button size="small" type="danger" @click="deleteChannel(scope.row)">删除</el-button>
          <el-button size="small" @click="testChannel(scope.row)">测试发送</el-button>
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
    <el-dialog v-model="dialogVisible" :title="editChannel.id ? '编辑渠道' : '新增渠道'" width="500px" :close-on-click-modal="false">
      <el-form :model="editChannel" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="editChannel.name" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="editChannel.type" placeholder="请选择类型" @change="onTypeChange">
            <el-option label="邮件" value="email" />
            <el-option label="钉钉" value="dingtalk" />
            <el-option label="企业微信" value="wechat" />
            <el-option label="飞书" value="feishu" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="editChannel.type === 'email'" label="邮箱" prop="target">
          <el-input v-model="editChannel.target" placeholder="请输入邮箱地址" />
        </el-form-item>
        <el-form-item v-else label="Webhook" prop="target">
          <el-input v-model="editChannel.target" type="textarea" :rows="2" placeholder="请输入Webhook地址" />
        </el-form-item>
        <el-form-item v-if="editChannel.type === 'dingtalk'" label="加签密钥" prop="dingtalk_secret">
          <el-input v-model="editChannel.dingtalk_secret" placeholder="请输入钉钉加签密钥" />
        </el-form-item>
        <el-form-item label="启用" prop="status">
          <el-switch v-model="editChannel.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="告警模板" prop="template_id">
          <el-select v-model="editChannel.template_id" placeholder="请选择告警模板" :loading="templatesLoading">
            <el-option v-for="tpl in filteredTemplates" :key="tpl.id" :label="tpl.name" :value="tpl.id" />
            <template #empty>
              <span v-if="templatesLoading">加载中...</span>
              <span v-else-if="editChannel.type">
                暂无{{ platformLabel(editChannel.type) }}类型的模板
              </span>
              <span v-else>请先选择渠道类型</span>
            </template>
          </el-select>
          <div v-if="editChannel.type && filteredTemplates.length > 0" style="font-size: 12px; color: #909399; margin-top: 4px;">
            显示 {{ platformLabel(editChannel.type) }} 类型的模板 ({{ filteredTemplates.length }}个)
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveChannel">保存</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'

const channels = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const editChannel = ref({})
const formRef = ref()
const filterType = ref('')
const templates = ref([])
const templatesLoading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const rules = {
  name: [{ required: true, message: '名称必填', trigger: 'blur' }],
  type: [{ required: true, message: '类型必选', trigger: 'change' }],
  target: [
    { required: true, message: '目标必填', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (editChannel.value.type === 'email') {
          const emailReg = /^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$/
          if (!emailReg.test(value)) {
            callback(new Error('请输入正确的邮箱地址'))
            return
          }
        }
        callback()
      },
      trigger: 'blur'
    }
  ],
}

const filteredChannels = computed(() => {
  if (!filterType.value) return channels.value
  return channels.value.filter(c => c.type === filterType.value)
})

// 根据渠道类型过滤模板
const filteredTemplates = computed(() => {
  if (!editChannel.value.type) return templates.value
  const channelType = editChannel.value.type
  const templateType = getTemplateTypeByChannelType(channelType)
  return templates.value.filter(tpl => tpl.type === templateType)
})

// 渠道类型到模板类型的映射
function getTemplateTypeByChannelType(channelType) {
  const typeMap = {
    'dingtalk': 'dingtalk',
    'wechat': 'weixin',  // 渠道类型是wechat，对应模板类型是weixin
    'feishu': 'feishu',
    'email': 'email'
  }
  return typeMap[channelType] || channelType
}

function platformLabel(val) {
  if (val === 'dingtalk') return '钉钉'
  if (val === 'feishu') return '飞书'
  if (val === 'wechat') return '企业微信'
  if (val === 'email') return '邮件'
  return val
}

function fetchChannels() {
  loading.value = true
  axios.get('/api/channels', { params: { page: page.value, size: pageSize.value } }).then(res => {
    channels.value = res.data.items || res.data
    total.value = res.data.total || res.data.length || 0
    loading.value = false
  }).catch(() => { loading.value = false })
}

function fetchTemplates() {
  templatesLoading.value = true
  return axios.get('/api/alert-templates').then(res => {
    templates.value = res.data
  }).finally(() => {
    templatesLoading.value = false
  })
}

function openDialog(row = null) {
  if (row) {
    editChannel.value = { ...row }
    // 如果没有可用模板，或者模板id不在模板列表中，强制设为null
    const validIds = filteredTemplates.value.map(tpl => tpl.id)
    if (!editChannel.value.template_id || !validIds.includes(editChannel.value.template_id)) {
      editChannel.value.template_id = null
    }
  } else {
    editChannel.value = { name: '', type: '', target: '', status: 1, dingtalk_secret: '', template_id: null }
  }
  dialogVisible.value = true
}

function onTypeChange() {
  // 清空与类型相关的字段
  if (editChannel.value.type !== 'dingtalk') {
    editChannel.value.dingtalk_secret = ''
  }
  
  // 检查当前选择的模板是否与新的渠道类型匹配
  if (editChannel.value.template_id) {
    const selectedTemplate = templates.value.find(tpl => tpl.id === editChannel.value.template_id)
    if (selectedTemplate) {
      const expectedTemplateType = getTemplateTypeByChannelType(editChannel.value.type)
      if (selectedTemplate.type !== expectedTemplateType) {
        editChannel.value.template_id = null
      }
    }
  }
}

function saveChannel() {
  formRef.value.validate(valid => {
    if (!valid) return
    if (editChannel.value.id) {
      axios.put(`/api/channels/${editChannel.value.id}`, editChannel.value).then(() => {
        ElMessage.success('更新成功')
        dialogVisible.value = false
        fetchChannels()
      })
    } else {
      axios.post('/api/channels', editChannel.value).then(() => {
        ElMessage.success('创建成功')
        dialogVisible.value = false
        fetchChannels()
      })
    }
  })
}

function deleteChannel(row) {
  ElMessageBox.confirm('确定要删除该渠道吗？', '提示', { type: 'warning' }).then(() => {
    axios.delete(`/api/channels/${row.id}`).then(() => {
      ElMessage.success('删除成功')
      fetchChannels()
    })
  })
}

function toggleStatus(row) {
  axios.put(`/api/channels/${row.id}`, { ...row, status: row.status }).then(() => {
    ElMessage.success(row.status ? '已启用' : '已禁用')
    fetchChannels()
  })
}

function testChannel(row) {
  axios.post(`/api/channels/${row.id}/test`).then(res => {
    ElMessage.success(res.data.message || '测试发送成功')
  }).catch(err => {
    ElMessage.error(err.response?.data?.error || '测试发送失败')
  })
}

onMounted(() => {
  fetchChannels()
  fetchTemplates()
})
</script> 