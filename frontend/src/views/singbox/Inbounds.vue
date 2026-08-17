<template>
  <div class="inbounds-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <el-icon><Upload /></el-icon>
          <span>Inbound 配置</span>
          <el-button type="primary" @click="showAddDialog">
            <el-icon><Plus /></el-icon>
            添加
          </el-button>
        </div>
      </template>

      <el-table :data="inbounds" v-loading="loading" stripe>
        <el-table-column prop="tag" label="标签" width="150" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="监听地址" width="150">
          <template #default="{ row }">
            <template v-if="row.type !== 'tun' && row.type !== 'tproxy'">
              {{ row.listen || '0.0.0.0' }}:{{ row.listenPort }}
            </template>
            <template v-else-if="row.type === 'tun'">
              {{ (row.options?.address || []).join(', ') || '-' }}
            </template>
            <template v-else-if="row.type === 'tproxy'">
              {{ row.listen || '0.0.0.0' }}:{{ row.listenPort }}
              <el-tag size="small" style="margin-left:4px">{{ row.options?.network ? '仅 ' + row.options.network.toUpperCase() : 'TCP+UDP' }}</el-tag>
            </template>
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
            <el-button type="primary" link size="small" @click="editInbound(row)">
              编辑
            </el-button>
            <el-button type="danger" link size="small" @click="deleteInbound(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Add/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="editingInbound ? '编辑 Inbound' : '添加 Inbound'"
      width="600px"
    >
      <el-form :model="form" label-width="100px" ref="formRef" :rules="rules">
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" placeholder="选择类型" @change="onTypeChange">
            <el-option
              v-for="item in inboundTypes"
              :key="item.type"
              :label="item.name"
              :value="item.type"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="标签" prop="tag">
          <el-input v-model="form.tag" placeholder="例如: http-in" />
        </el-form-item>

        <template v-if="form.type !== 'tun' && form.type !== 'tproxy'">
          <el-form-item label="监听地址">
            <el-input v-model="form.listen" placeholder="0.0.0.0" />
          </el-form-item>

          <el-form-item label="监听端口" prop="listenPort">
            <el-input-number v-model="form.listenPort" :min="1" :max="65535" />
          </el-form-item>
        </template>

        <template v-if="form.type === 'tproxy'">
          <el-form-item label="监听地址">
            <el-input v-model="form.listen" placeholder="0.0.0.0" />
          </el-form-item>

          <el-form-item label="监听端口" prop="listenPort">
            <el-input-number v-model="form.listenPort" :min="1" :max="65535" />
          </el-form-item>

          <el-form-item label="网络">
            <el-select v-model="tproxyNetwork" placeholder="自动 (TCP+UDP)" clearable>
              <el-option label="TCP" value="tcp" />
              <el-option label="UDP" value="udp" />
            </el-select>
          </el-form-item>

          <el-form-item label="UDP 超时">
            <el-input v-model="tproxyUdpTimeout" placeholder="60s（默认），留空则使用全局默认" />
            <div class="form-tip">UDP NAT 会话超时，调小可降低内存/goroutine/GC 开销</div>
          </el-form-item>
        </template>

        <!-- TUN specific fields -->
        <template v-if="form.type === 'tun'">
          <el-divider content-position="left">TUN 配置</el-divider>

          <el-form-item label="网段地址" required>
            <div class="tun-address-list">
              <div v-for="(addr, idx) in tunAddressList" :key="idx" class="tun-address-row">
                <el-input v-model="tunAddressList[idx]" placeholder="例: 172.19.0.1/30 或 fdfe:dcba:9876::1/126" />
                <el-button type="danger" link @click="removeTunAddress(idx)" :disabled="tunAddressList.length <= 1">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button type="primary" link @click="addTunAddress">
                <el-icon><Plus /></el-icon>
                添加地址
              </el-button>
            </div>
          </el-form-item>

          <el-form-item label="MTU">
            <el-input-number v-model="tunMtu" :min="576" :max="65535" />
          </el-form-item>

          <el-form-item label="Stack">
            <el-select v-model="tunStack">
              <el-option label="system" value="system" />
              <el-option label="gvisor" value="gvisor" />
              <el-option label="mixed" value="mixed" />
            </el-select>
          </el-form-item>

          <el-form-item label="自动路由">
            <el-switch v-model="tunAutoRoute" />
          </el-form-item>

          <el-form-item label="自动重定向">
            <el-switch v-model="tunAutoRedirect" />
          </el-form-item>

          <el-form-item label="严格路由">
            <el-switch v-model="tunStrictRoute" />
          </el-form-item>
        </template>

        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveInbound" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { singboxApi } from '../../api/singbox'
import { Upload, Plus, Delete } from '@element-plus/icons-vue'

const inbounds = ref([])
const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const editingInbound = ref(null)
const inboundTypes = ref([])
const formRef = ref(null)

const form = ref({
  type: 'http',
  tag: '',
  listen: '0.0.0.0',
  listenPort: 8080,
  enabled: true,
  options: {}
})

// TUN specific fields
const tunAddressList = ref(['172.19.0.1/30'])

// TProxy specific fields
const tproxyNetwork = ref('')
const tproxyUdpTimeout = ref('')
const tunMtu = ref(9000)
const tunStack = ref('gvisor')
const tunAutoRoute = ref(true)
const tunAutoRedirect = ref(true)
const tunStrictRoute = ref(true)

const rules = {
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  tag: [{ required: true, message: '请输入标签', trigger: 'blur' }],
  listenPort: [
    {
      required: true,
      trigger: 'blur',
      validator: (rule, value, callback) => {
        if (form.value.type === 'tun') {
          callback()
        } else if (value === undefined || value === null || value === '') {
          callback(new Error('请输入端口'))
        } else {
          callback()
        }
      }
    }
  ]
}

const addTunAddress = () => {
  tunAddressList.value.push('')
}

const removeTunAddress = (idx) => {
  if (tunAddressList.value.length > 1) {
    tunAddressList.value.splice(idx, 1)
  }
}

const loadInbounds = async () => {
  loading.value = true
  try {
    const res = await singboxApi.getInbounds()
    if (res.data.success) {
      inbounds.value = res.data.data || []
    }
  } catch (err) {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

const loadTypes = async () => {
  try {
    const res = await singboxApi.getInboundTypes()
    if (res.data.success) {
      inboundTypes.value = res.data.data || []
    }
  } catch (err) {
    console.error('Failed to load types:', err)
  }
}

const showAddDialog = () => {
  editingInbound.value = null
  form.value = {
    type: 'http',
    tag: '',
    listen: '0.0.0.0',
    listenPort: 8080,
    enabled: true,
    options: {}
  }
  resetTunFields()
  tproxyNetwork.value = ''
  tproxyUdpTimeout.value = '60s'
  dialogVisible.value = true
}

const resetTunFields = () => {
  tunAddressList.value = ['172.19.0.1/30']
  tunMtu.value = 9000
  tunStack.value = 'gvisor'
  tunAutoRoute.value = true
  tunAutoRedirect.value = true
  tunStrictRoute.value = true
}

const editInbound = (inbound) => {
  editingInbound.value = inbound
  form.value = { ...inbound, options: { ...inbound.options } }

  if (inbound.type === 'tun' && inbound.options) {
    tunAddressList.value = [...(inbound.options.address || ['172.19.0.1/30'])]
    tunMtu.value = inbound.options.mtu ?? 9000
    tunStack.value = inbound.options.stack || 'gvisor'
    tunAutoRoute.value = inbound.options.auto_route !== false
    tunAutoRedirect.value = inbound.options.auto_redirect !== false
    tunStrictRoute.value = inbound.options.strict_route !== false
  } else {
    resetTunFields()
  }

  if (inbound.type === 'tproxy') {
    tproxyNetwork.value = inbound.options?.network || ''
    tproxyUdpTimeout.value = inbound.options?.udp_timeout || '60s'
  } else {
    tproxyNetwork.value = ''
    tproxyUdpTimeout.value = ''
  }

  dialogVisible.value = true
}

const onTypeChange = (type) => {
  const portMap = {
    http: 8080,
    socks: 1080,
    mixed: 2080,
    tun: 0,
    shadowsocks: 8388,
    tproxy: 1080
  }
  form.value.listenPort = portMap[type] || 8080
  if (type === 'tun') {
    resetTunFields()
  }
}

const saveInbound = async () => {
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  const payload = { ...form.value }

  if (payload.type === 'tun') {
    delete payload.listen
    delete payload.listenPort

    const validAddresses = tunAddressList.value.filter(a => a.trim())
    if (validAddresses.length === 0) {
      ElMessage.warning('请至少添加一个网段地址')
      return
    }

    payload.options = {
      ...payload.options,
      address: validAddresses,
      mtu: tunMtu.value,
      stack: tunStack.value,
      auto_route: tunAutoRoute.value,
      auto_redirect: tunAutoRedirect.value,
      strict_route: tunStrictRoute.value,
    }
  } else if (payload.type === 'tproxy') {
    if (tproxyNetwork.value) {
      payload.options = { ...payload.options, network: tproxyNetwork.value }
    } else {
      if (payload.options) {
        delete payload.options.network
      }
    }
    if (tproxyUdpTimeout.value) {
      payload.options = { ...payload.options, udp_timeout: tproxyUdpTimeout.value }
    } else {
      if (payload.options) {
        delete payload.options.udp_timeout
      }
    }
  }

  saving.value = true
  try {
    if (editingInbound.value) {
      await singboxApi.updateInbound(payload)
      ElMessage.success('更新成功')
    } else {
      await singboxApi.addInbound(payload)
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    loadInbounds()
  } catch (err) {
    ElMessage.error('保存失败: ' + (err.response?.data?.error || err.message))
  } finally {
    saving.value = false
  }
}

const toggleEnabled = async (inbound) => {
  try {
    await singboxApi.updateInbound(inbound)
    ElMessage.success('状态已更新')
  } catch (err) {
    inbound.enabled = !inbound.enabled
    ElMessage.error('更新失败')
  }
}

const deleteInbound = async (inbound) => {
  try {
    await ElMessageBox.confirm(`确定要删除 "${inbound.tag}" 吗？`, '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await singboxApi.deleteInbound(inbound.id)
    ElMessage.success('删除成功')
    loadInbounds()
  } catch (err) {
    if (err !== 'cancel') {
      const message = err.response?.data?.error || ''
      if (message.includes('referenced')) {
        ElMessage.error('该入站已被路由规则使用，请先移除引用')
      } else {
        ElMessage.error('删除失败: ' + (message || err.message || '未知错误'))
      }
    }
  }
}

onMounted(() => {
  loadInbounds()
  loadTypes()
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

.tun-address-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.tun-address-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.form-tip {
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-secondary);
}
</style>
