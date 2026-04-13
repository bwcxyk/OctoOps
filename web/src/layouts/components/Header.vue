<template>
  <div :class="layoutCls">
    <t-head-menu :class="menuCls" :theme="menuTheme" expand-type="popup" :value="active">
      <template #logo>
        <span v-if="showLogo" class="header-logo-container" @click="handleNav('/dashboard/base')">
          <logo-full class="t-logo" />
        </span>
        <div v-else class="header-operate-left">
          <t-button theme="default" shape="square" variant="text" @click="changeCollapsed">
            <t-icon class="collapsed-icon" name="view-list" />
          </t-button>
          <search :layout="layout" />
        </div>
      </template>
      <template v-if="layout !== 'side'" #default>
        <menu-content class="header-menu" :nav-data="menu" />
      </template>
      <template #operations>
        <div class="operations-container">
          <!-- 搜索框 -->
          <search v-if="layout !== 'side'" :layout="layout" />

          <!-- 全局通知 -->
          <notice />

          <t-dropdown trigger="click">
            <t-button theme="default" shape="square" variant="text">
              <translate-icon />
            </t-button>
            <template #dropdown>
              <t-dropdown-item
                v-for="(lang, index) in langList"
                :key="index"
                :value="lang.value"
                @click="changeLang(String(lang.value))"
              >
                {{ lang.content }}
              </t-dropdown-item>
            </template>
          </t-dropdown>
          <t-dropdown :min-column-width="120" trigger="click">
            <template #dropdown>
              <t-dropdown-item class="operations-dropdown-container-item" @click="showChangePasswordDialog">
                <t-icon name="lock-on" />{{ t('layout.header.changePassword') }}
              </t-dropdown-item>
              <t-dropdown-item class="operations-dropdown-container-item" @click="handleLogout">
                <poweroff-icon />{{ t('layout.header.signOut') }}
              </t-dropdown-item>
            </template>
            <t-button class="header-user-btn" theme="default" variant="text">
              <template #icon>
                <t-icon class="header-user-avatar" name="user-circle" />
              </template>
              <div class="header-user-account">{{ user.userInfo.name }}</div>
              <template #suffix><chevron-down-icon /></template>
            </t-button>
          </t-dropdown>
          <t-tooltip placement="bottom" :content="t('layout.header.setting')">
            <t-button theme="default" shape="square" variant="text" @click="toggleSettingPanel">
              <setting-icon />
            </t-button>
          </t-tooltip>
        </div>
      </template>
    </t-head-menu>

    <t-dialog
      v-model:visible="changePasswordVisible"
      :header="t('layout.header.changePassword')"
      width="520px"
      :confirm-btn="{ content: '确定', theme: 'primary', loading: changePasswordSubmitting }"
      :cancel-btn="{ content: '取消' }"
      @confirm="submitChangePassword"
      @close="resetChangePasswordForm"
    >
      <t-form ref="changePasswordFormRef" :data="changePasswordForm" :rules="CHANGE_PASSWORD_RULES" label-width="100px">
        <t-form-item name="old_password" label="原密码">
          <t-input v-model="changePasswordForm.old_password" type="password" placeholder="请输入原密码" />
        </t-form-item>
        <t-form-item name="new_password" label="新密码">
          <t-input v-model="changePasswordForm.new_password" type="password" placeholder="请输入新密码" />
        </t-form-item>
        <t-form-item name="confirm_password" label="确认密码">
          <t-input v-model="changePasswordForm.confirm_password" type="password" placeholder="请再次输入新密码" />
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>
<script setup lang="ts">
import { ChevronDownIcon, PoweroffIcon, SettingIcon, TranslateIcon } from 'tdesign-icons-vue-next';
import type { FormInstanceFunctions, FormRule } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import type { PropType } from 'vue';
import { computed, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';

import { changePasswordApi } from '@/api/auth';
import LogoFull from '@/assets/assets-logo-full.svg?component';
import { prefix } from '@/config/global';
import { langList, t } from '@/locales';
import { useLocale } from '@/locales/useLocale';
import { getActive } from '@/router';
import { useSettingStore, useUserStore } from '@/store';
import type { MenuRoute, ModeType } from '@/types/interface';

import MenuContent from './MenuContent.vue';
import Notice from './Notice.vue';
import Search from './Search.vue';

const { theme, layout, showLogo, menu, isFixed, isCompact } = defineProps({
  theme: {
    type: String,
    default: 'light',
  },
  layout: {
    type: String,
    default: 'top',
  },
  showLogo: {
    type: Boolean,
    default: true,
  },
  menu: {
    type: Array as PropType<MenuRoute[]>,
    default: () => [],
  },
  isFixed: {
    type: Boolean,
    default: false,
  },
  isCompact: {
    type: Boolean,
    default: false,
  },
  maxLevel: {
    type: Number,
    default: 3,
  },
});

const router = useRouter();
const settingStore = useSettingStore();
const user = useUserStore();
const changePasswordVisible = ref(false);
const changePasswordSubmitting = ref(false);
const changePasswordFormRef = ref<FormInstanceFunctions>();
const changePasswordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
});
const CHANGE_PASSWORD_RULES: Record<string, FormRule[]> = {
  old_password: [{ required: true, message: '请输入原密码', type: 'error' }],
  new_password: [{ required: true, message: '请输入新密码', type: 'error' }],
  confirm_password: [{ required: true, message: '请确认新密码', type: 'error' }],
};

const toggleSettingPanel = () => {
  settingStore.updateConfig({
    showSettingPanel: true,
  });
};

const active = computed(() => getActive());

const layoutCls = computed(() => [`${prefix}-header-layout`]);

const menuCls = computed(() => {
  return [
    {
      [`${prefix}-header-menu`]: !isFixed,
      [`${prefix}-header-menu-fixed`]: isFixed,
      [`${prefix}-header-menu-fixed-side`]: layout === 'side' && isFixed,
      [`${prefix}-header-menu-fixed-side-compact`]: layout === 'side' && isFixed && isCompact,
    },
  ];
});
const menuTheme = computed(() => theme as ModeType);

// 切换语言
const { changeLocale } = useLocale();
const changeLang = (lang: string) => {
  changeLocale(lang);
};

const changeCollapsed = () => {
  settingStore.updateConfig({
    isSidebarCompact: !settingStore.isSidebarCompact,
  });
};

const handleNav = (url: string) => {
  router.push(url);
};

const resetChangePasswordForm = () => {
  changePasswordForm.old_password = '';
  changePasswordForm.new_password = '';
  changePasswordForm.confirm_password = '';
};

const showChangePasswordDialog = () => {
  changePasswordVisible.value = true;
};

const submitChangePassword = async () => {
  const valid = await changePasswordFormRef.value?.validate();
  if (valid !== true) return;
  if (changePasswordForm.new_password !== changePasswordForm.confirm_password) {
    MessagePlugin.warning('两次输入的新密码不一致');
    return;
  }

  changePasswordSubmitting.value = true;
  try {
    await changePasswordApi({
      old_password: changePasswordForm.old_password,
      new_password: changePasswordForm.new_password,
    });
    MessagePlugin.success('修改密码成功，请重新登录');
    changePasswordVisible.value = false;
    await user.logout();
    await router.push('/login');
  } catch (error: unknown) {
    MessagePlugin.error(error instanceof Error ? error.message : '修改密码失败');
  } finally {
    changePasswordSubmitting.value = false;
  }
};

const handleLogout = () => {
  router.push({
    path: '/login',
    query: { redirect: encodeURIComponent(router.currentRoute.value.fullPath) },
  });
};

</script>
<style lang="less" scoped>
.@{starter-prefix}-header {
  &-menu-fixed {
    position: fixed;
    top: 0;
    z-index: 1001;

    :deep(.t-head-menu__inner) {
      padding-right: var(--td-comp-margin-xl);
    }

    &-side {
      left: 232px;
      right: 0;
      z-index: 10;
      width: auto;
      transition: all 0.3s;

      &-compact {
        left: 64px;
      }
    }
  }

  &-logo-container {
    cursor: pointer;
    display: inline-flex;
  }
}

.header-menu {
  flex: 1 1 auto;
  display: inline-flex;

  :deep(.t-menu__item) {
    min-width: unset;
  }
}

.operations-container {
  display: flex;
  align-items: center;

  .t-popup__reference {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .t-button {
    margin-left: var(--td-comp-margin-l);
  }
}

.header-operate-left {
  display: flex;
  align-items: normal;
  line-height: 0;
}

.header-logo-container {
  width: 184px;
  height: 26px;
  display: flex;
  margin-left: 24px;
  color: var(--td-text-color-primary);

  .t-logo {
    width: 100%;
    height: 100%;

    &:hover {
      cursor: pointer;
    }
  }

  &:hover {
    cursor: pointer;
  }
}

.header-user-account {
  display: inline-flex;
  align-items: center;
  color: var(--td-text-color-primary);
}

:deep(.t-head-menu__inner) {
  border-bottom: 1px solid var(--td-component-stroke);
}

.t-menu--light {
  .header-user-account {
    color: var(--td-text-color-primary);
  }
}

.t-menu--dark {
  .t-head-menu__inner {
    border-bottom: 1px solid var(--td-gray-color-10);
  }

  .header-user-account {
    color: rgb(255 255 255 / 55%);
  }
}

.operations-dropdown-container-item {
  width: 100%;
  display: flex;
  align-items: center;

  :deep(.t-dropdown__item-text) {
    display: flex;
    align-items: center;
  }

  .t-icon {
    font-size: var(--td-comp-size-xxxs);
    margin-right: var(--td-comp-margin-s);
  }

  :deep(.t-dropdown__item) {
    width: 100%;
    margin-bottom: 0;
  }

  &:last-child {
    :deep(.t-dropdown__item) {
      margin-bottom: 8px;
    }
  }
}
</style>
<!-- eslint-disable-next-line vue-scoped-css/enforce-style-type -->
<style lang="less">
.operations-dropdown-container-item {
  .t-dropdown__item-text {
    display: flex;
    align-items: center;
  }
}
</style>
