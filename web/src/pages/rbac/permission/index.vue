<template>
  <div>
    <t-card class="list-card-container" :bordered="false">
      <t-row justify="space-between">
        <div class="left-operation-container">
          <t-button @click="onAdd">新增权限</t-button>
        </div>
      </t-row>

      <t-table :data="permissions" :columns="columns" row-key="id" :loading="loading" :hover="true">
        <template #type="{ row }">
          <t-tag :theme="row.type === 'menu' ? 'primary' : 'default'" variant="light">
            {{ row.type === 'menu' ? '菜单' : '接口' }}
          </t-tag>
        </template>
        <template #status="{ row }">
          <t-tag :theme="row.status === 1 ? 'success' : 'warning'" variant="light">
            {{ row.status === 1 ? '正常' : '禁用' }}
          </t-tag>
        </template>
        <template #op="{ row }">
          <t-space>
            <t-link theme="primary" @click="onEdit(row)">编辑</t-link>
            <t-popconfirm content="确定删除该权限吗？" @confirm="onDelete(row.id)">
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
          @current-change="fetchPermissions"
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
        <t-form-item name="name" label="权限名">
          <t-input v-model="editForm.name" placeholder="请输入权限名" />
        </t-form-item>
        <t-form-item name="code" label="标识">
          <t-input v-model="editForm.code" placeholder="请输入权限标识" />
        </t-form-item>
        <t-form-item name="type" label="类型">
          <t-select v-model="editForm.type" placeholder="请选择类型">
            <t-option label="菜单" value="menu" />
            <t-option label="接口" value="api" />
          </t-select>
        </t-form-item>
        <t-form-item name="path" label="路径">
          <t-input v-model="editForm.path" placeholder="请输入路径" />
        </t-form-item>
        <t-form-item v-if="editForm.type === 'api'" name="method" label="方法">
          <t-select v-model="editForm.method" placeholder="请选择方法">
            <t-option label="GET" value="GET" />
            <t-option label="POST" value="POST" />
            <t-option label="PUT" value="PUT" />
            <t-option label="DELETE" value="DELETE" />
          </t-select>
        </t-form-item>
        <t-form-item name="description" label="描述">
          <t-input v-model="editForm.description" placeholder="请输入描述" />
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>
<script setup lang="ts">
import type { FormRule, PrimaryTableCol, SubmitContext, TableRowData } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { onMounted, reactive, ref } from 'vue';

import type { PermissionItem } from '@/api/model/rbacModel';
import { createPermissionApi, deletePermissionApi, getPermissionsApi, updatePermissionApi } from '@/api/rbac';

defineOptions({
  name: 'RbacPermissionManage',
});

interface PermissionEditForm {
  id?: number;
  name: string;
  code: string;
  type: string;
  path: string;
  method: string;
  description: string;
}

const columns: PrimaryTableCol<TableRowData>[] = [
  { title: 'ID', colKey: 'id', width: 80 },
  { title: '权限名', colKey: 'name', minWidth: 180 },
  { title: '标识', colKey: 'code', minWidth: 220 },
  { title: '类型', colKey: 'type', width: 110 },
  { title: '路径', colKey: 'path', minWidth: 220 },
  { title: '方法', colKey: 'method', width: 110 },
  { title: '状态', colKey: 'status', width: 110 },
  { title: '操作', colKey: 'op', width: 140, fixed: 'right' },
];

const rules: Record<string, FormRule[]> = {
  name: [{ required: true, message: '请输入权限名', type: 'error' }],
  code: [{ required: true, message: '请输入权限标识', type: 'error' }],
  type: [{ required: true, message: '请选择权限类型', type: 'error' }],
  path: [{ required: true, message: '请输入路径', type: 'error' }],
  method: [{ required: true, message: '请选择请求方法', type: 'error' }],
};

const permissions = ref<PermissionItem[]>([]);
const loading = ref(false);
const submitLoading = ref(false);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);

const dialogVisible = ref(false);
const dialogTitle = ref('新增权限');
const formRef = ref();

const editForm = reactive<PermissionEditForm>({
  name: '',
  code: '',
  type: '',
  path: '',
  method: '',
  description: '',
});

function resetEditForm() {
  editForm.id = undefined;
  editForm.name = '';
  editForm.code = '';
  editForm.type = '';
  editForm.path = '';
  editForm.method = '';
  editForm.description = '';
}

async function fetchPermissions() {
  loading.value = true;
  try {
    const res = await getPermissionsApi({ page: page.value, page_size: pageSize.value });
    permissions.value = res.permissions;
    total.value = res.total;
  } catch (error) {
    console.error(error);
    MessagePlugin.error('获取权限列表失败');
  } finally {
    loading.value = false;
  }
}

function onPageSizeChange(size: number) {
  pageSize.value = size;
  page.value = 1;
  fetchPermissions();
}

function onAdd() {
  dialogTitle.value = '新增权限';
  resetEditForm();
  dialogVisible.value = true;
}

function onEdit(row: PermissionItem) {
  dialogTitle.value = '编辑权限';
  editForm.id = row.id;
  editForm.name = row.name;
  editForm.code = row.code;
  editForm.type = row.type;
  editForm.path = row.path;
  editForm.method = row.method;
  editForm.description = row.description || '';
  dialogVisible.value = true;
}

async function onDelete(id: number) {
  try {
    await deletePermissionApi(id);
    MessagePlugin.success('删除成功');
    await fetchPermissions();
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
      await updatePermissionApi(editForm.id, {
        name: editForm.name,
        code: editForm.code,
        type: editForm.type,
        path: editForm.path,
        method: editForm.method,
        description: editForm.description,
      });
      MessagePlugin.success('编辑成功');
    } else {
      await createPermissionApi({
        name: editForm.name,
        code: editForm.code,
        type: editForm.type,
        path: editForm.path,
        method: editForm.method,
        description: editForm.description,
      });
      MessagePlugin.success('新增成功');
    }
    dialogVisible.value = false;
    await fetchPermissions();
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

onMounted(fetchPermissions);
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
