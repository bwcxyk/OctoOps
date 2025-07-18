<template>
  <el-card>
    <el-form :model="form" :rules="rules" ref="formRef" label-width="120px">
      <el-form-item label="作业名称" prop="name">
        <el-input v-model="form.name" />
      </el-form-item>
      <el-form-item label="描述">
        <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入任务描述" />
      </el-form-item>
      <el-form-item label="状态">
        <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item label="Cron表达式" prop="cron_expr">
        <el-input v-model="form.cron_expr" placeholder="秒 分 时 日 月 周" />
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
import { ElMessage } from 'element-plus'
import { CronExpressionParser } from 'cron-parser'

const route = useRoute()
const router = useRouter()
const isEdit = ref(!!route.params.id)
const form = ref({
  name: '',
  description: '',
  cron_expr: '',
  config: '',
  config_format: 'json',
  status: 1,
  task_type: 'batch',
})
const formRef = ref()
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

onMounted(() => {
  if (isEdit.value) {
    // 获取任务详情
    getTasks('batch').then(res => {
      const arr = Array.isArray(res.data.data) ? res.data.data : []
      const task = arr.find(x => x.id == route.params.id) // 用 == 宽松比较
      if (task) {
        form.value = { ...task }
      }
    })
  }
})

function handleSave() {
  formRef.value.validate(valid => {
    if (!valid) return
    // 只取业务字段
    const payload = {
      id: form.value.id,
      name: form.value.name,
      description: form.value.description,
      cron_expr: form.value.cron_expr,
      config: form.value.config,
      config_format: form.value.config_format,
      status: form.value.status,
      task_type: form.value.task_type
      // 其他需要的字段
    }
    if (isEdit.value) {
      updateTask(form.value.id, payload).then(() => {
        ElMessage.success('更新成功')
        goBack()
      })
    } else {
      createTask(payload).then(() => {
        ElMessage.success('创建成功')
        goBack()
      })
    }
  })
}
function goBack() {
  router.push({ name: 'BatchTask' })
}
</script> 