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
        </el-sub-menu>
      </el-menu>
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
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useTheme } from './composables/useTheme'
import { Monitor, Setting, Connection, Odometer, Box, Sunny, Moon } from '@element-plus/icons-vue'

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
  margin-top: auto;
  padding: 12px 16px;
  border-top: 1px solid var(--sidebar-border);
  display: flex;
  justify-content: center;
  gap: 8px;
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
