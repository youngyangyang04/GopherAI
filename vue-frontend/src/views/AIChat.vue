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
          <div class="session-updated" v-if="session.updateAt">更新：{{ formatUpdateTime(session.updateAt) }}</div>
        </li>
      </ul>
    </div>

    <!-- 右侧聊天区域 -->
    <div class="chat-section">
      <div class="top-bar">
        <div class="top-actions">
          <button class="back-btn" @click="$router.push('/menu')">← 返回</button>
          <button class="sync-btn" @click="syncHistory" :disabled="!currentSessionId || tempSession">同步历史数据</button>
        </div>
        <div class="top-controls">
          <div class="select-group">
            <label for="modelType">选择模型</label>
            <select id="modelType" v-model="selectedModel" class="model-select">
              <option value="1">openai</option>
              <option value="2">ollama</option>
            </select>
          </div>
          <button
            type="button"
            class="chip-toggle"
            :class="{ active: isStreaming }"
            @click="isStreaming = !isStreaming"
          >
            <span class="chip-indicator google"></span>
            <span class="chip-text">
              <strong>流式响应</strong>
              <small>实时输出</small>
            </span>
          </button>
          <button
            type="button"
            class="chip-toggle"
            :class="{ active: isUsingGoogle }"
            @click="toggleGoogle"
          >
            <span class="chip-indicator"></span>
            <span class="chip-text">
              <strong>使用 Google</strong>
              <small>{{ isUsingGoogle ? '已启用' : '未启用' }}</small>
            </span>
          </button>
          <button
            type="button"
            class="chip-toggle"
            :class="{ active: isUsingRAG }"
            @click="toggleRAG"
          >
            <span class="chip-indicator rag"></span>
            <span class="chip-text">
              <strong>专家问诊</strong>
              <small>{{ isUsingRAG ? 'RAG 检索' : '默认模式' }}</small>
            </span>
          </button>
        </div>
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
          <div class="message-content">
            <MdPreview
              v-if="message.role === 'assistant'"
              :modelValue="message.content"
              previewTheme="github"
              :showCodeRowNumber="false"
            />
            <div v-else class="user-plain-text">{{ message.content }}</div>
          </div>
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
import { MdPreview } from 'md-editor-v3'
import api from '../utils/api'

export default {
  name: 'AIChat',
  components: {
    MdPreview
  },
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
    const isUsingGoogle = ref(false)
    const isUsingRAG = ref(false)

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

    const parseTimestamp = (value) => {
      if (!value) return 0
      const time = new Date(value).getTime()
      return Number.isNaN(time) ? 0 : time
    }

    const formatUpdateTime = (value) => {
      if (!value) return ''
      const date = new Date(value)
      if (Number.isNaN(date.getTime())) return value
      const pad = (num) => String(num).padStart(2, '0')
      return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
    }

    const touchSessionTimestamp = (sessionId, timestamp) => {
      if (!sessionId) return
      const sid = String(sessionId)
      if (!sessions.value[sid]) return
      sessions.value[sid].updateAt = timestamp || new Date().toISOString()
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

    const toggleGoogle = () => {
      if (!isUsingGoogle.value && isUsingRAG.value) {
        ElMessage.warning('Google 搜索和专家问诊不能同时启用，请先关闭专家问诊')
        return
      }
      isUsingGoogle.value = !isUsingGoogle.value
    }

    const toggleRAG = () => {
      if (!isUsingRAG.value && isUsingGoogle.value) {
        ElMessage.warning('Google 搜索和专家问诊不能同时启用，请先关闭 Google 搜索')
        return
      }
      isUsingRAG.value = !isUsingRAG.value
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
              updateAt: s.updateAt || s.updatedAt || '',
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
            sessions.value[normalizedId].updateAt = response.data.updateAt || new Date().toISOString()
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
          sessions.value[currentSessionId.value].updateAt = response.data.updateAt || new Date().toISOString()
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
        ? {
            question: question,
            modelType: selectedModel.value,
            usingGoogle: isUsingGoogle.value,
            usingRAG: isUsingRAG.value
          }
        : {
            question: question,
            modelType: selectedModel.value,
            sessionId: currentSessionId.value,
            usingGoogle: isUsingGoogle.value,
            usingRAG: isUsingRAG.value
          }

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
                        updateAt: parsed.updateAt || new Date().toISOString(),
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
        touchSessionTimestamp(currentSessionId.value)

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
          modelType: selectedModel.value,
          usingGoogle: isUsingGoogle.value,
          usingRAG: isUsingRAG.value
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
            updateAt: response.data.updateAt || new Date().toISOString(),
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
          sessionId: currentSessionId.value,
          usingGoogle: isUsingGoogle.value,
          usingRAG: isUsingRAG.value
        })
        if (response.data && response.data.status_code === 1000) {
          const aiMessage = { role: 'assistant', content: response.data.Information || '' }
          sessionMsgs.push(aiMessage)
          currentMessages.value = [...sessionMsgs]
          touchSessionTimestamp(currentSessionId.value, response.data.updateAt || new Date().toISOString())
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
      sessions: computed(() => {
        const list = Object.values(sessions.value)
        return list.sort((a, b) => parseTimestamp(b.updateAt) - parseTimestamp(a.updateAt))
      }),
      currentSessionId,
      tempSession,
      currentMessages,
      inputMessage,
      loading,
      messagesRef,
      messageInput,
      selectedModel,
      isStreaming,
      isUsingGoogle,
      isUsingRAG,
      canInteract,
      formatUpdateTime,
      playTTS,
      createNewSession,
      switchSession,
      syncHistory,
      sendMessage,
      toggleGoogle,
      toggleRAG
    }
  }
}
</script>

<style scoped>
.ai-chat-container {
  height: 100vh;
  display: flex;
  gap: 24px;
  padding: 32px;
  background: radial-gradient(circle at top, #f6f8ff 0%, #eef1fb 45%, #e4e7f1 100%);
  color: #1f2333;
  font-family: 'Google Sans', 'Roboto', 'PingFang SC', sans-serif;
  box-sizing: border-box;
}

.session-list {
  width: 280px;
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.85);
  border: 1px solid rgba(255, 255, 255, 0.7);
  box-shadow: 0 20px 45px rgba(71, 78, 114, 0.15);
  backdrop-filter: blur(16px);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.session-list-header {
  padding: 28px 24px 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  font-weight: 600;
}

.new-chat-btn {
  width: 100%;
  padding: 12px 0;
  border-radius: 18px;
  border: none;
  background: linear-gradient(120deg, #5f7afe, #7b8bff);
  color: #fff;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  box-shadow: 0 12px 24px rgba(95, 122, 254, 0.3);
}

.new-chat-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 18px 30px rgba(95, 122, 254, 0.35);
}

.session-list-ul {
  list-style: none;
  padding: 0;
  margin: 0;
  flex: 1;
  overflow-y: auto;
}

.session-item {
  padding: 16px 24px;
  cursor: pointer;
  color: #3b415b;
  transition: background 0.2s ease, color 0.2s ease;
  border-left: 4px solid transparent;
}

.session-name {
  font-weight: 600;
  font-size: 15px;
}

.session-model {
  font-size: 12px;
  color: #7a809c;
  margin-top: 6px;
}

.session-updated {
  margin-top: 4px;
  font-size: 11px;
  color: #9ca2c4;
}

.session-item.active {
  background: rgba(95, 122, 254, 0.12);
  color: #3d4ef7;
  border-left-color: #5f7afe;
}

.session-item:hover {
  background: rgba(95, 122, 254, 0.08);
}

.chat-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  border-radius: 32px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 25px 50px rgba(24, 32, 79, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.65);
  backdrop-filter: blur(20px);
}

.top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px 32px;
  border-bottom: 1px solid rgba(99, 110, 146, 0.08);
}

.top-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.top-controls {
  display: flex;
  align-items: center;
  gap: 20px;
}

.select-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 13px;
  color: #7d849f;
}

.select-group label {
  font-weight: 600;
}

.chip-toggle {
  border: none;
  border-radius: 20px;
  padding: 10px 14px;
  background: rgba(91, 101, 138, 0.08);
  color: #4c5170;
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  min-width: 140px;
  transition: background 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease;
}

.chip-toggle .chip-indicator {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: rgba(76, 81, 112, 0.4);
  position: relative;
  box-shadow: inset 0 0 0 2px rgba(76, 81, 112, 0.3);
}
.chip-toggle .chip-indicator.google {
  width: 12px;
  height: 12px;
  background: #c3c6d9;
  box-shadow: inset 0 0 0 2px rgba(255, 255, 255, 0.8);
}
.chip-toggle .chip-indicator.rag {
  background: #d7c3ff;
  box-shadow: inset 0 0 0 2px rgba(255, 255, 255, 0.8);
}

.chip-toggle .chip-text {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
}

.chip-toggle strong {
  font-size: 13px;
  font-weight: 600;
}

.chip-toggle small {
  font-size: 11px;
  color: #7b80a0;
}

.chip-toggle.active {
  background: rgba(95, 122, 254, 0.15);
  box-shadow: 0 10px 18px rgba(95, 122, 254, 0.2);
  transform: translateY(-1px);
}

.chip-toggle.active .chip-indicator {
  background: linear-gradient(135deg, #5f7afe, #7bb0ff);
  box-shadow: 0 0 8px rgba(95, 122, 254, 0.7);
}

.chip-toggle.active .chip-indicator.google {
  background: linear-gradient(135deg, #34a853, #fbbc04, #4285f4);
}
.chip-toggle.active .chip-indicator.rag {
  background: linear-gradient(135deg, #9c27b0, #e040fb);
}

.chip-toggle.active .chip-text small {
  color: #5260d6;
}

.back-btn,
.sync-btn {
  padding: 10px 18px;
  border-radius: 18px;
  border: none;
  background: rgba(93, 108, 231, 0.12);
  color: #3d4ef7;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.sync-btn {
  color: #fff;
  background: linear-gradient(120deg, #3d4ef7, #6a6ff9);
  box-shadow: 0 12px 24px rgba(61, 78, 247, 0.35);
}

.sync-btn:disabled {
  background: #cdd2f1;
  color: #fff;
  box-shadow: none;
  cursor: not-allowed;
}

.model-select {
  border-radius: 14px;
  border: 1px solid rgba(99, 110, 146, 0.25);
  padding: 6px 12px;
  background: #fff;
  color: #232742;
  font-weight: 600;
  outline: none;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 36px 40px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  background: linear-gradient(180deg, rgba(248, 249, 255, 0.9), rgba(240, 243, 255, 0.6));
}

.chat-messages::-webkit-scrollbar {
  width: 6px;
}

.chat-messages::-webkit-scrollbar-thumb {
  background: rgba(99, 110, 146, 0.3);
  border-radius: 3px;
}

.message {
  max-width: 700px;
  padding: 18px 22px;
  border-radius: 20px;
  line-height: 1.6;
  font-size: 15px;
  box-shadow: 0 16px 35px rgba(46, 57, 107, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.6);
  backdrop-filter: blur(5px);
}

.user-message {
  align-self: flex-end;
  background: linear-gradient(135deg, #4d7cfe, #6ac2ff);
  color: #fff;
}

.ai-message {
  align-self: flex-start;
  background: rgba(255, 255, 255, 0.95);
  color: #202437;
}

.message-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
  font-size: 13px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: rgba(32, 36, 55, 0.7);
}

.message-header b {
  font-size: 12px;
  font-weight: 700;
}

.message-content {
  white-space: normal;
  word-break: break-word;
}

.message-content :deep(.md-editor-preview) {
  background: transparent;
  padding: 0;
}

.message-content :deep(pre) {
  background: #101736;
  color: #f4f4ff;
  border-radius: 14px;
  padding: 14px;
  overflow-x: auto;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.user-plain-text {
  white-space: pre-wrap;
  word-break: break-word;
}

.streaming-indicator {
  color: #1ac0ff;
  font-weight: 600;
}

.tts-btn {
  border: none;
  border-radius: 12px;
  background: rgba(0, 0, 0, 0.08);
  color: inherit;
  padding: 4px 10px;
  cursor: pointer;
}

.chat-input {
  padding: 24px 32px 32px;
  background: transparent;
  display: flex;
  gap: 16px;
  align-items: flex-end;
  position: relative;
}

.chat-input textarea {
  flex: 1;
  border-radius: 18px;
  border: 1px solid rgba(108, 121, 170, 0.35);
  padding: 16px 18px;
  resize: none;
  min-height: 54px;
  max-height: 180px;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: inset 0 2px 8px rgba(23, 32, 90, 0.08);
}

.chat-input textarea:focus {
  border-color: #5f7afe;
  box-shadow: 0 0 0 3px rgba(95, 122, 254, 0.25);
  outline: none;
}

.send-btn {
  border-radius: 16px;
  border: none;
  padding: 14px 32px;
  background: linear-gradient(120deg, #5f7afe, #7bb0ff);
  color: #fff;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  box-shadow: 0 15px 30px rgba(87, 116, 255, 0.35);
}

.send-btn:disabled {
  background: #cdd2f1;
  box-shadow: none;
  cursor: not-allowed;
}

.send-btn:not(:disabled):hover {
  transform: translateY(-2px);
}

.chat-input-hint {
  position: absolute;
  top: 6px;
  left: 40px;
  color: #ea4335;
  font-size: 13px;
}

@media (max-width: 1100px) {
  .ai-chat-container {
    flex-direction: column;
    height: auto;
  }

  .session-list {
    width: 100%;
    flex-direction: column;
  }

  .chat-section {
    width: 100%;
  }

  .top-bar {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .top-controls {
    width: 100%;
    flex-wrap: wrap;
  }
}
</style>
