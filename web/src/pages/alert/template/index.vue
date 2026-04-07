<template>
  <div>
    <t-card class="list-card-container" :bordered="false">
      <t-row justify="space-between">
        <div class="left-operation-container">
          <t-button @click="openEditDialog()">新增</t-button>
          <t-popup trigger="hover" placement="bottom-left">
            <template #content>
              <div class="tips-content">
                <strong>模板格式说明（Go Template）</strong><br />
                <code v-pre>{{.JobID}}</code> 作业ID<br />
                <code v-pre>{{.JobName}}</code> 作业名称<br />
                <code v-pre>{{.Status}}</code> 状态<br />
                <code v-pre>{{.TaskType}}</code> 任务类型<br />
                <code v-pre>{{.StartTime}}</code> 开始时间<br />
                <code v-pre>{{.EndTime}}</code> 结束时间<br />
                <code v-pre>{{.Reason}}</code> 原因
              </div>
            </template>
            <t-button variant="text" theme="default" shape="square" class="tips-icon-btn">
              <t-icon name="info-circle" />
            </t-button>
          </t-popup>
        </div>
      </t-row>

      <t-table :data="pagedData" :columns="columns" row-key="id" :loading="loading" :hover="true">
        <template #index="{ rowIndex }">{{ rowIndex + 1 + (page - 1) * pageSize }}</template>
        <template #type="{ row }">{{ typeLabel(row.type) }}</template>
        <template #name="{ row }">
          <t-link theme="primary" hover="color" @click="showDetail(row)">{{ row.name }}</t-link>
        </template>
        <template #op="{ row }">
          <t-space>
            <t-link theme="primary" @click="openEditDialog(row)">编辑</t-link>
            <t-popconfirm content="确定要删除该模板吗？" @confirm="removeTemplate(row.id)">
              <t-link theme="danger">删除</t-link>
            </t-popconfirm>
          </t-space>
        </template>
      </t-table>

      <div class="list-pagination">
        <t-pagination
          v-model="page"
          v-model:page-size="pageSize"
          :total="templates.length"
          show-jumper
          show-page-size
          :page-size-options="[10, 20, 50, 100]"
        />
      </div>
    </t-card>

    <t-dialog
      v-model:visible="editDialogVisible"
      :header="editForm.id ? '编辑模板' : '新增模板'"
      width="860px"
      :confirm-btn="{ content: '保存', loading: submitLoading, theme: 'primary' }"
      @confirm="onSubmit"
    >
      <t-form ref="formRef" :data="editForm" :rules="rules" label-width="100px" @submit="onFormSubmit">
        <t-form-item label="模板名称" name="name">
          <t-input v-model="editForm.name" />
        </t-form-item>
        <t-form-item label="模板类型" name="type">
          <t-select v-model="editForm.type" placeholder="请选择模板类型">
            <t-option label="钉钉" value="dingtalk" />
            <t-option label="企业微信" value="weixin" />
            <t-option label="飞书" value="feishu" />
            <t-option label="邮件" value="email" />
          </t-select>
        </t-form-item>
        <t-form-item label="内容" name="content">
          <t-textarea v-model="editForm.content" :autosize="{ minRows: 12, maxRows: 16 }" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <t-dialog v-model:visible="detailDialogVisible" header="模板详情" width="640px" :footer="false">
      <t-descriptions :column="1" bordered>
        <t-descriptions-item label="模板名称">{{ detailForm.name }}</t-descriptions-item>
        <t-descriptions-item label="类型">{{ typeLabel(detailForm.type || '') }}</t-descriptions-item>
        <t-descriptions-item label="内容">
          <t-textarea :model-value="detailForm.content || ''" readonly :autosize="{ minRows: 8, maxRows: 12 }" />
        </t-descriptions-item>
      </t-descriptions>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import type { FormRule, PrimaryTableCol, SubmitContext, TableRowData } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, reactive, ref } from 'vue';

import {
  createAlertTemplateApi,
  deleteAlertTemplateApi,
  getAlertTemplatesApi,
  updateAlertTemplateApi,
} from '@/api/alert';
import type { AlertTemplate } from '@/api/model/alertModel';

defineOptions({ name: 'AlertTemplateManage' });

const columns: PrimaryTableCol<TableRowData>[] = [
  { title: '序号', colKey: 'index', width: 80 },
  { title: '模板名称', colKey: 'name', minWidth: 220 },
  { title: '类型', colKey: 'type', width: 140 },
  { title: '操作', colKey: 'op', width: 160 },
];

const rules: Record<string, FormRule[]> = {
  name: [{ required: true, message: '模板名称必填', type: 'error' }],
  type: [{ required: true, message: '模板类型必选', type: 'error' }],
  content: [{ required: true, message: '内容必填', type: 'error' }],
};

const loading = ref(false);
const submitLoading = ref(false);
const templates = ref<AlertTemplate[]>([]);
const page = ref(1);
const pageSize = ref(10);

const editDialogVisible = ref(false);
const detailDialogVisible = ref(false);
const formRef = ref();

const editForm = reactive<Partial<AlertTemplate>>({
  id: undefined,
  name: '',
  type: '',
  content: '',
});

const detailForm = ref<Partial<AlertTemplate>>({});

const pagedData = computed(() => {
  const start = (page.value - 1) * pageSize.value;
  return templates.value.slice(start, start + pageSize.value);
});

const typeLabel = (type: string) => {
  if (type === 'dingtalk') return '钉钉';
  if (type === 'weixin') return '企业微信';
  if (type === 'feishu') return '飞书';
  if (type === 'email') return '邮件';
  return type;
};

const fetchTemplates = async () => {
  loading.value = true;
  try {
    templates.value = await getAlertTemplatesApi();
  } catch (error) {
    console.error(error);
    MessagePlugin.error('获取模板列表失败');
  } finally {
    loading.value = false;
  }
};

const openEditDialog = (row?: AlertTemplate) => {
  if (row) {
    Object.assign(editForm, { ...row });
  } else {
    Object.assign(editForm, { id: undefined, name: '', type: '', content: '' });
  }
  editDialogVisible.value = true;
};

const showDetail = (row: AlertTemplate) => {
  detailForm.value = { ...row };
  detailDialogVisible.value = true;
};

const removeTemplate = async (id: number) => {
  try {
    await deleteAlertTemplateApi(id);
    MessagePlugin.success('删除成功');
    await fetchTemplates();
  } catch (error) {
    console.error(error);
    MessagePlugin.error('删除失败');
  }
};

const onSubmit = async () => {
  await formRef.value?.submit();
};

const onFormSubmit = async (ctx: SubmitContext) => {
  if (ctx.validateResult !== true) return;
  submitLoading.value = true;
  try {
    if (editForm.id) {
      await updateAlertTemplateApi(editForm.id, editForm);
      MessagePlugin.success('更新成功');
    } else {
      await createAlertTemplateApi(editForm);
      MessagePlugin.success('创建成功');
    }
    editDialogVisible.value = false;
    await fetchTemplates();
  } catch (error) {
    console.error(error);
    MessagePlugin.error('保存失败');
  } finally {
    submitLoading.value = false;
  }
};

onMounted(fetchTemplates);
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
  margin-bottom: var(--td-comp-margin-xxl);
}

.tips-icon-btn {
  padding: 0;
}

.tips-content {
  max-width: 420px;
  line-height: 1.8;
  padding: 8px 10px;
}

.list-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--td-comp-margin-xxl);
}
</style>
