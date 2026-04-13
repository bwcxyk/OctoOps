<template>
  <div>
    <t-card class="list-card-container" :bordered="false">
      <t-form :data="searchForm" layout="inline" label-width="auto" class="search-form">
        <t-form-item label="任务">
          <t-input v-model="searchForm.task_name" clearable placeholder="请输入任务名称" style="width: 220px" />
        </t-form-item>
        <t-form-item label="运行状态">
          <t-select v-model="searchForm.status" clearable placeholder="请选择状态" style="width: 140px">
            <t-option label="成功" value="success" />
            <t-option label="失败" value="failed" />
          </t-select>
        </t-form-item>
        <t-form-item label="开始时间">
          <t-date-picker v-model="searchForm.start_time" enable-time-picker clearable format="YYYY-MM-DD HH:mm:ss" />
        </t-form-item>
        <t-form-item label="结束时间">
          <t-date-picker v-model="searchForm.end_time" enable-time-picker clearable format="YYYY-MM-DD HH:mm:ss" />
        </t-form-item>
        <t-form-item>
          <t-button theme="primary" @click="onSearch">查询</t-button>
          <t-button variant="outline" @click="onReset">重置</t-button>
        </t-form-item>
      </t-form>

      <t-table :data="logs" :columns="columns" row-key="id" :loading="loading" :hover="true" table-layout="fixed">
        <template #task_name="{ row }">
          <t-tooltip v-if="row.task_name" :content="row.task_name">
            <t-link theme="primary" hover="color" @click="showDetail(row)">{{ formatTaskName(row.task_name) }}</t-link>
          </t-tooltip>
          <span v-else>-</span>
        </template>
        <template #status="{ row }">
          <t-tag :theme="row.status === 'success' ? 'success' : 'danger'" variant="light">
            {{ row.status === 'success' ? '成功' : row.status === 'failed' ? '失败' : row.status }}
          </t-tag>
        </template>
        <template #created_at="{ row }">{{ formatDateTime(row.created_at) }}</template>
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

    <t-dialog v-model:visible="detailDialogVisible" header="日志详情" width="720px" :footer="false">
      <t-descriptions :column="1" bordered>
        <t-descriptions-item label="任务名称">{{ detailLog.task_name }}</t-descriptions-item>
        <t-descriptions-item label="运行状态">{{ formatStatusText(detailLog.status) }}</t-descriptions-item>
        <t-descriptions-item label="返回内容">
          <t-textarea :model-value="detailLog.result || ''" readonly :autosize="{ minRows: 8, maxRows: 12 }" />
        </t-descriptions-item>
        <t-descriptions-item label="创建时间">{{ formatDateTime(detailLog.created_at) }}</t-descriptions-item>
      </t-descriptions>
    </t-dialog>
  </div>
</template>
<script setup lang="ts">
import type { PrimaryTableCol, TableRowData } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { onMounted, reactive, ref } from 'vue';

import type { TaskLogItem } from '@/api/model/seatunnelModel';
import { getTaskLogsApi } from '@/api/seatunnel';

defineOptions({ name: 'TaskLogManage' });

const columns: PrimaryTableCol<TableRowData>[] = [
  { title: 'ID', colKey: 'id', width: 80 },
  { title: '任务名称', colKey: 'task_name', width: 160, ellipsis: true },
  { title: '运行状态', colKey: 'status', width: 100 },
  { title: '创建时间', colKey: 'created_at', width: 160, ellipsis: true },
];

const logs = ref<TaskLogItem[]>([]);
const total = ref(0);
const loading = ref(false);
const page = ref(1);
const pageSize = ref(10);

const detailDialogVisible = ref(false);
const detailLog = ref<Partial<TaskLogItem & { result?: string }>>({});

const searchForm = reactive<{ task_name: string; status: string; start_time: string; end_time: string }>({
  task_name: '',
  status: '',
  start_time: '',
  end_time: '',
});

function formatDateTime(value?: string) {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString().replaceAll('/', '-');
}

function formatTaskName(value?: string) {
  if (!value) return '-';
  return value.length > 30 ? `${value.slice(0, 30)}...` : value;
}

function formatStatusText(status?: string) {
  if (!status) return '-';
  if (status === 'success') return '成功';
  if (status === 'failed') return '失败';
  return status;
}

async function fetchLogs() {
  loading.value = true;
  try {
    const params: Record<string, string | number> = {
      page: page.value,
      page_size: pageSize.value,
    };
    if (searchForm.task_name) params.task_name = searchForm.task_name;
    if (searchForm.status) params.status = searchForm.status;
    if (searchForm.start_time) params.start_time = searchForm.start_time;
    if (searchForm.end_time) params.end_time = searchForm.end_time;

    const res = await getTaskLogsApi(params as any);
    logs.value = Array.isArray(res.data) ? res.data : [];
    total.value = typeof res.total === 'number' ? res.total : 0;
  } catch (error) {
    console.error(error);
    MessagePlugin.error('获取日志失败');
  } finally {
    loading.value = false;
  }
}

function onSearch() {
  page.value = 1;
  fetchLogs();
}

function onReset() {
  searchForm.task_name = '';
  searchForm.status = '';
  searchForm.start_time = '';
  searchForm.end_time = '';
  page.value = 1;
  fetchLogs();
}

function onPageChange(current: number) {
  page.value = current;
  fetchLogs();
}

function onPageSizeChange(size: number) {
  pageSize.value = size;
  page.value = 1;
  fetchLogs();
}

function showDetail(row: TaskLogItem) {
  detailLog.value = { ...row } as any;
  detailDialogVisible.value = true;
}

onMounted(fetchLogs);
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

.list-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--td-comp-margin-xxl);
}
</style>
