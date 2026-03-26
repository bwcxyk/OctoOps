<template>
  <div class="login-bg">
    <el-card class="login-card">
      <img src="/src/assets/logo.png" class="login-logo" />
      <h2 class="login-title">OctoOps 运维平台</h2>
      <div style="height: 24px;"></div>
      <el-form :model="form" class="login-form" @keyup.enter.native="onLogin">
        <el-form-item>
          <el-input v-model="form.username" autocomplete="username" placeholder="用户名">
          </el-input>
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" autocomplete="current-password" placeholder="密码" show-password>
          </el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" style="width:100%" @click="onLogin">登录</el-button>
        </el-form-item>
        <div style="text-align:right;margin-top:-10px;">
          <el-link type="primary" @click="showForgot = true" underline="hover">忘记密码？</el-link>
        </div>
      </el-form>
    </el-card>
    <!-- 忘记密码弹窗 -->
    <el-dialog v-model="showForgot" title="忘记密码" width="400px" @close="resetForgot" :close-on-click-modal="false">
      <el-form :model="forgotForm" :rules="forgotRules" ref="forgotFormRef" label-width="90px">
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="forgotForm.email" autocomplete="off" />
        </el-form-item>
        <el-form-item label="验证码" prop="code">
          <div style="display: flex; gap: 8px; width: 100%;">
            <el-input v-model="forgotForm.code" autocomplete="off" />
            <el-button
              :disabled="codeTimer > 0 || !forgotForm.email"
              @click="onSendCode"
              type="primary"
              style="min-width: 100px;"
            >
              {{ codeTimer > 0 ? codeTimer + 's后重试' : '获取验证码' }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="新密码" prop="new_password">
          <el-input v-model="forgotForm.new_password" type="password" autocomplete="off" />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirm_password">
          <el-input v-model="forgotForm.confirm_password" type="password" autocomplete="off" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showForgot = false">取消</el-button>
        <el-button type="primary" @click="onForgotSubmit" :loading="forgotLoading">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/store/user'
import { ElMessage } from 'element-plus'
import { forgotPassword, sendResetCode } from '@/api/user'

const router = useRouter()
const userStore = useUserStore()
const form = ref({ username: '', password: '' })

async function onLogin() {
  try {
    await userStore.login(form.value)
    await userStore.fetchUserInfo()
    router.push('/')
  } catch (e) {
    ElMessage.error(e.message)
  }
}

// 忘记密码弹窗相关
const showForgot = ref(false)
const forgotLoading = ref(false)
const forgotFormRef = ref()
const codeTimer = ref(0)
let timer = null

function onSendCode() {
  forgotFormRef.value.validateField(['email'], async (valid) => {
    if (!valid) return
    try {
      await sendResetCode({
        email: forgotForm.email
      })
      ElMessage.success('验证码已发送到邮箱')
      codeTimer.value = 60
      timer = setInterval(() => {
        codeTimer.value--
        if (codeTimer.value <= 0) clearInterval(timer)
      }, 1000)
    } catch (e) {
      ElMessage.error(e?.response?.data?.message || '发送失败')
    }
  })
}

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

const forgotForm = reactive({
  email: '',
  code: '',
  new_password: '',
  confirm_password: ''
})
const forgotRules = {
  email: [{ required: true, message: '请输入邮箱', trigger: 'blur' }],
  code: [{ required: true, message: '请输入验证码', trigger: 'blur' }],
  new_password: [{ required: true, message: '请输入新密码', trigger: 'blur' }],
  confirm_password: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    { validator: (rule, value) => value === forgotForm.new_password, message: '两次输入不一致', trigger: 'blur' }
  ]
}
function resetForgot() {
  Object.assign(forgotForm, { email: '', code: '', new_password: '', confirm_password: '' })
  forgotFormRef.value && forgotFormRef.value.clearValidate()
}
function onForgotSubmit() {
  forgotFormRef.value.validate(async valid => {
    if (!valid) return
    forgotLoading.value = true
    try {
      await forgotPassword({
        email: forgotForm.email,
        code: forgotForm.code,
        new_password: forgotForm.new_password
      })
      ElMessage.success('密码重置成功，请重新登录')
      showForgot.value = false
    } catch (e) {
      ElMessage.error(e?.response?.data?.message || '重置失败')
    } finally {
      forgotLoading.value = false
    }
  })
}
</script>

<style scoped>
.login-bg {
  min-height: 100vh;
  background: linear-gradient(135deg, #4A90E2 0%, #1976d2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}
.login-card {
  width: 360px;
  padding: 32px 32px 18px 32px;
  border-radius: 16px;
  box-shadow: 0 8px 32px #1976d255;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.login-logo {
  width: 72px;
  height: 72px;
  margin-bottom: 12px;
  border-radius: 12px;
  object-fit: contain;
  background: #fff;
  box-shadow: 0 2px 8px #1976d233;
  display: block;
  margin-left: auto;
  margin-right: auto;
}
.login-title {
  margin: 0 0 6px 0;
  font-size: 1.5rem;
  font-weight: bold;
  color: #1976d2;
  letter-spacing: 2px;
  text-align: center;
}
.login-subtitle {
  color: #888;
  font-size: 1rem;
  margin-bottom: 18px;
}
.login-form {
  width: 100%;
  margin-top: 8px;
}

</style> 