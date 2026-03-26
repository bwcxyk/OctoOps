<template>
  <div>
    <t-card class="list-card-container" :bordered="false">
      <t-form :data="searchForm" layout="inline" label-width="auto" class="search-form">
        <t-form-item label="名称">
          <t-input v-model="searchForm.name" clearable placeholder="名称" style="width: 180px" />
        </t-form-item>
        <t-form-item label="AccessKey">
          <t-input v-model="searchForm.access_key" clearable placeholder="AccessKey" style="width: 220px" />
        </t-form-item>
        <t-form-item label="状态">
          <t-select v-model="searchForm.status" clearable placeholder="请选择状态" style="width: 140px">
            <t-option label="启用" :value="1" />
            <t-option label="禁用" :value="0" />
          </t-select>
        </t-form-item>
        <t-form-item>
          <t-button theme="primary" @click="onSearch">查询</t-button>
          <t-button variant="outline" @click="onReset">重置</t-button>
        </t-form-item>
      </t-form>

      <t-row justify="space-between">
        <div class="left-operation-container">
          <t-button @click="openDialog()">新增</t-button>
          <t-popup trigger="hover" placement="bottom-left">
            <template #content>
              <div class="tips-content">
                <div>RAM子账号需要以下权限才能正常操作安全组：</div>
                <pre>
ecs:AuthorizeSecurityGroup
ecs:AuthorizeSecurityGroupEgress
ecs:ModifySecurityGroupEgressRule
ecs:ModifySecurityGroupRule
ecs:RevokeSecurityGroup
ecs:RevokeSecurityGroupEgress
ecs:DescribeSecurityGroupAttribute</pre
                >
              </div>
            </template>
            <t-button variant="text" theme="default" shape="square" class="tips-icon-btn">
              <t-icon name="help-circle" />
            </t-button>
          </t-popup>
        </div>
      </t-row>

      <t-table :data="configs" :columns="columns" row-key="id" :loading="loading" :hover="true">
        <template #index="{ rowIndex }">{{ rowIndex + 1 }}</template>
        <template #last_ip_updated_at="{ row }">{{ formatDateTime(row.last_ip_updated_at) }}</template>
        <template #updated_at="{ row }">{{ formatDateTime(row.updated_at) }}</template>
        <template #status="{ row }">
          <t-switch
            :value="Number(row.status || 0)"
            :custom-value="[1, 0]"
            @change="(value) => onStatusChange(row, Number(value) as 0 | 1)"
          />
        </template>
        <template #op="{ row }">
          <t-space>
            <t-link theme="success" @click="onSyncOne(row)">同步</t-link>
            <t-link theme="primary" @click="openDialog(row)">编辑</t-link>
            <t-popconfirm content="确定要删除该配置吗？" @confirm="onDelete(row.id)">
              <t-link theme="danger">删除</t-link>
            </t-popconfirm>
          </t-space>
        </template>
      </t-table>
    </t-card>

    <t-dialog
      v-model:visible="dialogVisible"
      :header="editForm.id ? '编辑配置' : '新增配置'"
      width="640px"
      :confirm-btn="{ content: '保存', theme: 'primary', loading: submitLoading }"
      @confirm="onSubmit"
    >
      <t-form ref="formRef" :data="editForm" :rules="rules" label-width="120px" @submit="onFormSubmit">
        <t-form-item label="名称" name="name">
          <t-input v-model="editForm.name" />
        </t-form-item>
        <t-form-item label="AccessKey" name="access_key">
          <t-input v-model="editForm.access_key" />
        </t-form-item>
        <t-form-item label="AccessSecret" name="access_secret">
          <t-input v-model="editForm.access_secret" type="password" />
        </t-form-item>
        <t-form-item label="RegionId" name="region_id">
          <t-input v-model="editForm.region_id" />
        </t-form-item>
        <t-form-item label="安全组ID" name="security_group_id">
          <t-input v-model="editForm.security_group_id" />
        </t-form-item>
        <t-form-item label="端口列表" name="port_list">
          <t-input v-model="editForm.port_list" placeholder="如 22,80,443" />
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>
<script setup lang="ts">
import type { FormRule, PrimaryTableCol, SubmitContext, TableRowData } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { onMounted, reactive, ref } from 'vue';

import {
  createAliyunSgConfigApi,
  deleteAliyunSgConfigApi,
  getAliyunSgConfigsApi,
  syncAliyunSgConfigApi,
  updateAliyunSgConfigApi,
} from '@/api/aliyun';
import type { AliyunSgConfig } from '@/api/model/aliyunModel';

defineOptions({ name: 'AliyunEcsSecurityGroup' });

const columns: PrimaryTableCol<TableRowData>[] = [
  { title: '序号', colKey: 'index', width: 80 },
  { title: '名称', colKey: 'name', width: 140 },
  { title: 'AccessKey', colKey: 'access_key', width: 220 },
  { title: '安全组ID', colKey: 'security_group_id', width: 200 },
  { title: '端口列表', colKey: 'port_list', width: 150, ellipsis: true },
  { title: '上次授权IP', colKey: 'last_ip', width: 150 },
  { title: '同步时间', colKey: 'last_ip_updated_at', width: 180 },
  { title: '更新时间', colKey: 'updated_at', width: 180 },
  { title: '状态', colKey: 'status', width: 100 },
  { title: '操作', colKey: 'op', width: 180 },
];

const rules: Record<string, FormRule[]> = {
  name: [{ required: true, message: '名称必填', type: 'error' }],
  access_key: [{ required: true, message: 'AccessKey必填', type: 'error' }],
  access_secret: [{ required: true, message: 'AccessSecret必填', type: 'error' }],
  region_id: [{ required: true, message: 'RegionId必填', type: 'error' }],
  security_group_id: [{ required: true, message: '安全组ID必填', type: 'error' }],
  port_list: [{ required: true, message: '端口列表必填', type: 'error' }],
};

const configs = ref<AliyunSgConfig[]>([]);
const loading = ref(false);
const submitLoading = ref(false);
const dialogVisible = ref(false);
const formRef = ref();

const searchForm = reactive<{ name: string; access_key: string; status: number | undefined }>({
  name: '',
  access_key: '',
  status: undefined,
});

const editForm = reactive<Partial<AliyunSgConfig>>({
  id: undefined,
  name: '',
  access_key: '',
  access_secret: '',
  region_id: '',
  security_group_id: '',
  port_list: '',
  status: 1,
});

function formatDateTime(value?: string) {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString().replaceAll('/', '-');
}

async function fetchConfigs() {
  loading.value = true;
  try {
    const params: Record<string, string | number> = {};
    if (searchForm.name) params.name = searchForm.name;
    if (searchForm.access_key) params.access_key = searchForm.access_key;
    if (typeof searchForm.status === 'number') params.status = searchForm.status;
    configs.value = await getAliyunSgConfigsApi(params as any);
  } catch (error) {
    console.error(error);
    MessagePlugin.error('获取配置失败');
  } finally {
    loading.value = false;
  }
}

function resetEditForm() {
  editForm.id = undefined;
  editForm.name = '';
  editForm.access_key = '';
  editForm.access_secret = '';
  editForm.region_id = '';
  editForm.security_group_id = '';
  editForm.port_list = '';
  editForm.status = 1;
}

function openDialog(row?: AliyunSgConfig) {
  if (row) {
    Object.assign(editForm, { ...row });
  } else {
    resetEditForm();
  }
  dialogVisible.value = true;
}

function onSearch() {
  fetchConfigs();
}

function onReset() {
  searchForm.name = '';
  searchForm.access_key = '';
  searchForm.status = undefined;
  fetchConfigs();
}

async function onSubmit() {
  await formRef.value?.submit();
}

async function onFormSubmit(ctx: SubmitContext) {
  if (ctx.validateResult !== true) return;
  submitLoading.value = true;
  try {
    const payload: Partial<AliyunSgConfig> = {
      name: editForm.name,
      access_key: editForm.access_key,
      access_secret: editForm.access_secret,
      region_id: editForm.region_id,
      security_group_id: editForm.security_group_id,
      port_list: editForm.port_list,
      status: Number(editForm.status || 1) as 0 | 1,
    };
    if (editForm.id) {
      await updateAliyunSgConfigApi(editForm.id, payload);
      MessagePlugin.success('更新成功');
    } else {
      await createAliyunSgConfigApi(payload);
      MessagePlugin.success('创建成功');
    }
    dialogVisible.value = false;
    await fetchConfigs();
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(error instanceof Error ? error.message : '保存失败');
  } finally {
    submitLoading.value = false;
  }
}

async function onDelete(id: number) {
  try {
    await deleteAliyunSgConfigApi(id);
    MessagePlugin.success('删除成功');
    await fetchConfigs();
  } catch (error) {
    console.error(error);
    MessagePlugin.error('删除失败');
  }
}

async function onSyncOne(row: AliyunSgConfig) {
  try {
    await syncAliyunSgConfigApi(row.id);
    MessagePlugin.success('同步成功');
    await fetchConfigs();
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(error instanceof Error ? error.message : '同步失败');
  }
}

async function onStatusChange(row: AliyunSgConfig, status: 0 | 1) {
  try {
    await updateAliyunSgConfigApi(row.id, { status });
    MessagePlugin.success('状态更新成功');
    await fetchConfigs();
  } catch (error) {
    console.error(error);
    MessagePlugin.error('状态更新失败');
    await fetchConfigs();
  }
}

onMounted(fetchConfigs);
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
  gap: 12px;
  margin-bottom: var(--td-comp-margin-l);
}

.tips-icon-btn {
  padding: 0;
}

.search-form {
  margin-bottom: var(--td-comp-margin-l);
}

.tips-content {
  max-width: 340px;
  line-height: 1.6;
  padding: 8px 10px;

  pre {
    margin-top: 8px;
    padding: 8px;
    border-radius: 4px;
    background: var(--td-bg-color-container-hover);
    white-space: pre-wrap;
    font-size: 12px;
  }
}
</style>
