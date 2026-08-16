<template>
  <el-card class="section-card" shadow="never">
    <template #header>
      <div class="card-header">
        <el-icon><Coin /></el-icon>
        <span>数据库管理</span>
        <el-button type="primary" link size="small" @click="loadBuckets">
          <el-icon><RefreshRight /></el-icon>
          刷新
        </el-button>
        <el-button type="primary" link size="small" @click="exportDatabase">
          <el-icon><Download /></el-icon>
          导出
        </el-button>
        <el-button type="warning" link size="small" @click="triggerImport">
          <el-icon><Upload /></el-icon>
          导入
        </el-button>
        <input ref="importFileRef" type="file" accept=".json,application/json" style="display: none" @change="onImportFile" />
      </div>
    </template>

    <div class="data-section">
      <div class="tree-panel" v-loading="loadingBuckets">
        <el-tree
          ref="treeRef"
          :data="treeData"
          :props="treeProps"
          node-key="key"
          highlight-current
          default-expand-all
          @node-click="onNodeClick"
        >
          <template #default="{ node, data }">
            <span class="tree-node">
              <el-icon v-if="data.isBucket"><Folder /></el-icon>
              <el-icon v-else><Document /></el-icon>
              <span>{{ node.label }}</span>
              <el-button
                v-if="data.isBucket"
                class="bucket-delete"
                type="danger"
                link
                size="small"
                @click.stop="deleteBucket(data)"
                title="删除空目录"
              >
                <el-icon><Delete /></el-icon>
              </el-button>
            </span>
          </template>
        </el-tree>
        <el-empty v-if="treeData.length === 0 && !loadingBuckets" description="暂无数据" :image-size="60" />
      </div>

      <div class="value-panel">
        <template v-if="selectedKey">
          <div class="section-label">
            <span class="path-text">{{ selectedBucket }} / {{ selectedKey }}</span>
            <el-button type="danger" link size="small" @click="deleteCurrentKey">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
            <el-button type="primary" size="small" @click="saveValue" :loading="savingValue">
              <el-icon><Check /></el-icon>
              保存
            </el-button>
          </div>
          <div ref="editorRef" class="codemirror-container"></div>
        </template>
        <el-empty v-else description="选择一个 key 查看数据" :image-size="60" />
      </div>
    </div>
  </el-card>
</template>

<script setup>
import { ref, onMounted, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useTheme } from '../../composables/useTheme'
import axios from 'axios'
import { EditorView, keymap, lineNumbers, highlightActiveLine } from '@codemirror/view'
import { EditorState } from '@codemirror/state'
import { foldGutter, syntaxHighlighting, defaultHighlightStyle } from '@codemirror/language'
import { json } from '@codemirror/lang-json'
import { oneDark } from '@codemirror/theme-one-dark'
import { defaultKeymap, indentWithTab } from '@codemirror/commands'
import { Coin, RefreshRight, Delete, Check, Folder, Document, Download, Upload } from '@element-plus/icons-vue'

const { isDark } = useTheme()

const treeRef = ref(null)
const treeData = ref([])
const treeProps = { children: 'children', label: 'label' }
const selectedBucket = ref('')
const selectedKey = ref('')
const valueContent = ref('')
const loadingBuckets = ref(false)
const savingValue = ref(false)
const editorRef = ref(null)
let editorView = null
const dbApi = axios.create({ baseURL: '/api/db', timeout: 10000 })
const localApi = axios.create({ baseURL: '/api/instances', timeout: 10000 })

// When the panel has a sync token configured, the export/import endpoints
// require it via the X-Sync-Token header.
const attachSyncToken = async () => {
  try {
    const res = await localApi.get('/local-info')
    const token = res.data?.data?.syncToken
    if (token) {
      dbApi.defaults.headers.common['X-Sync-Token'] = token
    }
  } catch {
    // ignore: token protection stays as configured on the server side
  }
}

const initEditor = () => {
  if (editorView) {
    editorView.destroy()
    editorView = null
  }
  if (!editorRef.value) return

  const extensions = [
    lineNumbers(),
    foldGutter(),
    highlightActiveLine(),
    json(),
    syntaxHighlighting(defaultHighlightStyle),
    keymap.of([...defaultKeymap, indentWithTab]),
    EditorView.updateListener.of((update) => {
      if (update.docChanged) {
        valueContent.value = update.state.doc.toString()
      }
    })
  ]

  if (isDark.value) {
    extensions.push(oneDark)
  }

  editorView = new EditorView({
    parent: editorRef.value,
    state: EditorState.create({
      doc: valueContent.value || '',
      extensions
    })
  })
}

watch(isDark, async () => {
  if (selectedKey.value) {
    await nextTick()
    initEditor()
  }
})

const destroyEditor = () => {
  if (editorView) {
    editorView.destroy()
    editorView = null
  }
}

const loadBuckets = async () => {
  loadingBuckets.value = true
  selectedBucket.value = ''
  selectedKey.value = ''
  valueContent.value = ''
  destroyEditor()
  try {
    const res = await dbApi.get('/buckets')
    const buckets = res.data.data || []
    const tree = []
    for (const b of buckets) {
      const keysRes = await dbApi.get('/keys', { params: { bucket: b } })
      const keys = keysRes.data.data || []
      tree.push({
        label: b,
        key: b,
        isBucket: true,
        children: keys.map(k => ({ label: k, key: `${b}/${k}`, bucket: b, isBucket: false }))
      })
    }
    treeData.value = tree
  } catch (err) {
    ElMessage.error('加载数据失败')
  } finally {
    loadingBuckets.value = false
  }
}

const onNodeClick = async (data) => {
  if (data.isBucket) {
    selectedBucket.value = data.label
    selectedKey.value = ''
    valueContent.value = ''
    destroyEditor()
    return
  }
  selectedBucket.value = data.bucket
  selectedKey.value = data.label
  try {
    const res = await dbApi.get('/value', { params: { bucket: data.bucket, key: data.label } })
    if (res.data.success) {
      valueContent.value = res.data.data || ''
      await nextTick()
      initEditor()
    }
  } catch (err) {
    ElMessage.error('加载数据失败')
    valueContent.value = ''
    destroyEditor()
  }
}

const saveValue = async () => {
  savingValue.value = true
  try {
    await dbApi.put('/value', {
      bucket: selectedBucket.value,
      key: selectedKey.value,
      value: valueContent.value
    })
    ElMessage.success('保存成功')
  } catch (err) {
    ElMessage.error('保存失败: ' + (err.response?.data?.error || err.message))
  } finally {
    savingValue.value = false
  }
}

const deleteCurrentKey = async () => {
  if (!selectedBucket.value || !selectedKey.value) return
  try {
    await ElMessageBox.confirm(`确定要删除 "${selectedKey.value}" 吗？`, '确认删除', {
      confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
    })
    await dbApi.delete('/value', { params: { bucket: selectedBucket.value, key: selectedKey.value } })
    ElMessage.success('删除成功')
    selectedKey.value = ''
    valueContent.value = ''
    destroyEditor()
    loadBuckets()
  } catch (err) {
    if (err !== 'cancel') ElMessage.error('删除失败')
  }
}

const deleteBucket = async (data) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除目录 "${data.label}" 吗？\n仅允许删除空目录（不含任何 key）。`,
      '确认删除目录',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    await dbApi.delete('/bucket', { params: { bucket: data.label } })
    ElMessage.success('目录已删除')
    if (selectedBucket.value === data.label) {
      selectedBucket.value = ''
      selectedKey.value = ''
      valueContent.value = ''
      destroyEditor()
    }
    loadBuckets()
  } catch (err) {
    if (err !== 'cancel') ElMessage.error('删除失败: ' + (err.response?.data?.error || err.message || ''))
  }
}

const importFileRef = ref(null)

const exportDatabase = async () => {
  try {
    const res = await dbApi.get('/export')
    if (!res.data.success) throw new Error(res.data.error || '导出失败')
    const blob = new Blob([JSON.stringify(res.data.data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    const ts = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19)
    a.href = url
    a.download = `sing-panel-db-${ts}.json`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (err) {
    ElMessage.error('导出失败: ' + (err.message || ''))
  }
}

const triggerImport = () => {
  importFileRef.value?.click()
}

const onImportFile = async (e) => {
  const file = e.target.files?.[0]
  if (!file) return
  try {
    const text = await file.text()
    let parsed
    try {
      parsed = JSON.parse(text)
    } catch {
      ElMessage.error('导入文件不是有效的 JSON')
      return
    }
    if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) {
      ElMessage.error('导入文件格式不正确：应为 { bucket: { key: value } } 结构')
      return
    }
    await ElMessageBox.confirm(
      '导入将替换数据库中全部配置数据（本机运行状态、多实例管理数据除外），正在运行的配置会立即变更。确定继续？',
      '危险操作确认',
      { confirmButtonText: '确认导入', cancelButtonText: '取消', type: 'warning' }
    )
    await dbApi.post('/import', { data: parsed })
    ElMessage.success('导入成功')
    selectedKey.value = ''
    valueContent.value = ''
    destroyEditor()
    loadBuckets()
  } catch (err) {
    if (err !== 'cancel') ElMessage.error('导入失败: ' + (err.response?.data?.error || err.message || ''))
  } finally {
    e.target.value = ''
  }
}

onMounted(() => {
  attachSyncToken()
  loadBuckets()
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

.data-section {
  display: flex;
  gap: 16px;
  height: 500px;
}

.tree-panel {
  width: 240px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 8px;
  overflow-y: auto;
  flex-shrink: 0;
}

.tree-node {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}

.tree-node .bucket-delete {
  margin-left: auto;
  padding: 0 4px;
  opacity: 0;
  transition: opacity 0.2s;
}

.tree-node:hover .bucket-delete {
  opacity: 1;
}

.section-label {
  font-weight: 600;
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.path-text {
  flex: 1;
  font-size: 13px;
  color: var(--text-secondary);
}

.value-panel {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.codemirror-container {
  flex: 1;
  overflow: hidden;
  border: 1px solid var(--border-color);
  border-radius: 6px;
}

.codemirror-container :deep(.cm-editor) {
  height: 100%;
}
</style>
