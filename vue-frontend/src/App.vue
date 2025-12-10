<template>
  <div id="app">
    <router-view v-slot="{ Component }">
      <transition name="page" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
  </div>
</template>

<script>
export default {
  name: 'App'
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body {
  height: 100%;
  font-family: 'Google Sans', 'Roboto', 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', '微软雅黑', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  background: #f8f9fa;
  color: #1f1f1f;
}

#app {
  height: 100%;
  background: #f8f9fa;
}

/* 页面切换动画 */
.page-enter-active,
.page-leave-active {
  transition: all 0.4s cubic-bezier(0.55, 0, 0.1, 1);
}

.page-enter-from {
  opacity: 0;
  transform: translateX(30px);
}

.page-leave-to {
  opacity: 0;
  transform: translateX(-30px);
}

/* 全局滚动条样式 */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.05);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb {
  background: rgba(102, 126, 234, 0.3);
  border-radius: 4px;
  transition: background 0.3s ease;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(102, 126, 234, 0.5);
}

/* Element Plus 组件样式覆盖 */
.el-button {
  font-weight: 500;
  border-radius: 10px;
  transition: all 0.2s ease;
}

.el-button--primary {
  background-color: #1a73e8;
  border-color: #1a73e8;
  color: #fff;
}

.el-button--primary:hover {
  background-color: #1967d2;
  border-color: #1967d2;
  box-shadow: 0 8px 20px rgba(25, 103, 210, 0.25);
}

.el-button--text {
  color: #1a73e8;
}

.el-button--text:hover {
  color: #174ea6;
}

.el-input {
  border-radius: 10px;
  border-color: rgba(32, 33, 36, 0.12);
}

.el-card {
  border-radius: 16px;
  border-color: rgba(32, 33, 36, 0.08);
  box-shadow: 0 6px 18px rgba(32, 33, 36, 0.08);
}

.el-message {
  border-radius: 10px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .page-enter-from,
  .page-leave-to {
    transform: translateX(0);
    opacity: 0;
  }

  .page-enter-active,
  .page-leave-active {
    transition: opacity 0.3s ease;
  }
}
</style>
