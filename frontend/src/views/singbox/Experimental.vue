<template>
  <div class="experimental-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <el-icon><Setting /></el-icon>
          <span>Experimental 配置</span>
        </div>
      </template>

      <el-form :model="form" label-width="180px" v-loading="loading">
        <!-- Cache File Section -->
        <el-divider content-position="left">Cache File</el-divider>

        <el-form-item label="启用">
          <el-switch v-model="cacheFileEnabled" />
        </el-form-item>

        <template v-if="cacheFileEnabled">
          <el-form-item label="路径 (path)">
            <el-input v-model="form.cache_file.path" placeholder="留空使用默认" />
          </el-form-item>

          <el-form-item label="缓存 ID (cache_id)">
            <el-input v-model="form.cache_file.cache_id" placeholder="留空使用默认" />
          </el-form-item>
        </template>

        <!-- Customized fork features -->
        <el-divider content-position="left">定制化功能</el-divider>

        <el-form-item label="启用定制化功能">
          <el-switch v-model="customizedEnabled" />
          <div class="form-tip">启用后才会向内核下发 fallback 等 fork 独有功能。</div>
        </el-form-item>

        <!-- Clash API Section -->
        <el-divider content-position="left">Clash API</el-divider>

        <el-form-item label="外部控制器">
          <el-input v-model="form.clash_api.external_controller" placeholder="127.0.0.1:9090" />
        </el-form-item>

        <el-form-item label="UI 目录 (external_ui)">
          <el-input v-model="form.clash_api.external_ui" placeholder="留空使用默认" />
        </el-form-item>

        <el-form-item label="UI 下载地址">
          <el-input v-model="form.clash_api.external_ui_download_url" placeholder="Dashboard 下载地址" />
        </el-form-item>

        <el-form-item label="下载出口">
          <el-select v-model="form.clash_api.external_ui_download_detour" clearable placeholder="选择出口">
            <el-option label="(默认)" value="" />
            <el-option
              v-for="ob in outbounds"
              :key="ob.tag"
              :label="ob.tag"
              :value="ob.tag"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="允许来源">
          <el-select v-model="form.clash_api.access_control_allow_origin" multiple filterable allow-create placeholder="例如: http://localhost:*">
            <el-option label="*" value="*" />
            <el-option label="http://localhost:*" value="http://localhost:*" />
            <el-option label="http://127.0.0.1:*" value="http://127.0.0.1:*" />
          </el-select>
          <div class="form-tip">CORS 允许的来源，* 表示允许所有</div>
        </el-form-item>

        <el-form-item label="允许私网访问">
          <el-switch v-model="clashAPIAllowPrivateNetwork" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="saveConfig" :loading="saving">
            <el-icon><Check /></el-icon>
            保存配置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { singboxApi } from '../../api/singbox'
import { configApi } from '../../api/config'
import { Setting, Check } from '@element-plus/icons-vue'

const form = ref({
  cache_file: {
    path: '',
    cache_id: ''
  },
  clash_api: {
    external_controller: '',
    external_ui: '',
    external_ui_download_url: '',
    external_ui_download_detour: '',
    access_control_allow_origin: [],
    access_control_allow_private_network: false
  }
})

const outbounds = ref([])
const loading = ref(false)
const saving = ref(false)
const customizedEnabled = ref(false)

// Computed helpers for pointer fields
const cacheFileEnabled = computed({
  get: () => form.value.cache_file.enabled !== false,
  set: (val) => { form.value.cache_file.enabled = val }
})

const clashAPIAllowPrivateNetwork = computed({
  get: () => form.value.clash_api.access_control_allow_private_network === true,
  set: (val) => { form.value.clash_api.access_control_allow_private_network = val }
})

const loadConfig = async () => {
  loading.value = true
  try {
    const [expRes, outRes, appRes] = await Promise.all([
      singboxApi.getExperimental(),
      singboxApi.getOutbounds(),
      configApi.get()
    ])
    if (expRes.data.success) {
      const data = expRes.data.data || {}
      form.value = {
        cache_file: {
          enabled: data.cache_file?.enabled,
          path: data.cache_file?.path || '',
          cache_id: data.cache_file?.cache_id || ''
        },
        clash_api: {
          external_controller: data.clash_api?.external_controller || '',
          external_ui: data.clash_api?.external_ui || '',
          external_ui_download_url: data.clash_api?.external_ui_download_url || '',
          external_ui_download_detour: data.clash_api?.external_ui_download_detour || '',
          access_control_allow_origin: data.clash_api?.access_control_allow_origin || [],
          access_control_allow_private_network: data.clash_api?.access_control_allow_private_network
        }
      }
    }
    if (outRes.data.success) {
      outbounds.value = outRes.data.data || []
    }
    if (appRes.data.success) {
      customizedEnabled.value = appRes.data.data?.customizedFeaturesEnabled === true
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
    const res = await singboxApi.updateExperimental(form.value)
    if (res.data.success) {
      await configApi.update({ customizedFeaturesEnabled: customizedEnabled.value })
      ElMessage.success('配置已保存')
    }
  } catch (err) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
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

.form-tip {
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-secondary);
}
</style>
