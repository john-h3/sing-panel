<template>
  <div class="settings-page">
    <el-row :gutter="24">
      <el-col :span="24">
        <h2 class="page-title">系统设置</h2>
      </el-col>
    </el-row>

    <!-- Accelerate Domain -->
    <el-card class="section-card" shadow="never">
      <template #header>
        <div class="card-header">
          <el-icon><Link /></el-icon>
          <span>加速域名配置</span>
        </div>
      </template>

      <el-form :model="form" label-width="120px" v-loading="loading">
        <el-form-item label="加速域名">
          <el-input
            v-model="form.accelerateDomain"
            placeholder="https://mirror.ghproxy.com"
            clearable
          />
          <div class="form-tip">
            配置 GitHub 加速代理，留空则直连。例如: https://mirror.ghproxy.com
          </div>
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
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { configApi } from '../api/config'
import { Link, Check } from '@element-plus/icons-vue'

const form = ref({
  accelerateDomain: ''
})
const loading = ref(false)
const saving = ref(false)

const loadConfig = async () => {
  loading.value = true
  try {
    const res = await configApi.get()
    if (res.data.success) {
      form.value = res.data.data
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
    const res = await configApi.update(form.value)
    if (res.data.success) {
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
.settings-page {
  max-width: 800px;
  margin: 0 auto;
}

.page-title {
  margin: 0 0 24px 0;
  color: #303133;
  font-size: 24px;
  font-weight: 600;
}

.section-card {
  margin-bottom: 24px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.form-tip {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}
</style>
