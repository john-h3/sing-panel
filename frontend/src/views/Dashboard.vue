<template>
  <div class="dashboard-page">
    <el-row :gutter="24">
      <el-col :span="24">
        <h2 class="page-title">监控面板</h2>
      </el-col>
    </el-row>

    <template v-if="dashboards.length > 0">
      <el-tabs v-model="activeDashboard" type="border-card" class="dashboard-tabs">
        <el-tab-pane
          v-for="(dashboard, index) in dashboards"
          :key="index"
          :label="dashboard.name || `Dashboard ${index + 1}`"
          :name="String(index)"
        >
          <el-card shadow="never" class="dashboard-card">
            <template #header>
              <div class="card-header">
                <el-icon><Monitor /></el-icon>
                <span>{{ dashboard.name || `Dashboard ${index + 1}` }}</span>
                <span class="url-text">{{ dashboard.url }}</span>
                <el-button type="primary" size="small" @click="refreshIframe(index)">
                  <el-icon><RefreshRight /></el-icon>
                  刷新
                </el-button>
              </div>
            </template>
            <div class="iframe-container">
              <iframe
                :ref="(el) => setIframeRef(index, el)"
                :src="dashboard.url"
                class="dashboard-iframe"
                frameborder="0"
              />
            </div>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </template>

    <template v-else>
      <el-row :gutter="24">
        <el-col :span="24">
          <el-card shadow="never" class="dashboard-card">
            <div class="empty-state">
              <el-empty description="未配置 Dashboard">
                <el-button type="primary" @click="goToSettings">前往设置</el-button>
              </el-empty>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { configApi } from '../api/config'
import { Monitor, RefreshRight } from '@element-plus/icons-vue'

const router = useRouter()
const dashboards = ref([])
const activeDashboard = ref('0')
const iframeRefs = ref({})

const setIframeRef = (index, el) => {
  if (el) {
    iframeRefs.value[index] = el
  } else {
    delete iframeRefs.value[index]
  }
}

const refreshIframe = (index) => {
  const iframe = iframeRefs.value[index]
  if (iframe && dashboards.value[index]) {
    iframe.src = dashboards.value[index].url
  }
}

const goToSettings = () => {
  router.push('/settings/dashboard')
}

const loadData = async () => {
  try {
    const res = await configApi.get()
    if (res.data.success) {
      dashboards.value = (res.data.data.dashboards || []).filter(d => d.enabled !== false)
      if (dashboards.value.length > 0) {
        activeDashboard.value = '0'
      }
    }
  } catch (err) {}
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.dashboard-page { max-width: 1400px; margin: 0 auto; display: flex; flex-direction: column; height: calc(100vh - 48px); }
.page-title { margin: 0 0 24px 0; color: var(--text-primary); font-size: 24px; font-weight: 600; flex-shrink: 0; }
.dashboard-tabs { background: var(--bg-card); flex: 1; display: flex; flex-direction: column; }
.dashboard-tabs :deep(.el-tabs__content) { flex: 1; overflow: hidden; }
.dashboard-tabs :deep(.el-tab-pane) { height: 100%; }
.dashboard-card { background: var(--bg-card); border-color: var(--border-color); border: none; height: 100%; display: flex; flex-direction: column; }
.dashboard-card :deep(.el-card__body) { flex: 1; overflow: hidden; padding: 0; }
.card-header { display: flex; align-items: center; gap: 8px; font-weight: 600; color: var(--text-primary); }
.card-header .el-button { margin-left: auto; }
.url-text { font-size: 12px; color: var(--text-secondary); margin-left: 8px; }
.iframe-container { width: 100%; height: 100%; border: 1px solid var(--border-color); border-radius: 4px; overflow: hidden; }
.dashboard-iframe { width: 100%; height: 100%; border: none; }
.empty-state { display: flex; align-items: center; justify-content: center; min-height: 500px; }
</style>
