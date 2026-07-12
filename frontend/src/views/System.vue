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
              <el-popconfirm
                v-if="status.installed"
                title="确定要删除已安装的内核吗？"
                confirm-button-text="确定"
                cancel-button-text="取消"
                @confirm="removeKernel"
              >
                <template #reference>
                  <el-button type="danger" link size="small">
                    <el-icon><Delete /></el-icon>
                    删除
                  </el-button>
                </template>
              </el-popconfirm>
            </div>
            <div v-if="status.installed" class="status-info">
              <p><strong>版本:</strong> {{ status.version || '未知' }}</p>
              <p><strong>下载类型:</strong> {{ status.downloadType || '未知' }}</p>
              <p><strong>更新时间:</strong> {{ formatDate(status.lastUpdated) }}</p>
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
                :disabled="!status.installed"
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
                :disabled="!status.installed"
              >
                <el-icon><RefreshRight /></el-icon>
                重启
              </el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Download Progress -->
    <el-card v-if="downloadProgress.status === 'downloading' || downloadProgress.status === 'completed' || downloadProgress.status === 'failed'" class="section-card" shadow="never">
      <template #header>
        <div class="card-header">
          <el-icon v-if="downloadProgress.status === 'downloading'"><Loading /></el-icon>
          <el-icon v-else-if="downloadProgress.status === 'completed'"><CircleCheck /></el-icon>
          <el-icon v-else><WarningFilled /></el-icon>
          <span>下载进度</span>
        </div>
      </template>
      <div class="progress-content">
        <el-progress
          :percentage="Number(downloadProgress.progress.toFixed(3))"
          :status="downloadProgress.status === 'completed' ? 'success' : downloadProgress.status === 'failed' ? 'exception' : ''"
          :striped="downloadProgress.status === 'downloading'"
          :striped-flow="downloadProgress.status === 'downloading'"
        />
        <div class="progress-info">
          <span>{{ statusText }}</span>
          <span v-if="downloadProgress.version">版本: {{ downloadProgress.version }}</span>
        </div>
        <el-alert
          v-if="downloadProgress.error"
          :title="downloadProgress.error"
          type="error"
          show-icon
          :closable="false"
          class="progress-error"
        />
        <el-button
          v-if="downloadProgress.active"
          type="danger"
          @click="stopDownload"
          class="stop-button"
        >
          <el-icon><VideoPause /></el-icon>
          停止下载
        </el-button>
      </div>
    </el-card>

    <!-- Download Section -->
    <el-card class="section-card" shadow="never">
      <template #header>
        <div class="card-header">
          <el-icon><Download /></el-icon>
          <span>下载内核</span>
        </div>
      </template>

      <el-tabs v-model="downloadType" type="border-card">
        <el-tab-pane label="Latest (最新版)" name="latest">
          <div class="tab-content">
            <p class="tab-desc">下载最新的开发版本，包含最新功能和修复</p>
            <el-button
              type="primary"
              @click="startDownload('latest')"
              :loading="downloadProgress.active"
              :disabled="downloadProgress.active"
            >
              <el-icon><Download /></el-icon>
              下载 Latest 版本
            </el-button>
          </div>
        </el-tab-pane>

        <el-tab-pane label="Stable (稳定版)" name="stable">
          <div class="tab-content">
            <p class="tab-desc">下载经过测试的稳定版本，推荐生产环境使用</p>
            <el-button
              type="success"
              @click="startDownload('stable')"
              :loading="downloadProgress.active"
              :disabled="downloadProgress.active"
            >
              <el-icon><Download /></el-icon>
              下载 Stable 版本
            </el-button>
          </div>
        </el-tab-pane>

        <el-tab-pane label="Custom (自定义)" name="custom">
          <div class="tab-content">
            <p class="tab-desc">使用自定义下载链接</p>
            <el-form :model="customForm" label-width="100px">
              <el-form-item label="下载链接">
                <el-input
                  v-model="customForm.url"
                  placeholder="https://github.com/.../sing-box-xxx.tar.gz"
                  :disabled="downloadProgress.active"
                />
              </el-form-item>
              <el-form-item label="版本号">
                <el-input
                  v-model="customForm.version"
                  placeholder="可选，例如 1.8.0"
                  :disabled="downloadProgress.active"
                />
              </el-form-item>
              <el-form-item>
                <el-button
                  type="warning"
                  @click="startDownload('custom')"
                  :loading="downloadProgress.active"
                  :disabled="downloadProgress.active || !customForm.url"
                >
                  <el-icon><Download /></el-icon>
                  下载自定义版本
                </el-button>
              </el-form-item>
            </el-form>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- Available Versions -->
    <el-card class="section-card" shadow="never">
      <template #header>
        <div class="card-header">
          <el-icon><List /></el-icon>
          <span>可用版本</span>
          <span v-if="cacheTime" class="cache-time">上一次拉取时间: {{ formatDate(cacheTime) }}</span>
          <el-button type="primary" link @click="refreshVersions" :loading="loadingVersions">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <el-table :data="versions" v-loading="loadingVersions" stripe>
        <el-table-column prop="version" label="版本号" width="120" />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.prerelease ? 'warning' : 'success'" size="small">
              {{ row.prerelease ? 'Pre-release' : 'Stable' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发布时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.publishedAt) }}
          </template>
        </el-table-column>
        <el-table-column label="大小" width="100">
          <template #default="{ row }">
            {{ formatSize(row.assets?.[0]?.size) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              link
              size="small"
              @click="installVersion(row)"
              :disabled="downloadProgress.active"
            >
              安装
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { kernelApi } from '../api/kernel'
import { processApi } from '../api/process'
import {
  Box, Download, List, Refresh, Delete, Loading,
  CircleCheck, VideoPause, WarningFilled, InfoFilled,
  VideoPlay, RefreshRight, Upload
} from '@element-plus/icons-vue'

const status = ref({
  installed: false,
  version: '',
  path: '',
  lastUpdated: null,
  downloadType: ''
})

const systemInfo = ref({
  platform: '-',
  arch: '-',
  hostname: '-',
  kernelVersion: '-'
})

const runtimeStats = ref({
  startTime: null,
  traffic: {
    up: { total: 0, speed: 0 },
    down: { total: 0, speed: 0 }
  }
})

const processStatus = ref({
  running: false,
  pid: 0,
  status: 'stopped',
  version: ''
})

const controlling = ref(false)

const versions = ref([])
const loadingVersions = ref(false)
const cacheTime = ref(null)
const downloadType = ref('latest')
const downloadProgress = ref({
  active: false,
  progress: 0,
  status: '',
  version: '',
  error: ''
})

const customForm = ref({
  url: '',
  version: ''
})

let progressInterval = null
let uptimeInterval = null
const uptimeTick = ref(0)

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

const formatSize = (bytes) => {
  if (!bytes) return '-'
  const units = ['B', 'KB', 'MB', 'GB']
  let size = bytes
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }
  return `${size.toFixed(1)} ${units[unitIndex]}`
}

const formatBytes = formatSize

const statusText = computed(() => {
  const map = {
    downloading: '下载中',
    completed: '下载完成',
    failed: '下载失败',
    idle: '空闲'
  }
  return map[downloadProgress.value.status] || downloadProgress.value.status
})

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
      const data = res.data.data
      status.value = data

      // If download is active, show progress bar and start polling
      if (data.active) {
        downloadProgress.value = {
          active: true,
          progress: data.progress || 0,
          status: data.status || 'downloading',
          version: data.version || '',
          error: data.statusMsg || ''
        }
        startProgressPolling()
      }
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

const loadStats = async () => {
  try {
    const res = await statsApi.getServiceInfo()
    if (res.data.success && res.data.data) {
      runtimeStats.value = {
        ...runtimeStats.value,
        startTime: res.data.data.startTime
      }
    }
  } catch (err) {
    console.error('Failed to get stats:', err)
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

const loadVersions = async () => {
  loadingVersions.value = true
  try {
    const res = await kernelApi.getVersions()
    if (res.data.success) {
      versions.value = res.data.data || []
      if (res.data.cacheTime) {
        cacheTime.value = res.data.cacheTime
      }
    }
  } catch (err) {
    ElMessage.error('加载版本列表失败: ' + (err.response?.data?.error || err.message))
  } finally {
    loadingVersions.value = false
  }
}

const refreshVersions = async () => {
  loadingVersions.value = true
  try {
    const res = await kernelApi.refreshVersions()
    if (res.data.success) {
      if (res.data.cacheTime) {
        cacheTime.value = res.data.cacheTime
      }
      await loadVersions()
      ElMessage.success('版本列表已刷新')
    }
  } catch (err) {
    ElMessage.error('刷新失败: ' + (err.response?.data?.error || err.message))
  } finally {
    loadingVersions.value = false
  }
}

const startDownload = async (type) => {
  const data = { type }
  if (type === 'custom') {
    data.url = customForm.value.url
    data.version = customForm.value.version
  }

  try {
    await kernelApi.download(data)
    ElMessage.success('下载已开始')
    startProgressPolling()
  } catch (err) {
    ElMessage.error('下载失败: ' + (err.response?.data?.error || err.message))
  }
}

const stopDownload = async () => {
  try {
    await kernelApi.stopDownload()
    ElMessage.success('下载已停止')
    stopProgressPolling()
    loadStatus()
  } catch (err) {
    ElMessage.error('停止失败: ' + (err.response?.data?.error || err.message))
  }
}

const installVersion = (version) => {
  if (version.assets && version.assets.length > 0) {
    customForm.value.url = version.assets[0].downloadUrl
    customForm.value.version = version.version
    downloadType.value = 'custom'
    startDownload('custom')
  }
}

const removeKernel = async () => {
  try {
    await kernelApi.remove()
    ElMessage.success('内核已删除')
    loadStatus()
  } catch (err) {
    ElMessage.error('删除失败: ' + (err.response?.data?.error || err.message))
  }
}

const startProgressPolling = () => {
  stopProgressPolling()
  progressInterval = setInterval(async () => {
    try {
      const res = await kernelApi.getStatus()
      if (res.data.success) {
        const data = res.data.data
        status.value = data

        // Update download progress from API
        downloadProgress.value = {
          active: data.active,
          progress: data.progress || 0,
          status: data.status || 'idle',
          version: data.version || '',
          error: data.statusMsg || ''
        }

        // Stop polling when not active
        if (!data.active) {
          stopProgressPolling()
          if (data.status === 'completed') {
            ElMessage.success('下载完成')
          } else if (data.status === 'failed') {
            ElMessage.error('下载失败: ' + (data.statusMsg || '未知错误'))
          }
          loadStatus()
        }
      }
    } catch (err) {
      console.error('Failed to poll status:', err)
    }
  }, 1000)
}

const stopProgressPolling = () => {
  if (progressInterval) {
    clearInterval(progressInterval)
    progressInterval = null
  }
}

onMounted(() => {
  loadStatus()
  loadSystemInfo()
  loadProcessStatus()
  loadVersions()
  // Update uptime display every 60 seconds
  uptimeInterval = setInterval(() => {
    uptimeTick.value++
  }, 60000)
})

onUnmounted(() => {
  stopProgressPolling()
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
  color: #303133;
  font-size: 24px;
  font-weight: 600;
}

.overview-row {
  margin-bottom: 24px;
}

.status-card {
  min-height: 240px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.card-header.danger {
  color: #f56c6c;
}

.cache-time {
  margin-left: auto;
  margin-right: 16px;
  font-size: 12px;
  color: #909399;
  font-weight: normal;
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

.status-row .el-button {
  margin-left: auto;
}

.status-info {
  color: #606266;
  font-size: 14px;
}

.status-info p {
  margin: 4px 0;
}

.system-info {
  color: #606266;
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
  color: #909399;
}

.control-buttons {
  display: flex;
  gap: 8px;
}

.runtime-info {
  font-size: 12px;
  color: #606266;
}

.runtime-info p {
  margin: 2px 0;
}

.section-card {
  margin-bottom: 24px;
}

.tab-content {
  padding: 16px 0;
}

.tab-desc {
  color: #909399;
  margin-bottom: 16px;
}

.progress-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  color: #606266;
  font-size: 14px;
}

.progress-error {
  margin-top: 8px;
}

.stop-button {
  align-self: flex-start;
}

.danger-zone {
  border-color: #f56c6c;
}

.danger-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.danger-content p {
  color: #909399;
  margin: 0;
}
</style>
