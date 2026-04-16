<template>
  <div>
    <t-card class="list-card-container" :bordered="false">
      <t-row justify="space-between">
        <div class="left-operation-container">
          <t-button @click="openEditDialog()">新增</t-button>
        </div>
      </t-row>

      <t-table :data="pagedData" :columns="columns" row-key="id" :loading="loading" :hover="true">
        <template #index="{ rowIndex }">{{ rowIndex + 1 + (page - 1) * pageSize }}</template>
        <template #status="{ row }">
          <t-switch
            :value="row.status"
            :custom-value="[1, 0]"
            @change="(value) => onToggleStatus(row, Number(value) as 0 | 1)"
          />
        </template>
        <template #op="{ row }">
          <t-space>
            <t-link theme="primary" @click="openEditDialog(row)">编辑</t-link>
            <t-link theme="default" @click="openMemberDialog(row)">成员管理</t-link>
            <t-popconfirm content="确定要删除该告警组吗？" @confirm="removeGroup(row.id)">
              <t-link theme="danger">删除</t-link>
            </t-popconfirm>
          </t-space>
        </template>
      </t-table>

      <div class="list-pagination">
        <t-pagination
          v-model="page"
          v-model:page-size="pageSize"
          :total="groups.length"
          show-jumper
          show-page-size
          :page-size-options="[10, 20, 50, 100]"
        />
      </div>
    </t-card>

    <t-dialog
      v-model:visible="editDialogVisible"
      :header="editForm.id ? '编辑告警组' : '新增告警组'"
      width="560px"
      :confirm-btn="{ content: '保存', loading: submitLoading, theme: 'primary' }"
      @confirm="onSubmit"
    >
      <t-form ref="formRef" :data="editForm" :rules="rules" label-width="100px" @submit="onFormSubmit">
        <t-form-item label="名称" name="name">
          <t-input v-model="editForm.name" />
        </t-form-item>
        <t-form-item label="描述" name="description">
          <t-input v-model="editForm.description" />
        </t-form-item>
        <t-form-item label="启用" name="status">
          <t-switch v-model="editForm.status" :custom-value="[1, 0]" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <t-dialog v-model:visible="memberDialogVisible" header="成员管理" width="760px" :footer="false">
      <div class="member-toolbar">
        <t-select v-model="addMemberType" placeholder="选择类型" style="width: 140px">
          <t-option label="邮件" value="email" />
          <t-option label="钉钉" value="dingtalk" />
          <t-option label="企业微信" value="wechat" />
          <t-option label="飞书" value="feishu" />
        </t-select>
        <t-select v-model="addMemberId" filterable placeholder="选择渠道" style="width: 280px">
          <t-option
            v-for="item in availableChannels"
            :key="item.id"
            :label="item.name || item.target"
            :value="item.id"
          />
        </t-select>
        <t-button :disabled="!addMemberType || !addMemberId" @click="addMember">添加成员</t-button>
      </div>

      <t-table :data="members" :columns="memberColumns" row-key="id" :hover="true">
        <template #index="{ rowIndex }">{{ rowIndex + 1 }}</template>
        <template #channel_type="{ row }">{{ channelTypeLabel(row.channel_type) }}</template>
        <template #channel_id="{ row }">{{ getChannelName(row) }}</template>
        <template #op="{ row }">
          <t-popconfirm content="确定要删除该成员吗？" @confirm="removeMember(row.id)">
            <t-link theme="danger">删除</t-link>
          </t-popconfirm>
        </template>
      </t-table>
    </t-dialog>
  </div>
</template>
<script setup lang="ts">
import type { FormRule, PrimaryTableCol, SubmitContext, TableRowData } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, reactive, ref, watch } from 'vue';

import {
  addAlertGroupMemberApi,
  createAlertGroupApi,
  deleteAlertGroupApi,
  deleteAlertGroupMemberApi,
  getAlertChannelsApi,
  getAlertGroupMembersApi,
  getAlertGroupsApi,
  updateAlertGroupApi,
} from '@/api/alert';
import type { AlertChannel, AlertGroup, AlertGroupMember } from '@/api/model/alertModel';

defineOptions({ name: 'AlertGroupManage' });

const columns: PrimaryTableCol<TableRowData>[] = [
  { title: '序号', colKey: 'index', width: 80 },
  { title: '告警组名称', colKey: 'name', width: 180 },
  { title: '描述', colKey: 'description', minWidth: 240 },
  { title: '启用', colKey: 'status', width: 100 },
  { title: '操作', colKey: 'op', width: 260 },
];

const memberColumns: PrimaryTableCol<TableRowData>[] = [
  { title: '序号', colKey: 'index', width: 80 },
  { title: '类型', colKey: 'channel_type', width: 140 },
  { title: '渠道', colKey: 'channel_id', minWidth: 240 },
  { title: '操作', colKey: 'op', width: 120 },
];

const rules: Record<string, FormRule[]> = {
  name: [{ required: true, message: '请输入告警组名称', type: 'error' }],
};

const loading = ref(false);
const submitLoading = ref(false);
const groups = ref<AlertGroup[]>([]);
const channels = ref<AlertChannel[]>([]);
const members = ref<AlertGroupMember[]>([]);
const page = ref(1);
const pageSize = ref(10);

const editDialogVisible = ref(false);
const memberDialogVisible = ref(false);
const formRef = ref();

const currentGroup = ref<AlertGroup | null>(null);
const addMemberType = ref('');
const addMemberId = ref<number | null>(null);

const editForm = reactive<Partial<AlertGroup>>({
  id: undefined,
  name: '',
  description: '',
  status: 1,
});

const pagedData = computed(() => {
  const start = (page.value - 1) * pageSize.value;
  return groups.value.slice(start, start + pageSize.value);
});

const availableChannels = computed(() => {
  if (!addMemberType.value) return [];
  return channels.value.filter((item) => item.type === addMemberType.value && item.status === 1);
});

const channelTypeLabel = (type: string) => {
  if (type === 'email') return '邮件';
  if (type === 'dingtalk') return '钉钉';
  if (type === 'wechat') return '企业微信';
  if (type === 'feishu') return '飞书';
  return type;
};

const fetchGroups = async () => {
  loading.value = true;
  try {
    groups.value = await getAlertGroupsApi();
  } catch (error) {
    console.error(error);
    MessagePlugin.error('获取告警组失败');
  } finally {
    loading.value = false;
  }
};

const fetchChannels = async () => {
  try {
    channels.value = await getAlertChannelsApi();
  } catch (error) {
    console.error(error);
    MessagePlugin.error('获取渠道列表失败');
  }
};

const fetchMembers = async () => {
  if (!currentGroup.value) return;
  try {
    members.value = await getAlertGroupMembersApi(currentGroup.value.id);
  } catch (error) {
    console.error(error);
    MessagePlugin.error('获取成员失败');
  }
};

const openEditDialog = (row?: AlertGroup) => {
  if (row) {
    Object.assign(editForm, { ...row });
  } else {
    Object.assign(editForm, { id: undefined, name: '', description: '', status: 1 });
  }
  editDialogVisible.value = true;
};

const onSubmit = async () => {
  await formRef.value?.submit();
};

const onFormSubmit = async (ctx: SubmitContext) => {
  if (ctx.validateResult !== true) return;
  submitLoading.value = true;
  try {
    if (editForm.id) {
      await updateAlertGroupApi(editForm.id, editForm);
      MessagePlugin.success('更新成功');
    } else {
      await createAlertGroupApi(editForm);
      MessagePlugin.success('创建成功');
    }
    editDialogVisible.value = false;
    await fetchGroups();
  } catch (error) {
    console.error(error);
    MessagePlugin.error('保存失败');
  } finally {
    submitLoading.value = false;
  }
};

const removeGroup = async (id: number) => {
  try {
    await deleteAlertGroupApi(id);
    MessagePlugin.success('删除成功');
    await fetchGroups();
  } catch (error) {
    console.error(error);
    MessagePlugin.error('删除失败');
  }
};

const onToggleStatus = async (row: AlertGroup, status: 0 | 1) => {
  try {
    await updateAlertGroupApi(row.id, { ...row, status });
    MessagePlugin.success(status ? '已启用' : '已禁用');
    await fetchGroups();
  } catch (error) {
    console.error(error);
    MessagePlugin.error('更新状态失败');
  }
};

const openMemberDialog = async (group: AlertGroup) => {
  currentGroup.value = group;
  memberDialogVisible.value = true;
  addMemberType.value = '';
  addMemberId.value = null;
  await Promise.all([fetchMembers(), fetchChannels()]);
};

const addMember = async () => {
  if (!currentGroup.value || !addMemberType.value || !addMemberId.value) return;
  try {
    await addAlertGroupMemberApi(currentGroup.value.id, {
      channel_type: addMemberType.value as AlertGroupMember['channel_type'],
      channel_id: addMemberId.value,
    });
    MessagePlugin.success('添加成功');
    addMemberId.value = null;
    await fetchMembers();
  } catch (error) {
    console.error(error);
    MessagePlugin.error(error instanceof Error ? error.message : '添加成员失败');
  }
};

const removeMember = async (memberId: number) => {
  if (!currentGroup.value) return;
  try {
    await deleteAlertGroupMemberApi(currentGroup.value.id, memberId);
    MessagePlugin.success('删除成功');
    await fetchMembers();
  } catch (error) {
    console.error(error);
    MessagePlugin.error('删除成员失败');
  }
};

const getChannelName = (member: AlertGroupMember) => {
  const found = channels.value.find((item) => item.id === member.channel_id && item.type === member.channel_type);
  return found ? found.name || found.target : String(member.channel_id);
};

watch(addMemberType, () => {
  addMemberId.value = null;
});

onMounted(fetchGroups);
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

.member-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: var(--td-comp-margin-l);
}

.list-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--td-comp-margin-xxl);
}
</style>
