<template>
  <div>
    <t-card class="list-card-container" :bordered="false">
      <t-form :data="searchForm" layout="inline" label-width="auto" class="search-form">
        <t-form-item label="作业名称">
          <t-input v-model="searchForm.name" clearable placeholder="请输入作业名称" style="width: 220px" />
        </t-form-item>
        <t-form-item label="状态">
          <t-select v-model="searchForm.status" clearable placeholder="请选择状态" style="width: 160px">
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
          <t-button @click="openEditDialog()">新增</t-button>
        </div>
      </t-row>

      <t-table :data="tasks" :columns="columns" row-key="id" :loading="loading" :hover="true">
        <template #index="{ rowIndex }">{{ rowIndex + 1 + (page - 1) * pageSize }}</template>
        <template #name="{ row }">
          <t-link theme="primary" hover="color" @click="showDetail(row)">{{ row.name }}</t-link>
        </template>
        <template #status="{ row }">
          <t-switch
            :value="Number(row.status || 0)"
            :custom-value="[1, 0]"
            @change="(value) => onStatusChange(row, Number(value) as 0 | 1)"
          />
        </template>
        <template #next_run_time="{ row }">
          {{ row.next_run_time ? formatDateTime(row.next_run_time) : '未设置' }}
        </template>
        <template #last_run_time="{ row }">
          {{ row.last_run_time ? formatDateTime(row.last_run_time) : '未运行' }}
        </template>
        <template #op="{ row }">
          <t-space>
            <t-link theme="primary" @click="openEditDialog(row)">编辑</t-link>
            <t-link theme="default" @click="onManualExecute(row)">手动执行</t-link>
            <t-popconfirm content="确定要删除该任务吗？" @confirm="onDelete(row.id)">
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
          @current-change="onPageChange"
          @page-size-change="onPageSizeChange"
        />
      </div>
    </t-card>

    <t-dialog
      v-model:visible="editDialogVisible"
      :header="editForm.id ? '编辑批任务' : '新增批任务'"
      width="900px"
      :confirm-btn="{ content: '保存', theme: 'primary', loading: submitLoading }"
      @confirm="onSubmit"
    >
      <t-form ref="formRef" :data="editForm" :rules="rules" label-width="110px" @submit="onFormSubmit">
        <t-form-item label="作业名称" name="name">
          <t-input v-model="editForm.name" />
        </t-form-item>
        <t-form-item label="描述" name="description">
          <t-input v-model="editForm.description" />
        </t-form-item>
        <t-form-item label="状态" name="status">
          <t-switch v-model="editForm.status" :custom-value="[1, 0]" />
        </t-form-item>
        <t-form-item label="Cron表达式" name="cron_expr">
          <t-input v-model="editForm.cron_expr" placeholder="秒 分 时 日 月 周" />
        </t-form-item>
        <t-form-item label="配置风格" name="config_format">
          <t-select v-model="editForm.config_format" style="width: 180px">
            <t-option label="JSON" value="json" />
            <t-option label="HOCON" value="hocon" />
          </t-select>
        </t-form-item>
        <t-form-item label="作业配置" name="config">
          <t-textarea v-model="editForm.config" :autosize="{ minRows: 10, maxRows: 16 }" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <t-dialog v-model:visible="detailDialogVisible" header="任务详情" width="720px" :footer="false">
      <t-descriptions :column="1" bordered>
        <t-descriptions-item label="作业名称">{{ detailTask.name }}</t-descriptions-item>
        <t-descriptions-item label="描述">{{ detailTask.description || '-' }}</t-descriptions-item>
        <t-descriptions-item label="状态">{{
          Number(detailTask.status || 0) === 1 ? '启用' : '禁用'
        }}</t-descriptions-item>
        <t-descriptions-item label="任务类型">{{ detailTask.task_type }}</t-descriptions-item>
        <t-descriptions-item label="Cron表达式">{{ detailTask.cron_expr || '-' }}</t-descriptions-item>
        <t-descriptions-item label="下次执行时间">
          {{ detailTask.next_run_time ? formatDateTime(detailTask.next_run_time) : '未设置' }}
        </t-descriptions-item>
        <t-descriptions-item label="最后运行时间">
          {{ detailTask.last_run_time ? formatDateTime(detailTask.last_run_time) : '未运行' }}
        </t-descriptions-item>
        <t-descriptions-item label="创建时间">{{ formatDateTime(detailTask.created_at) }}</t-descriptions-item>
        <t-descriptions-item label="更新时间">{{ formatDateTime(detailTask.updated_at) }}</t-descriptions-item>
        <t-descriptions-item label="作业配置">
          <t-textarea :model-value="detailTask.config || ''" readonly :autosize="{ minRows: 10, maxRows: 14 }" />
        </t-descriptions-item>
      </t-descriptions>
    </t-dialog>
  </div>
</template>
<script setup lang="ts">
import type { FormRule, PrimaryTableCol, SubmitContext, TableRowData } from 'tdesign-vue-next';
import { DialogPlugin, MessagePlugin } from 'tdesign-vue-next';
import { onMounted, reactive, ref } from 'vue';

import type { SeatunnelTask } from '@/api/model/seatunnelModel';
import { createTaskApi, deleteTaskApi, getTasksApi, submitJobApi, updateTaskApi } from '@/api/seatunnel';
import { isValidCronExpression } from '@/utils/cron';

defineOptions({ name: 'SeatunnelBatchTask' });

interface BatchEditForm {
  id?: number;
  name: string;
  description: string;
  status: 0 | 1;
  cron_expr: string;
  config_format: 'json' | 'hocon';
  config: string;
}

const columns: PrimaryTableCol<TableRowData>[] = [
  { title: '序号', colKey: 'index', width: 80 },
  { title: '作业名称', colKey: 'name', minWidth: 200 },
  { title: '描述', colKey: 'description', minWidth: 220, ellipsis: true },
  { title: '状态', colKey: 'status', width: 110 },
  { title: 'Cron表达式', colKey: 'cron_expr', width: 180 },
  { title: '下次执行时间', colKey: 'next_run_time', width: 180 },
  { title: '最后运行时间', colKey: 'last_run_time', width: 180 },
  { title: '操作', colKey: 'op', width: 240 },
];

const rules: Record<string, FormRule[]> = {
  name: [{ required: true, message: '请输入作业名称', type: 'error' }],
  config: [{ required: true, message: '请输入作业配置', type: 'error' }],
  cron_expr: [
    {
      validator: (_val) => {
        const expr = (editForm.cron_expr || '').trim();
        // 禁用状态允许为空，启用状态必须填写
        if (editForm.status !== 1 && !expr) return true;
        if (!expr) return { result: false, message: '批处理任务启用时必须填写 Cron 表达式' };
        return isValidCronExpression(expr) ? true : { result: false, message: '无效的 Cron 表达式' };
      },
      type: 'warning',
    },
  ],
};

const loading = ref(false);
const submitLoading = ref(false);
const tasks = ref<SeatunnelTask[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);

const searchForm = reactive({
  name: '',
  status: undefined as number | undefined,
});

const editDialogVisible = ref(false);
const detailDialogVisible = ref(false);
const formRef = ref();

const editForm = reactive<BatchEditForm>({
  id: undefined,
  name: '',
  description: '',
  status: 0,
  cron_expr: '',
  config_format: 'json',
  config: '',
});

const detailTask = ref<Partial<SeatunnelTask>>({});

function formatDateTime(value?: string) {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString().replaceAll('/', '-');
}

async function fetchTasks() {
  loading.value = true;
  try {
    const params: Record<string, string | number> = {
      task_type: 'batch',
      page: page.value,
      page_size: pageSize.value,
    };
    if (searchForm.name) params.name = searchForm.name;
    if (typeof searchForm.status === 'number') params.status = searchForm.status;

    const res = await getTasksApi(params as any);
    tasks.value = Array.isArray(res.data) ? res.data : [];
    total.value = typeof res.total === 'number' ? res.total : 0;
  } catch (error) {
    console.error(error);
    MessagePlugin.error('获取任务列表失败');
  } finally {
    loading.value = false;
  }
}

function onSearch() {
  page.value = 1;
  fetchTasks();
}

function onReset() {
  searchForm.name = '';
  searchForm.status = undefined;
  page.value = 1;
  fetchTasks();
}

function onPageChange(current: number) {
  page.value = current;
  fetchTasks();
}

function onPageSizeChange(size: number) {
  pageSize.value = size;
  page.value = 1;
  fetchTasks();
}

function resetEditForm() {
  editForm.id = undefined;
  editForm.name = '';
  editForm.description = '';
  editForm.status = 0;
  editForm.cron_expr = '';
  editForm.config_format = 'json';
  editForm.config = '';
}

function openEditDialog(row?: SeatunnelTask) {
  if (row) {
    editForm.id = row.id;
    editForm.name = row.name || '';
    editForm.description = row.description || '';
    editForm.status = Number(row.status || 0) === 1 ? 1 : 0;
    editForm.cron_expr = row.cron_expr || '';
    editForm.config_format = (row.config_format as 'json' | 'hocon') || 'json';
    editForm.config = row.config || '';
  } else {
    resetEditForm();
  }
  editDialogVisible.value = true;
}

function showDetail(row: SeatunnelTask) {
  detailTask.value = { ...row };
  detailDialogVisible.value = true;
}

async function onSubmit() {
  await formRef.value?.submit();
}

async function onFormSubmit(ctx: SubmitContext) {
  if (ctx.validateResult !== true) return;
  if (editForm.status === 1 && !editForm.cron_expr.trim()) {
    MessagePlugin.error('批处理任务启用时必须填写 Cron 表达式');
    return;
  }
  if (editForm.cron_expr.trim() && !isValidCronExpression(editForm.cron_expr.trim())) {
    MessagePlugin.error('无效的 Cron 表达式');
    return;
  }

  submitLoading.value = true;
  try {
    const payload: Partial<SeatunnelTask> = {
      name: editForm.name,
      description: editForm.description,
      status: editForm.status,
      cron_expr: editForm.cron_expr,
      config: editForm.config,
      config_format: editForm.config_format,
      task_type: 'batch',
    };

    if (editForm.id) {
      await updateTaskApi(editForm.id, payload, 'batch');
      MessagePlugin.success('更新成功');
    } else {
      await createTaskApi(payload);
      MessagePlugin.success('创建成功');
    }
    editDialogVisible.value = false;
    await fetchTasks();
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(error instanceof Error ? error.message : '保存失败');
  } finally {
    submitLoading.value = false;
  }
}

async function onDelete(id: number) {
  try {
    await deleteTaskApi(id, 'batch');
    MessagePlugin.success('删除成功');
    await fetchTasks();
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(error instanceof Error ? error.message : '删除失败');
  }
}

async function onStatusChange(row: SeatunnelTask, status: 0 | 1) {
  if (status === 1 && !row.cron_expr?.trim()) {
    MessagePlugin.error('批处理任务启用时必须填写 Cron 表达式');
    await fetchTasks();
    return;
  }
  if (status === 1 && row.cron_expr?.trim() && !isValidCronExpression(row.cron_expr.trim())) {
    MessagePlugin.error('无效的 Cron 表达式');
    await fetchTasks();
    return;
  }

  try {
    await updateTaskApi(row.id, { status }, 'batch');
    MessagePlugin.success('状态更新成功');
    await fetchTasks();
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(error instanceof Error ? error.message : '状态更新失败');
  }
}

function onManualExecute(task: SeatunnelTask) {
  const dialog = DialogPlugin.confirm({
    header: '确认执行',
    body: `确定要手动执行任务「${task.name}」吗？`,
    confirmBtn: '确定',
    cancelBtn: '取消',
    onConfirm: async () => {
      try {
        await submitJobApi({ id: task.id });
        MessagePlugin.success('手动执行成功');
        await fetchTasks();
      } catch (error: unknown) {
        console.error(error);
        MessagePlugin.error(error instanceof Error ? error.message : '手动执行失败');
      } finally {
        dialog.destroy();
      }
    },
    onClose: () => {
      dialog.destroy();
    },
  });
}

onMounted(fetchTasks);
</script>
<style lang="less" scoped>
.list-card-container {
  padding: var(--td-comp-paddingTB-xxl) var(--td-comp-paddingLR-xxl);

  :deep(.t-card__body) {
    padding: 0;
  }
}

.search-form {
  margin-bottom: var(--td-comp-margin-l);
}

.left-operation-container {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: var(--td-comp-margin-xxl);
}

.list-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--td-comp-margin-xxl);
}
</style>
