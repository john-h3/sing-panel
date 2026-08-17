<template>
  <div class="route-rules-page">
    <!-- Route Config -->
    <el-card shadow="never" class="config-card">
      <template #header>
        <div class="card-header">
          <el-icon><Setting /></el-icon>
          <span>路由配置</span>
          <el-button type="primary" size="small" @click="saveRouteConfig" :loading="savingConfig">
            <el-icon><Check /></el-icon>
            保存
          </el-button>
        </div>
      </template>
      <el-form :model="routeConfig" label-width="100px">
        <el-form-item label="默认出站 (final)">
          <el-select v-model="routeConfig.final" placeholder="选择默认出站" filterable clearable>
            <el-option
              v-for="ob in enabledOutbounds"
              :key="ob.tag"
              :label="ob.tag"
              :value="ob.tag"
            />
          </el-select>
          <div class="form-tip">未匹配任何路由规则时使用的默认出站</div>
        </el-form-item>

        <el-form-item label="默认 HTTP 客户端">
          <el-select v-model="routeConfig.default_http_client" placeholder="选择 HTTP 客户端" filterable clearable>
            <el-option
              v-for="hc in httpClientList"
              :key="hc.tag"
              :label="hc.tag"
              :value="hc.tag"
            />
          </el-select>
          <div class="form-tip">路由默认使用的 HTTP 客户端，留空使用默认</div>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Route Rules -->
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <el-icon><Share /></el-icon>
          <span>路由规则配置</span>
          <el-button type="primary" @click="showAddDialog">
            <el-icon><Plus /></el-icon>
            添加
          </el-button>
        </div>
      </template>

      <el-table
        ref="tableRef"
        :data="routeRules"
        v-loading="loading"
        stripe
        @selection-change="onSelectionChange"
      >
        <el-table-column type="selection" width="45" />
        <el-table-column label="#" width="60">
          <template #default="{ row, $index }">
            <span
              class="drag-handle"
              draggable="true"
              @dragstart.stop="onDragStart($event, $index)"
            >{{ $index + 1 }}</span>
          </template>
        </el-table-column>
        <el-table-column label="动作" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="getActionTag(row.action)">{{ getActionLabel(row.action) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="规则集" min-width="200">
          <template #default="{ row }">
            <el-tag
              v-for="tag in row.rule_set"
              :key="tag"
              size="small"
              type="info"
              class="rule-tag"
            >{{ tag }}</el-tag>
            <span v-if="!row.rule_set || row.rule_set.length === 0" class="empty-text">-</span>
          </template>
        </el-table-column>
        <el-table-column label="入站" min-width="150">
          <template #default="{ row }">
            <el-tag
              v-for="tag in row.inbound"
              :key="tag"
              size="small"
              type="info"
              class="rule-tag"
            >{{ tag }}</el-tag>
            <span v-if="!row.inbound || row.inbound.length === 0" class="empty-text">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="outbound" label="出站" width="150">
          <template #default="{ row }">
            <el-tag v-if="row.outbound" size="small" type="success">{{ row.outbound }}</el-tag>
            <span v-else class="empty-text">-</span>
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
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row, $index }">
            <el-button type="primary" link size="small" @click="editRule(row, $index)">
              编辑
            </el-button>
            <el-button type="danger" link size="small" @click="deleteRule(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
        </el-table>

      <div v-if="selectedRules.length > 0" class="batch-actions">
        <span class="batch-count">已选择 {{ selectedRules.length }} 条</span>
        <el-select
          v-model="batchInboundTag"
          placeholder="选择入站"
          filterable
          clearable
          size="small"
          class="batch-select"
        >
          <el-option
            v-for="ib in allInbounds"
            :key="ib.tag"
            :label="ib.tag"
            :value="ib.tag"
          />
        </el-select>
        <el-button type="primary" size="small" @click="batchAddInbound" :disabled="!batchInboundTag">
          加入入站
        </el-button>
        <el-select
          v-model="batchOutboundTag"
          placeholder="选择出站"
          filterable
          clearable
          size="small"
          class="batch-select"
        >
          <el-option
            v-for="ob in enabledOutbounds"
            :key="ob.tag"
            :label="ob.tag"
            :value="ob.tag"
          />
        </el-select>
        <el-button type="success" size="small" @click="batchSetOutbound" :disabled="!batchOutboundTag">
          设置出站
        </el-button>
        <el-button type="danger" size="small" @click="batchDelete">
          批量删除
        </el-button>
      </div>
    </el-card>

    <!-- Add/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="editingIndex !== null ? '编辑路由规则' : '添加路由规则'"
      width="600px"
    >
      <el-form :model="form" label-width="100px">
        <el-form-item label="动作">
          <el-select v-model="form.action" placeholder="选择动作">
            <el-option label="route" value="route" />
            <el-option label="sniff" value="sniff" />
            <el-option label="bypass" value="bypass" />
            <el-option label="reject" value="reject" />
          </el-select>
          <div class="form-tip">route: 转发到出站, sniff: 协议嗅探, bypass: 跳过匹配, reject: 拒绝连接</div>
        </el-form-item>

        <el-form-item v-if="form.action === 'sniff'" label="超时时间">
          <el-input v-model="form.options.timeout" placeholder="例如: 500ms" />
          <div class="form-tip">嗅探超时时间，支持 ms、s 格式</div>
        </el-form-item>

        <el-form-item label="规则集">
          <el-select
            v-model="form.rule_set"
            multiple
            placeholder="选择规则集（可多选）"
            filterable
          >
            <el-option
              v-for="rs in allRulesets"
              :key="rs.tag"
              :label="rs.tag"
              :value="rs.tag"
            />
          </el-select>
          <div class="form-tip">匹配这些规则集的流量将应用此规则</div>
        </el-form-item>

        <el-form-item label="入站">
          <el-select
            v-model="form.inbound"
            multiple
            placeholder="选择入站（可多选，留空匹配全部）"
            filterable
            clearable
          >
            <el-option
              v-for="ib in allInbounds"
              :key="ib.tag"
              :label="ib.tag"
              :value="ib.tag"
            />
          </el-select>
          <div class="form-tip">仅匹配指定入站的流量，留空匹配全部入站</div>
        </el-form-item>

        <el-form-item v-if="form.action === 'route'" label="出站">
          <el-select v-model="form.outbound" placeholder="选择出站" filterable clearable>
            <el-option
              v-for="ob in enabledOutbounds"
              :key="ob.tag"
              :label="ob.tag"
              :value="ob.tag"
            />
          </el-select>
          <div class="form-tip">匹配的流量将转发到此出站</div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveRule" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { singboxApi } from '../../api/singbox'
import { Share, Plus, Setting, Check } from '@element-plus/icons-vue'

const routeRules = ref([])
const allRulesets = ref([])
const allInbounds = ref([])
const enabledOutbounds = ref([])
const httpClientList = ref([])
const loading = ref(false)
const saving = ref(false)
const savingConfig = ref(false)
const dialogVisible = ref(false)
const editingIndex = ref(null)
const tableRef = ref(null)
const selectedRules = ref([])
const batchInboundTag = ref('')
const batchOutboundTag = ref('')

const routeConfig = ref({ final: '', default_http_client: '' })

// Drag state
const dragIndex = ref(null)
let dragEventsSetup = false

const form = ref({
  action: 'route',
  inbound: [],
  rule_set: [],
  outbound: '',
  options: {}
})

const getActionLabel = (action) => {
  const map = { route: 'route', sniff: 'sniff', bypass: 'bypass', reject: 'reject' }
  return map[action] || 'route'
}

const getActionTag = (action) => {
  const map = { route: 'success', sniff: 'warning', bypass: 'warning', reject: 'danger' }
  return map[action] || 'info'
}

const saveRouteConfig = async () => {
  savingConfig.value = true
  try {
    await singboxApi.updateRouteConfig(routeConfig.value)
    ElMessage.success('路由配置已保存')
  } catch (err) {
    ElMessage.error('保存失败: ' + (err.response?.data?.error || err.message))
  } finally {
    savingConfig.value = false
  }
}

const loadData = async () => {
  loading.value = true
  try {
    const [rulesRes, rcRes, rsRes, ibRes, obRes, hcRes] = await Promise.all([
      singboxApi.getRouteRules(),
      singboxApi.getRouteConfig(),
      singboxApi.getRulesets(),
      singboxApi.getInbounds(),
      singboxApi.getOutbounds(),
      singboxApi.getHTTPClients()
    ])
    if (rulesRes.data.success) {
      routeRules.value = rulesRes.data.data || []
    }
    if (rcRes.data.success && rcRes.data.data) {
      routeConfig.value = {
        final: rcRes.data.data.final || '',
        default_http_client: rcRes.data.data.default_http_client || ''
      }
    }
    if (rsRes.data.success) {
      allRulesets.value = rsRes.data.data || []
    }
    if (ibRes.data.success) {
      allInbounds.value = (ibRes.data.data || []).filter(i => i.enabled)
    }
    if (obRes.data.success) {
      enabledOutbounds.value = (obRes.data.data || []).filter(o => o.enabled)
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
  editingIndex.value = null
  form.value = { action: 'route', inbound: [], rule_set: [], outbound: '', options: {} }
  dialogVisible.value = true
}

const editRule = (rule, index) => {
  editingIndex.value = index
  form.value = {
    id: rule.id,
    action: rule.action || 'route',
    inbound: [...(rule.inbound || [])],
    rule_set: [...(rule.rule_set || [])],
    outbound: rule.outbound || '',
    options: { ...(rule.options || {}) }
  }
  dialogVisible.value = true
}

const saveRule = async () => {
  saving.value = true
  try {
    const data = {
      id: form.value.id,
      action: form.value.action,
      inbound: form.value.inbound,
      rule_set: form.value.rule_set,
      outbound: form.value.action === 'route' ? form.value.outbound : '',
      options: form.value.options || {},
      enabled: true
    }

    if (editingIndex.value !== null) {
      data.id = routeRules.value[editingIndex.value].id
      data.enabled = routeRules.value[editingIndex.value].enabled
      await singboxApi.updateRouteRule(data)
      ElMessage.success('更新成功')
    } else {
      await singboxApi.addRouteRule(data)
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    loadData()
  } catch (err) {
    ElMessage.error('保存失败: ' + (err.response?.data?.error || err.message))
  } finally {
    saving.value = false
  }
}

const toggleEnabled = async (rule) => {
  try {
    await singboxApi.updateRouteRule(rule)
    ElMessage.success('状态已更新')
  } catch (err) {
    rule.enabled = !rule.enabled
    ElMessage.error('更新失败')
  }
}

const deleteRule = async (rule) => {
  try {
    await ElMessageBox.confirm('确定要删除此路由规则吗？', '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await singboxApi.deleteRouteRule(rule.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const onSelectionChange = (rules) => {
  selectedRules.value = rules
}

const updateSelectedRules = async (data, message) => {
  if (selectedRules.value.length === 0) return
  try {
    await singboxApi.batchUpdateRouteRules({
      ids: selectedRules.value.map(rule => rule.id),
      ...data
    })
    ElMessage.success(message)
    batchInboundTag.value = ''
    batchOutboundTag.value = ''
    await loadData()
  } catch (err) {
    ElMessage.error('批量更新失败: ' + (err.response?.data?.error || err.message))
  }
}

const batchAddInbound = () => {
  updateSelectedRules({ inbounds: [batchInboundTag.value] }, '已批量加入入站（重复项已自动忽略）')
}

const batchSetOutbound = () => {
  updateSelectedRules({ outbound: batchOutboundTag.value }, '已批量设置出站')
}

const batchDelete = async () => {
  if (selectedRules.value.length === 0) return
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedRules.value.length} 条路由规则吗？`,
      '批量删除',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    await singboxApi.batchDeleteRouteRules(selectedRules.value.map(rule => rule.id))
    ElMessage.success('批量删除成功')
    selectedRules.value = []
    await loadData()
  } catch (err) {
    if (err !== 'cancel') ElMessage.error('批量删除失败: ' + (err.response?.data?.error || err.message))
  }
}

const onDragStart = (e, index) => {
  dragIndex.value = index
  e.dataTransfer.effectAllowed = 'move'
  e.dataTransfer.setData('text/plain', String(index))
  // Add a slight delay to allow the drag image to be captured
  setTimeout(() => {
    e.target.style.opacity = '0.4'
  }, 0)
}

const onDragEnd = (e) => {
  e.target.style.opacity = '1'
  dragIndex.value = null
  clearDragOver()
}

const clearDragOver = () => {
  if (!tableRef.value?.$el) return
  const rows = tableRef.value.$el.querySelectorAll('.el-table__body-wrapper tbody tr')
  rows.forEach(tr => tr.classList.remove('drag-over-row'))
}

const setupTableDragEvents = () => {
  if (dragEventsSetup) return
  if (!tableRef.value?.$el) return
  const tbody = tableRef.value.$el.querySelector('.el-table__body-wrapper tbody')
  if (!tbody) return
  dragEventsSetup = true

  tbody.addEventListener('dragover', (e) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
    const tr = e.target.closest('tr')
    if (!tr) return
    clearDragOver()
    tr.classList.add('drag-over-row')
  })

  tbody.addEventListener('dragleave', (e) => {
    const tr = e.target.closest('tr')
    if (tr) tr.classList.remove('drag-over-row')
  })

  tbody.addEventListener('drop', (e) => {
    e.preventDefault()
    const tr = e.target.closest('tr')
    if (!tr) return
    clearDragOver()

    const fromIndex = dragIndex.value
    const rows = Array.from(tbody.querySelectorAll('tr'))
    const toIndex = rows.indexOf(tr)

    if (fromIndex === null || fromIndex === toIndex || toIndex < 0) return

    const rules = [...routeRules.value]
    const [moved] = rules.splice(fromIndex, 1)
    rules.splice(toIndex, 0, moved)
    routeRules.value = rules

    dragIndex.value = null
    persistOrder()
  })

  tbody.addEventListener('dragend', onDragEnd)
}

const persistOrder = async () => {
  try {
    const ids = routeRules.value.map(r => r.id)
    await singboxApi.reorderRouteRules(ids)
  } catch (err) {
    ElMessage.error('排序更新失败')
    loadData()
  }
}

onMounted(async () => {
  await loadData()
  // Setup drag events after table renders
  setTimeout(() => setupTableDragEvents(), 100)
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

.drag-handle {
  cursor: grab;
  user-select: none;
  display: inline-block;
  width: 100%;
  text-align: center;
}

.drag-handle:active {
  cursor: grabbing;
}

.drag-over-row {
  background-color: var(--el-color-primary-light-9) !important;
}

.rule-tag {
  margin-right: 4px;
  margin-bottom: 2px;
}

.empty-text {
  color: var(--text-secondary);
  font-size: 13px;
}

.form-tip {
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-secondary);
}

.config-card {
  margin-bottom: 16px;
}

.batch-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 16px;
}

.batch-count {
  color: var(--text-secondary);
  font-size: 13px;
  margin-right: 4px;
}

.batch-select {
  width: 180px;
}
</style>
