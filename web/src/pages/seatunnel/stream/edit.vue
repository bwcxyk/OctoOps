<template>
  <div class="edit-page">
    <t-card class="edit-card-container" :bordered="false">
      <div class="edit-page-header">
        <div>
          <div class="edit-page-title">{{ isEditMode ? '编辑流任务' : '新增流任务' }}</div>
        </div>
        <t-space>
          <t-button variant="outline" @click="goBack">返回</t-button>
          <t-button theme="primary" :loading="submitLoading" @click="onSubmit">保存</t-button>
        </t-space>
      </div>

      <t-loading :loading="pageLoading" class="edit-page-loading">
        <div class="edit-page-scroll">
          <t-form ref="formRef" :data="editForm" :rules="rules" label-width="110px" @submit="onFormSubmit">
            <div class="stream-edit-grid">
              <t-form-item label="作业名称" name="name">
                <t-input v-model="editForm.name" />
              </t-form-item>
              <t-form-item label="描述" name="description">
                <t-input v-model="editForm.description" />
              </t-form-item>
              <t-form-item label="失败告警" name="enable_alert">
                <t-checkbox v-model="editForm.enable_alert" @change="onEnableAlertChange">作业失败时发送告警</t-checkbox>
              </t-form-item>
              <t-form-item v-if="editForm.enable_alert" label="告警组" name="alert_group_ids">
                <t-select v-model="editForm.alert_group_ids" multiple clearable placeholder="请选择告警组">
                  <t-option v-for="group in alertGroups" :key="group.id" :label="group.name" :value="group.id" />
                </t-select>
              </t-form-item>
              <div v-else class="stream-edit-grid__placeholder" />
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

import { getAlertGroupsApi } from '@/api/alert';
import type { AlertGroup } from '@/api/model/alertModel';
import type { SeatunnelTask } from '@/api/model/seatunnelModel';
import { createTaskApi, getTaskApi, updateTaskApi } from '@/api/seatunnel';
import CodeEditor from '@/components/code-editor/index.vue';

defineOptions({ name: 'SeatunnelStreamTaskEditor' });

interface StreamEditForm {
  id?: number;
  name: string;
  description: string;
  config_format: 'json' | 'hocon';
  config: string;
  enable_alert: boolean;
  alert_group_ids: number[];
}

const route = useRoute();
const router = useRouter();

const rules: Record<string, FormRule[]> = {
  name: [{ required: true, message: '请输入作业名称', type: 'error' }],
  config: [{ required: true, message: '请输入作业配置', type: 'error' }],
  alert_group_ids: [
    {
      required: true,
      validator: () => {
        if (!editForm.enable_alert) return true;
        return editForm.alert_group_ids.length > 0 ? true : { result: false, message: '请至少选择一个告警组' };
      },
      type: 'error',
    },
  ],
};

const formRef = ref();
const pageLoading = ref(false);
const submitLoading = ref(false);
const alertGroups = ref<AlertGroup[]>([]);

const editForm = reactive<StreamEditForm>({
  id: undefined,
  name: '',
  description: '',
  config_format: 'json',
  config: '',
  enable_alert: false,
  alert_group_ids: [],
});

const taskId = computed(() => {
  const rawId = route.params.id;
  if (!rawId) return undefined;
  const id = Number(Array.isArray(rawId) ? rawId[0] : rawId);
  return Number.isFinite(id) && id > 0 ? id : undefined;
});
const isEditMode = computed(() => typeof taskId.value === 'number');

function parseAlertGroupIds(alertGroup?: string) {
  if (!alertGroup) return [];
  return alertGroup
    .split(',')
    .map((item) => Number(item.trim()))
    .filter((item) => Number.isFinite(item) && item > 0);
}

function fillEditForm(task: SeatunnelTask) {
  editForm.id = task.id;
  editForm.name = task.name || '';
  editForm.description = task.description || '';
  editForm.config_format = (task.config_format as 'json' | 'hocon') || 'json';
  editForm.config = task.config || '';
  editForm.alert_group_ids = parseAlertGroupIds(task.alert_group);
  editForm.enable_alert = editForm.alert_group_ids.length > 0;
}

async function fetchAlertGroups() {
  const groups = await getAlertGroupsApi();
  alertGroups.value = groups.filter((group) => group.status === 1);
}

async function fetchTaskDetail() {
  if (!taskId.value) return;
  const task = await getTaskApi(taskId.value, 'stream');
  fillEditForm(task);
}

function onEnableAlertChange(value: boolean) {
  if (!value) {
    editForm.alert_group_ids = [];
    formRef.value?.clearValidate?.();
  }
}

async function onSubmit() {
  await formRef.value?.submit();
}

async function onFormSubmit(ctx: SubmitContext) {
  if (ctx.validateResult !== true) return;
  if (editForm.enable_alert && editForm.alert_group_ids.length === 0) {
    MessagePlugin.warning('请至少选择一个告警组');
    return;
  }

  submitLoading.value = true;
  try {
    const payload: Partial<SeatunnelTask> = {
      name: editForm.name,
      description: editForm.description,
      config: editForm.config,
      config_format: editForm.config_format,
      task_type: 'stream',
      alert_group: editForm.enable_alert ? editForm.alert_group_ids.join(',') : '',
    };

    if (editForm.id) {
      await updateTaskApi(editForm.id, payload, 'stream');
      MessagePlugin.success('更新成功');
    } else {
      await createTaskApi(payload);
      MessagePlugin.success('创建成功');
    }
    router.push('/seatunnel/stream');
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(error instanceof Error ? error.message : '保存失败');
  } finally {
    submitLoading.value = false;
  }
}

function goBack() {
  router.push('/seatunnel/stream');
}

onMounted(async () => {
  pageLoading.value = true;
  try {
    await Promise.all([fetchAlertGroups(), fetchTaskDetail()]);
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

.stream-edit-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 24px;
  row-gap: 2px;
}

.stream-edit-grid__placeholder {
  min-height: 1px;
}

.editor-form-item {
  align-items: flex-start;
}

:deep(.stream-edit-grid .t-form__item) {
  margin-bottom: 14px;
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

  .stream-edit-grid {
    grid-template-columns: 1fr;
    column-gap: 0;
  }
}
</style>
