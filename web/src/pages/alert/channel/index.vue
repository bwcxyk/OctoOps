<template>
  <div>
    <t-card class="list-card-container" :bordered="false">
      <t-row justify="space-between">
        <div class="left-operation-container">
          <t-select v-model="filterType" placeholder="全部类型" clearable style="width: 160px">
            <t-option label="邮件" value="email" />
            <t-option label="钉钉" value="dingtalk" />
            <t-option label="企业微信" value="wechat" />
            <t-option label="飞书" value="feishu" />
          </t-select>
          <t-button @click="openDialog()">新增</t-button>
        </div>
      </t-row>

      <t-table :data="pagedData" :columns="columns" row-key="id" :loading="loading" :hover="true">
        <template #index="{ rowIndex }">{{ rowIndex + 1 + (page - 1) * pageSize }}</template>
        <template #type="{ row }">{{ platformLabel(row.type) }}</template>
        <template #status="{ row }">
          <t-switch
            :value="row.status"
            :custom-value="[1, 0]"
            @change="(value) => onToggleStatus(row, Number(value) as 0 | 1)"
          />
        </template>
        <template #op="{ row }">
          <t-space>
            <t-link theme="primary" @click="openDialog(row)">编辑</t-link>
            <t-popconfirm content="确定要删除该渠道吗？" @confirm="removeChannel(row.id)">
              <t-link theme="danger">删除</t-link>
            </t-popconfirm>
            <t-link theme="default" @click="openTestDialog(row)">测试发送</t-link>
          </t-space>
        </template>
      </t-table>

      <div class="list-pagination">
        <t-pagination
          v-model="page"
          v-model:page-size="pageSize"
          :total="filteredChannels.length"
          show-jumper
          show-page-size
          :page-size-options="[10, 20, 50, 100]"
        />
      </div>
    </t-card>

    <t-dialog
      v-model:visible="dialogVisible"
      :header="editChannel.id ? '编辑渠道' : '新增渠道'"
      width="560px"
      :confirm-btn="{ content: '保存', loading: submitLoading, theme: 'primary' }"
      @confirm="onSubmit"
    >
      <t-form ref="formRef" :data="editChannel" :rules="rules" label-width="100px" @submit="onFormSubmit">
        <t-form-item label="名称" name="name">
          <t-input v-model="editChannel.name" />
        </t-form-item>
        <t-form-item label="类型" name="type">
          <t-select v-model="editChannel.type" placeholder="请选择类型" @change="onTypeChange">
            <t-option label="邮件" value="email" />
            <t-option label="钉钉" value="dingtalk" />
            <t-option label="企业微信" value="wechat" />
            <t-option label="飞书" value="feishu" />
          </t-select>
        </t-form-item>
        <t-form-item :label="editChannel.type === 'email' ? '邮箱' : 'Webhook'" name="target">
          <t-textarea
            v-if="editChannel.type !== 'email'"
            v-model="editChannel.target"
            :autosize="{ minRows: 2, maxRows: 4 }"
            placeholder="请输入Webhook地址"
          />
          <t-input v-else v-model="editChannel.target" placeholder="请输入邮箱地址" />
        </t-form-item>
        <t-form-item v-if="editChannel.type === 'dingtalk'" label="加签密钥" name="dingtalk_secret">
          <t-input v-model="editChannel.dingtalk_secret" placeholder="请输入钉钉加签密钥" />
        </t-form-item>
        <t-form-item label="启用" name="status">
          <t-switch v-model="editChannel.status" :custom-value="[1, 0]" />
        </t-form-item>
        <t-form-item label="告警模板" name="template_id">
          <t-select v-model="editChannel.template_id" placeholder="请选择告警模板" :loading="templatesLoading">
            <t-option v-for="tpl in filteredTemplates" :key="tpl.id" :label="tpl.name" :value="tpl.id" />
          </t-select>
        </t-form-item>
      </t-form>
    </t-dialog>

    <t-dialog
      v-model:visible="testDialogVisible"
      header="测试发送"
      width="680px"
      :confirm-btn="{ content: '发送测试', loading: testing, theme: 'primary' }"
      @confirm="submitTest"
    >
      <t-form label-width="100px">
        <t-form-item label="渠道名称">
          <t-input :model-value="testForm.channelName" readonly />
        </t-form-item>
        <t-form-item label="模板内容">
          <t-textarea
            v-model="testForm.templateContent"
            :autosize="{ minRows: 10, maxRows: 14 }"
            placeholder="可选：支持 {{ .time }} / {{ .message }} / {{ .channel }}"
          />
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>
<script setup lang="ts">
import type { FormRule, PrimaryTableCol, SubmitContext, TableRowData } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, reactive, ref } from 'vue';

import {
  createAlertChannelApi,
  deleteAlertChannelApi,
  getAlertChannelsApi,
  getAlertTemplatesApi,
  testAlertChannelApi,
  updateAlertChannelApi,
} from '@/api/alert';
import type { AlertChannel, AlertTemplate } from '@/api/model/alertModel';

defineOptions({ name: 'AlertChannelManage' });

const columns: PrimaryTableCol<TableRowData>[] = [
  { title: '序号', colKey: 'index', width: 80 },
  { title: '名称', colKey: 'name', width: 180 },
  { title: '类型', colKey: 'type', width: 120 },
  { title: '目标', colKey: 'target', minWidth: 200, ellipsis: true },
  { title: '启用', colKey: 'status', width: 100 },
  { title: '操作', colKey: 'op', width: 240 },
];

const rules: Record<string, FormRule[]> = {
  name: [{ required: true, message: '名称必填', type: 'error' }],
  type: [{ required: true, message: '类型必选', type: 'error' }],
  target: [
    { required: true, message: '目标必填', type: 'error' },
    {
      validator: (_val) => {
        if (editChannel.type === 'email') {
          const emailReg = /^[\w.%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i;
          return emailReg.test(editChannel.target || '') ? true : { result: false, message: '请输入正确的邮箱地址' };
        }
        return true;
      },
      type: 'warning',
    },
  ],
};

const loading = ref(false);
const submitLoading = ref(false);
const channels = ref<AlertChannel[]>([]);
const templates = ref<AlertTemplate[]>([]);
const templatesLoading = ref(false);
const filterType = ref('');
const page = ref(1);
const pageSize = ref(10);

const dialogVisible = ref(false);
const testDialogVisible = ref(false);
const testing = ref(false);
const formRef = ref();

const editChannel = reactive<Partial<AlertChannel>>({
  id: undefined,
  name: '',
  type: '',
  target: '',
  status: 1,
  dingtalk_secret: '',
  template_id: null,
});

const testForm = reactive({
  channelId: 0,
  channelName: '',
  templateContent: '',
});

const filteredChannels = computed(() => {
  if (!filterType.value) return channels.value;
  return channels.value.filter((item) => item.type === filterType.value);
});

const pagedData = computed(() => {
  const start = (page.value - 1) * pageSize.value;
  return filteredChannels.value.slice(start, start + pageSize.value);
});

const filteredTemplates = computed(() => {
  if (!editChannel.type) return templates.value;
  const templateType = getTemplateTypeByChannelType(editChannel.type);
  return templates.value.filter((tpl) => tpl.type === templateType);
});

function getTemplateTypeByChannelType(channelType: string) {
  const map: Record<string, string> = {
    dingtalk: 'dingtalk',
    wechat: 'weixin',
    feishu: 'feishu',
    email: 'email',
  };
  return map[channelType] || channelType;
}

function platformLabel(val: string) {
  if (val === 'dingtalk') return '钉钉';
  if (val === 'feishu') return '飞书';
  if (val === 'wechat') return '企业微信';
  if (val === 'email') return '邮件';
  return val;
}

async function fetchChannels() {
  loading.value = true;
  try {
    channels.value = await getAlertChannelsApi();
  } catch (error) {
    console.error(error);
    MessagePlugin.error('获取渠道列表失败');
  } finally {
    loading.value = false;
  }
}

async function fetchTemplates() {
  templatesLoading.value = true;
  try {
    templates.value = await getAlertTemplatesApi();
  } catch (error) {
    console.error(error);
    MessagePlugin.error('获取模板列表失败');
  } finally {
    templatesLoading.value = false;
  }
}

function openDialog(row?: AlertChannel) {
  if (row) {
    Object.assign(editChannel, { ...row });
  } else {
    Object.assign(editChannel, {
      id: undefined,
      name: '',
      type: '',
      target: '',
      status: 1,
      dingtalk_secret: '',
      template_id: null,
    });
  }
  dialogVisible.value = true;
}

function onTypeChange() {
  if (editChannel.type !== 'dingtalk') {
    editChannel.dingtalk_secret = '';
  }
  if (editChannel.template_id) {
    const selected = templates.value.find((tpl) => tpl.id === editChannel.template_id);
    if (selected && selected.type !== getTemplateTypeByChannelType(editChannel.type || '')) {
      editChannel.template_id = null;
    }
  }
}

async function onSubmit() {
  await formRef.value?.submit();
}

async function onFormSubmit(ctx: SubmitContext) {
  if (ctx.validateResult !== true) return;
  submitLoading.value = true;
  try {
    if (editChannel.id) {
      await updateAlertChannelApi(editChannel.id, editChannel);
      MessagePlugin.success('更新成功');
    } else {
      await createAlertChannelApi(editChannel);
      MessagePlugin.success('创建成功');
    }
    dialogVisible.value = false;
    await fetchChannels();
  } catch (error) {
    console.error(error);
    MessagePlugin.error('保存失败');
  } finally {
    submitLoading.value = false;
  }
}

async function removeChannel(id: number) {
  try {
    await deleteAlertChannelApi(id);
    MessagePlugin.success('删除成功');
    await fetchChannels();
  } catch (error) {
    console.error(error);
    MessagePlugin.error('删除失败');
  }
}

async function onToggleStatus(row: AlertChannel, status: 0 | 1) {
  try {
    await updateAlertChannelApi(row.id, { ...row, status });
    MessagePlugin.success(status ? '已启用' : '已禁用');
    await fetchChannels();
  } catch (error) {
    console.error(error);
    MessagePlugin.error('更新状态失败');
  }
}

function openTestDialog(row: AlertChannel) {
  testForm.channelId = row.id;
  testForm.channelName = row.name;
  testForm.templateContent = '';
  testDialogVisible.value = true;
}

async function submitTest() {
  if (!testForm.channelId) return;
  testing.value = true;
  try {
    const res = await testAlertChannelApi(testForm.channelId, testForm.templateContent);
    MessagePlugin.success(res.message || '测试发送成功');
    testDialogVisible.value = false;
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(error instanceof Error ? error.message : '测试发送失败');
  } finally {
    testing.value = false;
  }
}

onMounted(async () => {
  await Promise.all([fetchChannels(), fetchTemplates()]);
});
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

.list-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--td-comp-margin-xxl);
}
</style>
