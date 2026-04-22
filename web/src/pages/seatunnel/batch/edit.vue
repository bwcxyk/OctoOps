<template>
  <div class="edit-page">
    <t-card class="edit-card-container" :bordered="false">
      <div class="edit-page-header">
        <div>
          <div class="edit-page-title">{{ isEditMode ? '编辑批任务' : '新增批任务' }}</div>
        </div>
        <t-space>
          <t-button variant="outline" @click="goBack">返回</t-button>
          <t-button theme="primary" :loading="submitLoading" @click="onSubmit">保存</t-button>
        </t-space>
      </div>

      <t-loading :loading="pageLoading" class="edit-page-loading">
        <div class="edit-page-scroll">
          <t-form ref="formRef" :data="editForm" :rules="rules" label-width="110px" @submit="onFormSubmit">
            <div class="batch-edit-grid">
              <t-form-item label="作业名称" name="name">
                <t-input v-model="editForm.name" />
              </t-form-item>
              <t-form-item label="状态" name="status">
                <t-switch v-model="editForm.status" :custom-value="[1, 0]" />
              </t-form-item>
              <t-form-item label="描述" name="description">
                <t-input v-model="editForm.description" />
              </t-form-item>
              <t-form-item label="Cron表达式" name="cron_expr">
                <t-input v-model="editForm.cron_expr" placeholder="秒 分 时 日 月 周" />
              </t-form-item>
            </div>

            <t-form-item label="配置风格" name="config_format">
              <t-select v-model="editForm.config_format" style="width: 180px">
                <t-option label="JSON" value="json" />
                <t-option label="HOCON" value="hocon" />
              </t-select>
            </t-form-item>

            <t-form-item label="作业配置" name="config" class="editor-form-item">
              <CodeEditor v-model="editForm.config" active :language="editForm.config_format" :height="360" />
            </t-form-item>
          </t-form>
        </div>
      </t-loading>
    </t-card>
  </div>
</template>

<script setup lang="ts">
import type { FormRule, SubmitContext } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, reactive, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import type { SeatunnelTask } from '@/api/model/seatunnelModel';
import { createTaskApi, getTaskApi, updateTaskApi } from '@/api/seatunnel';
import CodeEditor from '@/components/code-editor/index.vue';
import { isValidCronExpression } from '@/utils/cron';

defineOptions({ name: 'SeatunnelBatchTaskEditor' });

interface BatchEditForm {
  id?: number;
  name: string;
  description: string;
  status: 0 | 1;
  cron_expr: string;
  config_format: 'json' | 'hocon';
  config: string;
}

const route = useRoute();
const router = useRouter();

const rules: Record<string, FormRule[]> = {
  name: [{ required: true, message: '请输入作业名称', type: 'error' }],
  config: [{ required: true, message: '请输入作业配置', type: 'error' }],
  cron_expr: [
    {
      validator: () => {
        const expr = (editForm.cron_expr || '').trim();
        if (editForm.status !== 1 && !expr) return true;
        if (!expr) return { result: false, message: '批处理任务启用时必须填写 Cron 表达式' };
        return isValidCronExpression(expr) ? true : { result: false, message: '无效的 Cron 表达式' };
      },
      type: 'warning',
    },
  ],
};

const formRef = ref();
const pageLoading = ref(false);
const submitLoading = ref(false);

const editForm = reactive<BatchEditForm>({
  id: undefined,
  name: '',
  description: '',
  status: 0,
  cron_expr: '',
  config_format: 'json',
  config: '',
});

const taskId = computed(() => {
  const rawId = route.params.id;
  if (!rawId) return undefined;
  const id = Number(Array.isArray(rawId) ? rawId[0] : rawId);
  return Number.isFinite(id) && id > 0 ? id : undefined;
});
const isEditMode = computed(() => typeof taskId.value === 'number');

function fillEditForm(task: SeatunnelTask) {
  editForm.id = task.id;
  editForm.name = task.name || '';
  editForm.description = task.description || '';
  editForm.status = Number(task.status || 0) === 1 ? 1 : 0;
  editForm.cron_expr = task.cron_expr || '';
  editForm.config_format = (task.config_format as 'json' | 'hocon') || 'json';
  editForm.config = task.config || '';
}

async function fetchTaskDetail() {
  if (!taskId.value) return;
  const task = await getTaskApi(taskId.value, 'batch');
  fillEditForm(task);
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
    router.push('/seatunnel/batch');
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(error instanceof Error ? error.message : '保存失败');
  } finally {
    submitLoading.value = false;
  }
}

function goBack() {
  router.push('/seatunnel/batch');
}

onMounted(async () => {
  pageLoading.value = true;
  try {
    await fetchTaskDetail();
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(error instanceof Error ? error.message : '加载任务信息失败');
    goBack();
  } finally {
    pageLoading.value = false;
  }
});
</script>

<style lang="less" scoped>
.edit-page {
  height: calc(100vh - 220px);
  min-height: 640px;
}

.edit-card-container {
  height: 100%;
  padding: var(--td-comp-paddingTB-xxl) var(--td-comp-paddingLR-xxl);
  overflow: hidden;

  :deep(.t-card__body) {
    padding: 0;
    height: 100%;
    display: flex;
    flex-direction: column;
  }
}

.edit-page-header {
  flex-shrink: 0;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 24px;
}

.edit-page-title {
  font-size: 20px;
  font-weight: 600;
  line-height: 28px;
  color: var(--td-text-color-primary);
}

.edit-page-subtitle {
  margin-top: 6px;
  color: var(--td-text-color-secondary);
}

.edit-page-loading {
  flex: 1;
  min-height: 0;

  :deep(.t-loading__parent) {
    height: 100%;
  }
}

.edit-page-scroll {
  height: 100%;
  min-height: 0;
  overflow-y: auto;
  padding-right: 4px;
}

.batch-edit-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 24px;
}

.editor-form-item {
  align-items: flex-start;
}

@media (max-width: 900px) {
  .edit-page {
    height: auto;
    min-height: auto;
  }

  .edit-page-header {
    flex-direction: column;
    align-items: stretch;
  }

  .batch-edit-grid {
    grid-template-columns: 1fr;
    column-gap: 0;
  }
}
</style>
