<template>
  <div class="app-layout">
    <aside class="sidebar">
      <div class="logo">
        <el-icon :size="24"><Monitor /></el-icon>
        <span>Sing Box</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        router
        class="sidebar-menu"
      >
        <el-menu-item index="/">
          <el-icon><Odometer /></el-icon>
          <span>监控面板</span>
        </el-menu-item>
        <el-menu-item index="/kernel">
          <el-icon><Box /></el-icon>
          <span>内核管理</span>
        </el-menu-item>
        <el-menu-item index="/singbox">
          <el-icon><Connection /></el-icon>
          <span>Sing-Box 配置</span>
        </el-menu-item>
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <span>系统设置</span>
        </el-menu-item>
      </el-menu>
    </aside>
    <main class="app-main">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useClashStream } from './composables/useClashStream'
import { Monitor, Setting, Connection, Odometer, Box } from '@element-plus/icons-vue'

const route = useRoute()
const { startTrafficStream, startConnectionsStream } = useClashStream()

const activeMenu = computed(() => {
  const path = route.path
  if (path.startsWith('/singbox')) return '/singbox'
  if (path.startsWith('/settings')) return '/settings'
  if (path.startsWith('/kernel')) return '/kernel'
  return '/'
})

// Initialize WebSocket streams globally on app start
onMounted(() => {
  startTrafficStream()
  startConnectionsStream()
})
</script>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
}

.sidebar {
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  width: 200px;
  background: #1d1e1f;
  display: flex;
  flex-direction: column;
  z-index: 100;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px;
  color: white;
  font-size: 18px;
  font-weight: 600;
  border-bottom: 1px solid #333;
}

.sidebar-menu {
  border-right: none;
  background: transparent;
}

.sidebar-menu .el-menu-item {
  color: rgba(255, 255, 255, 0.65);
}

.sidebar-menu .el-menu-item:hover {
  background: rgba(255, 255, 255, 0.08);
  color: white;
}

.sidebar-menu .el-menu-item.is-active {
  background: #409eff;
  color: white;
}

.app-main {
  flex: 1;
  margin-left: 200px;
  background: #f5f7fa;
  min-height: 100vh;
  padding: 24px;
}
</style>
