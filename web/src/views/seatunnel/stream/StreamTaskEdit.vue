<template>
  <el-card>
    <el-form :model="form" :rules="rules" ref="formRef" label-width="120px">
      <el-form-item label="作业名称" prop="name">
        <el-input v-model="form.name" />
      </el-form-item>
      <el-form-item label="描述">
        <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入任务描述" />
      </el-form-item>
      <el-form-item label="配置风格" prop="config_format">
        <el-select v-model="form.config_format" placeholder="请选择配置风格">
          <el-option label="JSON" value="json" />
          <el-option label="HOCON" value="hocon" />
        </el-select>
      </el-form-item>
      <el-form-item label="作业配置" prop="config">
        <el-input type="textarea" v-model="form.config" :rows="8" placeholder="请输入完整作业配置（json或hocon）" />
      </el-form-item>
      <el-form-item label="通知" prop="enable_alert">
        <el-checkbox v-model="form.enable_alert">作业失败时发送通知</el-checkbox>
      </el-form-item>
      <el-form-item label="告警组" prop="alert_group" v-if="form.enable_alert">
        <el-select v-model="form.alert_group" multiple placeholder="请选择告警组">
          <el-option v-for="group in alertGroups" :key="group.id" :label="group.name" :value="group.id" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button @click="goBack">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getTasks, createTask, updateTask } from '../../../api/task'
import axios from 'axios'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const isEdit = ref(!!route.params.id)
const form = ref({
  name: '',
  description: '',
  config: '',
  config_format: 'json',
  task_type: 'stream',
  enable_alert: false,
  alert_group: [],
})
const formRef = ref()
const rules = {
  name: [
    { required: true, message: '任务名称为必填', trigger: 'blur' }
  ],
  config: [
    { required: true, message: '作业配置为必填', trigger: 'blur' }
  ],
  alert_group: [
    {
      required: true,
      message: '请选择告警组',
      trigger: 'change',
      validator: (rule, value, callback) => {
        if (form.value.enable_alert && (!value || value.length === 0)) {
          callback(new Error('请选择告警组'))
        } else {
          callback()
        }
      }
    }
  ]
}
const alertGroups = ref([])

onMounted(() => {
  if (isEdit.value) {
    getTasks('stream').then(res => {
      const arr = Array.isArray(res.data.data) ? res.data.data : []
      const task = arr.find(t => t.id == route.params.id)
      if (task) {
        form.value = { ...task, enable_alert: !!(task.alert_group && task.alert_group.length > 0), alert_group: task.alert_group ? task.alert_group.split(',').map(id => Number(id)) : [] }
      }
    })
  }
  axios.get('/api/alert-groups').then(res => {
    alertGroups.value = res.data.filter(g => g.status === 1)
  })
})

function handleSave() {
  formRef.value.validate(valid => {
    if (!valid) return
    const submitForm = {
      name: form.value.name,
      description: form.value.description,
      config: form.value.config,
      config_format: form.value.config_format,
      task_type: form.value.task_type,
      alert_group: form.value.enable_alert ? (Array.isArray(form.value.alert_group) ? form.value.alert_group.join(',') : form.value.alert_group) : ''
    }
    if (isEdit.value) {
      updateTask(form.value.id, submitForm).then(() => {
        ElMessage.success('更新成功')
        goBack()
      })
    } else {
      createTask(submitForm).then(() => {
        ElMessage.success('创建成功')
        goBack()
      })
    }
  })
}
function goBack() {
  router.push({ name: 'StreamTask' })
}
</script> 