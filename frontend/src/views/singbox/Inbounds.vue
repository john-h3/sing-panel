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

        <el-form-item label="监听地址">
          <el-input v-model="form.listen" placeholder="0.0.0.0" />
        </el-form-item>

        <el-form-item label="监听端口" prop="listenPort">
          <el-input-number v-model="form.listenPort" :min="1" :max="65535" />
        </el-form-item>

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
import { Upload, Plus } from '@element-plus/icons-vue'

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

const rules = {
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  tag: [{ required: true, message: '请输入标签', trigger: 'blur' }],
  listenPort: [{ required: true, message: '请输入端口', trigger: 'blur' }]
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
  dialogVisible.value = true
}

const editInbound = (inbound) => {
  editingInbound.value = inbound
  form.value = { ...inbound, options: { ...inbound.options } }
  dialogVisible.value = true
}

const onTypeChange = (type) => {
  // Set default port based on type
  const portMap = {
    http: 8080,
    socks: 1080,
    mixed: 2080,
    tun: 0,
    shadowsocks: 8388
  }
  form.value.listenPort = portMap[type] || 8080
}

const saveInbound = async () => {
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  saving.value = true
  try {
    if (editingInbound.value) {
      await singboxApi.updateInbound(form.value)
      ElMessage.success('更新成功')
    } else {
      await singboxApi.addInbound(form.value)
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
      ElMessage.error('删除失败')
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
</style>
