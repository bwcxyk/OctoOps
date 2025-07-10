<template>
  <div>
    <el-card>
      <div style="margin-bottom: 16px; display: flex; justify-content: space-between; align-items: center;">
        <el-button type="primary" @click="onAdd">新增</el-button>
      </div>
      <el-table :data="permissions" style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="权限名" />
        <el-table-column prop="code" label="标识" />
        <el-table-column prop="type" label="类型" />
        <el-table-column prop="path" label="路径" />
        <el-table-column prop="method" label="方法" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="status" label="状态" :formatter="statusFormatter" />
        <el-table-column label="操作" width="180">
          <template #default="scope">
            <el-button size="small" @click="onEdit(scope.row)">编辑</el-button>
            <el-button size="small" type="danger" @click="onDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next, jumper"
        @current-change="fetchPermissions"
        @size-change="fetchPermissions"
        style="margin-top: 16px; text-align: right;"
      />
    </el-card>
    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="400px" :close-on-click-modal="false">
      <el-form :model="editForm" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="权限名" prop="name">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="标识" prop="code">
          <el-input v-model="editForm.code" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="editForm.type" placeholder="请选择类型">
            <el-option label="菜单" value="menu" />
            <el-option label="接口" value="api" />
          </el-select>
        </el-form-item>
        <el-form-item label="路径" prop="path">
          <el-input v-model="editForm.path" />
        </el-form-item>
        <el-form-item label="方法" prop="method" v-if="editForm.type === 'api'">
          <el-select v-model="editForm.method" placeholder="请选择方法">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="editForm.description" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="onSubmit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPermissionsApi, createPermissionApi, updatePermissionApi, deletePermissionApi } from '@/api/user'

const permissions = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const dialogVisible = ref(false)
const dialogTitle = ref('新增权限')
const editForm = ref({})
const formRef = ref()
const rules = {
  name: [{ required: true, message: '请输入权限名', trigger: 'blur' }],
  code: [{ required: true, message: '请输入标识', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  path: [{ required: true, message: '请输入路径', trigger: 'blur' }],
  method: [{ required: true, message: '请选择方法', trigger: 'change' }],
}

function statusFormatter(row) {
  return row.status === 1 ? '正常' : '禁用'
}

async function fetchPermissions() {
  try {
    const res = await getPermissionsApi({ page: page.value, page_size: pageSize.value })
    permissions.value = res.data.permissions
    total.value = res.data.total
  } catch (e) {
    ElMessage.error('获取权限列表失败')
  }
}

function onAdd() {
  dialogTitle.value = '新增权限'
  editForm.value = { name: '', code: '', type: '', path: '', method: '', description: '' }
  dialogVisible.value = true
}

function onEdit(row) {
  dialogTitle.value = '编辑权限'
  editForm.value = { ...row }
  dialogVisible.value = true
}

function onDelete(row) {
  ElMessageBox.confirm('确定要删除该权限吗？', '提示', { type: 'warning' })
    .then(async () => {
      try {
        await deletePermissionApi(row.id)
        ElMessage.success('删除成功')
        fetchPermissions()
      } catch (e) {
        ElMessage.error('删除失败')
      }
    })
    .catch(() => {})
}

function onSubmit() {
  formRef.value.validate(async valid => {
    if (!valid) return
    try {
      if (editForm.value.id) {
        await updatePermissionApi(editForm.value.id, editForm.value)
        ElMessage.success('编辑成功')
      } else {
        await createPermissionApi(editForm.value)
        ElMessage.success('新增成功')
      }
      dialogVisible.value = false
      fetchPermissions()
    } catch (e) {
      ElMessage.error('操作失败')
    }
  })
}
onMounted(fetchPermissions)
</script>
