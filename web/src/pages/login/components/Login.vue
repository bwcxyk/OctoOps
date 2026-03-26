<template>
  <div>
    <t-form
      ref="form"
      class="item-container"
      :class="[`login-${type}`]"
      :data="formData"
      :rules="FORM_RULES"
      label-width="0"
      @submit="onSubmit"
    >
      <template v-if="type === 'password'">
        <t-form-item name="account">
          <t-input v-model="formData.account" size="large" :placeholder="t('pages.login.input.account')">
            <template #prefix-icon>
              <t-icon name="user" />
            </template>
          </t-input>
        </t-form-item>

        <t-form-item name="password">
          <t-input
            v-model="formData.password"
            size="large"
            :type="showPsw ? 'text' : 'password'"
            clearable
            :placeholder="t('pages.login.input.password')"
          >
            <template #prefix-icon>
              <t-icon name="lock-on" />
            </template>
            <template #suffix-icon>
              <t-icon :name="showPsw ? 'browse' : 'browse-off'" @click="showPsw = !showPsw" />
            </template>
          </t-input>
        </t-form-item>

        <div class="check-container remember-pwd">
          <t-checkbox v-model="formData.checked">{{ t('pages.login.remember') }}</t-checkbox>
          <span class="tip" @click="forgotVisible = true">{{ t('pages.login.forget') }}</span>
        </div>
      </template>

      <!-- 扫码登录 -->
      <template v-else-if="type === 'qrcode'">
        <div class="tip-container">
          <span class="tip">{{ t('pages.login.wechatLogin') }}</span>
          <span class="refresh">{{ t('pages.login.refresh') }} <t-icon name="refresh" /> </span>
        </div>
        <qrcode-vue value="" :size="160" level="H" />
      </template>

      <!-- 手机号登录 -->
      <template v-else>
        <t-form-item name="phone">
          <t-input v-model="formData.phone" size="large" :placeholder="t('pages.login.input.phone')">
            <template #prefix-icon>
              <t-icon name="mobile" />
            </template>
          </t-input>
        </t-form-item>

        <t-form-item class="verification-code" name="verifyCode">
          <t-input v-model="formData.verifyCode" size="large" :placeholder="t('pages.login.input.verification')" />
          <t-button size="large" variant="outline" :disabled="countDown > 0" @click="sendCode">
            {{ countDown === 0 ? t('pages.login.sendVerification') : `${countDown}秒后可重发` }}
          </t-button>
        </t-form-item>
      </template>

      <t-form-item v-if="type !== 'qrcode'" class="btn-container">
        <t-button block size="large" type="submit"> {{ t('pages.login.signIn') }} </t-button>
      </t-form-item>

      <div class="switch-container">
        <span v-if="type !== 'password'" class="tip" @click="switchType('password')">{{
          t('pages.login.accountLogin')
        }}</span>
        <span v-if="type !== 'qrcode'" class="tip" @click="switchType('qrcode')">{{
          t('pages.login.wechatLogin')
        }}</span>
        <span v-if="type !== 'phone'" class="tip" @click="switchType('phone')">{{ t('pages.login.phoneLogin') }}</span>
      </div>
    </t-form>

    <t-dialog
      v-model:visible="forgotVisible"
      header="忘记密码"
      width="520px"
      :confirm-btn="{ content: '重置密码', theme: 'primary', loading: forgotSubmitting }"
      @confirm="onForgotSubmit"
    >
      <t-form ref="forgotFormRef" :data="forgotForm" :rules="FORGOT_RULES" label-width="100px">
        <t-form-item name="email" label="邮箱">
          <t-input v-model="forgotForm.email" placeholder="请输入邮箱" />
        </t-form-item>
        <t-form-item class="verification-code" name="code" label="验证码">
          <t-input v-model="forgotForm.code" placeholder="请输入验证码" />
          <t-button variant="outline" :disabled="countDown > 0" @click="sendResetCode">
            {{ countDown === 0 ? '发送验证码' : `${countDown}秒后可重发` }}
          </t-button>
        </t-form-item>
        <t-form-item name="new_password" label="新密码">
          <t-input v-model="forgotForm.new_password" type="password" placeholder="请输入新密码" />
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>
<script setup lang="ts">
import QrcodeVue from 'qrcode.vue';
import type { FormInstanceFunctions, FormRule, SubmitContext } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { onMounted, reactive, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { forgotPasswordApi, sendResetCodeApi } from '@/api/auth';
import { useCounter } from '@/hooks';
import { t } from '@/locales';
import { useUserStore } from '@/store';

const userStore = useUserStore();

const INITIAL_DATA = {
  phone: '',
  account: '',
  password: '',
  verifyCode: '',
  checked: false,
};
const REMEMBER_ACCOUNT_KEY = 'remembered_account';

const FORM_RULES: Record<string, FormRule[]> = {
  phone: [{ required: true, message: t('pages.login.required.phone'), type: 'error' }],
  account: [{ required: true, message: t('pages.login.required.account'), type: 'error' }],
  password: [{ required: true, message: t('pages.login.required.password'), type: 'error' }],
  verifyCode: [{ required: true, message: t('pages.login.required.verification'), type: 'error' }],
};

const FORGOT_RULES: Record<string, FormRule[]> = {
  email: [
    { required: true, message: '邮箱必填', type: 'error' },
    { email: true, message: '请输入正确的邮箱', type: 'warning' },
  ],
  code: [{ required: true, message: '验证码必填', type: 'error' }],
  new_password: [{ required: true, message: '新密码必填', type: 'error' }],
};

const type = ref('password');

const form = ref<FormInstanceFunctions>();
const forgotFormRef = ref<FormInstanceFunctions>();
const formData = ref({ ...INITIAL_DATA });
const showPsw = ref(false);
const forgotVisible = ref(false);
const forgotSubmitting = ref(false);
const forgotForm = reactive({
  email: '',
  code: '',
  new_password: '',
});

const [countDown, handleCounter] = useCounter();

onMounted(() => {
  const rememberedAccount = localStorage.getItem(REMEMBER_ACCOUNT_KEY);
  if (rememberedAccount) {
    formData.value.account = rememberedAccount;
    formData.value.checked = true;
  }
});

const switchType = (val: string) => {
  if (val === 'qrcode') {
    MessagePlugin.info('使用微信扫一扫登录功能待开发');
    return;
  }
  if (val === 'phone') {
    MessagePlugin.info('使用手机号登录功能待开发');
    return;
  }
  type.value = val;
};

const router = useRouter();
const route = useRoute();

/**
 * 手机号验证码发送（待开发提示）
 */
const sendCode = () => {
  MessagePlugin.info('使用手机号登录功能待开发');
};

const sendResetCode = async () => {
  const valid = await forgotFormRef.value?.validate({ fields: ['email'] });
  if (valid !== true) return;
  try {
    await sendResetCodeApi(forgotForm.email);
    MessagePlugin.success('验证码已发送');
    handleCounter();
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(error instanceof Error ? error.message : '发送验证码失败');
  }
};

const onForgotSubmit = async () => {
  const valid = await forgotFormRef.value?.validate();
  if (valid !== true) return;
  forgotSubmitting.value = true;
  try {
    await forgotPasswordApi({
      email: forgotForm.email,
      code: forgotForm.code,
      new_password: forgotForm.new_password,
    });
    MessagePlugin.success('密码重置成功，请重新登录');
    forgotVisible.value = false;
  } catch (error: unknown) {
    console.error(error);
    MessagePlugin.error(error instanceof Error ? error.message : '重置密码失败');
  } finally {
    forgotSubmitting.value = false;
  }
};

const onSubmit = async (ctx: SubmitContext) => {
  if (ctx.validateResult === true) {
    try {
      await userStore.login(formData.value);

      if (formData.value.checked && formData.value.account) {
        localStorage.setItem(REMEMBER_ACCOUNT_KEY, formData.value.account);
      } else {
        localStorage.removeItem(REMEMBER_ACCOUNT_KEY);
      }

      MessagePlugin.success('登录成功');
      const redirect = route.query.redirect as string;
      const redirectUrl = redirect ? decodeURIComponent(redirect) : '/dashboard';
      router.push(redirectUrl);
    } catch (error: unknown) {
      console.log(error);
      MessagePlugin.error(error instanceof Error ? error.message : '登录失败');
    }
  }
};
</script>
<style lang="less" scoped>
@import '../index.less';
</style>
