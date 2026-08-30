<template>
  <div class="dns-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <el-icon><Coin /></el-icon>
          <span>DNS 配置</span>
          <el-button type="primary" @click="saveDNS" :loading="saving">
            <el-icon><Check /></el-icon>
            保存
          </el-button>
        </div>
      </template>

      <el-form :model="form" label-width="140px" ref="formRef" :rules="rules" v-loading="loading">
        <!-- Global Settings -->
        <el-divider content-position="left">全局设置</el-divider>

        <el-form-item label="默认 DNS (final)">
          <el-select v-model="form.final" placeholder="选择默认 DNS 服务器" filterable clearable>
            <el-option
              v-for="s in form.servers"
              :key="s.tag"
              :label="s.tag"
              :value="s.tag"
            />
          </el-select>
          <div class="form-tip">未匹配任何规则时使用的 DNS 服务器 tag</div>
        </el-form-item>

        <el-form-item label="策略 (strategy)">
          <el-select v-model="form.strategy" placeholder="选择策略" clearable>
            <el-option label="prefer_ipv4" value="prefer_ipv4" />
            <el-option label="prefer_ipv6" value="prefer_ipv6" />
            <el-option label="ipv4_only" value="ipv4_only" />
            <el-option label="ipv6_only" value="ipv6_only" />
          </el-select>
        </el-form-item>

        <el-form-item label="超时 (timeout)">
          <el-input v-model="form.timeout" placeholder="例如: 1s" />
        </el-form-item>

        <el-form-item label="缓存容量">
          <el-input-number v-model="form.cache_capacity" :min="0" />
        </el-form-item>

        <el-form-item label="客户端子网">
          <el-input v-model="form.client_subnet" placeholder="例如: 1.1.1.1" />
        </el-form-item>

        <el-form-item label="禁用缓存">
          <el-switch v-model="form.disable_cache" />
        </el-form-item>

        <el-form-item label="禁用过期">
          <el-switch v-model="form.disable_expire" />
        </el-form-item>

        <el-form-item label="独立缓存">
          <el-switch v-model="form.independent_cache" />
        </el-form-item>

        <el-form-item label="乐观缓存">
          <el-switch v-model="optimisticEnabled" />
        </el-form-item>

        <el-form-item label="反向映射">
          <el-switch v-model="form.reverse_mapping" />
        </el-form-item>

        <!-- FakeIP -->
        <el-divider content-position="left">FakeIP</el-divider>

        <el-form-item label="启用 FakeIP">
          <el-switch v-model="fakeipEnabled" />
        </el-form-item>

        <template v-if="fakeipEnabled">
          <el-form-item label="FakeIP 模式">
            <el-select v-model="form.fakeip.mode" placeholder="选择模式">
              <el-option label="strict" value="strict" />
              <el-option label="ip" value="ip" />
            </el-select>
          </el-form-item>
          <el-form-item label="FakeIP 池">
            <el-input v-model="form.fakeip.ips" placeholder="例如: 198.18.0.0/15" />
          </el-form-item>
          <el-form-item label="FakeIP 排除">
            <el-input v-model="form.fakeip.excludes" type="textarea" :rows="3" placeholder="每行一条域名，支持通配符" />
          </el-form-item>
        </template>

        <!-- DNS Servers -->
        <el-divider content-position="left">DNS 服务器</el-divider>

        <div class="list-section">
          <div v-for="(server, index) in form.servers" :key="index" class="list-item">
            <el-row :gutter="12" align="middle">
              <el-col :span="6">
                <el-select v-model="server.type" placeholder="类型">
                  <el-option label="local" value="local" />
                </el-select>
              </el-col>
              <el-col :span="14">
                <el-form-item :prop="`servers.${index}.tag`" :rules="serverTagRules">
                  <el-input v-model="server.tag" placeholder="tag" />
                </el-form-item>
              </el-col>
              <el-col :span="4">
                <el-button type="danger" :icon="Delete" circle size="small" @click="removeServer(index)" />
              </el-col>
            </el-row>
          </div>
          <el-button type="primary" link @click="addServer" class="add-btn">
            <el-icon><Plus /></el-icon>
            添加 DNS 服务器
          </el-button>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { singboxApi } from '../../api/singbox'
import { Coin, Check, Plus, Delete } from '@element-plus/icons-vue'

const loading = ref(false)
const saving = ref(false)
const formRef = ref(null)
const optimisticEnabled = ref(false)
const fakeipEnabled = ref(false)

const defaultForm = () => ({
  servers: [],
  final: '',
  strategy: '',
  disable_cache: false,
  disable_expire: false,
  independent_cache: false,
  cache_capacity: 0,
  optimistic: false,
  timeout: '',
  reverse_mapping: false,
  client_subnet: '',
  fakeip: { mode: '', ips: '', excludes: '' }
})

const form = ref(defaultForm())

const serverTagRules = [
  { required: true, message: '请输入 DNS 服务器标签', trigger: 'blur' }
]

const rules = {}

const loadDNS = async () => {
  loading.value = true
  try {
    const dnsRes = await singboxApi.getDNS()
    if (dnsRes.data.success && dnsRes.data.data) {
      const d = dnsRes.data.data
      optimisticEnabled.value = d.optimistic === true || (typeof d.optimistic === 'object' && d.optimistic !== null)
      fakeipEnabled.value = d.fakeip && Object.keys(d.fakeip).length > 0
      form.value = {
        servers: d.servers || [],
        final: d.final || '',
        strategy: d.strategy || '',
        disable_cache: d.disable_cache || false,
        disable_expire: d.disable_expire || false,
        independent_cache: d.independent_cache || false,
        cache_capacity: d.cache_capacity || 0,
        optimistic: d.optimistic || false,
        timeout: d.timeout || '',
        reverse_mapping: d.reverse_mapping || false,
        client_subnet: d.client_subnet || '',
        fakeip: d.fakeip || { mode: '', ips: '', excludes: '' }
      }
    }
  } catch (err) {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

const addServer = () => {
  form.value.servers.push({ type: 'local', tag: '' })
}

const removeServer = (index) => {
  form.value.servers.splice(index, 1)
}

const saveDNS = async () => {
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  const servers = form.value.servers.map(server => ({
    ...server,
    tag: String(server.tag || '').trim()
  }))
  const duplicateTags = [...new Set(
    servers.filter((server, index) => servers.some((other, otherIndex) =>
      otherIndex < index && other.tag === server.tag
    )).map(server => server.tag)
  )]
  if (duplicateTags.length > 0) {
    try {
      await ElMessageBox.confirm(
        `DNS 服务器标签「${duplicateTags.join('、')}」已存在，是否覆盖旧配置？`,
        '名称重复',
        { confirmButtonText: '覆盖', cancelButtonText: '取消', type: 'warning' }
      )
    } catch (err) {
      if (err === 'cancel' || err === 'close') return
      throw err
    }
    const seen = new Set()
    const deduplicated = []
    for (let i = servers.length - 1; i >= 0; i -= 1) {
      if (!seen.has(servers[i].tag)) {
        seen.add(servers[i].tag)
        deduplicated.unshift(servers[i])
      }
    }
    servers.splice(0, servers.length, ...deduplicated)
  }

  saving.value = true
  try {
    const data = {
      ...form.value,
      servers,
      optimistic: optimisticEnabled.value,
      fakeip: fakeipEnabled.value ? form.value.fakeip : {}
    }
    await singboxApi.updateDNS(data)
    ElMessage.success('DNS 配置已保存')
  } catch (err) {
    ElMessage.error('保存失败: ' + (err.response?.data?.error || err.message))
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadDNS()
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

.list-section {
  margin-bottom: 16px;
}

.list-item {
  margin-bottom: 12px;
  padding: 12px;
  background: var(--bg-page);
  border-radius: 6px;
}

.add-btn {
  margin-top: 4px;
}
</style>
