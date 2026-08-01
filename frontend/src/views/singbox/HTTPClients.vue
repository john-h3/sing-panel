<template>
  <div class="http-clients-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <el-icon><Connection /></el-icon>
          <span>HTTP 客户端配置</span>
          <el-button type="primary" @click="showAddDialog">
            <el-icon><Plus /></el-icon>
            添加
          </el-button>
        </div>
      </template>

      <el-table :data="httpClients" v-loading="loading" stripe>
        <el-table-column prop="tag" label="标签" width="180" />
        <el-table-column label="详情" min-width="300">
          <template #default="{ row }">
            <span class="detail-text">{{ getClientInfo(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="editClient(row)">
              编辑
            </el-button>
            <el-button type="danger" link size="small" @click="deleteClient(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Add/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="editingClient ? '编辑 HTTP 客户端' : '添加 HTTP 客户端'"
      width="600px"
    >
      <el-form :model="form" label-width="120px" ref="formRef" :rules="rules">
        <el-form-item label="标签" prop="tag">
          <el-input v-model="form.tag" placeholder="例如: my-http-client" />
        </el-form-item>

        <el-divider content-position="left">HTTP 配置</el-divider>

        <el-form-item label="出站 (detour)">
          <el-select v-model="httpDetour" placeholder="选择出站" filterable clearable>
            <el-option
              v-for="ob in enabledOutbounds"
              :key="ob.tag"
              :label="ob.tag"
              :value="ob.tag"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="Host">
          <div class="host-list">
            <div v-for="(h, idx) in httpHostList" :key="idx" class="host-row">
              <el-input v-model="httpHostList[idx]" placeholder="例如: www.example.com" />
              <el-button type="danger" :icon="Delete" circle size="small" @click="removeHost(idx)" />
            </div>
            <el-button type="primary" link @click="addHost">
              <el-icon><Plus /></el-icon>
              添加 Host
            </el-button>
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveClient" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { singboxApi } from '../../api/singbox'
import { Connection, Plus, Delete } from '@element-plus/icons-vue'

const httpClients = ref([])
const enabledOutbounds = ref([])
const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const editingClient = ref(null)
const formRef = ref(null)

const form = ref({
  tag: '',
  options: {}
})

// HTTP specific fields
const httpDetour = ref('')
const httpHostList = ref([])

const rules = {
  tag: [{ required: true, message: '请输入标签', trigger: 'blur' }]
}

const resetHttpFields = () => {
  httpDetour.value = ''
  httpHostList.value = []
}

const addHost = () => {
  httpHostList.value.push('')
}

const removeHost = (idx) => {
  httpHostList.value.splice(idx, 1)
}

const getClientInfo = (client) => {
  const opts = client.options || {}
  const parts = []
  if (opts.detour) parts.push(`出站: ${opts.detour}`)
  if (opts.host?.length) parts.push(`Host: ${opts.host.join(', ')}`)
  return parts.join(' | ') || '默认配置'
}

const loadClients = async () => {
  loading.value = true
  try {
    const [clientsRes, obRes] = await Promise.all([
      singboxApi.getHTTPClients(),
      singboxApi.getOutbounds()
    ])
    if (clientsRes.data.success) {
      httpClients.value = clientsRes.data.data || []
    }
    if (obRes.data.success) {
      enabledOutbounds.value = (obRes.data.data || []).filter(o => o.enabled)
    }
  } catch (err) {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

const showAddDialog = () => {
  editingClient.value = null
  form.value = { tag: '', options: {} }
  resetHttpFields()
  dialogVisible.value = true
}

const editClient = (client) => {
  editingClient.value = client
  form.value = { ...client, options: { ...client.options } }

  const opts = client.options || {}
  httpDetour.value = opts.detour || ''
  httpHostList.value = [...(opts.host || [])]

  dialogVisible.value = true
}

const saveClient = async () => {
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  const opts = {}
  if (httpDetour.value) opts.detour = httpDetour.value
  if (httpHostList.value.length > 0) opts.host = httpHostList.value.filter(h => h.trim())

  const payload = { ...form.value, options: opts }

  saving.value = true
  try {
    if (editingClient.value) {
      payload.id = editingClient.value.id
      await singboxApi.updateHTTPClient(payload)
      ElMessage.success('更新成功')
    } else {
      await singboxApi.addHTTPClient(payload)
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    loadClients()
  } catch (err) {
    ElMessage.error('保存失败: ' + (err.response?.data?.error || err.message))
  } finally {
    saving.value = false
  }
}

const deleteClient = async (client) => {
  try {
    await ElMessageBox.confirm(`确定要删除 "${client.tag}" 吗？`, '确认删除', {
      confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
    })
    await singboxApi.deleteHTTPClient(client.id)
    ElMessage.success('删除成功')
    loadClients()
  } catch (err) {
    if (err !== 'cancel') ElMessage.error('删除失败')
  }
}

onMounted(() => {
  loadClients()
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

.host-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.host-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.host-row .el-input {
  flex: 1;
}
</style>
