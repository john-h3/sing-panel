<template>
  <el-card class="section-card" shadow="never">
    <template #header>
      <div class="card-header">
        <el-icon><Monitor /></el-icon>
        <span>Dashboard 配置</span>
        <el-button type="primary" size="small" @click="addDashboard">
          <el-icon><Plus /></el-icon>
          添加
        </el-button>
      </div>
    </template>

    <div v-loading="loading">
      <div v-if="form.dashboards.length === 0" class="empty-dashboards">
        <el-empty description="暂无 Dashboard 配置" :image-size="60" />
      </div>

      <div v-else class="dashboard-list">
        <el-card
          v-for="(dashboard, index) in form.dashboards"
          :key="index"
          class="dashboard-item"
          shadow="hover"
        >
          <div class="dashboard-item-header">
            <el-switch v-model="dashboard.enabled" />
            <el-input
              v-model="dashboard.name"
              placeholder="Dashboard 名称"
              class="name-input"
            />
            <el-button
              type="danger"
              :icon="Delete"
              circle
              size="small"
              @click="removeDashboard(index)"
            />
          </div>
          <el-input
            v-model="dashboard.url"
            placeholder="http://127.0.0.1:9090/ui"
            clearable
          >
            <template #prepend>
              <el-select
                v-model="dashboard._serviceTag"
                placeholder="选择服务"
                clearable
                filterable
                @change="(val) => onServiceSelect(index, val)"
                style="width: 120px"
              >
                <el-option
                  v-for="svc in apiServices"
                  :key="svc.tag"
                  :label="svc.tag"
                  :value="svc.tag"
                />
              </el-select>
            </template>
          </el-input>
        </el-card>
      </div>

      <div class="form-tip" style="margin-top: 12px">
        配置多个 Dashboard 地址，可在监控面板中切换显示。选择服务配置中的 API 面板可自动填充地址。
      </div>
    </div>

    <el-form-item style="margin-top: 16px">
      <el-button type="primary" @click="saveConfig" :loading="saving">
        <el-icon><Check /></el-icon>
        保存配置
      </el-button>
    </el-form-item>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { configApi } from '../../api/config'
import { singboxApi } from '../../api/singbox'
import { Monitor, Check, Plus, Delete } from '@element-plus/icons-vue'

const form = ref({
  dashboards: []
})
const apiServices = ref([])
const loading = ref(false)
const saving = ref(false)

const loadConfig = async () => {
  loading.value = true
  try {
    const [configRes, servicesRes] = await Promise.all([
      configApi.get(),
      singboxApi.getServices()
    ])
    if (configRes.data.success) {
      const dashboards = (configRes.data.data.dashboards || []).map(d => ({
        ...d,
        enabled: d.enabled !== false,
        _serviceTag: ''
      }))
      form.value = { dashboards }
    }
    if (servicesRes.data.success) {
      apiServices.value = (servicesRes.data.data || []).filter(s => s.type === 'api')
    }
  } catch (err) {
    ElMessage.error('加载配置失败')
  } finally {
    loading.value = false
  }
}

const saveConfig = async () => {
  saving.value = true
  try {
    const data = {
      dashboards: form.value.dashboards.map(({ _serviceTag, ...rest }) => rest)
    }
    const res = await configApi.update(data)
    if (res.data.success) {
      ElMessage.success('配置已保存')
    }
  } catch (err) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

const addDashboard = () => {
  form.value.dashboards.push({ name: '', url: '', enabled: true, _serviceTag: '' })
}

const removeDashboard = (index) => {
  form.value.dashboards.splice(index, 1)
}

const onServiceSelect = (index, tag) => {
  if (!tag) return
  const svc = apiServices.value.find(s => s.tag === tag)
  if (svc) {
    const listen = svc.listen || '127.0.0.1'
    const host = (listen === '0.0.0.0' || listen === '::') ? window.location.hostname : listen
    const port = svc.listenPort
    form.value.dashboards[index].url = `http://${host}:${port}/ui`
    form.value.dashboards[index].name = svc.tag
  }
}

onMounted(() => {
  loadConfig()
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

.card-header .el-button {
  margin-left: auto;
}

.form-tip {
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-secondary);
}

.empty-dashboards {
  padding: 20px 0;
}

.dashboard-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.dashboard-item {
  background: var(--bg-card);
  border-color: var(--border-color);
}

.dashboard-item :deep(.el-card__body) {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.dashboard-item-header {
  display: flex;
  gap: 8px;
  align-items: center;
}

.dashboard-item-header .name-input {
  flex: 1;
}
</style>
