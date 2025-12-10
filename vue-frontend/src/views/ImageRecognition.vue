<template>
  <div class="image-recognition-container">
    <!-- 左侧会话列表 -->
    <div class="session-list">
      <div class="session-list-header">
        <span>图像识别</span>
      </div>
      <ul class="session-list-ul">
        <li class="session-item active">
          图像识别助手
        </li>
      </ul>
    </div>

    <!-- 右侧聊天区域 -->
    <div class="chat-section">
      <div class="top-bar">
        <button class="back-btn" @click="$router.push('/menu')">← 返回</button>
        <h2>AI 图像识别助手</h2>
      </div>

      <div class="chat-messages" ref="chatContainerRef">
        <div
          v-for="(message, index) in messages"
          :key="index"
          :class="['message', message.role === 'user' ? 'user-message' : 'ai-message']"
        >
          <div class="message-header">
            <b>{{ message.role === 'user' ? '你' : 'AI' }}:</b>
          </div>
          <div class="message-content">
            <span>{{ message.content }}</span>
            <img v-if="message.imageUrl" :src="message.imageUrl" alt="上传的图片" />
          </div>
        </div>
      </div>

      <div class="chat-input">
        <form @submit.prevent="handleSubmit">
          <input
            ref="fileInputRef"
            type="file"
            accept="image/*"
            required
            @change="handleFileSelect"
          />
          <button type="submit" :disabled="!selectedFile">发送图片</button>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, nextTick } from 'vue'
import api from '../utils/api'

export default {
  name: 'ImageRecognition',
  setup() {
    const messages = ref([])
    const selectedFile = ref(null)
    const fileInputRef = ref()
    const chatContainerRef = ref()

    const handleFileSelect = (event) => {
      selectedFile.value = event.target.files[0]
    }

    const handleSubmit = async () => {
      if (!selectedFile.value) return

      const file = selectedFile.value
      const imageUrl = URL.createObjectURL(file)

      // Add user message to UI
      messages.value.push({
        role: 'user',
        content: `已上传图片: ${file.name}`,
        imageUrl: imageUrl,
      })

      await nextTick()
      scrollToBottom()

      // Create FormData
      const formData = new FormData()
      formData.append('image', file)

      try {
        const response = await api.post('/image/recognize', formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        })


        if (response.data && response.data.class_name) {
             const aiText = `识别结果: ${response.data.class_name}`
            messages.value.push({
                role: 'assistant',
                content: aiText,
            })
        } else {
             messages.value.push({
                 role: 'assistant',
                 content: `[错误] ${response.data.status_msg || '识别失败'}`,
             })
        }
      } catch (error) {
        console.error('Upload error:', error)
        messages.value.push({
          role: 'assistant',
          content: `[错误] 无法连接到服务器或上传失败: ${error.message}`,
        })
      } finally {

        URL.revokeObjectURL(imageUrl)

            await nextTick()
        scrollToBottom()


        selectedFile.value = null
        if (fileInputRef.value) {
          fileInputRef.value.value = ''
        }
      }
    }

    const scrollToBottom = () => {
      if (chatContainerRef.value) {
        chatContainerRef.value.scrollTop = chatContainerRef.value.scrollHeight
      }
    }

    return {
      messages,
      selectedFile,
      fileInputRef,
      chatContainerRef,
      handleFileSelect,
      handleSubmit
    }
  }
}
</script>

<style scoped>
.image-recognition-container {
  height: 100vh;
  display: flex;
  background: #f8f9fa;
  color: #202124;
  font-family: 'Google Sans', 'Roboto', sans-serif;
}

.session-list {
  width: 280px;
  border-right: 1px solid rgba(60, 64, 67, 0.12);
  background: #fff;
  display: flex;
  flex-direction: column;
}

.session-list-header {
  padding: 24px;
  border-bottom: 1px solid rgba(60, 64, 67, 0.12);
  font-weight: 600;
}

.session-list-ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.session-item {
  padding: 16px 24px;
  color: #3c4043;
  border-bottom: 1px solid rgba(60, 64, 67, 0.08);
}

.session-item.active {
  color: #1a73e8;
  font-weight: 600;
  background: rgba(26, 115, 232, 0.08);
}

.chat-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #f8f9fa;
}

.top-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 28px;
  border-bottom: 1px solid rgba(60, 64, 67, 0.12);
  background: #fff;
}

.back-btn {
  border: 1px solid rgba(60, 64, 67, 0.16);
  border-radius: 999px;
  padding: 8px 16px;
  background: #fff;
  color: #1a73e8;
  font-weight: 600;
  cursor: pointer;
}

.top-bar h2 {
  margin: 0;
  font-size: 22px;
  color: #202124;
}

.chat-messages {
  flex: 1;
  padding: 32px;
  display: flex;
  flex-direction: column;
  gap: 18px;
  overflow-y: auto;
}

.chat-messages::-webkit-scrollbar {
  width: 6px;
}

.chat-messages::-webkit-scrollbar-thumb {
  background: rgba(60, 64, 67, 0.2);
  border-radius: 3px;
}

.message {
  max-width: 640px;
  padding: 16px 20px;
  border-radius: 18px;
  line-height: 1.6;
  box-shadow: 0 8px 20px rgba(60, 64, 67, 0.08);
  background: #fff;
}

.user-message {
  align-self: flex-end;
  background: #1a73e8;
  color: #fff;
}

.ai-message {
  align-self: flex-start;
  color: #202124;
}

.message-content {
  white-space: pre-wrap;
}

.message-content img {
  max-width: 280px;
  border-radius: 16px;
  margin-top: 12px;
  box-shadow: 0 8px 24px rgba(32, 33, 36, 0.18);
}

.chat-input {
  padding: 24px 28px;
  border-top: 1px solid rgba(60, 64, 67, 0.12);
  background: #fff;
}

.chat-input form {
  display: flex;
  gap: 16px;
  align-items: center;
}

.chat-input input[type='file'] {
  flex: 1;
  border: 1px dashed rgba(60, 64, 67, 0.3);
  border-radius: 16px;
  padding: 12px 16px;
  background: #f8f9fa;
  cursor: pointer;
}

.chat-input input[type='file']::file-selector-button {
  border: none;
  border-radius: 999px;
  padding: 8px 16px;
  background: #1a73e8;
  color: #fff;
  cursor: pointer;
}

.chat-input button {
  border: none;
  border-radius: 999px;
  padding: 12px 24px;
  background: #1a73e8;
  color: #fff;
  font-weight: 600;
  cursor: pointer;
}

.chat-input button:disabled {
  background: #dadce0;
  cursor: not-allowed;
}

@media (max-width: 960px) {
  .image-recognition-container {
    flex-direction: column;
  }

  .session-list {
    width: 100%;
    border-right: none;
    border-bottom: 1px solid rgba(60, 64, 67, 0.12);
  }
}
</style>
