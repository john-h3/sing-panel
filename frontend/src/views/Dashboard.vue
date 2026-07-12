<template>
  <div class="dashboard-page">
    <el-row :gutter="24">
      <el-col :span="24">
        <h2 class="page-title">监控面板</h2>
      </el-col>
    </el-row>

    <el-row :gutter="24" class="overview-row">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <el-icon :size="32" :color="processStatus.running ? '#67c23a' : '#909399'"><VideoPlay /></el-icon>
            <div class="stat-info">
              <div class="stat-label">内核状态</div>
              <div class="stat-value">{{ processStatus.running ? '运行中' : '已停止' }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <el-icon :size="32" color="#409eff"><Timer /></el-icon>
            <div class="stat-info">
              <div class="stat-label">运行时长</div>
              <div class="stat-value">{{ processStatus.running && processStatus.startTime ? formatUpTime(processStatus.startTime) : '-' }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <el-icon :size="32" color="#409eff"><Top /></el-icon>
            <div class="stat-info">
              <div class="stat-label">上行速度</div>
              <div class="stat-value">{{ formatBytes(currentSpeed.up) }}/s</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <el-icon :size="32" color="#67c23a"><Bottom /></el-icon>
            <div class="stat-info">
              <div class="stat-label">下行速度</div>
              <div class="stat-value">{{ formatBytes(currentSpeed.down) }}/s</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="24" class="overview-row">
      <el-col :span="12">
        <el-card shadow="never" class="chart-card">
          <template #header>
            <div class="card-header">
              <el-icon :size="20"><DataLine /></el-icon>
              <span>实时速率</span>
              <el-select v-model="speedInterval" size="small" class="interval-select" @change="onSpeedIntervalChange">
                <el-option label="1s" :value="1" />
                <el-option label="3s" :value="3" />
                <el-option label="5s" :value="5" />
                <el-option label="10s" :value="10" />
                <el-option label="30s" :value="30" />
              </el-select>
              <span class="traffic-speed-summary">
                <span class="up-speed">↑ {{ formatBytes(currentSpeed.up) }}/s</span>
                <span class="down-speed">↓ {{ formatBytes(currentSpeed.down) }}/s</span>
              </span>
            </div>
          </template>
          <TrafficChart :data="speedHistory" mode="speed" />
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never" class="chart-card">
          <template #header>
            <div class="card-header">
              <el-icon :size="20"><DataLine /></el-icon>
              <span>累计流量</span>
              <el-select v-model="totalInterval" size="small" class="interval-select" @change="onTotalIntervalChange">
                <el-option label="1s" :value="1" />
                <el-option label="3s" :value="3" />
                <el-option label="5s" :value="5" />
                <el-option label="10s" :value="10" />
                <el-option label="30s" :value="30" />
              </el-select>
              <span class="traffic-total-summary">
                <span class="up-total">↑ {{ formatBytes(cumulativeTraffic.up) }}</span>
                <span class="down-total">↓ {{ formatBytes(cumulativeTraffic.down) }}</span>
              </span>
            </div>
          </template>
          <TrafficChart :data="totalHistory" mode="total" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { processApi } from '../api/process'
import { useClashStream } from '../composables/useClashStream'
import TrafficChart from '../components/TrafficChart.vue'
import { VideoPlay, Timer, Top, Bottom, DataLine } from '@element-plus/icons-vue'

const { currentSpeed, cumulativeTraffic, speedHistory, totalHistory, startTrafficStream, startConnectionsStream, setSpeedInterval, setTotalInterval } = useClashStream()

const processStatus = ref({ running: false, pid: 0, status: 'stopped', startTime: null })
const speedInterval = ref(3)
const totalInterval = ref(3)
const uptimeTick = ref(0)
let uptimeInterval = null

const onSpeedIntervalChange = (val) => {
  setSpeedInterval(val)
}

const onTotalIntervalChange = (val) => {
  setTotalInterval(val)
}

const formatBytes = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let size = bytes, i = 0
  while (size >= 1024 && i < units.length - 1) { size /= 1024; i++ }
  return `${size.toFixed(1)} ${units[i]}`
}

const formatUpTime = (startTime) => {
  if (!startTime) return '-'
  const diff = Math.floor((Date.now() - new Date(startTime).getTime()) / 1000)
  if (diff < 0) return '-'
  const y = Math.floor(diff / 31536000), mo = Math.floor((diff % 31536000) / 2592000)
  const d = Math.floor((diff % 2592000) / 86400), h = Math.floor((diff % 86400) / 3600), m = Math.floor((diff % 3600) / 60)
  const p = []
  if (y) p.push(`${y}y`); if (mo) p.push(`${mo}m`); if (d) p.push(`${d}d`); if (h) p.push(`${h}h`); if (m) p.push(`${m}m`)
  return p.length ? p.join(' ') : '<1m'
}

const loadData = async () => {
  try {
    const res = await processApi.getStatus()
    if (res.data.success) processStatus.value = res.data.data
  } catch (err) {}
}

onMounted(() => {
  loadData()
  startTrafficStream()
  startConnectionsStream()
  uptimeInterval = setInterval(() => uptimeTick.value++, 60000)
})

onUnmounted(() => {
  if (uptimeInterval) clearInterval(uptimeInterval)
})
</script>

<style scoped>
.dashboard-page { max-width: 1400px; margin: 0 auto; }
.page-title { margin: 0 0 24px 0; color: #303133; font-size: 24px; font-weight: 600; }
.overview-row { margin-bottom: 24px; }
.stat-card { min-height: 100px; }
.stat-content { display: flex; align-items: center; gap: 16px; }
.stat-info { display: flex; flex-direction: column; }
.stat-label { font-size: 13px; color: #909399; }
.stat-value { font-size: 18px; font-weight: 600; color: #303133; }
.chart-card { min-height: 350px; }
.card-header { display: flex; align-items: center; gap: 8px; font-weight: 600; }
.traffic-speed-summary, .traffic-total-summary { margin-left: auto; margin-right: 16px; font-size: 13px; font-weight: normal; display: flex; gap: 12px; }
.up-speed, .up-total { color: #409eff; }
.down-speed, .down-total { color: #67c23a; }
.interval-select { width: 70px; margin-left: 12px; }
</style>
