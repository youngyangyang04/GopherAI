<template>
  <div class="menu-container">
    <el-header class="header">
      <h1>AI应用平台</h1>
      <el-button type="danger" @click="handleLogout">退出登录</el-button>
    </el-header>
    <el-main class="main">
      <div class="menu-grid">
        <el-card class="menu-item" @click="$router.push('/ai-chat')">
          <div class="card-content">
            <el-icon size="48" color="#409eff"><ChatDotRound /></el-icon>
            <h3>AI聊天</h3>
            <p>与AI进行智能对话</p>
          </div>
        </el-card>
        <el-card class="menu-item" @click="$router.push('/image-recognition')">
          <div class="card-content">
            <el-icon size="48" color="#67c23a"><Camera /></el-icon>
            <h3>图像识别</h3>
            <p>上传图片进行AI识别</p>
          </div>
        </el-card>
      </div>
    </el-main>
  </div>
</template>

<script>
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ChatDotRound, Camera } from '@element-plus/icons-vue'

export default {
  name: 'MenuView',
  components: {
    ChatDotRound,
    Camera
  },
  setup() {
    const router = useRouter()

    const handleLogout = async () => {
      try {
        await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        localStorage.removeItem('token')
        ElMessage.success('退出登录成功')
        router.push('/login')
      } catch {
        // 用户取消操作
      }
    }

    return {
      handleLogout
    }
  }
}
</script>

<style scoped>
.menu-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f8f9fa;
  padding: 24px;
  gap: 24px;
}

.header {
  background: #fff;
  border-radius: 20px;
  padding: 20px 32px;
  box-shadow: 0 12px 30px rgba(60, 64, 67, 0.15);
  display: flex;
  justify-content: space-between;
  align-items: center;
  border: 1px solid rgba(60, 64, 67, 0.08);
}

.header h1 {
  margin: 0;
  font-size: 26px;
  font-weight: 600;
  color: #202124;
}

.main {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
}

.menu-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 24px;
  width: 100%;
  max-width: 960px;
}

.menu-item {
  cursor: pointer;
  border-radius: 20px;
  border: 1px solid rgba(60, 64, 67, 0.12);
  box-shadow: 0 12px 30px rgba(32, 33, 36, 0.08);
  transition: transform 0.15s ease, box-shadow 0.15s ease, border-color 0.15s ease;
  background: #fff;
}

.menu-item:hover {
  transform: translateY(-4px);
  border-color: rgba(26, 115, 232, 0.4);
  box-shadow: 0 18px 36px rgba(26, 115, 232, 0.15);
}

.card-content {
  text-align: center;
  padding: 42px 32px;
}

.el-icon {
  display: block;
  margin: 0 auto 16px;
  color: #1a73e8;
}

.card-content h3 {
  margin: 0 0 8px 0;
  color: #202124;
  font-size: 22px;
  font-weight: 600;
}

.card-content p {
  margin: 0;
  color: #5f6368;
  font-size: 15px;
  line-height: 1.5;
}

@media (max-width: 768px) {
  .menu-container {
    padding: 16px;
  }

  .header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
}
</style>
