<template>
  <div class="services-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <el-icon><Service /></el-icon>
          <span>服务配置</span>
          <el-button type="primary" @click="showAddDialog">
            <el-icon><Plus /></el-icon>
            添加
          </el-button>
        </div>
      </template>

      <el-table :data="services" v-loading="loading" stripe>
        <el-table-column prop="tag" label="标签" width="180" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="监听地址" width="150">
          <template #default="{ row }">
            {{ row.listen || '0.0.0.0' }}:{{ row.listenPort }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              @change="toggleEnabled(row)"
              size="small"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="editService(row)">
              编辑
            </el-button>
            <el-button type="danger" link size="small" @click="deleteService(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Add/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="editingService ? '编辑服务' : '添加服务'"
      width="600px"
    >
      <el-form :model="form" label-width="100px" ref="formRef" :rules="rules">
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" placeholder="选择类型">
            <el-option label="API 控制面板" value="api" />
          </el-select>
        </el-form-item>

        <el-form-item label="标签" prop="tag">
          <el-input v-model="form.tag" placeholder="例如: api" />
        </el-form-item>

        <el-form-item label="监听地址">
          <el-input v-model="form.listen" placeholder="0.0.0.0" />
        </el-form-item>

        <el-form-item label="监听端口">
          <el-input-number v-model="form.listenPort" :min="1" :max="65535" />
        </el-form-item>

        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>

        <!-- API specific fields -->
        <template v-if="form.type === 'api'">
          <el-divider content-position="left">API 配置</el-divider>

          <el-form-item label="密钥 (secret)">
            <el-input v-model="apiSecret" placeholder="留空则无认证" show-password />
          </el-form-item>

          <el-form-item label="允许来源">
            <el-select v-model="apiAllowOrigin" multiple filterable allow-create placeholder="例如: http://localhost:*">
              <el-option label="*" value="*" />
              <el-option label="http://localhost:*" value="http://localhost:*" />
              <el-option label="http://127.0.0.1:*" value="http://127.0.0.1:*" />
            </el-select>
            <div class="form-tip">CORS 允许的来源，* 表示允许所有</div>
          </el-form-item>

          <el-form-item label="允许私网">
            <el-switch v-model="apiAllowPrivateNetwork" />
            <div class="form-tip">允许局域网访问 API</div>
          </el-form-item>

          <el-divider content-position="left">Dashboard 配置</el-divider>

          <el-form-item label="启用 Dashboard">
            <el-switch v-model="apiDashboardEnabled" />
          </el-form-item>

          <template v-if="apiDashboardEnabled">
            <el-form-item label="HTTP 客户端">
              <el-select v-model="apiHttpClient" placeholder="选择 HTTP 客户端" filterable clearable>
                <el-option
                  v-for="hc in httpClientList"
                  :key="hc.tag"
                  :label="hc.tag"
                  :value="hc.tag"
                />
              </el-select>
              <div class="form-tip">用于下载 Dashboard 的 HTTP 客户端，留空使用默认</div>
            </el-form-item>

            <el-form-item label="下载地址">
              <el-input v-model="apiDashboardDownloadURL" placeholder="Dashboard 下载地址" />
              <div class="form-tip">Dashboard 静态文件下载地址</div>
            </el-form-item>

            <el-form-item label="下载目录">
              <el-input v-model="apiDashboardPath" placeholder="留空使用默认路径" />
            </el-form-item>

            <el-form-item label="更新间隔">
              <el-input v-model="apiDashboardUpdateInterval" placeholder="例如: 24h" />
              <div class="form-tip">Dashboard 自动更新间隔</div>
            </el-form-item>
          </template>
        </template>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveService" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { singboxApi } from '../../api/singbox'
import { Service, Plus } from '@element-plus/icons-vue'

const services = ref([])
const httpClientList = ref([])
const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const editingService = ref(null)
const formRef = ref(null)

const form = ref({
  type: 'api',
  tag: '',
  listen: '0.0.0.0',
  listenPort: 9090,
  enabled: true,
  options: {}
})

// API specific fields
const apiHttpClient = ref('')
const apiSecret = ref('')
const apiAllowOrigin = ref([])
const apiAllowPrivateNetwork = ref(false)
const apiDashboardEnabled = ref(true)
const apiDashboardDownloadURL = ref('')
const apiDashboardPath = ref('')
const apiDashboardUpdateInterval = ref('')

const rules = {
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  tag: [{ required: true, message: '请输入标签', trigger: 'blur' }]
}

const resetApiFields = () => {
  apiHttpClient.value = ''
  apiSecret.value = ''
  apiAllowOrigin.value = []
  apiAllowPrivateNetwork.value = false
  apiDashboardEnabled.value = true
  apiDashboardDownloadURL.value = ''
  apiDashboardPath.value = ''
  apiDashboardUpdateInterval.value = ''
}

const getApiInfo = (svc) => {
  const opts = svc.options || {}
  const parts = []
  if (opts.secret) parts.push('secret: ***')
  if (opts.access_control_allow_private_network) parts.push('允许私网')
  const dashboard = opts.dashboard || {}
  if (dashboard.enabled !== false) parts.push('Dashboard 已启用')
  return parts.join(' | ') || '默认配置'
}

const loadServices = async () => {
  loading.value = true
  try {
    const [svcRes, hcRes] = await Promise.all([
      singboxApi.getServices(),
      singboxApi.getHTTPClients()
    ])
    if (svcRes.data.success) {
      services.value = svcRes.data.data || []
    }
    if (hcRes.data.success) {
      httpClientList.value = hcRes.data.data || []
    }
  } catch (err) {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

const showAddDialog = () => {
  editingService.value = null
  form.value = { type: 'api', tag: '', listen: '0.0.0.0', listenPort: 9090, enabled: true, options: {} }
  resetApiFields()
  dialogVisible.value = true
}

const editService = (svc) => {
  editingService.value = svc
  form.value = { ...svc, options: { ...svc.options } }

  if (svc.type === 'api' && svc.options) {
    apiSecret.value = svc.options.secret || ''
    apiAllowOrigin.value = [...(svc.options.access_control_allow_origin || [])]
    apiAllowPrivateNetwork.value = svc.options.access_control_allow_private_network || false
    const dashboard = svc.options.dashboard || {}
    apiDashboardEnabled.value = dashboard.enabled !== false
    apiHttpClient.value = dashboard.http_client || ''
    apiDashboardDownloadURL.value = dashboard.download_url || ''
    apiDashboardPath.value = dashboard.path || ''
    apiDashboardUpdateInterval.value = dashboard.update_interval || ''
  } else {
    resetApiFields()
  }

  dialogVisible.value = true
}

const saveService = async () => {
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  const payload = { ...form.value }

  if (payload.type === 'api') {
    const opts = {}
    if (apiSecret.value) opts.secret = apiSecret.value
    if (apiAllowOrigin.value.length > 0) opts.access_control_allow_origin = apiAllowOrigin.value
    if (apiAllowPrivateNetwork.value) opts.access_control_allow_private_network = true

    const dashboard = {}
    dashboard.enabled = apiDashboardEnabled.value
    if (apiHttpClient.value) dashboard.http_client = apiHttpClient.value
    if (apiDashboardDownloadURL.value) dashboard.download_url = apiDashboardDownloadURL.value
    if (apiDashboardPath.value) dashboard.path = apiDashboardPath.value
    if (apiDashboardUpdateInterval.value) dashboard.update_interval = apiDashboardUpdateInterval.value
    if (Object.keys(dashboard).length > 0) opts.dashboard = dashboard

    payload.options = opts
  }

  saving.value = true
  try {
    if (editingService.value) {
      payload.id = editingService.value.id
      await singboxApi.updateService(payload)
      ElMessage.success('更新成功')
    } else {
      await singboxApi.addService(payload)
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    loadServices()
  } catch (err) {
    ElMessage.error('保存失败: ' + (err.response?.data?.error || err.message))
  } finally {
    saving.value = false
  }
}

const toggleEnabled = async (svc) => {
  try {
    await singboxApi.updateService(svc)
    ElMessage.success('状态已更新')
  } catch (err) {
    svc.enabled = !svc.enabled
    ElMessage.error('更新失败')
  }
}

const deleteService = async (svc) => {
  try {
    await ElMessageBox.confirm(`确定要删除 "${svc.tag}" 吗？`, '确认删除', {
      confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
    })
    await singboxApi.deleteService(svc.id)
    ElMessage.success('删除成功')
    loadServices()
  } catch (err) {
    if (err !== 'cancel') ElMessage.error('删除失败')
  }
}

onMounted(() => {
  loadServices()
})
</script>

<style scoped>
.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.card-header .el-button {
  margin-left: auto;
}

.form-tip {
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-secondary);
}

.detail-text {
  font-size: 13px;
  color: var(--text-secondary);
}
</style>
