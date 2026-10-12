<template>
  <div>
    <t-card class="setting-card" :bordered="false" :loading="loading">
      <t-tabs v-model="activeTab" class="setting-tabs">
        <t-tab-panel value="task" label="系统任务">
          <div class="setting-panel">
            <p class="section-hint">
              Cron 表达式为 6 位（秒 分 时 日 月 周），例如 <code>0 */10 * * * *</code> 每 10 分钟执行一次，
              <code>0 0 8 * * *</code> 每天 8 点执行一次。开关关闭时任务不会调度。
            </p>
            <div v-for="item in taskItems" :key="item.key" class="setting-item">
              <div class="setting-head">
                <t-switch v-model="draft[item.key].enabled" />
                <span class="setting-name">{{ item.name }}</span>
              </div>
              <p class="setting-desc">{{ item.description }}</p>
              <div v-if="draft[item.key].enabled" class="setting-cron">
                <t-input
                  v-model="draft[item.key].cron"
                  class="cron-input"
                  placeholder="如 0 */10 * * * *"
                  clearable
                />
              </div>
            </div>
          </div>
        </t-tab-panel>

        <t-tab-panel value="mail" label="邮件设置">
          <div class="setting-panel">
            <p class="section-hint">
              用于告警通知与找回密码验证码邮件，关闭后不会发送任何邮件。出于安全考虑，已保存的密码不会回显，
              显示为 <code>******</code>，保持原样或留空表示不修改。
            </p>
            <div class="setting-item">
              <div class="setting-head">
                <t-switch v-model="mailDraft.enable" />
                <span class="setting-name">启用邮件通知</span>
              </div>
            </div>
            <div v-if="mailDraft.enable" class="mail-fields">
              <div class="mail-field">
                <span class="mail-label">SMTP 服务器</span>
                <t-input v-model="mailDraft.smtp_address" placeholder="如 smtp.feishu.cn" clearable />
              </div>
              <div class="mail-field">
                <span class="mail-label">端口</span>
                <t-input-number
                  v-model="mailDraft.smtp_port"
                  theme="normal"
                  :min="1"
                  :max="65535"
                  placeholder="如 465"
                  style="width: 120px"
                />
              </div>
              <div class="mail-field">
                <span class="mail-label">邮箱账号</span>
                <t-input v-model="mailDraft.smtp_user" placeholder="发件邮箱" clearable />
              </div>
              <div class="mail-field">
                <span class="mail-label">密码 / 授权码</span>
                <t-input v-model="mailDraft.smtp_password" type="password" placeholder="未修改请保留原值" clearable />
              </div>
              <div class="mail-field">
                <span class="mail-label">发件人名称</span>
                <t-input v-model="mailDraft.display_name" placeholder="如 OctoOps" clearable />
              </div>
              <div class="mail-field">
                <span class="mail-label">使用 SSL</span>
                <t-switch v-model="mailDraft.ssl" />
              </div>
            </div>
          </div>
        </t-tab-panel>

        <t-tab-panel value="log" label="日志设置">
          <div class="setting-panel">
            <p class="section-hint">
              每天 03:10 自动清理超过保留天数的任务日志，设为 <code>-1</code> 表示不清理。
            </p>
            <div class="setting-item">
              <div class="setting-head">
                <span class="setting-name">任务日志保留天数</span>
              </div>
              <div class="setting-field">
                <t-input-number v-model="retentionDays" :min="-1" :max="3650" style="width: 140px" />
              </div>
            </div>
          </div>
        </t-tab-panel>
      </t-tabs>

      <div class="setting-actions">
        <t-button theme="primary" :disabled="!dirty" :loading="saving" @click="onSubmit">提交</t-button>
        <t-button variant="outline" :disabled="!dirty" @click="onReset">还原</t-button>
      </div>
    </t-card>
  </div>
</template>
<script setup lang="ts">
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, reactive, ref } from 'vue';

import type { TaskSettingValue } from '@/api/model/systemModel';
import { getSystemSettingsApi, updateSystemSettingsApi } from '@/api/system';

defineOptions({ name: 'SystemSetting' });

interface SettingItem {
  key: string;
  name: string;
  description: string;
}

// 系统任务设置项：后续新增任务在此登记（同时后端 service/system 里登记对应的默认值与校验）
const taskItems: SettingItem[] = [
  {
    key: 'task.job_status_sync',
    name: 'ETL 实时作业状态同步',
    description: '定时同步 SeaTunnel 实时作业的运行状态，关闭后作业状态将不再自动更新',
  },
  {
    key: 'task.ecs_sg_sync',
    name: 'ECS 安全组同步',
    description: '定时同步阿里云 ECS 安全组规则，关闭后需手动触发同步',
  },
];

const allItems = taskItems;

// 当前设置分组
const activeTab = ref('task');

// 日志保留天数设置项
const retentionKey = 'log.retention_days';
const defaultRetentionDays = 30;
const retentionDays = ref(defaultRetentionDays);
const savedRetentionDays = ref(defaultRetentionDays);

// 邮件设置项
const mailKey = 'mail.smtp';
interface MailSettingForm {
  enable: boolean;
  smtp_address: string;
  smtp_port: number;
  smtp_user: string;
  smtp_password: string;
  display_name: string;
  ssl: boolean;
}
function defaultMail(): MailSettingForm {
  return {
    enable: false,
    smtp_address: '',
    smtp_port: 465,
    smtp_user: '',
    smtp_password: '',
    display_name: '',
    ssl: true,
  };
}
const mailDraft = reactive<MailSettingForm>(defaultMail());
const savedMail = ref<MailSettingForm>(defaultMail());

function defaultValue(): TaskSettingValue {
  return { enabled: false, cron: '' };
}

const draft = reactive<Record<string, TaskSettingValue>>(
  Object.fromEntries(allItems.map((item) => [item.key, defaultValue()])),
);
const saved = ref<Record<string, TaskSettingValue>>(
  Object.fromEntries(allItems.map((item) => [item.key, defaultValue()])),
);
const loading = ref(false);
const saving = ref(false);

const dirty = computed(
  () =>
    allItems.some((item) => JSON.stringify(draft[item.key]) !== JSON.stringify(saved.value[item.key])) ||
    retentionDays.value !== savedRetentionDays.value ||
    JSON.stringify(mailDraft) !== JSON.stringify(savedMail.value),
);

function getErrorMessage(error: unknown, fallback: string) {
  const data = (error as { response?: { data?: { error?: string } } } | null)?.response?.data;
  if (data?.error) return data.error;
  return error instanceof Error && error.message ? error.message : fallback;
}

async function fetchSettings() {
  loading.value = true;
  try {
    const res = await getSystemSettingsApi();
    const data = res?.data || {};
    allItems.forEach((item) => {
      const value = data[item.key] as Partial<TaskSettingValue> | undefined;
      draft[item.key] = {
        enabled: Boolean(value?.enabled),
        cron: value?.cron ?? '',
      };
      saved.value[item.key] = { ...draft[item.key] };
    });
    const retention = data[retentionKey] as { retention_days?: number } | undefined;
    retentionDays.value = typeof retention?.retention_days === 'number' ? retention.retention_days : defaultRetentionDays;
    savedRetentionDays.value = retentionDays.value;
    const mail = data[mailKey] as Partial<MailSettingForm> | undefined;
    Object.assign(mailDraft, defaultMail(), {
      enable: Boolean(mail?.enable),
      smtp_address: mail?.smtp_address ?? '',
      smtp_port: typeof mail?.smtp_port === 'number' ? mail.smtp_port : 465,
      smtp_user: mail?.smtp_user ?? '',
      smtp_password: mail?.smtp_password ?? '',
      display_name: mail?.display_name ?? '',
      ssl: mail?.ssl ?? true,
    });
    savedMail.value = { ...mailDraft };
  } catch (error) {
    console.error(error);
    MessagePlugin.error('获取系统设置失败');
  } finally {
    loading.value = false;
  }
}

async function onSubmit() {
  const invalid = allItems.find((item) => draft[item.key].enabled && !draft[item.key].cron.trim());
  if (invalid) {
    MessagePlugin.warning(`请填写「${invalid.name}」的调度周期`);
    return;
  }
  const payload: Record<string, unknown> = Object.fromEntries(
    allItems
      .filter((item) => JSON.stringify(draft[item.key]) !== JSON.stringify(saved.value[item.key]))
      .map((item) => [item.key, draft[item.key]]),
  );
  if (retentionDays.value !== savedRetentionDays.value) {
    if (retentionDays.value !== -1 && retentionDays.value < 1) {
      MessagePlugin.warning('日志保留天数需不小于 1，-1 表示不清理');
      return;
    }
    payload[retentionKey] = { retention_days: retentionDays.value };
  }
  if (JSON.stringify(mailDraft) !== JSON.stringify(savedMail.value)) {
    if (mailDraft.enable && !mailDraft.smtp_address.trim()) {
      MessagePlugin.warning('启用邮件时必须填写 SMTP 服务器地址');
      return;
    }
    if (mailDraft.enable && !mailDraft.smtp_user.trim()) {
      MessagePlugin.warning('启用邮件时必须填写邮箱账号');
      return;
    }
    payload[mailKey] = { ...mailDraft };
  }
  if (!Object.keys(payload).length) return;
  saving.value = true;
  try {
    await updateSystemSettingsApi(payload);
    MessagePlugin.success('系统设置已保存');
    await fetchSettings();
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(getErrorMessage(error, '保存失败'));
  } finally {
    saving.value = false;
  }
}

function onReset() {
  allItems.forEach((item) => {
    draft[item.key] = { ...saved.value[item.key] };
  });
  retentionDays.value = savedRetentionDays.value;
  Object.assign(mailDraft, savedMail.value);
}

onMounted(fetchSettings);
</script>
<style lang="less" scoped>
.setting-card {
  margin: var(--td-comp-margin-s);
  padding: var(--td-comp-paddingTB-xxl) var(--td-comp-paddingLR-xxl);

  :deep(.t-card__body) {
    padding: 0;
  }
}

.setting-tabs {
  max-width: 720px;

  :deep(.t-tabs__content) {
    padding-top: var(--td-comp-margin-xxl);
  }
}

.setting-panel {
  min-height: 240px;
}

.section-hint {
  margin: 0 0 var(--td-comp-margin-xxl);
  font: var(--td-font-body-small);
  color: var(--td-text-color-secondary);
  line-height: 22px;

  code {
    padding: 0 4px;
    font-family: var(--td-font-family-monospace);
    background: var(--td-bg-color-container-hover);
    border-radius: var(--td-radius-small);
  }
}

.setting-item {
  &:not(:last-child) {
    margin-bottom: var(--td-comp-margin-xxl);
  }
}

.setting-head {
  display: flex;
  align-items: center;
  gap: var(--td-comp-margin-s);
}

.setting-name {
  font: var(--td-font-body-large);
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.setting-desc {
  margin: var(--td-comp-margin-xs) 0 0 calc(var(--td-comp-margin-s) + 32px);
  font: var(--td-font-body-small);
  color: var(--td-text-color-secondary);
  white-space: pre-line;
}

.setting-cron {
  margin-top: var(--td-comp-margin-s);
  padding-left: calc(var(--td-comp-margin-s) + 32px);
}

.setting-field {
  margin-top: var(--td-comp-margin-s);
}

.mail-fields {
  margin-top: var(--td-comp-margin-l);
  max-width: 520px;
}

.mail-field {
  display: flex;
  align-items: center;
  gap: var(--td-comp-margin-l);
  margin-bottom: var(--td-comp-margin-l);

  :deep(.t-input) {
    flex: 1;
  }
}

.mail-label {
  width: 110px;
  flex-shrink: 0;
  color: var(--td-text-color-secondary);
}

.cron-input {
  width: 320px;
}

.setting-actions {
  margin-top: var(--td-comp-margin-xxl);
  display: flex;
  gap: var(--td-comp-margin-s);
}
</style>
