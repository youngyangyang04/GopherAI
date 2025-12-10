<template>
  <div class="register-container">
    <el-card class="register-card">
      <template #header>
        <div class="card-header">
          <h2>注册</h2>
        </div>
      </template>
      <el-form
        ref="registerFormRef"
        :model="registerForm"
        :rules="registerRules"
        label-width="80px"
      >
        <el-form-item label="邮箱" prop="email">
          <el-input
            v-model="registerForm.email"
            placeholder="请输入邮箱"
            type="email"
          />
        </el-form-item>
        <el-form-item label="验证码" prop="captcha">
          <el-row :gutter="10">
            <el-col :span="16">
              <el-input
                v-model="registerForm.captcha"
                placeholder="请输入验证码"
              />
            </el-col>
            <el-col :span="8">
              <el-button
                type="primary"
                :loading="codeLoading"
                :disabled="countdown > 0"
                @click="sendCode"
                style="width: 100%"
              >
                {{ countdown > 0 ? `${countdown}s` : '发送验证码' }}
              </el-button>
            </el-col>
          </el-row>
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="registerForm.password"
            placeholder="请输入密码"
            type="password"
            show-password
          />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="registerForm.confirmPassword"
            placeholder="请再次输入密码"
            type="password"
            show-password
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            @click="handleRegister"
            style="width: 100%"
          >
            注册
          </el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            type="text"
            @click="$router.push('/login')"
            style="width: 100%"
          >
            已有账号？去登录
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '../utils/api'

export default {
  name: 'RegisterView',
  setup() {
    const router = useRouter()
    const registerFormRef = ref()
    const loading = ref(false)
    const codeLoading = ref(false)
    const countdown = ref(0)

    const registerForm = reactive({
      email: '',
      captcha: '',
      password: '',
      confirmPassword: ''
    })

    const validateConfirmPassword = (rule, value, callback) => {
      if (value !== registerForm.password) {
        callback(new Error('两次输入密码不一致'))
      } else {
        callback()
      }
    }

    const registerRules = {
      email: [
        { required: true, message: '请输入邮箱', trigger: 'blur' },
        { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
      ],
      captcha: [
        { required: true, message: '请输入验证码', trigger: 'blur' }
      ],
      password: [
        { required: true, message: '请输入密码', trigger: 'blur' },
        { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
      ],
      confirmPassword: [
        { required: true, message: '请确认密码', trigger: 'blur' },
        { validator: validateConfirmPassword, trigger: 'blur' }
      ]
    }

    const sendCode = async () => {
      if (!registerForm.email) {
        ElMessage.warning('请先输入邮箱')
        return
      }
      try {
        codeLoading.value = true
        const response = await api.post('/user/captcha', { email: registerForm.email })
        if (response.data.status_code === 1000) {
          ElMessage.success('验证码发送成功')
          countdown.value = 60
          const timer = setInterval(() => {
            countdown.value--
            if (countdown.value <= 0) {
              clearInterval(timer)
            }
          }, 1000)
        } else {
          ElMessage.error(response.data.status_msg || '验证码发送失败')
        }
      } catch (error) {
        console.error('Send code error:', error)
        ElMessage.error('验证码发送失败，请重试')
      } finally {
        codeLoading.value = false
      }
    }

    const handleRegister = async () => {
      try {
        await registerFormRef.value.validate()
        loading.value = true
        const response = await api.post('/user/register', {
              email: registerForm.email,
              captcha: registerForm.captcha,
              password: registerForm.password
        })
        if (response.data.status_code === 1000) {
          ElMessage.success('注册成功，请登录')
          router.push('/login')
        } else {
          ElMessage.error(response.data.status_msg || '注册失败')
        }
      } catch (error) {
        console.error('Register error:', error)
        ElMessage.error('注册失败，请重试')
      } finally {
        loading.value = false
      }
    }

    return {
      registerFormRef,
      loading,
      codeLoading,
      countdown,
      registerForm,
      registerRules,
      sendCode,
      handleRegister
    }
  }
}
</script>

<style scoped>
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: radial-gradient(circle at top, #e8f0fe 0%, #f8f9fa 60%, #ffffff 100%);
  padding: 24px;
}

.register-card {
  width: 460px;
  border-radius: 24px;
  border: 1px solid rgba(60, 64, 67, 0.12);
  box-shadow: 0 24px 60px rgba(60, 64, 67, 0.15);
  background: #fff;
  padding-bottom: 8px;
}

.card-header {
  text-align: center;
  padding: 24px 0 8px 0;
}

.card-header h2 {
  margin: 0;
  font-size: 28px;
  font-weight: 600;
  color: #202124;
}

.el-form {
  padding: 0 12px 16px;
}

.el-form-item {
  margin-bottom: 20px;
}

.el-input :deep(.el-input__wrapper) {
  box-shadow: none !important;
  border-radius: 14px;
  border: 1px solid rgba(60, 64, 67, 0.16);
  padding: 0 12px;
  background: #f8f9fa;
}

.el-input :deep(.el-input__wrapper.is-focus) {
  border-color: #1a73e8;
  box-shadow: 0 0 0 3px rgba(26, 115, 232, 0.12);
}

.el-input :deep(input) {
  font-size: 15px;
  background: transparent;
}

.el-row {
  align-items: center;
}

.el-button {
  height: 44px;
  font-weight: 600;
}

.el-button--text {
  color: #1a73e8;
  font-weight: 500;
}

@media (max-width: 520px) {
  .register-card {
    width: 100%;
  }
}
</style>
