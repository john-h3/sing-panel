<template>
  <div class="system-page">
    <el-row :gutter="24">
      <el-col :span="24">
        <h2 class="page-title">系统管理</h2>
      </el-col>
    </el-row>

    <!-- Overview -->
    <el-row :gutter="24" class="overview-row">
      <el-col :span="8">
        <el-card class="status-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <el-icon :size="20"><Box /></el-icon>
              <span>内核状态</span>
            </div>
          </template>
          <div class="status-content">
            <div class="status-row">
              <el-tag :type="status.installed ? 'success' : 'info'" size="large">
                {{ status.installed ? '已安装' : '未安装' }}
              </el-tag>
            </div>
            <div v-if="status.installed" class="status-info">
              <p><strong>版本:</strong> {{ status.version || '未知' }}</p>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card class="status-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <el-icon :size="20"><InfoFilled /></el-icon>
              <span>系统信息</span>
            </div>
          </template>
          <div class="system-info">
            <p><strong>平台:</strong> {{ systemInfo.platform }}</p>
            <p><strong>架构:</strong> {{ systemInfo.arch }}</p>
            <p><strong>主机名:</strong> {{ systemInfo.hostname }}</p>
            <p><strong>内核版本:</strong> {{ systemInfo.kernelVersion }}</p>
          </div>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card class="status-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <el-icon :size="20"><VideoPlay /></el-icon>
              <span>内核控制</span>
            </div>
          </template>
          <div class="control-content">
            <div class="process-status">
              <el-tag :type="processStatus.running ? 'success' : 'info'" size="large">
                {{ processStatus.running ? '运行中' : '已停止' }}
              </el-tag>
              <span v-if="processStatus.running" class="pid-info">PID: {{ processStatus.pid }}</span>
            </div>
            <div v-if="processStatus.running && processStatus.startTime" class="runtime-info">
              <p><strong>启动时间:</strong> {{ formatDate(processStatus.startTime) }}</p>
              <p v-if="uptimeTick >= 0"><strong>运行时长:</strong> {{ formatUpTime(processStatus.startTime) }}</p>
            </div>
            <div class="control-buttons">
              <el-button
                v-if="!processStatus.running"
                type="primary"
                size="small"
                @click="startProcess"
                :loading="controlling"
              >
                <el-icon><VideoPlay /></el-icon>
                启动
              </el-button>
              <el-button
                v-if="processStatus.running"
                type="danger"
                size="small"
                @click="stopProcess"
                :loading="controlling"
              >
                <el-icon><VideoPause /></el-icon>
                停止
              </el-button>
              <el-button
                type="warning"
                size="small"
                @click="restartProcess"
                :loading="controlling"
              >
                <el-icon><RefreshRight /></el-icon>
                重启
              </el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { kernelApi } from '../api/kernel'
import { processApi } from '../api/process'
import { Box, VideoPause, InfoFilled, VideoPlay, RefreshRight } from '@element-plus/icons-vue'

const status = ref({
  installed: false,
  version: '',
  path: ''
})

const systemInfo = ref({
  platform: '-',
  arch: '-',
  hostname: '-',
  kernelVersion: '-'
})

const processStatus = ref({
  running: false,
  pid: 0,
  status: 'stopped',
  version: ''
})

const controlling = ref(false)

let uptimeInterval = null
const uptimeTick = ref(0)

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

const formatUpTime = (startTime) => {
  if (!startTime) return '-'
  const start = new Date(startTime)
  const now = new Date()
  const diff = Math.floor((now - start) / 1000)

  if (diff < 0) return '-'

  const years = Math.floor(diff / 31536000)
  const months = Math.floor((diff % 31536000) / 2592000)
  const days = Math.floor((diff % 2592000) / 86400)
  const hours = Math.floor((diff % 86400) / 3600)
  const minutes = Math.floor((diff % 3600) / 60)

  const parts = []
  if (years > 0) parts.push(`${years}y`)
  if (months > 0) parts.push(`${months}m`)
  if (days > 0) parts.push(`${days}d`)
  if (hours > 0) parts.push(`${hours}h`)
  if (minutes > 0) parts.push(`${minutes}m`)

  return parts.length > 0 ? parts.join(' ') : '<1m'
}

const loadStatus = async () => {
  try {
    const res = await kernelApi.getStatus()
    if (res.data.success) {
      status.value = res.data.data
    }
  } catch (err) {
    console.error('Failed to load status:', err)
  }
}

const loadSystemInfo = async () => {
  try {
    const res = await kernelApi.getSystemInfo()
    if (res.data.success) {
      systemInfo.value = res.data.data
    }
  } catch (err) {
    console.error('Failed to get system info:', err)
  }
}

const loadProcessStatus = async () => {
  try {
    const res = await processApi.getStatus()
    if (res.data.success) {
      processStatus.value = { ...res.data.data }
    }
  } catch (err) {
    console.error('Failed to get process status:', err)
  }
}

const startProcess = async () => {
  controlling.value = true
  try {
    await processApi.start()
    ElMessage.success('内核已启动')
    await loadProcessStatus()
  } catch (err) {
    ElMessage.error('启动失败: ' + (err.response?.data?.error || err.message))
  } finally {
    controlling.value = false
  }
}

const stopProcess = async () => {
  controlling.value = true
  try {
    await processApi.stop()
    ElMessage.success('内核已停止')
    await loadProcessStatus()
  } catch (err) {
    ElMessage.error('停止失败: ' + (err.response?.data?.error || err.message))
  } finally {
    controlling.value = false
  }
}

const restartProcess = async () => {
  controlling.value = true
  try {
    await processApi.restart()
    ElMessage.success('内核已重启')
    await loadProcessStatus()
  } catch (err) {
    ElMessage.error('重启失败: ' + (err.response?.data?.error || err.message))
  } finally {
    controlling.value = false
  }
}

let processPollInterval = null

const startProcessPolling = () => {
  stopProcessPolling()
  processPollInterval = setInterval(() => {
    loadProcessStatus()
  }, 5000)
}

const stopProcessPolling = () => {
  if (processPollInterval) {
    clearInterval(processPollInterval)
    processPollInterval = null
  }
}

onMounted(() => {
  loadStatus()
  loadSystemInfo()
  loadProcessStatus()
  startProcessPolling()
  uptimeInterval = setInterval(() => {
    uptimeTick.value++
  }, 60000)
})

onUnmounted(() => {
  stopProcessPolling()
  if (uptimeInterval) {
    clearInterval(uptimeInterval)
  }
})
</script>

<style scoped>
.system-page {
  max-width: 1200px;
  margin: 0 auto;
}

.page-title {
  margin: 0 0 24px 0;
  color: var(--text-primary);
  font-size: 24px;
  font-weight: 600;
}

.overview-row {
  margin-bottom: 24px;
}

.status-card {
  min-height: 240px;
  background: var(--bg-card);
  border-color: var(--border-color);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: var(--text-primary);
}

.status-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.status-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-info {
  color: var(--text-regular);
  font-size: 14px;
}

.status-info p {
  margin: 4px 0;
}

.system-info {
  color: var(--text-regular);
  font-size: 14px;
}

.system-info p {
  margin: 4px 0;
}

.control-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.process-status {
  display: flex;
  align-items: center;
  gap: 12px;
}

.pid-info {
  font-size: 12px;
  color: var(--text-secondary);
}

.control-buttons {
  display: flex;
  gap: 8px;
}

.runtime-info {
  font-size: 12px;
  color: var(--text-regular);
}

.runtime-info p {
  margin: 2px 0;
}
</style>
