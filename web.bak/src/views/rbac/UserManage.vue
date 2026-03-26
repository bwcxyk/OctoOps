<template>
  <div>
    <el-card>
      <div style="margin-bottom: 16px; display: flex-start; justify-content: space-between; align-items: center;">
        <el-button type="primary" @click="onAdd">新增</el-button>
      </div>
      <el-table :data="users" style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="username" label="用户名" />
        <el-table-column prop="email" label="邮箱" />
        <el-table-column prop="nickname" label="昵称" />
        <el-table-column prop="roles" label="角色" :formatter="roleFormatter" />
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
        @current-change="fetchUsers"
        @size-change="fetchUsers"
        style="margin-top: 16px; text-align: right;"
      />
    </el-card>
    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="400px" :close-on-click-modal="false">
      <el-form :model="editForm" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="editForm.username" :disabled="!!editForm.id" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!editForm.id">
          <el-input v-model="editForm.password" type="password" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="editForm.email" />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="editForm.nickname" />
        </el-form-item>
        <el-form-item label="角色" prop="role_ids">
          <el-select v-model="editForm.role_ids" multiple placeholder="请选择角色">
            <el-option v-for="role in allRoles" :key="role.id" :label="role.name" :value="role.id" />
          </el-select>
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
import { getUsersApi, createUserApi, updateUserApi, deleteUserApi, getRolesApi } from '@/api/user'

const users = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const dialogVisible = ref(false)
const dialogTitle = ref('新增用户')
const editForm = ref({})
const formRef = ref()
const allRoles = ref([])
const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9]+$/, message: '用户名只能包含数字或英文字母', trigger: ['blur', 'change'] }
  ],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' }
  ],
  role_ids: [{ required: true, message: '请选择角色', trigger: 'change' }],
}

function roleFormatter(row) {
  return row.roles ? row.roles.map(r => r.name).join(', ') : ''
}
function statusFormatter(row) {
  return row.status === 1 ? '正常' : '禁用'
}

async function fetchUsers() {
  try {
    const res = await getUsersApi({ page: page.value, page_size: pageSize.value })
    users.value = res.data.users
    total.value = res.data.total
  } catch (e) {
    ElMessage.error('获取用户列表失败')
  }
}

async function fetchRoles() {
  try {
    const res = await getRolesApi({ page: 1, page_size: 100 })
    allRoles.value = res.data.roles || []
  } catch (e) {
    ElMessage.error('获取角色列表失败')
  }
}

function onAdd() {
  dialogTitle.value = '新增用户'
  editForm.value = { username: '', password: '', email: '', nickname: '', role_ids: [] }
  dialogVisible.value = true
}

function onEdit(row) {
  dialogTitle.value = '编辑用户'
  editForm.value = { ...row, password: '', role_ids: (row.roles || []).map(r => r.id) }
  dialogVisible.value = true
}

function onDelete(row) {
  ElMessageBox.confirm('确定要删除该用户吗？', '提示', { type: 'warning' })
    .then(async () => {
      try {
        await deleteUserApi(row.id)
        ElMessage.success('删除成功')
        fetchUsers()
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
        await updateUserApi(editForm.value.id, editForm.value)
        ElMessage.success('编辑成功')
      } else {
        await createUserApi(editForm.value)
        ElMessage.success('新增成功')
      }
      dialogVisible.value = false
      fetchUsers()
    } catch (e) {
      ElMessage.error(e?.response?.data?.message || e?.message || '操作失败')
    }
  })
}

onMounted(() => {
  fetchUsers()
  fetchRoles()
})
</script>
