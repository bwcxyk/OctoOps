<template>
  <div>
    <t-card class="setting-card" :bordered="false" :loading="loading">
      <t-tabs v-model="activeTab" class="setting-tabs">
        <t-tab-panel value="task" label="系统任务">
          <div class="setting-panel">
            <t-alert theme="info" class="setting-notice">
              Cron 表达式为 6 位（秒 分 时 日 月 周），例如 <code>0 */10 * * * *</code> 每 10 分钟执行一次，
              <code>0 0 8 * * *</code> 每天 8 点执行一次。开关关闭时任务不会调度。
            </t-alert>
            <div v-for="item in taskItems" :key="item.key" class="setting-item">
              <div class="setting-head">
                <t-switch v-model="draft[item.key].enabled" />
                <span class="setting-name">{{ item.name }}</span>
              </div>
              <div class="section-body">
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
          </div>
        </t-tab-panel>

        <t-tab-panel value="mail" label="邮件设置">
          <div class="setting-panel">
            <t-alert theme="info" class="setting-notice">
              修改配置后请先点击「提交」保存，再发送测试邮件，以确保使用的是已保存的配置。
            </t-alert>
            <div class="setting-item">
              <div class="setting-head">
                <t-switch v-model="mailDraft.enable" />
                <span class="setting-name">启用邮件通知</span>
              </div>
              <div class="section-body">
                <p class="setting-desc">
                  用于告警通知与找回密码验证码邮件，关闭后不会发送任何邮件。出于安全考虑，已保存的密码不会回显，
                  显示为 <code>******</code>，保持原样或留空表示不修改。
                </p>
              </div>
            </div>
            <template v-if="mailDraft.enable">
              <h3 class="section-title">服务</h3>
              <div class="mail-fields section-body">
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
                    :placeholder="portPlaceholder"
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
                  <span class="mail-label">使用 SSL/TLS</span>
                  <t-radio-group v-model="mailDraft.security">
                    <t-radio value="none">无</t-radio>
                    <t-radio value="ssl">SSL</t-radio>
                    <t-radio value="tls">TLS</t-radio>
                  </t-radio-group>
                </div>
              </div>
              <h3 class="section-title">测试</h3>
              <div class="mail-fields section-body">
                <div class="mail-field">
                  <span class="mail-label">收件人</span>
                  <t-input v-model="testRecipient" class="test-input" placeholder="接收测试邮件的邮箱" clearable />
                </div>
              </div>
            </template>
          </div>
        </t-tab-panel>

        <t-tab-panel value="log" label="日志设置">
          <div class="setting-panel">
            <t-alert theme="info" class="setting-notice">
              每天 03:10 自动清理超过保留天数的任务日志，设为 <code>-1</code> 表示不清理。
            </t-alert>
            <h3 class="section-title">日志保留</h3>
            <div class="mail-fields section-body">
              <div class="mail-field">
                <span class="mail-label">保留天数</span>
                <t-input-number v-model="retentionDays" :min="-1" :max="3650" style="width: 140px" />
              </div>
            </div>
          </div>
        </t-tab-panel>
      </t-tabs>

      <div class="setting-actions">
        <t-button theme="primary" :disabled="!dirty" :loading="saving" @click="onSubmit">提交</t-button>
        <t-button variant="outline" :disabled="!dirty" @click="onReset">还原</t-button>
        <t-button
          v-if="activeTab === 'mail' && mailDraft.enable"
          variant="outline"
          :loading="testing"
          @click="onSendTestMail"
        >
          测试发送
        </t-button>
      </div>
    </t-card>
  </div>
</template>
<script setup lang="ts">
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, reactive, ref } from 'vue';

import type { TaskSettingValue } from '@/api/model/systemModel';
import { getSystemSettingsApi, testMailSettingsApi, updateSystemSettingsApi } from '@/api/system';

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
type MailSecurity = 'none' | 'ssl' | 'tls';
interface MailSettingForm {
  enable: boolean;
  smtp_address: string;
  smtp_port: number;
  smtp_user: string;
  smtp_password: string;
  display_name: string;
  security: MailSecurity;
}
// 加密方式：none 不加密 / ssl 隐式 SSL（465）/ tls STARTTLS（587）
const defaultSecurity: MailSecurity = 'ssl';
const portHint: Record<MailSecurity, string> = {
  none: '如 25',
  ssl: '如 465',
  tls: '如 587',
};

function defaultMail(): MailSettingForm {
  return {
    enable: false,
    smtp_address: '',
    smtp_port: 465,
    smtp_user: '',
    smtp_password: '',
    display_name: '',
    security: defaultSecurity,
  };
}
const mailDraft = reactive<MailSettingForm>(defaultMail());
const savedMail = ref<MailSettingForm>(defaultMail());
// 端口占位提示跟随加密方式变化
const portPlaceholder = computed(() => portHint[mailDraft.security]);

// 测试邮件
const testRecipient = ref('');
const testing = ref(false);

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
      security: (mail?.security as MailSecurity) || defaultSecurity,
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

async function onSendTestMail() {
  const to = testRecipient.value.trim();
  if (!to) {
    MessagePlugin.warning('请填写收件人邮箱');
    return;
  }
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(to)) {
    MessagePlugin.warning('收件人邮箱格式不正确');
    return;
  }
  testing.value = true;
  try {
    await testMailSettingsApi({ to });
    MessagePlugin.success('测试邮件已发送，请查收');
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(getErrorMessage(error, '测试邮件发送失败'));
  } finally {
    testing.value = false;
  }
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

/* 各分组顶部统一提示条 */
.setting-notice {
  margin-bottom: var(--td-comp-margin-xxl);

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

/* 分组内标题：比分节标题（title-large）小一档 */
.setting-name {
  font: var(--td-font-title-small);
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.setting-desc {
  margin: var(--td-comp-margin-xs) 0 0;
  font: var(--td-font-body-small);
  color: var(--td-text-color-secondary);
  white-space: pre-line;
}

.setting-cron {
  margin-top: var(--td-comp-margin-s);
}

.section-title {
  margin: 0 0 var(--td-comp-margin-l);
  padding-top: var(--td-comp-margin-xxl);
  border-top: 1px solid var(--td-component-stroke);
  font: var(--td-font-title-large);
  font-weight: 600;
  color: var(--td-text-color-primary);
}

/* 标题与配置区错开：字段整体缩进 */
.section-body {
  padding-left: 24px;
}

.mail-fields {
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

.test-input {
  max-width: 260px;
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
