<template>
  <el-card>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <el-button type="primary" @click="showEditDialog()">新增</el-button>
    </div>
    <el-table :data="templates" style="width: 100%" v-loading="loading" empty-text="暂无数据">
      <el-table-column type="index" label="序号" width="60" />
      <el-table-column prop="name" label="模板名称" width="180">
        <template #default="scope">
          <span style="color: #409EFF; cursor: pointer;" @click="showDetailDialog(scope.row)">{{ scope.row.name }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="type" label="类型" width="120">
        <template #default="scope">
          {{ typeLabel(scope.row.type) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="scope">
          <el-button size="small" @click="showEditDialog(scope.row)">编辑</el-button>
          <el-button size="small" type="danger" @click="deleteTemplate(scope.row)">删除</el-button>
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
    <el-dialog v-model="editDialogVisible" :title="editForm.id ? '编辑模板' : '新增模板'" width="500px" :close-on-click-modal="false">
      <el-form :model="editForm" :rules="rules" label-width="100px">
        <el-form-item label="模板名称" prop="name">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="模板类型" prop="type">
          <el-select v-model="editForm.type" placeholder="请选择模板类型">
            <el-option label="钉钉" value="dingtalk" />
            <el-option label="企业微信" value="weixin" />
            <el-option label="飞书" value="feishu" />
            <el-option label="邮件" value="email" />
          </el-select>
        </el-form-item>
        <el-form-item label="内容" prop="content">
          <el-input type="textarea" v-model="editForm.content" :rows="6" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveTemplate">保存</el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="detailDialogVisible" title="模板详情" width="500px" :close-on-click-modal="false">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="模板名称">{{ detailForm.name }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ detailForm.type }}</el-descriptions-item>
        <el-descriptions-item label="内容">
          <el-input type="textarea" :rows="8" :model-value="detailForm.content" readonly />
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'

const templates = ref([])
const loading = ref(false)
const editDialogVisible = ref(false)
const editForm = ref({})
const detailDialogVisible = ref(false)
const detailForm = ref({})
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const rules = {
  name: [{ required: true, message: '模板名称必填', trigger: 'blur' }],
  type: [{ required: true, message: '模板类型必选', trigger: 'change' }],
  content: [{ required: true, message: '内容必填', trigger: 'blur' }]
}

function fetchTemplates() {
  loading.value = true
  axios.get('/api/alert-templates', { params: { page: page.value, size: pageSize.value } }).then(res => {
    templates.value = res.data.items || res.data // 兼容老接口
    total.value = res.data.total || res.data.length || 0
    loading.value = false
  }).catch(() => { loading.value = false })
}

function showEditDialog(row) {
  if (row) {
    editForm.value = { ...row }
  } else {
    editForm.value = { name: '', type: '', content: '' }
  }
  editDialogVisible.value = true
}

function saveTemplate() {
  if (editForm.value.id) {
    axios.put(`/api/alert-templates/${editForm.value.id}`, editForm.value).then(() => {
      ElMessage.success('更新成功')
      editDialogVisible.value = false
      fetchTemplates()
    })
  } else {
    axios.post('/api/alert-templates', editForm.value).then(() => {
      ElMessage.success('创建成功')
      editDialogVisible.value = false
      fetchTemplates()
    })
  }
}

function deleteTemplate(row) {
  axios.delete(`/api/alert-templates/${row.id}`).then(() => {
    ElMessage.success('删除成功')
    fetchTemplates()
  })
}

function showDetailDialog(row) {
  detailForm.value = { ...row }
  detailDialogVisible.value = true
}

function typeLabel(type) {
  if (type === 'dingtalk') return '钉钉'
  if (type === 'weixin') return '企业微信'
  if (type === 'feishu') return '飞书'
  if (type === 'email') return '邮件'
  return type
}

onMounted(fetchTemplates)
</script> 