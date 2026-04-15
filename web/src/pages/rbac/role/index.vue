<template>
  <div>
    <t-card class="list-card-container" :bordered="false">
      <t-row justify="space-between">
        <div class="left-operation-container">
          <t-button @click="onAdd">新增角色</t-button>
        </div>
      </t-row>

      <t-table :data="roles" :columns="columns" row-key="id" :loading="loading" :hover="true">
        <template #index="{ rowIndex }">{{ rowIndex + 1 + (page - 1) * pageSize }}</template>
        <template #status="{ row }">
          <t-tag :theme="row.status === 1 ? 'success' : 'warning'" variant="light">
            {{ row.status === 1 ? '正常' : '禁用' }}
          </t-tag>
        </template>
        <template #op="{ row }">
          <t-space>
            <t-link theme="primary" @click="onEdit(row)">编辑</t-link>
            <t-popconfirm content="确定删除该角色吗？" @confirm="onDelete(row.id)">
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
          @current-change="fetchRoles"
          @page-size-change="onPageSizeChange"
        />
      </div>
    </t-card>
    <t-dialog
      v-model:visible="dialogVisible"
      :header="dialogTitle"
      width="920px"
      :confirm-btn="{ content: '保存', theme: 'primary', loading: submitLoading }"
      @confirm="onSubmit"
    >
      <t-form ref="formRef" :data="editForm" :rules="rules" label-width="90px" @submit="onFormSubmit">
        <t-form-item name="name" label="角色名">
          <t-input v-model="editForm.name" placeholder="请输入角色名" />
        </t-form-item>
        <t-form-item name="description" label="描述">
          <t-input v-model="editForm.description" placeholder="请输入描述" />
        </t-form-item>
        <t-form-item name="permission_ids" label="权限">
          <div class="permission-sections">
            <div class="permission-section">
              <div class="permission-section-title">菜单权限</div>
              <div class="permission-tree-wrapper">
                <t-tree
                  :data="menuPermissionTree"
                  :keys="{ label: 'name', value: 'id', children: 'children' }"
                  checkable
                  hover
                  expand-on-click-node
                  value-mode="all"
                  :value="menuPermissionIds"
                  @change="onMenuPermissionChange"
                />
              </div>
            </div>
            <div class="permission-section">
              <div class="permission-section-title">接口权限</div>
              <div class="permission-tree-wrapper">
                <t-tree
                  :data="apiPermissionTree"
                  :keys="{ label: 'name', value: 'id', children: 'children' }"
                  checkable
                  hover
                  expand-on-click-node
                  value-mode="all"
                  :value="apiPermissionIds"
                  @change="onApiPermissionChange"
                />
              </div>
            </div>
          </div>
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>
<script setup lang="ts">
import type { FormRule, PrimaryTableCol, SubmitContext, TableRowData, TreeNodeValue } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { onMounted, reactive, ref } from 'vue';

import type { PermissionTreeNode, RoleItem } from '@/api/model/rbacModel';
import { createRoleApi, deleteRoleApi, getPermissionTreeApi, getRolesApi, updateRoleApi } from '@/api/rbac';

defineOptions({
  name: 'RbacRoleManage',
});

interface RoleEditForm {
  id?: number;
  name: string;
  description: string;
  permission_ids: number[];
}

type PermissionTreeDisplayNode = Omit<PermissionTreeNode, 'children'> & {
  children: PermissionTreeDisplayNode[];
  disabled?: boolean;
};

const columns: PrimaryTableCol<TableRowData>[] = [
  { title: '序号', colKey: 'index', width: 80 },
  { title: '角色名', colKey: 'name', minWidth: 180 },
  { title: '描述', colKey: 'description', minWidth: 220 },
  { title: '状态', colKey: 'status', width: 110 },
  { title: '操作', colKey: 'op', width: 140, fixed: 'right' },
];

const rules: Record<string, FormRule[]> = {
  name: [{ required: true, message: '请输入角色名', type: 'error' }],
  permission_ids: [{ required: true, message: '请选择权限', type: 'error' }],
};

const roles = ref<RoleItem[]>([]);
const permissionTree = ref<PermissionTreeNode[]>([]);
const menuPermissionTree = ref<PermissionTreeDisplayNode[]>([]);
const apiPermissionTree = ref<PermissionTreeDisplayNode[]>([]);
const permissionTypeMap = ref<Record<number, string>>({});
const menuPermissionIds = ref<number[]>([]);
const apiPermissionIds = ref<number[]>([]);
const loading = ref(false);
const submitLoading = ref(false);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);

const dialogVisible = ref(false);
const dialogTitle = ref('新增角色');
const formRef = ref();

const editForm = reactive<RoleEditForm>({
  name: '',
  description: '',
  permission_ids: [],
});

function resetEditForm() {
  editForm.id = undefined;
  editForm.name = '';
  editForm.description = '';
  editForm.permission_ids = [];
  menuPermissionIds.value = [];
  apiPermissionIds.value = [];
}

async function fetchRoles() {
  loading.value = true;
  try {
    const res = await getRolesApi({ page: page.value, page_size: pageSize.value });
    roles.value = res.roles;
    total.value = res.total;
  } catch (error) {
    console.error(error);
    MessagePlugin.error('获取角色列表失败');
  } finally {
    loading.value = false;
  }
}

async function fetchPermissions() {
  try {
    const res = await getPermissionTreeApi();
    permissionTree.value = res || [];
    permissionTypeMap.value = buildPermissionTypeMap(permissionTree.value);
    menuPermissionTree.value = buildMenuPermissionTree(permissionTree.value);
    apiPermissionTree.value = buildApiPermissionTree(permissionTree.value);
  } catch (error) {
    console.error(error);
    MessagePlugin.error('获取权限树失败');
  }
}

function syncPermissionIds() {
  editForm.permission_ids = Array.from(new Set([...menuPermissionIds.value, ...apiPermissionIds.value]));
}

function onMenuPermissionChange(value: TreeNodeValue[]) {
  menuPermissionIds.value = value
    .map((item) => Number(item))
    .filter((id) => permissionTypeMap.value[id] === 'menu');
  syncPermissionIds();
}

function onApiPermissionChange(value: TreeNodeValue[]) {
  apiPermissionIds.value = value
    .map((item) => Number(item))
    .filter((id) => permissionTypeMap.value[id] === 'api');
  syncPermissionIds();
}

function buildPermissionTypeMap(tree: PermissionTreeNode[]) {
  const map: Record<number, string> = {};
  const walk = (nodes: PermissionTreeNode[]) => {
    nodes.forEach((node) => {
      map[node.id] = node.type;
      if (node.children?.length) {
        walk(node.children);
      }
    });
  };
  walk(tree);
  return map;
}

function buildMenuPermissionTree(nodes: PermissionTreeNode[]): PermissionTreeDisplayNode[] {
  const result: PermissionTreeDisplayNode[] = [];
  nodes.forEach((node) => {
    if (node.type !== 'menu') {
      return;
    }
    result.push({
      ...node,
      children: buildMenuPermissionTree(node.children || []),
    });
  });
  return result;
}

function buildApiPermissionTree(nodes: PermissionTreeNode[]): PermissionTreeDisplayNode[] {
  const walk = (list: PermissionTreeNode[]): PermissionTreeDisplayNode[] => {
    const result: PermissionTreeDisplayNode[] = [];
    list.forEach((node) => {
      if (node.type === 'api') {
        result.push({
          ...node,
          children: [],
          name: `${node.name} [${node.method || '-'} ${node.path || '-'}]`,
        });
        return;
      }

      const apiChildren = walk(node.children || []);
      if (node.type === 'menu' && apiChildren.length > 0) {
        result.push({
          ...node,
          children: apiChildren,
        });
      }
    });
    return result;
  };

  const result = walk(nodes);
  return result;
}

function splitPermissionIds(allIds: number[]) {
  const menuIds: number[] = [];
  const apiIds: number[] = [];
  allIds.forEach((id) => {
    const type = permissionTypeMap.value[id];
    if (type === 'menu') {
      menuIds.push(id);
      return;
    }
    if (type === 'api') {
      apiIds.push(id);
    }
  });
  menuPermissionIds.value = menuIds;
  apiPermissionIds.value = apiIds;
  syncPermissionIds();
}

function onPageSizeChange(size: number) {
  pageSize.value = size;
  page.value = 1;
  fetchRoles();
}

function onAdd() {
  dialogTitle.value = '新增角色';
  resetEditForm();
  dialogVisible.value = true;
}

function onEdit(row: RoleItem) {
  dialogTitle.value = '编辑角色';
  editForm.id = row.id;
  editForm.name = row.name;
  editForm.description = row.description || '';
  splitPermissionIds((row.permissions || []).map((item) => item.id));
  dialogVisible.value = true;
}

async function onDelete(id: number) {
  try {
    await deleteRoleApi(id);
    MessagePlugin.success('删除成功');
    await fetchRoles();
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
      await updateRoleApi(editForm.id, {
        name: editForm.name,
        description: editForm.description,
        permission_ids: editForm.permission_ids,
      });
      MessagePlugin.success('编辑成功');
    } else {
      await createRoleApi({
        name: editForm.name,
        description: editForm.description,
        permission_ids: editForm.permission_ids,
      });
      MessagePlugin.success('新增成功');
    }
    dialogVisible.value = false;
    await fetchRoles();
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
  await Promise.all([fetchRoles(), fetchPermissions()]);
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

.permission-tree-wrapper {
  width: 100%;
  max-height: 320px;
  padding: var(--td-comp-paddingTB-s) var(--td-comp-paddingLR-s);
  overflow: auto;
  border: 1px solid var(--td-border-level-1-color);
  border-radius: var(--td-radius-small);

  :deep(.t-tree__item) {
    background-color: transparent;
  }

  :deep(.t-tree__item:hover) {
    background-color: transparent;
  }

  :deep(.t-tree__label.t-is-checked) {
    background-color: transparent;
  }
}

.permission-sections {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.permission-section-title {
  margin-bottom: 6px;
  font-size: 12px;
  color: var(--td-text-color-secondary);
}

@media (max-width: 900px) {
  .permission-sections {
    grid-template-columns: 1fr;
  }
}
</style>
