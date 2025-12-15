<template>
  <div class="ai-chat-container">
    <!-- 左侧会话列表 -->
    <div class="session-list">
      <div class="session-list-header">
        <span>会话列表</span>
        <button class="new-chat-btn" @click="createNewSession">＋ 新聊天</button>
      </div>
      <ul class="session-list-ul">
        <li
          v-for="session in sessions"
          :key="session.id"
          :class="['session-item', { active: currentSessionId === session.id }]"
          @click="switchSession(session.id)"
        >
          <div class="session-name">{{ session.name || `会话 ${session.id}` }}</div>
          <div class="session-model" v-if="session.modelType">模型：{{ session.modelType }}</div>
        </li>
      </ul>
    </div>

    <!-- 右侧聊天区域 -->
    <div class="chat-section">
      <div class="top-bar">
        <button class="back-btn" @click="$router.push('/menu')">← 返回</button>
        <button class="sync-btn" @click="syncHistory" :disabled="!currentSessionId || tempSession">同步历史数据</button>
        <label for="modelType">选择模型：</label>
        <select id="modelType" v-model="selectedModel" class="model-select">
          <option value="1">openai</option>
          <option value="2">ollama</option>
        </select>
        <label for="streamingMode" style="margin-left: 20px;">
          <input type="checkbox" id="streamingMode" v-model="isStreaming" />
          流式响应
        </label>
      </div>

      <div class="chat-messages" ref="messagesRef">
        <div
          v-for="(message, index) in currentMessages"
          :key="index"
          :class="['message', message.role === 'user' ? 'user-message' : 'ai-message']"
        >
          <div class="message-header">
            <b>{{ message.role === 'user' ? '你' : 'AI' }}:</b>
            <button v-if="message.role === 'assistant'" class="tts-btn" @click="playTTS(message.content)">🔊</button>
            <span v-if="message.meta && message.meta.status === 'streaming'" class="streaming-indicator"> ··</span>
          </div>
          <div class="message-content" v-html="renderMarkdown(message.content)"></div>
        </div>
      </div>

      <div class="chat-input">
        <div class="chat-input-hint" v-if="!canInteract">
          请点击“新聊天”或选择历史会话后再输入
        </div>
        <textarea
          v-model="inputMessage"
          placeholder="请输入你的问题..."
          @keydown.enter.exact.prevent="sendMessage"
          :disabled="loading || !canInteract"
          ref="messageInput"
          rows="1"
        ></textarea>
        <button
          type="button"
          :disabled="!inputMessage.trim() || loading || !canInteract"
          @click="sendMessage"
          class="send-btn"
        >
          {{ loading ? '发送中...' : '发送' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script>


import { ref, nextTick, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../utils/api'

export default {
  name: 'AIChat',
  setup() {

    const sessions = ref({})               
    const currentSessionId = ref(null)    
    const tempSession = ref(false)        
    const currentMessages = ref([])      
    const inputMessage = ref('')
    const loading = ref(false)
    const messagesRef = ref(null)
    const messageInput = ref(null)
    const selectedModel = ref('1')
    const isStreaming = ref(false)

    const modelValueToLabel = (value) => {
      const normalized = String(value ?? '').toLowerCase()
      if (normalized === '1' || normalized === 'openai') return 'openai'
      if (normalized === '2' || normalized === 'ollama') return 'ollama'
      return normalized || ''
    }

    const modelLabelToValue = (label) => {
      const normalized = String(label ?? '').toLowerCase()
      if (!normalized) return selectedModel.value
      if (normalized === 'openai' || normalized === '1') return '1'
      if (normalized === 'ollama' || normalized === '2') return '2'
      return String(label)
    }


    const renderMarkdown = (text) => {
      if (!text && text !== '') return ''
      return String(text)
        .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
        .replace(/\*(.*?)\*/g, '<em>$1</em>')
        .replace(/`(.*?)`/g, '<code>$1</code>')
        .replace(/\n/g, '<br>')
    }

    const playTTS = async (text) => {
      try {
        const response = await api.post('/chat/tts', { text })
        if (response.data && response.data.status_code === 1000 && response.data.url) {
          const audio = new Audio(response.data.url)
          audio.play()
        } else {
          ElMessage.error('无法获取语音')
        }
      } catch (error) {
        console.error('TTS error:', error)
        ElMessage.error('请求语音接口失败')
      }
    }

    const loadSessions = async () => {
      try {
        const response = await api.get('/AI/chat/sessions')
        if (response.data && response.data.status_code === 1000 && Array.isArray(response.data.sessions)) {
          const sessionMap = {}
          response.data.sessions.forEach(s => {
            const sid = String(s.sessionId)
            sessionMap[sid] = {
              id: sid,
              name: s.name || `会话 ${sid}`,
              modelType: s.modelType || '',
              messages: [] // lazy load
            }
          })
          sessions.value = sessionMap
        }
      } catch (error) {
        console.error('Load sessions error:', error)
      }
    }

    const createNewSession = () => {
      currentSessionId.value = 'temp'
      tempSession.value = true
      currentMessages.value = []
      // focus input
      nextTick(() => {
        if (messageInput.value) messageInput.value.focus()
      })
    }

    const switchSession = async (sessionId) => {
      if (!sessionId) return
      const normalizedId = String(sessionId)
      currentSessionId.value = normalizedId
      tempSession.value = false

      const sessionData = sessions.value[normalizedId]
      if (!sessionData) return
      if (sessionData.modelType) {
        selectedModel.value = modelLabelToValue(sessionData.modelType)
      }

      // lazy load history if not present
      if (!sessionData.messages || sessionData.messages.length === 0) {
        try {
          const response = await api.post('/AI/chat/history', { sessionId: currentSessionId.value })
          if (response.data && response.data.status_code === 1000 && Array.isArray(response.data.history)) {
            const messages = response.data.history.map(item => ({
              role: item.is_user ? 'user' : 'assistant',
              content: item.content
            }))
            sessions.value[normalizedId].messages = messages
          }
        } catch (err) {
          console.error('Load history error:', err)
        }
      }


      currentMessages.value = [...(sessions.value[normalizedId].messages || [])]
      await nextTick()
      scrollToBottom()
    }

    const syncHistory = async () => {
      if (!currentSessionId.value || tempSession.value) {
        ElMessage.warning('请选择已有会话进行同步')
        return
      }
      try {
        const response = await api.post('/AI/chat/history', { sessionId: currentSessionId.value })
        if (response.data && response.data.status_code === 1000 && Array.isArray(response.data.history)) {
          const messages = response.data.history.map(item => ({
            role: item.is_user ? 'user' : 'assistant',
            content: item.content
          }))
          sessions.value[currentSessionId.value].messages = messages
          currentMessages.value = [...messages]
          await nextTick()
          scrollToBottom()
        } else {
          ElMessage.error('无法获取历史数据')
        }
      } catch (err) {
        console.error('Sync history error:', err)
        ElMessage.error('请求历史数据失败')
      }
    }


    const sendMessage = async () => {
      if (!tempSession.value && !currentSessionId.value) {
        ElMessage.warning('请先新建或选择会话')
        return
      }
      if (!inputMessage.value || !inputMessage.value.trim()) {
        ElMessage.warning('请输入消息内容')
        return
      }

      const userMessage = {
        role: 'user',
        content: inputMessage.value
      }
      const currentInput = inputMessage.value
      inputMessage.value = ''


      currentMessages.value.push(userMessage)
      await nextTick()
      scrollToBottom()

      try {
        loading.value = true
        if (isStreaming.value) {

          await handleStreaming(currentInput)
        } else {

          await handleNormal(currentInput)
        }
      } catch (err) {
        console.error('Send message error:', err)
        ElMessage.error('发送失败，请重试')

        if (!tempSession.value && currentSessionId.value && sessions.value[currentSessionId.value] && sessions.value[currentSessionId.value].messages) {

          const sessionArr = sessions.value[currentSessionId.value].messages
          if (sessionArr && sessionArr.length) sessionArr.pop()
        }
        currentMessages.value.pop()
      } finally {
        if (!isStreaming.value) {
          loading.value = false
        }
        await nextTick()
        scrollToBottom()
      }
    }


    async function handleStreaming(question) {

      const aiMessage = {
        role: 'assistant',
        content: '',
        meta: { status: 'streaming' } // mark streaming
      }


      const aiMessageIndex = currentMessages.value.length
      currentMessages.value.push(aiMessage)

      if (!tempSession.value && currentSessionId.value && sessions.value[currentSessionId.value]) {
        if (!sessions.value[currentSessionId.value].messages) sessions.value[currentSessionId.value].messages = []
        sessions.value[currentSessionId.value].messages.push({ role: 'assistant', content: '' })
      }


      const url = tempSession.value
        ? '/api/AI/chat/send-stream-new-session'  
        : '/api/AI/chat/send-stream'           

      const headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
      }

      const body = tempSession.value
        ? { question: question, modelType: selectedModel.value }
        : { question: question, modelType: selectedModel.value, sessionId: currentSessionId.value }

      try {
        // 创建 fetch 连接读取 SSE 流
        const response = await fetch(url, {
          method: 'POST',
          headers,
          body: JSON.stringify(body)
        })

        if (!response.ok) {
          loading.value = false
          throw new Error('Network response was not ok')
        }

        const reader = response.body.getReader()
        const decoder = new TextDecoder()
        let buffer = ''

        // 读取流数据
        // eslint-disable-next-line no-constant-condition
        while (true) {
          const { done, value } = await reader.read()
          if (done) break

          const chunk = decoder.decode(value, { stream: true })
          buffer += chunk

          // 按行分割
          const lines = buffer.split('\n')
          buffer = lines.pop() || '' // 保留未完成的行

          for (const line of lines) {
            const trimmedLine = line.trim()
            if (!trimmedLine) continue

            // 处理 SSE 格式：data: <content>
            if (trimmedLine.startsWith('data:')) {
              const data = trimmedLine.slice(5).trim()
              console.log('[SSE] Received:', data) // 调试日志

              if (data === '[DONE]') {
                // 流结束
                console.log('[SSE] Stream done')
                loading.value = false
                currentMessages.value[aiMessageIndex].meta = { status: 'done' }
                currentMessages.value = [...currentMessages.value]
              } else if (data.startsWith('{')) {
                // 尝试解析 JSON（如 sessionId）
                try {
                  const parsed = JSON.parse(data)
                  if (parsed.sessionId) {
                    const newSid = String(parsed.sessionId)
                    console.log('[SSE] Session ID:', newSid)
                    if (tempSession.value) {
                      sessions.value[newSid] = {
                        id: newSid,
                        name: '新会话',
                        modelType: parsed.modelType || modelValueToLabel(selectedModel.value),
                        messages: [...currentMessages.value]
                      }
                      currentSessionId.value = newSid
                      tempSession.value = false
                    }
                  }
                } catch (e) {
                  // 不是 JSON，当作普通文本处理
                  currentMessages.value[aiMessageIndex].content += data
                  console.log('[SSE] Content updated:', currentMessages.value[aiMessageIndex].content.length)
                }
              } else {
                // 普通文本数据，直接追加
                // 使用数组索引直接更新，强制 Vue 响应式系统检测变化
                currentMessages.value[aiMessageIndex].content += data
                console.log('[SSE] Content updated:', currentMessages.value[aiMessageIndex].content.length)
              }

              // 每收到一条数据就立即更新 DOM
              // 强制更新整个数组以触发响应式
              currentMessages.value = [...currentMessages.value]
              
              // 使用 requestAnimationFrame 强制浏览器重排
              await new Promise(resolve => {
                requestAnimationFrame(() => {
                  scrollToBottom()
                  resolve()
                })
              })
            }
          }
        }

        // 流读取完成后的处理
        loading.value = false
        currentMessages.value[aiMessageIndex].meta = { status: 'done' }
        currentMessages.value = [...currentMessages.value]

        // 同步到 sessions 存储
        if (!tempSession.value && currentSessionId.value && sessions.value[currentSessionId.value]) {
          const sessMsgs = sessions.value[currentSessionId.value].messages
          if (Array.isArray(sessMsgs) && sessMsgs.length) {
            const lastIndex = sessMsgs.length - 1
            if (sessMsgs[lastIndex] && sessMsgs[lastIndex].role === 'assistant') {
              sessMsgs[lastIndex].content = currentMessages.value[aiMessageIndex].content
            }
          }
        }
      } catch (err) {
        console.error('Stream error:', err)
        loading.value = false
        currentMessages.value[aiMessageIndex].meta = { status: 'error' }
        currentMessages.value = [...currentMessages.value]
        ElMessage.error('流式传输出错')
      }
    }


    async function handleNormal(question) {
      if (tempSession.value) {

        const response = await api.post('/AI/chat/send-new-session', {
          question: question,
          modelType: selectedModel.value
        })
        if (response.data && response.data.status_code === 1000) {
          const sessionId = String(response.data.sessionId)
          const aiMessage = {
            role: 'assistant',
            content: response.data.Information || ''
          }

          sessions.value[sessionId] = {
            id: sessionId,
            name: '新会话',
            modelType: response.data.modelType || modelValueToLabel(selectedModel.value),
            messages: [ { role: 'user', content: question }, aiMessage ]
          }
          currentSessionId.value = sessionId
          tempSession.value = false
          currentMessages.value = [...sessions.value[sessionId].messages]
        } else {
          ElMessage.error(response.data?.status_msg || '发送失败')

          currentMessages.value.pop()
        }
      } else {

        const sessionMsgs = sessions.value[currentSessionId.value].messages

        sessionMsgs.push({ role: 'user', content: question })

        const response = await api.post('/AI/chat/send', {
          question: question,
          modelType: selectedModel.value,
          sessionId: currentSessionId.value
        })
        if (response.data && response.data.status_code === 1000) {
          const aiMessage = { role: 'assistant', content: response.data.Information || '' }
          sessionMsgs.push(aiMessage)
          currentMessages.value = [...sessionMsgs]
        } else {
          ElMessage.error(response.data?.status_msg || '发送失败')
          sessionMsgs.pop() // rollback
          currentMessages.value.pop()
        }
      }
    }


    const scrollToBottom = () => {
      if (messagesRef.value) {
        try {
          messagesRef.value.scrollTop = messagesRef.value.scrollHeight
        } catch (e) {
          // ignore
        }
      }
    }

    onMounted(() => {
      loadSessions()
    })

    const canInteract = computed(() => tempSession.value || !!currentSessionId.value)

    // expose to template
    return {
      sessions: computed(() => Object.values(sessions.value)),
      currentSessionId,
      tempSession,
      currentMessages,
      inputMessage,
      loading,
      messagesRef,
      messageInput,
      selectedModel,
      isStreaming,
      canInteract,
      renderMarkdown,
      playTTS,
      createNewSession,
      switchSession,
      syncHistory,
      sendMessage
    }
  }
}
</script>

<style scoped>
.ai-chat-container {
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
  height: 100%;
  overflow: hidden;
}

.session-list-header {
  padding: 24px;
  border-bottom: 1px solid rgba(60, 64, 67, 0.12);
  display: flex;
  flex-direction: column;
  gap: 12px;
  font-weight: 600;
}

.new-chat-btn {
  width: 100%;
  padding: 12px 0;
  border-radius: 20px;
  border: 1px solid rgba(60, 64, 67, 0.16);
  background: #fff;
  color: #1a73e8;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.new-chat-btn:hover {
  background: rgba(26, 115, 232, 0.08);
}

.session-list-ul {
  list-style: none;
  padding: 0;
  margin: 0;
  flex: 1;
  overflow-y: auto;
}

.session-item {
  padding: 14px 24px;
  cursor: pointer;
  border-bottom: 1px solid rgba(60, 64, 67, 0.08);
  color: #3c4043;
  transition: background 0.2s ease;
}

.session-name {
  font-weight: 600;
}

.session-model {
  font-size: 12px;
  color: #5f6368;
  margin-top: 4px;
}

.session-item.active {
  background: rgba(26, 115, 232, 0.1);
  color: #1a73e8;
  font-weight: 600;
}

.session-item:hover {
  background: rgba(26, 115, 232, 0.08);
}

.chat-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  background: #f8f9fa;
}

.top-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 24px;
  border-bottom: 1px solid rgba(60, 64, 67, 0.12);
  background: #fff;
}

.back-btn,
.sync-btn {
  padding: 8px 16px;
  border-radius: 999px;
  border: 1px solid rgba(60, 64, 67, 0.16);
  background: #fff;
  color: #1a73e8;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s ease;
}

.sync-btn {
  color: #fff;
  background: #1a73e8;
  border-color: #1a73e8;
}

.sync-btn:disabled {
  background: #dadce0;
  border-color: #dadce0;
  color: #fff;
  cursor: not-allowed;
}

.model-select {
  border-radius: 999px;
  border: 1px solid rgba(60, 64, 67, 0.16);
  padding: 6px 12px;
  background: #fff;
  color: #202124;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 32px;
  display: flex;
  flex-direction: column;
  gap: 18px;
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
  font-size: 15px;
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
  background: #fff;
  color: #202124;
}

.message-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.message-content {
  white-space: pre-wrap;
  word-break: break-word;
}

.streaming-indicator {
  color: #5f6368;
  font-weight: 600;
}

.tts-btn {
  border: none;
  border-radius: 12px;
  background: rgba(26, 115, 232, 0.1);
  color: #1a73e8;
  padding: 4px 10px;
  cursor: pointer;
}

.chat-input {
  padding: 24px;
  border-top: 1px solid rgba(60, 64, 67, 0.12);
  background: #fff;
  display: flex;
  gap: 16px;
  align-items: flex-end;
  position: relative;
}

.chat-input textarea {
  flex: 1;
  border-radius: 16px;
  border: 1px solid rgba(60, 64, 67, 0.24);
  padding: 14px 16px;
  resize: none;
  min-height: 48px;
  max-height: 180px;
}

.chat-input textarea:focus {
  border-color: #1a73e8;
  box-shadow: 0 0 0 3px rgba(26, 115, 232, 0.15);
  outline: none;
}

.send-btn {
  border-radius: 999px;
  border: none;
  padding: 12px 28px;
  background: #1a73e8;
  color: #fff;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s ease;
}

.send-btn:disabled {
  background: #dadce0;
  cursor: not-allowed;
}

.chat-input-hint {
  position: absolute;
  top: 8px;
  left: 24px;
  color: #ea4335;
  font-size: 13px;
}

@media (max-width: 960px) {
  .ai-chat-container {
    flex-direction: column;
  }

  .session-list {
    width: 100%;
    height: 220px;
    border-right: none;
    border-bottom: 1px solid rgba(60, 64, 67, 0.12);
  }

  .chat-section {
    height: calc(100vh - 220px);
  }
}
</style>
