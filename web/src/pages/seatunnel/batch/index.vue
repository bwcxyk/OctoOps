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
          <t-button @click="goCreatePage">新增</t-button>
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
            <t-link theme="primary" @click="goEditPage(row)">编辑</t-link>
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

    <t-dialog v-model:visible="detailDialogVisible" header="任务详情" width="720px" :footer="false">
      <div class="dialog-scroll-body">
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
            <CodeEditor
              :model-value="detailTask.config || ''"
              :active="detailDialogVisible"
              :language="detailConfigLanguage"
              :height="220"
              readonly
            />
          </t-descriptions-item>
        </t-descriptions>
      </div>
    </t-dialog>
  </div>
</template>
<script setup lang="ts">
import type { PrimaryTableCol, TableRowData } from 'tdesign-vue-next';
import { DialogPlugin, MessagePlugin } from 'tdesign-vue-next';
import { computed, onActivated, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';

import type { SeatunnelTask } from '@/api/model/seatunnelModel';
import { deleteTaskApi, getTasksApi, submitJobApi, updateTaskApi } from '@/api/seatunnel';
import CodeEditor from '@/components/code-editor/index.vue';
import { isValidCronExpression } from '@/utils/cron';

defineOptions({ name: 'SeatunnelBatchTask' });

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

const loading = ref(false);
const tasks = ref<SeatunnelTask[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const router = useRouter();

const searchForm = reactive({
  name: '',
  status: undefined as number | undefined,
});

const detailDialogVisible = ref(false);

const detailTask = ref<Partial<SeatunnelTask>>({});
const detailConfigLanguage = computed(() => ((detailTask.value.config_format as 'json' | 'hocon') || 'json'));
let skipFirstActivated = true;

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

function showDetail(row: SeatunnelTask) {
  detailTask.value = { ...row };
  detailDialogVisible.value = true;
}

function goCreatePage() {
  router.push('/seatunnel/batch/create');
}

function goEditPage(row: SeatunnelTask) {
  router.push(`/seatunnel/batch/${row.id}/edit`);
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

onActivated(() => {
  if (skipFirstActivated) {
    skipFirstActivated = false;
    return;
  }
  fetchTasks();
});
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

.dialog-scroll-body {
  overflow: hidden;
}
</style>
