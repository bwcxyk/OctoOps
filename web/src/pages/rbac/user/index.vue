<template>
  <div>
    <t-card class="list-card-container" :bordered="false">
      <t-row justify="space-between">
        <div class="left-operation-container">
          <t-button @click="onAdd">新增用户</t-button>
        </div>
      </t-row>

      <t-table :data="users" :columns="columns" row-key="id" :loading="loading" :hover="true">
        <template #roles="{ row }">
          {{ formatRoles(row.roles) }}
        </template>
        <template #status="{ row }">
          <t-tag :theme="row.status === 1 ? 'success' : 'warning'" variant="light">
            {{ row.status === 1 ? '正常' : '禁用' }}
          </t-tag>
        </template>
        <template #op="{ row }">
          <t-space>
            <t-link theme="primary" @click="onEdit(row)">编辑</t-link>
            <t-popconfirm content="确定删除该用户吗？" @confirm="onDelete(row.id)">
              <t-link theme="danger">删除</t-link>
            </t-popconfirm>
          </t-space>
        </template>
      </t-table>

      <div class="list-pagination">
        <t-pagination
          v-model="page"
          v-model:page-size="pageSize"
          :total="total"
          show-jumper
          show-page-size
          :page-size-options="[10, 20, 50, 100]"
          @current-change="fetchUsers"
          @page-size-change="onPageSizeChange"
        />
      </div>
    </t-card>
    <t-dialog
      v-model:visible="dialogVisible"
      :header="dialogTitle"
      width="560px"
      :confirm-btn="{ content: '保存', theme: 'primary', loading: submitLoading }"
      @confirm="onSubmit"
    >
      <t-form ref="formRef" :data="editForm" :rules="rules" label-width="90px" @submit="onFormSubmit">
        <t-form-item name="username" label="用户名">
          <t-input v-model="editForm.username" :disabled="Boolean(editForm.id)" placeholder="请输入用户名" />
        </t-form-item>
        <t-form-item v-if="!editForm.id" name="password" label="密码">
          <t-input v-model="editForm.password" type="password" placeholder="请输入密码" />
        </t-form-item>
        <t-form-item name="email" label="邮箱">
          <t-input v-model="editForm.email" placeholder="请输入邮箱" />
        </t-form-item>
        <t-form-item name="nickname" label="昵称">
          <t-input v-model="editForm.nickname" placeholder="请输入昵称" />
        </t-form-item>
        <t-form-item name="role_ids" label="角色">
          <t-select v-model="editForm.role_ids" multiple clearable placeholder="请选择角色">
            <t-option v-for="role in allRoles" :key="role.id" :label="role.name" :value="role.id" />
          </t-select>
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>
<script setup lang="ts">
import type { FormRule, PrimaryTableCol, SubmitContext, TableRowData } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { onMounted, reactive, ref } from 'vue';

import type { RoleItem, UserItem } from '@/api/model/rbacModel';
import { createUserApi, deleteUserApi, getRolesApi, getUsersApi, updateUserApi } from '@/api/rbac';

defineOptions({
  name: 'RbacUserManage',
});

interface UserEditForm {
  id?: number;
  username: string;
  password: string;
  email: string;
  nickname: string;
  role_ids: number[];
}

const columns: PrimaryTableCol<TableRowData>[] = [
  { title: 'ID', colKey: 'id', width: 80 },
  { title: '用户名', colKey: 'username', minWidth: 160 },
  { title: '邮箱', colKey: 'email', minWidth: 220 },
  { title: '昵称', colKey: 'nickname', minWidth: 140 },
  { title: '角色', colKey: 'roles', minWidth: 220 },
  { title: '状态', colKey: 'status', width: 110 },
  { title: '操作', colKey: 'op', width: 140, fixed: 'right' },
];

const rules: Record<string, FormRule[]> = {
  username: [
    { required: true, message: '请输入用户名', type: 'error' },
    { pattern: /^[a-z0-9]+$/i, message: '用户名只能包含数字或英文字母', type: 'warning' },
  ],
  password: [{ required: true, message: '请输入密码', type: 'error' }],
  email: [
    { required: true, message: '请输入邮箱', type: 'error' },
    { email: true, message: '邮箱格式不正确', type: 'warning' },
  ],
  role_ids: [{ required: true, message: '请选择角色', type: 'error' }],
};

const users = ref<UserItem[]>([]);
const allRoles = ref<RoleItem[]>([]);
const loading = ref(false);
const submitLoading = ref(false);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);

const dialogVisible = ref(false);
const dialogTitle = ref('新增用户');
const formRef = ref();

const editForm = reactive<UserEditForm>({
  username: '',
  password: '',
  email: '',
  nickname: '',
  role_ids: [],
});

function resetEditForm() {
  editForm.id = undefined;
  editForm.username = '';
  editForm.password = '';
  editForm.email = '';
  editForm.nickname = '';
  editForm.role_ids = [];
}

function formatRoles(roles?: RoleItem[]) {
  return (roles || []).map((role) => role.name).join(', ');
}

async function fetchUsers() {
  loading.value = true;
  try {
    const res = await getUsersApi({ page: page.value, page_size: pageSize.value });
    users.value = res.users;
    total.value = res.total;
  } catch (error) {
    console.error(error);
    MessagePlugin.error('获取用户列表失败');
  } finally {
    loading.value = false;
  }
}

async function fetchRoles() {
  try {
    const res = await getRolesApi({ page: 1, page_size: 1000 });
    allRoles.value = res.roles || [];
  } catch (error) {
    console.error(error);
    MessagePlugin.error('获取角色列表失败');
  }
}

function onPageSizeChange(size: number) {
  pageSize.value = size;
  page.value = 1;
  fetchUsers();
}

function onAdd() {
  dialogTitle.value = '新增用户';
  resetEditForm();
  dialogVisible.value = true;
}

function onEdit(row: UserItem) {
  dialogTitle.value = '编辑用户';
  editForm.id = row.id;
  editForm.username = row.username;
  editForm.password = '';
  editForm.email = row.email || '';
  editForm.nickname = row.nickname || '';
  editForm.role_ids = (row.roles || []).map((role) => role.id);
  dialogVisible.value = true;
}

async function onDelete(id: number) {
  try {
    await deleteUserApi(id);
    MessagePlugin.success('删除成功');
    await fetchUsers();
  } catch (error) {
    console.error(error);
    MessagePlugin.error('删除失败');
  }
}

async function onSubmit() {
  const form = formRef.value;
  if (form) {
    await form.submit();
  }
}

async function onFormSubmit(context: SubmitContext) {
  if (context.validateResult !== true) {
    return;
  }

  submitLoading.value = true;
  try {
    if (editForm.id) {
      await updateUserApi(editForm.id, {
        email: editForm.email,
        nickname: editForm.nickname,
        role_ids: editForm.role_ids,
      });
      MessagePlugin.success('编辑成功');
    } else {
      await createUserApi({
        username: editForm.username,
        password: editForm.password,
        email: editForm.email,
        nickname: editForm.nickname,
        role_ids: editForm.role_ids,
      });
      MessagePlugin.success('新增成功');
    }
    dialogVisible.value = false;
    await fetchUsers();
  } catch (error: unknown) {
    console.error(error);
    if (error instanceof Error) {
      MessagePlugin.error(error.message);
    } else {
      MessagePlugin.error('操作失败');
    }
  } finally {
    submitLoading.value = false;
  }
}

onMounted(async () => {
  await Promise.all([fetchUsers(), fetchRoles()]);
});
</script>
<style lang="less" scoped>
.list-card-container {
  padding: var(--td-comp-paddingTB-xxl) var(--td-comp-paddingLR-xxl);

  :deep(.t-card__body) {
    padding: 0;
  }
}

.left-operation-container {
  display: flex;
  align-items: center;
  margin-bottom: var(--td-comp-margin-xxl);
}

.list-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--td-comp-margin-xxl);
}
</style>
