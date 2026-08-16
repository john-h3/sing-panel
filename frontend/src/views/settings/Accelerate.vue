<template>
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

      <el-form-item label="加速匹配">
        <div class="match-list">
          <div
            v-for="(domain, index) in form.accelerateDomains"
            :key="index"
            class="match-item"
          >
            <el-input
              v-model="form.accelerateDomains[index]"
              placeholder="github.com"
              @blur="validateDomain(index)"
            />
            <el-button
              type="danger"
              :icon="Delete"
              circle
              size="small"
              @click="removeDomain(index)"
            />
          </div>
        </div>
        <el-button type="primary" link @click="addDomain" class="add-btn">
          <el-icon><Plus /></el-icon>
          添加匹配域名
        </el-button>
        <div class="form-tip">
           传给内核的 http(s) 地址中，域名匹配（精确或子域名后缀）的将使用加速域名拼接，如 github.com 可匹配 github.com 与 www.github.com；GitHub 的 raw.githubusercontent.com 等域名需额外添加 githubusercontent.com。未配置匹配域名时不会匹配任何地址。
        </div>
        <div class="common-domains">
          <span class="common-label">常用域名:</span>
          <el-tag
            v-for="domain in commonDomains"
            :key="domain"
            class="common-tag"
            @click="addCommonDomain(domain)"
          >
            {{ domain }}
          </el-tag>
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
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { configApi } from '../../api/config'
import { Link, Check, Plus, Delete } from '@element-plus/icons-vue'

const form = ref({
  accelerateDomain: '',
  accelerateDomains: []
})
const loading = ref(false)
const saving = ref(false)
const commonDomains = ['github.com', 'raw.githubusercontent.com']

const loadConfig = async () => {
  loading.value = true
  try {
    const res = await configApi.get()
    if (res.data.success) {
      form.value = {
        accelerateDomain: res.data.data.accelerateDomain || '',
        accelerateDomains: res.data.data.accelerateDomains || []
      }
    }
  } catch (err) {
    ElMessage.error('加载配置失败')
  } finally {
    loading.value = false
  }
}

const saveConfig = async () => {
  form.value.accelerateDomains = form.value.accelerateDomains
    .map(domain => normalizeDomain(domain))
    .filter(Boolean)
    .filter((domain, index, domains) => domains.indexOf(domain) === index)
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

const addDomain = () => {
  form.value.accelerateDomains.push('')
}

const removeDomain = (index) => {
  form.value.accelerateDomains.splice(index, 1)
}

const normalizeDomain = (d) => (d || '').trim().toLowerCase()

const hasDuplicate = (domain, excludeIndex) => {
  const norm = normalizeDomain(domain)
  return form.value.accelerateDomains.some((d, i) => i !== excludeIndex && normalizeDomain(d) === norm)
}

const addCommonDomain = (domain) => {
  if (!form.value.accelerateDomains.some(d => normalizeDomain(d) === normalizeDomain(domain))) {
    form.value.accelerateDomains.push(domain)
  }
}

const validateDomain = (index) => {
  const current = form.value.accelerateDomains[index]
  if (!current || !current.trim()) {
    form.value.accelerateDomains.splice(index, 1)
    return
  }
  if (hasDuplicate(current, index)) {
    ElMessage.warning('该域名已存在')
    form.value.accelerateDomains.splice(index, 1)
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

.match-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.match-item {
  display: flex;
  gap: 8px;
  align-items: center;
}

.match-item .el-input {
  flex: 1;
}

.add-btn {
  margin-top: 4px;
}

.common-domains {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
}

.common-label {
  font-size: 12px;
  color: var(--text-secondary);
}

.common-tag {
  cursor: pointer;
}
</style>
