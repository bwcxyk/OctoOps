<template>
  <div class="app-header">
    <div class="header-left">
      <img src="/src/assets/logo.png" alt="logo" class="logo" />
      <span class="title">OctoOps</span>
    </div>
    <div class="header-right">
      <el-dropdown @command="onCommand">
        <span class="user-dropdown">
          <el-avatar :size="32" :src="user?.avatar" icon="el-icon-user" style="background:#1976d2;">
            <template #icon><el-icon><User /></el-icon></template>
          </el-avatar>
          <span class="username">{{ user?.nickname || user?.username || '用户' }}</span>
          <el-icon><ArrowDown /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="password">
              <el-icon><Lock /></el-icon> 修改密码
            </el-dropdown-item>
            <el-dropdown-item divided command="logout">
              <el-icon><SwitchButton /></el-icon> 退出登录
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
  <!-- 修改密码弹窗 -->
  <el-dialog v-model="showChangePwd" title="修改密码" width="400px" @close="resetChangePwd">
    <el-form :model="changePwdForm" :rules="changePwdRules" ref="changePwdFormRef" label-width="90px">
      <el-form-item label="原密码" prop="old_password">
        <el-input v-model="changePwdForm.old_password" type="password" autocomplete="off" />
      </el-form-item>
      <el-form-item label="新密码" prop="new_password">
        <el-input v-model="changePwdForm.new_password" type="password" autocomplete="off" />
      </el-form-item>
      <el-form-item label="确认密码" prop="confirm_password">
        <el-input v-model="changePwdForm.confirm_password" type="password" autocomplete="off" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="showChangePwd = false">取消</el-button>
      <el-button type="primary" @click="onChangePwdSubmit" :loading="changePwdLoading">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { useUserStore } from '@/store/user'
import { useRouter } from 'vue-router'
import { computed, ref, reactive } from 'vue'
import { ArrowDown, User, Lock, SwitchButton, DArrowRight } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { changePassword } from '@/api/user'

const userStore = useUserStore()
const router = useRouter()
const user = computed(() => userStore.user)

const showChangePwd = ref(false)
const changePwdLoading = ref(false)
const changePwdFormRef = ref()
const changePwdForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})
const changePwdRules = {
  old_password: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
  new_password: [{ required: true, message: '请输入新密码', trigger: 'blur' }],
  confirm_password: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    { validator: (rule, value) => value === changePwdForm.new_password, message: '两次输入不一致', trigger: 'blur' }
  ]
}
function resetChangePwd() {
  Object.assign(changePwdForm, { old_password: '', new_password: '', confirm_password: '' })
  changePwdFormRef.value && changePwdFormRef.value.clearValidate()
}
function onChangePwdSubmit() {
  changePwdFormRef.value.validate(async valid => {
    if (!valid) return
    changePwdLoading.value = true
    try {
      await changePassword({
        old_password: changePwdForm.old_password,
        new_password: changePwdForm.new_password
      })
      ElMessage.success('修改密码成功，请重新登录')
      showChangePwd.value = false
      userStore.logout()
      router.push('/login')
    } catch (e) {
      ElMessage.error(e?.response?.data?.message || '修改失败')
    } finally {
      changePwdLoading.value = false
    }
  })
}

function onCommand(cmd) {
  if (cmd === 'logout') {
    userStore.logout()
    router.push('/login')
  } else if (cmd === 'profile') {
    // 跳转到个人信息页
  } else if (cmd === 'password') {
    showChangePwd.value = true
  }
  // 其它命令可扩展
}
</script>

<style scoped>
.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #4A90E2;
  color: #fff;
  padding: 0 32px;
  height: 56px;
  width: 100vw;
  min-width: 0;
  box-sizing: border-box;
  margin: 0;
  border-radius: 0;
}
.header-left {
  display: flex;
  align-items: center;
}
.logo {
  width: 65px;
  height: 65px;
  margin-right: 14px;
  background: transparent;
  border-radius: 8px;
  padding: 0;
  object-fit: contain;
  display: block;
}
.title {
  font-size: 1.25rem;
  font-weight: bold;
  letter-spacing: 2px;
}
.header-right {
  font-size: 1rem;
  opacity: 0.95;
  display: flex;
  align-items: center;
  gap: 24px;
}
.user-dropdown {
  display: flex;
  align-items: center;
  cursor: pointer;
  color: #fff;
  font-weight: 500;
}
.username {
  margin: 0 8px;
  font-size: 1rem;
}
</style> 