<template>
  <div>
    <el-card>
      <div style="margin-bottom: 16px; display: flex; justify-content: space-between; align-items: center;">
        <el-button type="primary" @click="onAdd" style="margin-bottom: 12px;">新增</el-button>
      </div>
      <el-table :data="roles" style="width: 100%" v-loading="loading" empty-text="暂无数据">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="角色名" />
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
        @current-change="fetchRoles"
        @size-change="fetchRoles"
        style="margin-top: 16px; text-align: right;"
      />
    </el-card>
    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="400px" :close-on-click-modal="false">
      <el-form :model="editForm" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="角色名" prop="name">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="editForm.description" />
        </el-form-item>
        <el-form-item label="权限" prop="permission_ids">
          <el-tree
            :data="permissionTree"
            show-checkbox
            node-key="id"
            :default-checked-keys="editForm.permission_ids"
            :props="{ label: 'name', children: 'children' }"
            @check="onTreeCheck"
            ref="treeRef"
            style="width: 100%;"
          />
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
import { getRolesApi, createRoleApi, updateRoleApi, deleteRoleApi, getPermissionsApi, getPermissionTreeApi } from '@/api/user'
import { nextTick } from 'vue'

const roles = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('新增角色')
const editForm = ref({})
const formRef = ref()
const allPermissions = ref([])
const permissionTree = ref([])
const treeRef = ref()
const rules = {
  name: [{ required: true, message: '请输入角色名', trigger: 'blur' }],
  permission_ids: [{ required: true, message: '请选择权限', trigger: 'change' }],
}

function statusFormatter(row) {
  return row.status === 1 ? '正常' : '禁用'
}
function permFormatter(row) {
  return row.permissions ? row.permissions.map(p => p.name).join(', ') : ''
}

async function fetchRoles() {
  loading.value = true
  try {
    const res = await getRolesApi({ page: page.value, page_size: pageSize.value })
    roles.value = res.data.roles
    total.value = res.data.total
  } catch (e) {
    ElMessage.error('获取角色列表失败')
  } finally {
    loading.value = false
  }
}
async function fetchPermissions() {
  try {
    const res = await getPermissionsApi({ page: 1, page_size: 1000 })
    allPermissions.value = res.data.permissions || []
  } catch (e) {
    ElMessage.error('获取权限列表失败')
  }
}
async function fetchPermissionTree() {
  try {
    const res = await getPermissionTreeApi()
    permissionTree.value = res.data || []
  } catch (e) {
    ElMessage.error('获取权限树失败')
  }
}
function onTreeCheck(checkedKeys) {
  editForm.value.permission_ids = checkedKeys
}
function onAdd() {
  dialogTitle.value = '新增角色'
  editForm.value = { name: '', description: '', permission_ids: [] }
  dialogVisible.value = true
  nextTick(() => {
    treeRef.value && treeRef.value.setCheckedKeys([])
  })
}
function onEdit(row) {
  dialogTitle.value = '编辑角色'
  editForm.value = { ...row, permission_ids: (row.permissions || []).map(p => p.id) }
  dialogVisible.value = true
  nextTick(() => {
    treeRef.value && treeRef.value.setCheckedKeys(editForm.value.permission_ids)
  })
}
function onDelete(row) {
  ElMessageBox.confirm('确定要删除该角色吗？', '提示', { type: 'warning' })
    .then(async () => {
      try {
        await deleteRoleApi(row.id)
        ElMessage.success('删除成功')
        fetchRoles()
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
        await updateRoleApi(editForm.value.id, editForm.value)
        ElMessage.success('编辑成功')
      } else {
        await createRoleApi(editForm.value)
        ElMessage.success('新增成功')
      }
      dialogVisible.value = false
      fetchRoles()
    } catch (e) {
      ElMessage.error('操作失败')
    }
  })
}
onMounted(() => {
  fetchRoles()
  fetchPermissionTree()
})
</script>
