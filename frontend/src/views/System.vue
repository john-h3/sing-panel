<template>
  <div class="system-page">
    <el-row :gutter="24">
      <el-col :span="24">
        <h2 class="page-title">内核管理</h2>
      </el-col>
    </el-row>

    <!-- Overview -->
    <el-row :gutter="24" class="overview-row">
      <el-col :span="8">
        <el-card class="status-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <el-icon :size="20"><Box /></el-icon>
              <span>内核信息</span>
            </div>
          </template>
          <div class="status-content">
            <div class="status-row">
              <el-tag type="success" size="large">内嵌内核</el-tag>
            </div>
            <div class="status-info">
              <p><strong>版本:</strong> {{ status.version || '未知' }}</p>
              <p><strong>构建时间:</strong> {{ systemInfo.buildTime || '未知' }}</p>
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
              <div class="auto-start-setting">
                <span class="setting-label">跟随启动</span>
                <el-tooltip
                  content="开启后，面板启动完成时会自动启动内嵌 sing-box 内核；关闭则需要手动启动。"
                  placement="top"
                >
                  <el-icon class="setting-info"><InfoFilled /></el-icon>
                </el-tooltip>
                <el-switch
                  v-model="autoStartKernel"
                  :loading="savingAutoStart"
                  @change="updateAutoStartKernel"
                  size="small"
                />
              </div>
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

    <!-- Runtime Monitor -->
    <el-row :gutter="24" class="monitor-row">
      <el-col :span="24">
        <el-card class="monitor-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <el-icon :size="20"><DataLine /></el-icon>
              <span>运行监控</span>
              <span class="monitor-hint">Go 运行时指标 · 每 5 秒刷新</span>
            </div>
          </template>
          <div class="monitor-grid">
            <div class="monitor-item">
              <div class="monitor-label">面板运行时长</div>
              <div class="monitor-value">{{ formatUpTime(new Date(Date.now() - monitor.uptimeSeconds * 1000).toISOString()) }}</div>
            </div>
            <div class="monitor-item">
              <div class="monitor-label">Goroutine 数</div>
              <div class="monitor-value">{{ monitor.goroutines }}</div>
            </div>
            <div class="monitor-item">
              <div class="monitor-label">堆内存使用</div>
              <div class="monitor-value">
                {{ formatBytes(monitor.heapAlloc) }}
                <span class="monitor-sub">/ {{ formatBytes(monitor.heapSys) }}</span>
              </div>
              <el-progress :percentage="heapPercent" :stroke-width="8" />
            </div>
            <div class="monitor-item">
              <div class="monitor-label">进程总内存 (Sys)</div>
              <div class="monitor-value">{{ formatBytes(monitor.sys) }}</div>
            </div>
            <div class="monitor-item">
              <div class="monitor-label">GC 次数</div>
              <div class="monitor-value">{{ monitor.numGC }}</div>
            </div>
            <div class="monitor-item">
              <div class="monitor-label">上次 GC 暂停</div>
              <div class="monitor-value">{{ formatNs(monitor.lastPauseNs) }}</div>
            </div>
            <div class="monitor-item">
              <div class="monitor-label">GC 累计暂停</div>
              <div class="monitor-value">{{ formatNs(monitor.pauseTotalNs) }}</div>
            </div>
            <div class="monitor-item">
              <div class="monitor-label">GOGC / CPU</div>
              <div class="monitor-value">
                {{ monitor.gcPercent }}%
                <span class="monitor-sub">/ {{ monitor.gomaxprocs }} 核</span>
              </div>
            </div>
            <div class="monitor-item">
              <div class="monitor-label">堆对象数</div>
              <div class="monitor-value">{{ monitor.heapObjects }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Runtime Config (only shown while the kernel is running) -->
    <el-row v-if="processStatus.running" :gutter="24" class="config-row">
      <el-col :span="24">
        <el-card class="config-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <el-icon :size="20"><Document /></el-icon>
              <span>内核配置</span>
              <span class="config-hint">实际传给内核的配置（只读）</span>
              <el-button type="primary" link size="small" @click="loadRuntimeConfig">
                <el-icon><RefreshRight /></el-icon>
                刷新
              </el-button>
            </div>
          </template>
          <div ref="configEditorRef" class="config-editor"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { kernelApi } from '../api/kernel'
import { processApi } from '../api/process'
import { configApi } from '../api/config'
import { useTheme } from '../composables/useTheme'
import { EditorView, lineNumbers, highlightActiveLine } from '@codemirror/view'
import { EditorState } from '@codemirror/state'
import { foldGutter, syntaxHighlighting, defaultHighlightStyle } from '@codemirror/language'
import { json } from '@codemirror/lang-json'
import { oneDark } from '@codemirror/theme-one-dark'
import { Box, VideoPause, InfoFilled, VideoPlay, RefreshRight, DataLine, Document } from '@element-plus/icons-vue'

const { isDark } = useTheme()

const status = ref({
  version: ''
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
const autoStartKernel = ref(false)
const savingAutoStart = ref(false)

const monitor = ref({
  uptimeSeconds: 0,
  goroutines: 0,
  numCPU: 0,
  gomaxprocs: 0,
  heapAlloc: 0,
  heapSys: 0,
  heapInuse: 0,
  heapObjects: 0,
  sys: 0,
  numGC: 0,
  lastGC: 0,
  pauseTotalNs: 0,
  lastPauseNs: 0,
  gcPercent: 0
})

const heapPercent = computed(() => {
  if (!monitor.value.heapSys) return 0
  return Math.min(100, Math.round((monitor.value.heapAlloc / monitor.value.heapSys) * 100))
})

const formatBytes = (bytes) => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  let v = bytes
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024
    i++
  }
  return `${v.toFixed(i === 0 ? 0 : 1)} ${units[i]}`
}

const formatNs = (ns) => {
  if (!ns) return '0 μs'
  if (ns < 1000) return `${ns} ns`
  if (ns < 1e6) return `${(ns / 1000).toFixed(1)} μs`
  return `${(ns / 1e6).toFixed(2)} ms`
}

const loadMonitor = async () => {
  try {
    const res = await kernelApi.getMonitor()
    if (res.data.success) {
      monitor.value = { ...res.data.data }
    }
  } catch (err) {
    console.error('Failed to load monitor stats:', err)
  }
}

const configEditorRef = ref(null)
let configEditorView = null

const initConfigEditor = (content) => {
  if (configEditorView) {
    configEditorView.destroy()
    configEditorView = null
  }
  if (!configEditorRef.value) return
  const extensions = [
    lineNumbers(),
    foldGutter(),
    highlightActiveLine(),
    json(),
    syntaxHighlighting(defaultHighlightStyle),
    EditorState.readOnly.of(true),
    EditorView.editable.of(false)
  ]
  if (isDark.value) extensions.push(oneDark)
  configEditorView = new EditorView({
    parent: configEditorRef.value,
    state: EditorState.create({ doc: content, extensions })
  })
}

const loadRuntimeConfig = async () => {
  try {
    const res = await processApi.getRuntimeConfig()
    if (res.data.success && res.data.running && res.data.data) {
      await nextTick()
      initConfigEditor(JSON.stringify(res.data.data, null, 2))
    }
  } catch (err) {
    console.error('Failed to load runtime config:', err)
  }
}

// Load the runtime config only when the kernel transitions to running,
// and tear the editor down when it stops.
watch(() => processStatus.value.running, (running) => {
  if (running) {
    loadRuntimeConfig()
  } else if (configEditorView) {
    configEditorView.destroy()
    configEditorView = null
  }
})

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

const loadConfig = async () => {
  try {
    const res = await configApi.get()
    if (res.data.success) {
      autoStartKernel.value = res.data.data?.autoStartKernel === true
    }
  } catch (err) {
    console.error('Failed to load panel config:', err)
  }
}

const updateAutoStartKernel = async (enabled) => {
  savingAutoStart.value = true
  try {
    const res = await configApi.update({ autoStartKernel: enabled })
    if (!res.data.success) {
      throw new Error(res.data.error || '配置保存失败')
    }
    ElMessage.success(enabled ? '已开启内核跟随面板启动' : '已关闭内核跟随面板启动')
  } catch (err) {
    autoStartKernel.value = !enabled
    ElMessage.error('保存失败: ' + (err.response?.data?.error || err.message))
  } finally {
    savingAutoStart.value = false
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
    loadMonitor()
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
  loadConfig()
  loadProcessStatus()
  loadMonitor()
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
  if (configEditorView) {
    configEditorView.destroy()
    configEditorView = null
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

.control-buttons {
  display: flex;
  gap: 8px;
}

.auto-start-setting {
  display: flex;
  align-items: center;
  gap: 5px;
  margin-left: auto;
}

.setting-label {
  font-size: 13px;
  color: var(--text-primary);
}

.setting-info {
  font-size: 14px;
  color: var(--text-secondary);
  cursor: help;
}

.runtime-info {
  font-size: 12px;
  color: var(--text-regular);
}

.runtime-info p {
  margin: 2px 0;
}

.monitor-row {
  margin-bottom: 24px;
}

.monitor-card {
  background: var(--bg-card);
  border-color: var(--border-color);
}

.monitor-hint {
  font-size: 12px;
  font-weight: 400;
  color: var(--text-secondary);
  margin-left: auto;
}

.monitor-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px 24px;
}

.monitor-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.monitor-label {
  font-size: 12px;
  color: var(--text-secondary);
}

.monitor-value {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.monitor-sub {
  font-size: 12px;
  font-weight: 400;
  color: var(--text-secondary);
}

.config-row {
  margin-bottom: 24px;
}

.config-card {
  background: var(--bg-card);
  border-color: var(--border-color);
}

.config-hint {
  font-size: 12px;
  font-weight: 400;
  color: var(--text-secondary);
}

.config-card .el-button {
  margin-left: auto;
}

.config-editor {
  height: 420px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  overflow: hidden;
}

.config-editor :deep(.cm-editor) {
  height: 100%;
}
</style>
