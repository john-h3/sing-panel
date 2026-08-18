<template>
  <div class="app-layout">
    <aside class="sidebar">
      <div class="logo">
        <el-icon :size="24"><Monitor /></el-icon>
        <span>Sing Panel</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :default-openeds="openeds"
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
        <el-sub-menu index="/singbox">
          <template #title>
            <el-icon><Connection /></el-icon>
            <span>Sing-Box 配置</span>
          </template>
          <el-menu-item index="/singbox/inbounds">Inbound 配置</el-menu-item>
          <el-menu-item index="/singbox/outbounds">Outbound 配置</el-menu-item>
          <el-menu-item index="/singbox/rulesets">Ruleset 配置</el-menu-item>
          <el-menu-item index="/singbox/route-rules">路由规则</el-menu-item>
          <el-menu-item index="/singbox/dns">DNS 配置</el-menu-item>
          <el-menu-item index="/singbox/services">服务配置</el-menu-item>
          <el-menu-item index="/singbox/http-clients">HTTP 客户端</el-menu-item>
          <el-menu-item index="/singbox/experimental">Experimental</el-menu-item>
        </el-sub-menu>
        <el-sub-menu index="/settings">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>系统设置</span>
          </template>
          <el-menu-item index="/settings/accelerate">加速域名</el-menu-item>
          <el-menu-item index="/settings/dashboard">Dashboard</el-menu-item>
          <el-menu-item index="/settings/data">数据管理</el-menu-item>
          <el-menu-item index="/settings/instances">多实例管理</el-menu-item>
        </el-sub-menu>
      </el-menu>
      <div class="sidebar-actions">
        <el-tooltip content="重启 sing-panel 服务" placement="top">
          <button
            class="action-btn"
            :disabled="actionLoading"
            @click="confirmRestartService"
          >
            <el-icon :size="16"><RefreshRight /></el-icon>
          </button>
        </el-tooltip>
        <el-tooltip content="重启机器（操作系统）" placement="top">
          <button
            class="action-btn danger"
            :disabled="actionLoading"
            @click="confirmRebootMachine"
          >
            <el-icon :size="16"><SwitchButton /></el-icon>
          </button>
        </el-tooltip>
      </div>
      <div class="theme-toggle">
        <span
          class="theme-icon"
          :class="{ active: theme === 'light' }"
          @click="setTheme('light')"
        >
          <el-icon :size="16"><Sunny /></el-icon>
        </span>
        <span
          class="theme-icon"
          :class="{ active: theme === 'dark' }"
          @click="setTheme('dark')"
        >
          <el-icon :size="16"><Moon /></el-icon>
        </span>
        <span
          class="theme-icon"
          :class="{ active: theme === 'system' }"
          @click="setTheme('system')"
        >
          <el-icon :size="16"><Monitor /></el-icon>
        </span>
      </div>
    </aside>
    <main class="app-main">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import axios from 'axios'
import { useTheme } from './composables/useTheme'
import { systemApi } from './api/system'
import {
  Monitor,
  Setting,
  Connection,
  Odometer,
  Box,
  Sunny,
  Moon,
  RefreshRight,
  SwitchButton
} from '@element-plus/icons-vue'

const route = useRoute()
const { theme, setTheme } = useTheme()

const activeMenu = computed(() => {
  return route.path
})

const openeds = computed(() => {
  const path = route.path
  const opens = []
  if (path.startsWith('/singbox')) opens.push('/singbox')
  if (path.startsWith('/settings')) opens.push('/settings')
  return opens
})

const actionLoading = ref(false)

// waitForServer polls the panel until it responds again, then re-enables
// the action buttons. Used after service/machine restart so the UI recovers
// automatically without a manual page refresh.
const waitForServer = (onBack) => {
  let attempts = 0
  const maxAttempts = 600 // ~10 min at 1s intervals
  const poll = async () => {
    attempts++
    try {
      await axios.get('/api/system/init', { timeout: 3000 })
      actionLoading.value = false
      onBack(true)
    } catch (e) {
      // Any HTTP response means the server is back up.
      if (e.response) {
        actionLoading.value = false
        onBack(true)
        return
      }
      if (attempts > maxAttempts) {
        actionLoading.value = false
        ElMessage.warning('等待服务上线超时，请手动刷新页面')
        return
      }
      setTimeout(poll, 1000)
    }
  }
  // Delay the first probe so a restarting/killed service doesn't answer
  // before it actually goes down.
  setTimeout(poll, 2000)
}

const confirmRestartService = () => {
  ElMessageBox.confirm(
    '确定要重启 sing-panel 服务吗？重启期间面板将短暂不可用。',
    '重启服务',
    {
      confirmButtonText: '确定重启',
      cancelButtonText: '取消',
      type: 'warning'
    }
  )
    .then(async () => {
      actionLoading.value = true
      try {
        await systemApi.restartService()
        ElMessage.success('sing-panel 服务正在重启，请稍候...')
        waitForServer((ok) => {
          if (ok) ElMessage.success('sing-panel 服务已恢复在线')
        })
      } catch (e) {
        ElMessage.error(e.response?.data?.error || '重启服务失败')
        actionLoading.value = false
      }
    })
    .catch(() => {})
}

const confirmRebootMachine = () => {
  ElMessageBox.confirm(
    '确定要重启机器（操作系统）吗？机器重启期间所有服务都会中断，请谨慎操作。',
    '重启机器',
    {
      confirmButtonText: '确定重启',
      cancelButtonText: '取消',
      type: 'error',
      confirmButtonClass: 'el-button--danger'
    }
  )
    .then(async () => {
      actionLoading.value = true
      try {
        await systemApi.rebootMachine()
        ElMessage.success('机器正在重启，请稍后重新连接...')
        waitForServer((ok) => {
          if (ok) ElMessage.success('机器已恢复在线')
        })
      } catch (e) {
        ElMessage.error(e.response?.data?.error || '重启机器失败')
        actionLoading.value = false
      }
    })
    .catch(() => {})
}
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
  background: var(--sidebar-bg);
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
  border-bottom: 1px solid var(--sidebar-border);
}

.sidebar-menu {
  border-right: none;
  background: transparent;
}

.sidebar-menu .el-menu-item,
.sidebar-menu :deep(.el-sub-menu__title) {
  color: rgba(255, 255, 255, 0.65);
}

.sidebar-menu .el-menu-item:hover,
.sidebar-menu :deep(.el-sub-menu__title:hover) {
  background: rgba(255, 255, 255, 0.08);
  color: white;
}

.sidebar-menu .el-menu-item.is-active {
  background: #409eff;
  color: white;
}

.sidebar-menu :deep(.el-sub-menu .el-menu) {
  background: rgba(0, 0, 0, 0.15);
}

.sidebar-menu :deep(.el-sub-menu .el-menu .el-menu-item) {
  padding-left: 52px !important;
  min-width: auto;
}

.sidebar-menu :deep(.el-sub-menu .el-menu .el-menu-item:hover) {
  background: rgba(255, 255, 255, 0.08);
}

.sidebar-menu :deep(.el-sub-menu .el-menu .el-menu-item.is-active) {
  background: #409eff;
  color: white;
}

.theme-toggle {
  padding: 12px 16px;
  display: flex;
  justify-content: center;
  gap: 8px;
}

.sidebar-actions {
  margin-top: auto;
  padding: 12px 16px;
  border-top: 1px solid var(--sidebar-border);
  display: flex;
  justify-content: center;
  gap: 12px;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 6px;
  border: none;
  background: rgba(255, 255, 255, 0.06);
  color: rgba(255, 255, 255, 0.7);
  cursor: pointer;
  transition: all 0.2s;
}

.action-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
}

.action-btn.danger:hover:not(:disabled) {
  background: #f56c6c;
  color: #fff;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.theme-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 6px;
  color: rgba(255, 255, 255, 0.45);
  cursor: pointer;
  transition: all 0.2s;
}

.theme-icon:hover {
  background: rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.85);
}

.theme-icon.active {
  background: #409eff;
  color: white;
}

.app-main {
  flex: 1;
  margin-left: 200px;
  background: var(--bg-page);
  min-height: 100vh;
  padding: 24px;
}
</style>
