<template>
  <div>
    <t-card class="list-card-container" :bordered="false" :loading="loading">
      <div class="header-bar">
        <t-button @click="fetchStatus">刷新</t-button>
        <t-space>
          <t-button theme="primary" @click="onReload">重新加载调度器</t-button>
          <t-button theme="warning" @click="onStop">停止调度器</t-button>
          <t-button theme="success" @click="onStart">启动调度器</t-button>
        </t-space>
      </div>

      <div class="status-row">
        <div class="status-item">
          <div class="status-item__label">活跃任务数量</div>
          <div class="status-item__value">{{ scheduler.active_tasks_count || 0 }}</div>
        </div>
        <div class="status-item">
          <div class="status-item__label">调度器状态</div>
          <t-tag :theme="scheduler.scheduler_running ? 'success' : 'warning'" variant="light">
            {{ schedulerRunningText }}
          </t-tag>
        </div>
      </div>

      <div class="toolbar">
        <t-select v-model="taskTypeFilter" clearable placeholder="任务类型" style="width: 160px">
          <t-option label="ETL任务" value="etl" />
          <t-option label="自定义任务" value="custom" />
        </t-select>
      </div>

      <t-table :data="pagedTasks" :columns="columns" row-key="entry_id" :hover="true">
        <template #index="{ rowIndex }">{{ rowIndex + 1 + (page - 1) * pageSize }}</template>
        <template #task_type="{ row }">
          <t-tag :theme="row.task_type === 'etl' ? 'success' : 'primary'" variant="light">
            {{ row.task_type === 'etl' ? 'ETL任务' : '自定义任务' }}
          </t-tag>
        </template>
        <template #next_run="{ row }">{{ formatDateTime(row.next_run) }}</template>
        <template #remain="{ row }">{{ getTimeRemaining(row.next_run) }}</template>
      </t-table>

      <div class="list-pagination">
        <t-pagination
          v-model="page"
          v-model:page-size="pageSize"
          :total="filteredTasks.length"
          show-jumper
          show-page-size
          :page-size-options="[10, 20, 50, 100]"
        />
      </div>
    </t-card>
  </div>
</template>
<script setup lang="ts">
import type { PrimaryTableCol, TableRowData } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref } from 'vue';

import type { SchedulerStatusResult } from '@/api/model/taskModel';
import { getSchedulerStatusApi, reloadSchedulerApi, startSchedulerApi, stopSchedulerApi } from '@/api/task';

defineOptions({ name: 'SchedulerManage' });

const columns: PrimaryTableCol<TableRowData>[] = [
  { title: '序号', colKey: 'index', width: 80 },
  { title: '任务名称', colKey: 'task_name', minWidth: 220 },
  { title: '任务类型', colKey: 'task_type', width: 120 },
  { title: '下次执行时间', colKey: 'next_run', width: 200 },
  { title: '剩余时间', colKey: 'remain', width: 160 },
];

const scheduler = ref<SchedulerStatusResult>({
  scheduler_running: true,
  active_tasks_count: 0,
  active_tasks: [],
});
const loading = ref(false);
const taskTypeFilter = ref('');
const page = ref(1);
const pageSize = ref(10);

const filteredTasks = computed(() => {
  const tasks = scheduler.value.active_tasks || [];
  if (!taskTypeFilter.value) return tasks;
  return tasks.filter((item) => item.task_type === taskTypeFilter.value);
});

const pagedTasks = computed(() => {
  const start = (page.value - 1) * pageSize.value;
  return filteredTasks.value.slice(start, start + pageSize.value);
});

const schedulerRunningText = computed(() => (scheduler.value.scheduler_running ? '运行中' : '已停止'));

function formatDateTime(value?: string) {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return String(value);
  return date.toLocaleString().replaceAll('/', '-');
}

function getTimeRemaining(nextRun?: string) {
  if (!nextRun) return '-';
  const diff = new Date(nextRun).getTime() - Date.now();
  if (Number.isNaN(diff)) return '-';
  if (diff <= 0) return '即将执行';
  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
  const seconds = Math.floor((diff % (1000 * 60)) / 1000);
  if (hours > 0) return `${hours}小时${minutes}分钟`;
  if (minutes > 0) return `${minutes}分钟${seconds}秒`;
  return `${seconds}秒`;
}

async function fetchStatus() {
  loading.value = true;
  try {
    scheduler.value = await getSchedulerStatusApi();
  } catch (error) {
    console.error(error);
    MessagePlugin.error('获取调度器状态失败');
  } finally {
    loading.value = false;
  }
}

async function doAction(action: () => Promise<unknown>, successMessage: string) {
  try {
    await action();
    MessagePlugin.success(successMessage);
    await fetchStatus();
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(error instanceof Error ? error.message : '操作失败');
  }
}

function onReload() {
  doAction(reloadSchedulerApi, '调度器重新加载成功');
}

function onStop() {
  doAction(stopSchedulerApi, '调度器已停止');
}

function onStart() {
  doAction(startSchedulerApi, '调度器已启动');
}

onMounted(fetchStatus);
</script>
<style lang="less" scoped>
.list-card-container {
  padding: var(--td-comp-paddingTB-xxl) var(--td-comp-paddingLR-xxl);

  :deep(.t-card__body) {
    padding: 0;
  }
}

.header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: var(--td-comp-margin-l);
}

.status-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 24px;
  margin-bottom: var(--td-comp-margin-l);
}

.status-item {
  min-width: 140px;
}

.status-item__label {
  margin-bottom: 6px;
  color: var(--td-text-color-secondary);
}

.status-item__value {
  font-size: 28px;
  line-height: 1;
}

.toolbar {
  display: flex;
  align-items: center;
  margin-bottom: var(--td-comp-margin-l);
}

@media (max-width: 992px) {
  .header-bar {
    flex-wrap: wrap;
    justify-content: flex-start;
  }
}

.list-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--td-comp-margin-xxl);
}
</style>
