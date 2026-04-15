<template>
  <div>
    <t-card class="list-card-container" :bordered="false">
      <t-row justify="space-between">
        <div class="left-operation-container">
          <t-button @click="fetchTasks">刷新</t-button>
        </div>
      </t-row>

      <t-table :data="tasks" :columns="columns" row-key="id" :loading="loading" :hover="true">
        <template #index="{ rowIndex }">{{ rowIndex + 1 + (page - 1) * pageSize }}</template>
        <template #custom_type="{ row }">{{ row.custom_type || '-' }}</template>
        <template #status="{ row }">
          <t-switch
            :value="row.status"
            :custom-value="[1, 0]"
            @change="(v) => onToggleStatus(row, Number(v) as 0 | 1)"
          />
        </template>
        <template #last_run_time="{ row }">{{ formatDateTime(row.last_run_time) || '-' }}</template>
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
  </div>
</template>
<script setup lang="ts">
import type { PrimaryTableCol, TableRowData } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { onMounted, ref } from 'vue';

import type { CustomTaskItem } from '@/api/model/taskModel';
import { getCustomTasksApi, updateCustomTaskApi } from '@/api/task';

defineOptions({ name: 'CustomTaskManage' });

const columns: PrimaryTableCol<TableRowData>[] = [
  { title: '序号', colKey: 'index', width: 80 },
  { title: '任务名称', colKey: 'name', minWidth: 180 },
  { title: '类型', colKey: 'custom_type', width: 160 },
  { title: '调度周期', colKey: 'cron_expr', width: 180 },
  { title: '状态', colKey: 'status', width: 100 },
  { title: '上次执行', colKey: 'last_run_time', width: 180 },
  { title: '上次结果', colKey: 'last_result', minWidth: 260, ellipsis: true },
];

const tasks = ref<CustomTaskItem[]>([]);
const loading = ref(false);
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);

function formatDateTime(value?: string) {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString().replaceAll('/', '-');
}

async function fetchTasks() {
  loading.value = true;
  try {
    const res = await getCustomTasksApi({ page: page.value, page_size: pageSize.value });
    tasks.value = Array.isArray(res.data) ? res.data : [];
    total.value = typeof res.total === 'number' ? res.total : 0;
  } catch (error) {
    console.error(error);
    MessagePlugin.error('获取任务失败');
  } finally {
    loading.value = false;
  }
}

async function onToggleStatus(row: CustomTaskItem, status: 0 | 1) {
  try {
    await updateCustomTaskApi(row.id, { status });
    MessagePlugin.success(status === 1 ? '已启用' : '已禁用');
    await fetchTasks();
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(error instanceof Error ? error.message : '操作失败');
    await fetchTasks();
  }
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

onMounted(fetchTasks);
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
