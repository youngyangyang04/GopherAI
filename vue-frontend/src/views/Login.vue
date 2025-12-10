<template>
  <div class="login-container">
    <el-card class="login-card">
      <template #header>
        <div class="card-header">
          <h2>登录</h2>
        </div>
      </template>
      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        label-width="80px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="请输入用户名"
          />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="loginForm.password"
            placeholder="请输入密码"
            type="password"
            show-password
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            @click="handleLogin"
            style="width: 100%"
          >
            登录
          </el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            type="text"
            @click="$router.push('/register')"
            style="width: 100%"
          >
            还没有账号？去注册
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '../utils/api'

export default {
  name: 'LoginView',
  setup() {
    const router = useRouter()
    const loginFormRef = ref()
    const loading = ref(false)
    const loginForm = ref({
      username: '',
      password: ''
    })

    const loginRules = {
      username: [
        { required: true, message: '请输入用户名', trigger: 'blur' }
      ],
      password: [
        { required: true, message: '请输入密码', trigger: 'blur' },
        { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
      ]
    }

    const handleLogin = async () => {
      try {
        await loginFormRef.value.validate()
        loading.value = true
        const response = await api.post('/user/login', {
          username: loginForm.value.username,
          password: loginForm.value.password
        })
        if (response.data.status_code === 1000) {
          localStorage.setItem('token', response.data.token)
          ElMessage.success('登录成功')
          router.push('/menu')
        } else {
          ElMessage.error(response.data.status_msg || '登录失败')
        }
      } catch (error) {
        console.error('Login error:', error)
        ElMessage.error('登录失败，请重试')
      } finally {
        loading.value = false
      }
    }

    return {
      loginFormRef,
      loading,
      loginForm,
      loginRules,
      handleLogin
    }
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: radial-gradient(circle at top, #e8f0fe 0%, #f8f9fa 60%, #ffffff 100%);
  padding: 24px;
}

.login-card {
  width: 420px;
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
  margin-bottom: 22px;
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

.el-button {
  height: 46px;
  font-weight: 600;
}

.el-button--text {
  color: #1a73e8;
  font-weight: 500;
}

@media (max-width: 480px) {
  .login-card {
    width: 100%;
  }
}
</style>
