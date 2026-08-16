<template>
  <div class="instances-page">
    <el-card class="section-card" shadow="never">
      <template #header>
        <div class="card-header">
          <el-icon><Share /></el-icon>
          <span>多实例管理</span>
          <el-button type="primary" size="small" @click="openDialog()" class="add-btn">
            <el-icon><Plus /></el-icon>
            添加实例
          </el-button>
        </div>
      </template>

      <div class="local-panel" v-loading="loadingLocal">
        <div class="local-info">
          <span class="local-label">本机面板</span>
          <el-tag v-if="localInfo" size="small" effect="plain">{{ localInfo.info.hostname }}</el-tag>
          <el-tag v-if="localInfo" size="small" type="info" effect="plain">
            {{ localInfo.info.platform }}/{{ localInfo.info.arch }}
          </el-tag>
          <el-tag v-if="localInfo" size="small" type="info" effect="plain">v{{ localInfo.info.version }}</el-tag>
          <el-tag v-if="localInfo" size="small" :type="localInfo.info.singboxRunning ? 'success' : 'danger'" effect="plain">
            sing-box {{ localInfo.info.singboxRunning ? '运行中' : '已停止' }}
          </el-tag>
          <el-tag v-if="localInfo && localInfo.info.syncTokenEnabled" size="small" effect="plain">同步令牌已启用</el-tag>
          <span v-if="localInfo" class="local-meta">
            数据库 {{ formatBytes(localInfo.info.dbSize) }}
          </span>
        </div>
        <div class="local-fingerprint">
          <span class="fp-label">本机配置指纹</span>
          <code v-if="localInfo" class="fp-value" @click="copyLocalFingerprint" title="点击复制">
            {{ localInfo.fingerprint ? localInfo.fingerprint.slice(0, 12) : '—' }}
          </code>
          <span v-else class="fp-value">—</span>
        </div>
      </div>

      <div class="token-row">
        <span class="token-label">同步令牌</span>
        <el-input
          v-model="tokenDraft"
          placeholder="留空则不启用令牌保护"
          class="token-input"
          clearable
        />
        <el-button type="primary" :loading="savingToken" @click="saveToken">保存</el-button>
        <span class="token-tip">
          设置后，其他面板访问本面板的同步接口（导出 / 导入 / 面板信息）需携带此令牌；远端实例在「编辑」中配置相同令牌。
        </span>
      </div>

      <el-table :data="rows" v-loading="checking" class="instances-table" empty-text="暂无实例，点击右上角「添加实例」">
        <el-table-column label="名称" width="140">
          <template #default="{ row }">
            <span class="inst-name">{{ row.name || row.status.instance?.name || '—' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="地址" min-width="200">
          <template #default="{ row }">{{ row.url }}</template>
        </el-table-column>
        <el-table-column label="状态" width="130">
          <template #default="{ row }">
            <template v-if="row.status.reachable">
              <el-tag size="small" type="success">在线</el-tag>
              <span class="latency">{{ row.status.latencyMs }}ms</span>
            </template>
            <el-tooltip v-else :content="row.status.error || '未检测'" placement="top">
              <el-tag size="small" type="danger">不可达</el-tag>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column label="远端面板" min-width="200">
          <template #default="{ row }">
            <template v-if="row.status.info">
              <div class="remote-line">{{ row.status.info.hostname }} · v{{ row.status.info.version }}</div>
              <div class="remote-sub">
                {{ row.status.info.platform }}/{{ row.status.info.arch }} · sing-box
                {{ row.status.info.singboxRunning ? '运行中' : '已停止' }}
              </div>
            </template>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="配置一致性" width="150">
          <template #default="{ row }">
            <el-tag v-if="row.status.inSync === true" size="small" type="success">一致</el-tag>
            <template v-else-if="row.status.inSync === false">
              <el-tag size="small" type="danger">不一致</el-tag>
              <el-button type="primary" link size="small" 
                @click="showDiff(row)" :loading="row.diffLoading">
                查看差异
              </el-button>
            </template>
            <el-tag v-else-if="row.status.reachable" size="small" type="info">未知</el-tag>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="远端指纹" width="110">
          <template #default="{ row }">
            <code v-if="row.status.fingerprint" class="fp-value" @click="copyFingerprint(row)"
              :title="row.status.fingerprint">
              {{ row.status.fingerprint.slice(0, 8) }}
            </code>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="检测时间" width="170">
          <template #default="{ row }">
            <span v-if="row.status.checkedAt" class="muted">{{ formatTime(row.status.checkedAt) }}</span>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" :loading="row.syncing" @click="checkOne(row)">
              <el-icon><RefreshRight /></el-icon>
              检查
            </el-button>
            <el-button type="success" link size="small" :loading="row.syncing" @click="push(row)">
              <el-icon><Upload /></el-icon>
              推送
            </el-button>
            <el-button type="warning" link size="small" :loading="row.syncing" @click="pull(row)">
              <el-icon><Download /></el-icon>
              拉取
            </el-button>
            <el-button link size="small" @click="openDialog(row)">
              <el-icon><Edit /></el-icon>
            </el-button>
            <el-button type="danger" link size="small" @click="remove(row)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="table-toolbar">
        <el-button :loading="checking" @click="checkAll">
          <el-icon><RefreshRight /></el-icon>
          全部检查
        </el-button>
        <el-button type="primary" :loading="pushingAll" @click="pushAll">
          <el-icon><Upload /></el-icon>
          推送本机配置到全部实例
        </el-button>
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑实例' : '添加实例'" width="520px">
      <el-form :model="dialogForm" label-width="90px">
        <el-form-item label="名称">
          <el-input v-model="dialogForm.name" placeholder="留空则使用远端面板主机名" />
        </el-form-item>
        <el-form-item label="地址" required>
          <el-input v-model="dialogForm.url" placeholder="198.51.100.10:8080（默认 http://）" />
          <div class="form-tip">面板的访问地址（不含末尾斜杠），未写协议时默认使用 http://。</div>
        </el-form-item>
        <el-form-item label="同步令牌">
          <el-input v-model="dialogForm.token" placeholder="与远端面板设置的同步令牌一致，否则留空" clearable />
          <div class="form-tip">远端要求令牌保护时必填，请求将以 X-Sync-Token 头发送。</div>
        </el-form-item>
        <el-form-item label="超时">
          <el-input-number v-model="dialogForm.timeout" :min="3" :max="60" />
          <span class="token-tip">请求超时（秒），默认 10</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="savingDialog" @click="saveDialog">保存</el-button>
      </template>
    </el-dialog>

    <!-- Config Diff Dialog -->
    <el-dialog v-model="diffDialogVisible" title="配置差异" width="900px" class="diff-dialog">
      <div class="diff-header">
        <div class="diff-fingerprint">
          <span class="diff-label">本机指纹:</span>
          <code class="fp-value">{{ diffResult.localFingerprint?.slice(0, 16) || '—' }}</code>
        </div>
        <div class="diff-fingerprint">
          <span class="diff-label">远端指纹:</span>
          <code class="fp-value">{{ diffResult.remoteFingerprint?.slice(0, 16) || '—' }}</code>
        </div>
        <div class="diff-tools">
          <el-radio-group v-model="diffViewMode" size="small">
            <el-radio-button value="split">分栏</el-radio-button>
            <el-radio-button value="unified">统一</el-radio-button>
          </el-radio-group>
          <span class="diff-stats" v-if="diffResult.differences?.length">
            共 {{ diffResult.differences.length }} 处差异
          </span>
        </div>
      </div>
      
      <div class="diff-content" v-loading="loadingDiff">
        <template v-if="groupedDifferences.length > 0">
          <div v-for="group in groupedDifferences" :key="group.bucket" class="diff-group">
            <div class="diff-group-header">
              <el-icon><Folder /></el-icon>
              <span class="diff-bucket-name">{{ group.bucket }}</span>
              <el-tag size="small" type="info">{{ group.items.length }} 处差异</el-tag>
            </div>
            <div class="diff-items">
              <div v-for="item in group.items" :key="item.key" class="diff-item">
                <div class="diff-item-header">
                  <el-tag size="small" :type="getDiffTypeTag(item.type)" class="diff-type-tag">
                    {{ getDiffTypeName(item.type) }}
                  </el-tag>
                  <code class="diff-key">{{ item.key }}</code>
                </div>
                <DiffView
                  :local="formatDiffValue(item.localValue)"
                  :remote="formatDiffValue(item.remoteValue)"
                  :view="diffViewMode"
                />
              </div>
            </div>
          </div>
        </template>
        <el-empty v-else-if="!loadingDiff" description="暂无差异" />
      </div>
      
      <template #footer>
        <el-button @click="diffDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { instancesApi } from '../../api/instances'
import DiffView from '../../components/DiffView.vue'
import {
  Share, RefreshRight, Plus, Upload, Download, Edit, Delete
} from '@element-plus/icons-vue'

const instances = ref([])
const statusMap = ref({})
const checking = ref(false)
const loadingLocal = ref(false)
const savingToken = ref(false)
const pushingAll = ref(false)
const localInfo = ref(null)
const tokenDraft = ref('')

const dialogVisible = ref(false)
const editingId = ref('')
const savingDialog = ref(false)
const dialogForm = ref({ name: '', url: '', token: '', timeout: 10 })

const diffDialogVisible = ref(false)
const loadingDiff = ref(false)
const diffViewMode = ref('split')
const diffResult = ref({ differences: [], localFingerprint: '', remoteFingerprint: '' })
const diffLoadingMap = ref({})

const rows = computed(() => {
  return instances.value.map(inst => ({
    ...inst,
    status: statusMap.value[inst.id] || {},
    syncing: false,
    diffLoading: diffLoadingMap.value[inst.id] || false
  }))
})

const formatBytes = (n) => {
  if (!n && n !== 0) return '—'
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / 1024 / 1024).toFixed(1)} MB`
}

const formatTime = (t) => {
  if (!t) return '—'
  const d = new Date(t)
  const pad = n => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

const loadLocalInfo = async () => {
  const res = await instancesApi.localInfo()
  if (res.data.success) {
    localInfo.value = res.data.data
    tokenDraft.value = res.data.data.syncToken || ''
  }
}

const loadInstances = async () => {
  const res = await instancesApi.list()
  if (res.data.success) {
    instances.value = res.data.data || []
  }
}

const checkAll = async () => {
  checking.value = true
  try {
    const res = await instancesApi.checkAll()
    if (res.data.success) {
      const map = {}
      for (const st of res.data.data) map[st.instance.id] = st
      statusMap.value = map
    }
  } catch (err) {
    ElMessage.error('检查失败: ' + (err.response?.data?.error || err.message))
  } finally {
    checking.value = false
  }
}

const checkOne = async (row) => {
  row.syncing = true
  try {
    const res = await instancesApi.checkOne(row.id)
    if (res.data.success) {
      statusMap.value = { ...statusMap.value, [row.id]: res.data.data }
    }
  } catch (err) {
    ElMessage.error('检查失败: ' + (err.response?.data?.error || err.message))
  } finally {
    row.syncing = false
  }
}

const refreshAfterSync = async () => {
  await Promise.all([loadLocalInfo(), checkAll()])
}

const push = async (row) => {
  row.syncing = true
  try {
    const res = await instancesApi.sync(row.id, 'push')
    if (res.data.success) {
      ElMessage.success(`配置已推送到 ${row.name}`)
      await checkOne(row)
    }
  } catch (err) {
    ElMessage.error(`推送失败: ${err.response?.data?.error || err.message}`)
  } finally {
    row.syncing = false
  }
}

const pull = async (row) => {
  try {
    await ElMessageBox.confirm(
      `将从 "${row.name}" 拉取配置并覆盖本机 sing-box 配置（本机运行状态与其他实例不受影响）。确定继续？`,
      '拉取配置确认',
      { confirmButtonText: '确认拉取', cancelButtonText: '取消', type: 'warning' }
    )
  } catch {
    return
  }
  row.syncing = true
  try {
    const res = await instancesApi.sync(row.id, 'pull')
    if (res.data.success) {
      ElMessage.success(`已从 ${row.name} 拉取配置`)
      await refreshAfterSync()
    }
  } catch (err) {
    ElMessage.error(`拉取失败: ${err.response?.data?.error || err.message}`)
  } finally {
    row.syncing = false
  }
}

const pushAll = async () => {
  if (instances.value.length === 0) {
    ElMessage.warning('暂无实例')
    return
  }
  try {
    await ElMessageBox.confirm(
      `将本机配置推送到全部 ${instances.value.length} 个实例，远端配置将被覆盖。确定继续？`,
      '推送全部确认',
      { confirmButtonText: '确认推送', cancelButtonText: '取消', type: 'warning' }
    )
  } catch {
    return
  }
  pushingAll.value = true
  try {
    const res = await instancesApi.syncAll()
    let failed = 0
    let ok = 0
    for (const st of res.data.data || []) {
      if (st.error) failed++
      else ok++
    }
    if (failed > 0) ElMessage.warning(`推送完成: 成功 ${ok} 个，失败 ${failed} 个`)
    else ElMessage.success(`已推送 ${ok} 个实例`)
    await checkAll()
  } catch (err) {
    ElMessage.error('推送失败: ' + (err.response?.data?.error || err.message))
  } finally {
    pushingAll.value = false
  }
}

const remove = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除实例 "${row.name}" 吗？`, '确认删除', {
      confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
    })
  } catch {
    return
  }
  try {
    await instancesApi.remove(row.id)
    ElMessage.success('实例已删除')
    statusMap.value = { ...statusMap.value, [row.id]: undefined }
    await loadInstances()
  } catch (err) {
    ElMessage.error('删除失败: ' + (err.response?.data?.error || err.message))
  }
}

const openDialog = (row) => {
  editingId.value = row?.id || ''
  dialogForm.value = row
    ? { name: row.name, url: row.url, token: row.token || '', timeout: row.timeout || 10 }
    : { name: '', url: '', token: '', timeout: 10 }
  dialogVisible.value = true
}

const saveDialog = async () => {
  if (!dialogForm.value.url.trim()) {
    ElMessage.warning('地址不能为空')
    return
  }
  savingDialog.value = true
  try {
    if (editingId.value) {
      await instancesApi.update(editingId.value, dialogForm.value)
      ElMessage.success('实例已更新')
    } else {
      await instancesApi.create(dialogForm.value)
      ElMessage.success('实例已添加')
    }
    dialogVisible.value = false
    await loadInstances()
    await checkAll()
  } catch (err) {
    ElMessage.error('保存失败: ' + (err.response?.data?.error || err.message))
  } finally {
    savingDialog.value = false
  }
}

const saveToken = async () => {
  savingToken.value = true
  try {
    const res = await instancesApi.setSyncToken(tokenDraft.value)
    if (res.data.success) {
      ElMessage.success('同步令牌已保存')
      await loadLocalInfo()
    }
  } catch (err) {
    ElMessage.error('保存失败: ' + (err.response?.data?.error || err.message))
  } finally {
    savingToken.value = false
  }
}

const copyLocalFingerprint = async () => {
  if (!localInfo.value?.fingerprint) return
  try {
    await navigator.clipboard.writeText(localInfo.value.fingerprint)
    ElMessage.success('本机指纹已复制')
  } catch {
    ElMessage.info(localInfo.value.fingerprint)
  }
}

const copyFingerprint = async (row) => {
  if (!row.status.fingerprint) return
  try {
    await navigator.clipboard.writeText(row.status.fingerprint)
    ElMessage.success('指纹已复制')
  } catch {
    ElMessage.info(row.status.fingerprint)
  }
}

const loadAll = async () => {
  loadingLocal.value = true
  try {
    await Promise.all([loadInstances(), loadLocalInfo()])
    await checkAll()
  } catch (err) {
    ElMessage.error('加载失败: ' + (err.response?.data?.error || err.message))
  } finally {
    loadingLocal.value = false
  }
}

const showDiff = async (row) => {
  diffLoadingMap.value = { ...diffLoadingMap.value, [row.id]: true }
  diffDialogVisible.value = true
  loadingDiff.value = true
  try {
    const res = await instancesApi.getConfigDiff(row.id)
    if (res.data.success) {
      diffResult.value = res.data.data
    }
  } catch (err) {
    ElMessage.error('获取差异失败: ' + (err.response?.data?.error || err.message))
    diffDialogVisible.value = false
  } finally {
    loadingDiff.value = false
    diffLoadingMap.value = { ...diffLoadingMap.value, [row.id]: false }
  }
}

const getDiffTypeTag = (type) => {
  const map = {
    added: 'success',
    removed: 'danger',
    modified: 'warning'
  }
  return map[type] || 'info'
}

// Group differences by bucket for git diff style display
const groupedDifferences = computed(() => {
  const groups = {}
  for (const item of (diffResult.value.differences || [])) {
    if (!groups[item.bucket]) {
      groups[item.bucket] = { bucket: item.bucket, items: [] }
    }
    groups[item.bucket].items.push(item)
  }
  return Object.values(groups).sort((a, b) => a.bucket.localeCompare(b.bucket))
})

// Format diff value for display (handle JSON)
const formatDiffValue = (value) => {
  if (!value) return '(空)'
  try {
    const obj = JSON.parse(value)
    return JSON.stringify(obj, null, 2)
  } catch {
    return value
  }
}

const getDiffTypeName = (type) => {
  const map = {
    added: '新增',
    removed: '删除',
    modified: '修改'
  }
  return map[type] || type
}

onMounted(loadAll)
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

.local-panel {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.local-info {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.local-label {
  font-size: 13px;
  color: var(--text-primary);
  font-weight: 600;
}

.local-meta {
  font-size: 12px;
  color: var(--text-secondary);
}

.local-fingerprint {
  display: flex;
  align-items: center;
  gap: 8px;
}

.fp-label {
  font-size: 12px;
  color: var(--text-secondary);
}

.fp-value {
  font-family: monospace;
  font-size: 12px;
  background: var(--bg-page);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: 2px 6px;
  cursor: pointer;
}

.token-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.token-label {
  font-size: 13px;
  color: var(--text-primary);
  font-weight: 600;
  white-space: nowrap;
}

.token-input {
  width: 220px;
}

.token-tip {
  font-size: 12px;
  color: var(--text-secondary);
}

.table-toolbar {
  margin-top: 12px;
  display: flex;
  gap: 8px;
}

.inst-name {
  font-weight: 600;
}

.latency {
  margin-left: 6px;
  font-size: 12px;
  color: var(--text-secondary);
}

.remote-line {
  font-size: 13px;
}

.remote-sub {
  font-size: 12px;
  color: var(--text-secondary);
}

.muted {
  color: var(--text-secondary);
  font-size: 12px;
}

.form-tip {
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.6;
}

.instances-table {
  width: 100%;
}

.diff-header {
  display: flex;
  gap: 24px;
  margin-bottom: 16px;
  padding: 12px;
  background: var(--bg-page);
  border-radius: 6px;
}

.diff-fingerprint {
  display: flex;
  align-items: center;
  gap: 8px;
}

.diff-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.diff-tools {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 12px;
}

.diff-stats {
  font-size: 13px;
  color: var(--text-secondary);
}

.diff-content {
  max-height: 55vh;
  overflow-y: auto;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: var(--bg-page);
}

.diff-group {
  border-bottom: 1px solid var(--border-color);
}

.diff-group:last-child {
  border-bottom: none;
}

.diff-group-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: var(--el-fill-color-light);
  font-weight: 600;
  font-size: 14px;
  border-bottom: 1px solid var(--border-color);
  position: sticky;
  top: 0;
  z-index: 1;
}

.diff-bucket-name {
  font-family: monospace;
}

.diff-items {
  padding: 8px 0;
}

.diff-item {
  padding: 8px 16px;
  border-bottom: 1px dashed var(--border-color);
}

.diff-item:last-child {
  border-bottom: none;
}

.diff-item-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.diff-type-tag {
  flex-shrink: 0;
}

.diff-key {
  font-family: monospace;
  font-size: 13px;
  font-weight: 500;
}

.diff-dialog .el-dialog__body {
  padding-top: 12px;
}
</style>