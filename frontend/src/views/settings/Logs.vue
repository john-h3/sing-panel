<template>
  <el-card class="section-card" shadow="never">
    <template #header>
      <div class="card-header">
        <el-icon><Document /></el-icon>
        <span>日志管理</span>
      </div>
    </template>

    <el-form :model="form" label-width="120px" v-loading="loading">
      <el-form-item label="日志级别">
        <el-select v-model="form.logLevel" style="width: 220px">
          <el-option
            v-for="item in levels"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <div class="form-tip">
          面板日志和嵌入式 sing-box 内核日志共用此级别，保存后立即生效。当前日志文件仍由应用按大小自动滚动。
        </div>
      </el-form-item>

      <el-form-item>
        <el-button type="primary" @click="saveConfig" :loading="saving">
          <el-icon><Check /></el-icon>
          保存配置
        </el-button>
      </el-form-item>
    </el-form>

    <el-alert
      title="推荐使用 warn"
      description="info 会记录每次入站连接，流量较大时会快速增加日志体积；debug 仅建议在排查问题时临时开启。"
      type="info"
      :closable="false"
      show-icon
    />

    <el-card class="memory-card" shadow="never">
      <template #header>
        <div class="card-header">
          <el-icon><Monitor /></el-icon>
          <span>内存日志</span>
          <span class="memory-stats">显示 {{ visibleLogs.length }} 条</span>
          <el-button type="danger" link size="small" @click="clearLogs">清空</el-button>
        </div>
      </template>
      <div class="filters">
        <el-select v-model="sourceFilter" clearable placeholder="全部来源" size="small" class="log-filter">
          <el-option label="面板" value="panel" />
          <el-option label="Gin" value="gin" />
          <el-option label="sing-box" value="singbox" />
        </el-select>
        <el-select v-model="levelFilter" clearable placeholder="最低级别" size="small" class="log-filter">
          <el-option v-for="item in levels" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
        <el-button size="small" :type="paused ? 'warning' : 'default'" @click="togglePaused">
          {{ paused ? '继续接收' : '暂停接收' }}
        </el-button>
        <el-switch v-model="autoScroll" active-text="自动滚动" />
        <el-switch v-model="filterHealthLogs" active-text="过滤检测 API" />
      </div>
      <el-table ref="logTable" :data="visibleLogs" height="480" size="small" empty-text="暂无内存日志">
        <el-table-column prop="time" label="时间" width="190">
          <template #default="scope">{{ formatTime(scope.row.time) }}</template>
        </el-table-column>
        <el-table-column prop="level" label="级别" width="90">
          <template #default="scope"><el-tag :type="levelTag(scope.row.level)" size="small">{{ scope.row.level }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="source" label="来源" width="100" />
        <el-table-column prop="message" label="内容" min-width="400" show-overflow-tooltip />
      </el-table>
      <div class="form-tip">页面只显示最近 {{ LOG_TAIL_SIZE }} 条；暂停时停止接收实时日志，继续后按序号补齐期间日志。服务重启后内存日志清空。</div>
    </el-card>
  </el-card>
</template>

<script setup>
import { ref, computed, nextTick, onMounted, onUnmounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { configApi } from '../../api/config'
import axios from 'axios'
import { Check, Document, Monitor } from '@element-plus/icons-vue'

const levels = [
  { value: 'debug', label: 'Debug（调试）' },
  { value: 'info', label: 'Info（信息）' },
  { value: 'warn', label: 'Warn（警告，推荐）' },
  { value: 'error', label: 'Error（错误）' }
]
const LOG_TAIL_SIZE = 100
const LOG_BUFFER_SIZE = 2048
const LEVEL_RANKS = { trace: 0, debug: 1, info: 2, warn: 3, error: 4 }
const form = ref({ logLevel: 'warn' })
const loading = ref(false)
const saving = ref(false)
const logs = ref([])
const logTable = ref(null)
const sourceFilter = ref('')
const levelFilter = ref('')
const autoScroll = ref(true)
const FILTER_HEALTH_KEY = 'logs.filterHealthLogs'
const filterHealthLogs = ref(localStorage.getItem(FILTER_HEALTH_KEY) !== '0')
const paused = ref(false)
const logStats = ref({ capacity: 2048, count: 0, dropped: 0 })
let logStream = null
let streamStarted = false
let pendingEntries = []
let flushFrame = 0
const isHealthLog = (entry) => {
  if (entry?.source !== 'gin') return false
  return entry.message.split(/\s+/).some(field => {
    field = field.replace(/["'|,]/g, '')
    return field === '/health' || field.startsWith('/health?')
  })
}

const entryMatchesFilter = (entry) => {
  if (sourceFilter.value && entry.source !== sourceFilter.value) return false
  const rank = LEVEL_RANKS[entry.level] ?? -1
  return rank >= (LEVEL_RANKS[levelFilter.value] ?? 0)
}

watch(filterHealthLogs, value => {
  localStorage.setItem(FILTER_HEALTH_KEY, value ? '1' : '0')
})

const visibleLogs = computed(() => {
  const filtered = filterHealthLogs.value
    ? logs.value.filter(entry => !isHealthLog(entry))
    : logs.value
  return filtered.slice(-LOG_TAIL_SIZE)
})

const loadConfig = async () => {
  loading.value = true
  try {
    const res = await configApi.get()
    if (res.data.success) {
      form.value.logLevel = res.data.data.logLevel || 'warn'
    }
  } catch {
    ElMessage.error('加载日志配置失败')
  } finally {
    loading.value = false
  }
}

const saveConfig = async () => {
  saving.value = true
  try {
    const res = await configApi.update({ logLevel: form.value.logLevel })
    if (!res.data.success) throw new Error(res.data.error || '保存失败')
    form.value.logLevel = res.data.data.logLevel || form.value.logLevel
    ElMessage.success('日志级别已更新')
  } catch (err) {
    ElMessage.error('保存失败: ' + (err.response?.data?.error || err.message))
  } finally {
    saving.value = false
  }
}

const formatTime = (value) => new Date(value).toLocaleString('zh-CN')
const levelTag = (level) => ({ debug: 'info', info: '', warn: 'warning', error: 'danger' }[level] || 'info')

const loadLogs = async () => {
  try {
    const res = await axios.get('/api/logs', {
      params: {
        limit: LOG_BUFFER_SIZE,
        level: levelFilter.value || undefined,
        source: sourceFilter.value || undefined
      }
    })
    if (res.data.success) {
      logs.value = res.data.data.entries || []
      logStats.value = res.data.data
    }
  } catch {
    ElMessage.error('加载内存日志失败')
  }
}

const clearLogs = async () => {
  try {
    await axios.delete('/api/logs')
    logs.value = []
    logStats.value.count = 0
  } catch {
    ElMessage.error('清空内存日志失败')
  }
}

const startLogStream = () => {
	if (logStream) logStream.close()
	pendingEntries = []
	if (flushFrame) {
		cancelAnimationFrame(flushFrame)
		flushFrame = 0
	}
	const params = new URLSearchParams()
	if (levelFilter.value) params.set('level', levelFilter.value)
	if (sourceFilter.value) params.set('source', sourceFilter.value)
	const lastSeq = logs.value.length ? logs.value[logs.value.length - 1].seq : 0
	if (lastSeq) params.set('after', String(lastSeq))
	const query = params.toString()
	logStream = new EventSource(`/api/logs/stream${query ? `?${query}` : ''}`)
	logStream.addEventListener('log', event => {
		try {
			const entry = JSON.parse(event.data)
			pendingEntries.push(entry)
			if (!flushFrame) flushFrame = requestAnimationFrame(flushPendingEntries)
		} catch {
      // Ignore malformed events and keep the stream alive.
    }
	})
}

const flushPendingEntries = () => {
	flushFrame = 0
	if (!pendingEntries.length) return
	const entries = pendingEntries.filter(entryMatchesFilter)
	pendingEntries = []
	if (!entries.length) return
  logs.value = [...logs.value, ...entries].slice(-LOG_BUFFER_SIZE)
	if (autoScroll.value) {
		nextTick(() => logTable.value?.setScrollTop(Number.MAX_SAFE_INTEGER))
	}
}

const togglePaused = async () => {
	paused.value = !paused.value
	if (paused.value) {
		if (logStream) {
			logStream.close()
			logStream = null
		}
		return
	}
	await loadLogsAfterLastSeq()
	startLogStream()
}

const loadLogsAfterLastSeq = async () => {
	const lastSeq = logs.value.length ? logs.value[logs.value.length - 1].seq : 0
	try {
		const res = await axios.get('/api/logs', {
      params: {
        after: lastSeq || undefined,
        limit: LOG_BUFFER_SIZE,
				level: levelFilter.value || undefined,
				source: sourceFilter.value || undefined
			}
		})
		if (res.data.success) {
			const entries = res.data.data.entries || []
      logs.value = [...logs.value, ...entries].slice(-LOG_BUFFER_SIZE)
			logStats.value = res.data.data
		}
	} catch {
		ElMessage.error('恢复日志失败')
	}
}

watch([sourceFilter, levelFilter], async () => {
	if (!streamStarted) return
	await loadLogs()
	if (!paused.value) startLogStream()
})

onMounted(async () => {
  await Promise.all([loadConfig(), loadLogs()])
  startLogStream()
	streamStarted = true
})

onUnmounted(() => {
  if (logStream) logStream.close()
	if (flushFrame) cancelAnimationFrame(flushFrame)
})
</script>

<style scoped>
.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: var(--text-primary);
}

.form-tip {
  margin-top: 4px;
  color: var(--text-secondary);
  font-size: 12px;
}

.memory-card {
  margin-top: 16px;
  background: var(--bg-card);
  border-color: var(--border-color);
}

.memory-stats {
  margin-left: auto;
  font-size: 12px;
  font-weight: 400;
  color: var(--text-secondary);
}

.filters {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}

.filters :deep(.log-filter) {
  width: 120px;
}

.filters :deep(.el-switch) {
  margin-left: 4px;
}
</style>
