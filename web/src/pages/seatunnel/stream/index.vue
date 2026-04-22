<template>
  <div>
    <t-card class="list-card-container" :bordered="false">
      <t-form :data="searchForm" layout="inline" label-width="auto" class="search-form">
        <t-form-item label="作业ID">
          <t-input v-model="searchForm.job_id" clearable placeholder="请输入作业ID" style="width: 200px" />
        </t-form-item>
        <t-form-item label="作业名称">
          <t-input v-model="searchForm.name" clearable placeholder="请输入作业名称" style="width: 200px" />
        </t-form-item>
        <t-form-item label="作业状态">
          <t-select v-model="searchForm.job_status" clearable placeholder="请选择作业状态" style="width: 180px">
            <t-option label="运行中" value="RUNNING" />
            <t-option label="已完成" value="FINISHED" />
            <t-option label="失败" value="FAILED" />
            <t-option label="已取消" value="CANCEL" />
            <t-option label="未知" value="UNKNOWN" />
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
          <t-button theme="success" :loading="syncing" @click="onSyncStatus">同步作业状态</t-button>
        </div>
      </t-row>

      <t-table :data="tasks" :columns="columns" row-key="id" :loading="loading" :hover="true">
        <template #index="{ rowIndex }">{{ rowIndex + 1 + (page - 1) * pageSize }}</template>
        <template #job_id="{ row }">
          <t-tooltip v-if="row.job_id" :content="row.job_id">
            <span>{{ formatJobId(row.job_id) }}</span>
          </t-tooltip>
          <span v-else>-</span>
        </template>
        <template #name="{ row }">
          <t-link theme="primary" hover="color" @click="showDetail(row)">{{ row.name }}</t-link>
        </template>
        <template #job_status="{ row }">
          <t-tag :theme="jobStatusTagTheme(row.job_status)" variant="light">{{ jobStatusText(row.job_status) }}</t-tag>
        </template>
        <template #op="{ row }">
          <t-space>
            <t-link theme="primary" @click="goEditPage(row)">编辑</t-link>
            <t-popconfirm content="确定要删除该任务吗？" @confirm="onDelete(row.id)">
              <t-link theme="danger">删除</t-link>
            </t-popconfirm>
            <t-link theme="default" :disabled="!canSubmitJob(row)" @click="openSubmitDialog(row)">提交作业</t-link>
            <t-link theme="warning" :disabled="!canStopJob(row)" @click="openStopDialog(row)">停止作业</t-link>
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
      v-model:visible="submitDialogVisible"
      header="提交作业"
      width="420px"
      :confirm-btn="{ content: '确定', theme: 'primary', loading: actionLoading }"
      @confirm="doSubmitJob"
    >
      <t-form label-width="180px">
        <t-form-item label="是否使用 SavePoint 启动">
          <t-switch v-model="isStartWithSavePoint" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <t-dialog
      v-model:visible="stopDialogVisible"
      header="停止作业"
      width="420px"
      :confirm-btn="{ content: '确定', theme: 'primary', loading: actionLoading }"
      @confirm="doStopJob"
    >
      <t-form label-width="180px">
        <t-form-item label="是否使用 SavePoint 停止">
          <t-switch v-model="isStopWithSavePoint" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <t-dialog v-model:visible="detailDialogVisible" header="任务详情" width="720px" :footer="false">
      <div class="dialog-scroll-body">
        <t-descriptions :column="1" bordered>
          <t-descriptions-item label="作业ID">
            <t-tooltip v-if="detailTask.job_id" :content="detailTask.job_id">
              <span>{{ formatJobId(detailTask.job_id) }}</span>
            </t-tooltip>
            <span v-else>-</span>
          </t-descriptions-item>
          <t-descriptions-item label="作业名称">{{ detailTask.name }}</t-descriptions-item>
          <t-descriptions-item label="描述">{{ detailTask.description || '-' }}</t-descriptions-item>
          <t-descriptions-item label="作业状态">{{ jobStatusText(detailTask.job_status) }}</t-descriptions-item>
          <t-descriptions-item label="任务类型">{{ detailTask.task_type }}</t-descriptions-item>
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
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onActivated, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';

import type { SeatunnelTask } from '@/api/model/seatunnelModel';
import {
  deleteTaskApi,
  getTasksApi,
  stopJobApi,
  submitJobApi,
  syncJobStatusApi,
} from '@/api/seatunnel';
import CodeEditor from '@/components/code-editor/index.vue';

defineOptions({ name: 'SeatunnelStreamTask' });

const columns: PrimaryTableCol<TableRowData>[] = [
  { title: '序号', colKey: 'index', width: 80 },
  { title: '作业ID', colKey: 'job_id', width: 220, ellipsis: true },
  { title: '作业名称', colKey: 'name', width: 240, ellipsis: true },
  { title: '描述', colKey: 'description', width: 300, ellipsis: true },
  { title: '作业状态', colKey: 'job_status', width: 150 },
  { title: '操作', colKey: 'op', width: 240 },
];

const loading = ref(false);
const actionLoading = ref(false);
const syncing = ref(false);

const tasks = ref<SeatunnelTask[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const router = useRouter();

const searchForm = reactive({
  job_id: '',
  name: '',
  job_status: '',
});

const submitDialogVisible = ref(false);
const stopDialogVisible = ref(false);
const detailDialogVisible = ref(false);

const currentTask = ref<SeatunnelTask | null>(null);
const detailTask = ref<Partial<SeatunnelTask>>({});
const detailConfigLanguage = computed(() => ((detailTask.value.config_format as 'json' | 'hocon') || 'json'));
const isStartWithSavePoint = ref(false);
const isStopWithSavePoint = ref(false);
let skipFirstActivated = true;

function formatDateTime(value?: string) {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString().replaceAll('/', '-');
}

function formatJobId(value?: string) {
  if (!value) return '-';
  return value.length > 30 ? `${value.slice(0, 30)}...` : value;
}

function jobStatusTagTheme(status?: string) {
  switch (status) {
    case 'RUNNING':
      return 'success';
    case 'FAILED':
      return 'danger';
    case 'CANCEL':
      return 'warning';
    case 'FINISHED':
      return 'default';
    default:
      return 'primary';
  }
}

function jobStatusText(status?: string) {
  switch (status) {
    case 'RUNNING':
      return '运行中';
    case 'FAILED':
      return '失败';
    case 'CANCEL':
      return '已取消';
    case 'FINISHED':
      return '已完成';
    case 'UNKNOWN':
      return '未知';
    default:
      return status || '未知';
  }
}

function canSubmitJob(task: SeatunnelTask) {
  return task.job_status !== 'RUNNING';
}

function canStopJob(task: SeatunnelTask) {
  return task.job_status === 'RUNNING';
}

async function fetchTasks() {
  loading.value = true;
  try {
    const params: Record<string, string | number> = {
      task_type: 'stream',
      page: page.value,
      page_size: pageSize.value,
    };
    if (searchForm.job_id) params.job_id = searchForm.job_id;
    if (searchForm.name) params.name = searchForm.name;
    if (searchForm.job_status) params.job_status = searchForm.job_status;

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
  searchForm.job_id = '';
  searchForm.name = '';
  searchForm.job_status = '';
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

async function onDelete(id: number) {
  try {
    await deleteTaskApi(id, 'stream');
    MessagePlugin.success('删除成功');
    await fetchTasks();
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(error instanceof Error ? error.message : '删除失败');
  }
}

function goCreatePage() {
  router.push('/seatunnel/stream/create');
}

function goEditPage(row: SeatunnelTask) {
  router.push(`/seatunnel/stream/${row.id}/edit`);
}

function openSubmitDialog(task: SeatunnelTask) {
  currentTask.value = task;
  isStartWithSavePoint.value = false;
  submitDialogVisible.value = true;
}

function openStopDialog(task: SeatunnelTask) {
  currentTask.value = task;
  isStopWithSavePoint.value = false;
  stopDialogVisible.value = true;
}

async function doSubmitJob() {
  if (!currentTask.value) return;
  actionLoading.value = true;
  try {
    await submitJobApi({
      id: currentTask.value.id,
      isStartWithSavePoint: isStartWithSavePoint.value,
    });
    MessagePlugin.success('提交作业成功');
    submitDialogVisible.value = false;
    await fetchTasks();
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(error instanceof Error ? error.message : '提交作业失败');
  } finally {
    actionLoading.value = false;
  }
}

async function doStopJob() {
  if (!currentTask.value) return;
  actionLoading.value = true;
  try {
    await stopJobApi({
      id: currentTask.value.id,
      isStopWithSavePoint: isStopWithSavePoint.value,
    });
    MessagePlugin.success('停止作业成功');
    stopDialogVisible.value = false;
    await fetchTasks();
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(error instanceof Error ? error.message : '停止作业失败');
  } finally {
    actionLoading.value = false;
  }
}

async function onSyncStatus() {
  syncing.value = true;
  try {
    await syncJobStatusApi();
    MessagePlugin.success('同步作业状态已触发');
    await fetchTasks();
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(error instanceof Error ? error.message : '同步失败');
  } finally {
    syncing.value = false;
  }
}

onMounted(async () => {
  await fetchTasks();
});

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
