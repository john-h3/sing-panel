<template>
  <div class="outbounds-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <el-icon><Download /></el-icon>
          <span>Outbound 配置</span>
          <el-button type="success" @click="showImportDialog">
            <el-icon><Upload /></el-icon>
            导入链接
          </el-button>
          <el-button type="primary" @click="showAddDialog">
            <el-icon><Plus /></el-icon>
            添加
          </el-button>
        </div>
      </template>

      <el-table :data="outbounds" v-loading="loading" stripe>
        <el-table-column prop="tag" label="标签" width="150" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag size="small" :type="getTypeTag(row.type)">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              :disabled="isCustomizedOutbound(row) && !customizedEnabled"
              @change="toggleEnabled(row)"
              size="small"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="editOutbound(row)">
              编辑
            </el-button>
            <el-button type="danger" link size="small" @click="deleteOutbound(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Add/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="editingOutbound ? '编辑 Outbound' : '添加 Outbound'"
      width="600px"
    >
      <el-form :model="form" label-width="100px" ref="formRef" :rules="rules">
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" placeholder="选择类型">
            <el-option
              v-for="item in outboundTypes"
              :key="item.type"
              :label="item.name"
              :value="item.type"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="标签" prop="tag">
          <el-input v-model="form.tag" placeholder="例如: proxy-out" />
        </el-form-item>

        <el-form-item label="启用">
          <el-switch
            v-model="form.enabled"
            :disabled="isCustomizedOutbound(form) && !customizedEnabled"
          />
        </el-form-item>

        <!-- Dynamic options based on type -->
        <template v-if="form.type === 'shadowsocks'">
          <el-form-item label="服务器">
            <el-input v-model="form.options.server" placeholder="server address" />
          </el-form-item>
          <el-form-item label="服务器端口">
            <el-input-number v-model="form.options.server_port" :min="1" :max="65535" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="form.options.password" type="password" show-password />
          </el-form-item>
          <el-form-item label="加密方式">
            <el-select v-model="form.options.method">
              <el-option label="2022-blake3-aes-128-gcm" value="2022-blake3-aes-128-gcm" />
              <el-option label="2022-blake3-aes-256-gcm" value="2022-blake3-aes-256-gcm" />
              <el-option label="2022-blake3-chacha20-poly1305" value="2022-blake3-chacha20-poly1305" />
              <el-option label="aes-128-gcm" value="aes-128-gcm" />
              <el-option label="aes-256-gcm" value="aes-256-gcm" />
              <el-option label="chacha20-ietf-poly1305" value="chacha20-ietf-poly1305" />
              <el-option label="xchacha20-ietf-poly1305" value="xchacha20-ietf-poly1305" />
            </el-select>
          </el-form-item>
        </template>

        <template v-if="form.type === 'vmess'">
          <el-form-item label="服务器">
            <el-input v-model="form.options.server" placeholder="server address" />
          </el-form-item>
          <el-form-item label="服务器端口">
            <el-input-number v-model="form.options.server_port" :min="1" :max="65535" />
          </el-form-item>
          <el-form-item label="UUID">
            <el-input v-model="form.options.uuid" placeholder="uuid" />
          </el-form-item>
          <el-form-item label="传输层">
            <el-select v-model="form.options.transport">
              <el-option label="tcp" value="tcp" />
              <el-option label="ws" value="ws" />
              <el-option label="grpc" value="grpc" />
              <el-option label="h2" value="h2" />
            </el-select>
          </el-form-item>
        </template>

        <template v-if="form.type === 'vless'">
          <el-form-item label="服务器">
            <el-input v-model="form.options.server" placeholder="server address" />
          </el-form-item>
          <el-form-item label="服务器端口">
            <el-input-number v-model="form.options.server_port" :min="1" :max="65535" />
          </el-form-item>
          <el-form-item label="UUID">
            <el-input v-model="form.options.uuid" placeholder="uuid" />
          </el-form-item>
          <el-form-item label="Flow">
            <el-select v-model="form.options.flow">
              <el-option label="无" value="" />
              <el-option label="xtls-rprx-vision" value="xtls-rprx-vision" />
            </el-select>
          </el-form-item>

          <!-- TLS Settings -->
          <el-divider content-position="left">TLS / Reality</el-divider>

          <el-form-item label="安全层">
            <el-select v-model="vlessSecurity" @change="onVlessSecurityChange">
              <el-option label="无" value="" />
              <el-option label="TLS" value="tls" />
              <el-option label="Reality" value="reality" />
            </el-select>
          </el-form-item>

          <template v-if="vlessSecurity">
            <el-form-item label="SNI">
              <el-input v-model="form.options.tls.server_name" placeholder="server name" />
            </el-form-item>
            <el-form-item label="指纹">
              <el-select v-model="vlessFingerprint">
                <el-option label="无" value="" />
                <el-option label="chrome" value="chrome" />
                <el-option label="firefox" value="firefox" />
                <el-option label="safari" value="safari" />
                <el-option label="edge" value="edge" />
                <el-option label="ios" value="ios" />
                <el-option label="android" value="android" />
                <el-option label="random" value="random" />
                <el-option label="randomized" value="randomized" />
              </el-select>
            </el-form-item>
            <el-form-item label="ALPN">
              <el-input v-model="vlessAlpn" placeholder="h2,http/1.1" />
            </el-form-item>
          </template>

          <template v-if="vlessSecurity === 'reality'">
            <el-form-item label="公钥 (pbk)">
              <el-input v-model="form.options.tls.reality.public_key" placeholder="reality public key" />
            </el-form-item>
            <el-form-item label="Short ID">
              <el-input v-model="form.options.tls.reality.short_id" placeholder="short id" />
            </el-form-item>
          </template>

          <!-- Transport Settings -->
          <template v-if="vlessSecurity !== 'reality'">
            <el-divider content-position="left">传输协议</el-divider>

            <el-form-item label="传输类型">
              <el-select v-model="vlessTransport" @change="onVlessTransportChange">
                <el-option label="tcp" value="tcp" />
                <el-option label="ws" value="ws" />
                <el-option label="grpc" value="grpc" />
                <el-option label="h2" value="h2" />
                <el-option label="quic" value="quic" />
              </el-select>
            </el-form-item>
          </template>

          <template v-if="vlessTransport === 'ws'">
            <el-form-item label="Host">
              <el-input v-model="form.options.transport.host" placeholder="伪装域名" />
            </el-form-item>
            <el-form-item label="Path">
              <el-input v-model="form.options.transport.path" placeholder="/path" />
            </el-form-item>
          </template>

          <template v-if="vlessTransport === 'grpc'">
            <el-form-item label="服务名">
              <el-input v-model="form.options.transport.service_name" placeholder="service name" />
            </el-form-item>
          </template>

          <template v-if="vlessTransport === 'h2'">
            <el-form-item label="Host">
              <el-input v-model="form.options.transport.host" placeholder="伪装域名" />
            </el-form-item>
            <el-form-item label="Path">
              <el-input v-model="form.options.transport.path" placeholder="/path" />
            </el-form-item>
          </template>
        </template>

        <template v-if="form.type === 'trojan'">
          <el-form-item label="服务器">
            <el-input v-model="form.options.server" placeholder="server address" />
          </el-form-item>
          <el-form-item label="服务器端口">
            <el-input-number v-model="form.options.server_port" :min="1" :max="65535" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="form.options.password" type="password" show-password />
          </el-form-item>
        </template>

        <template v-if="form.type === 'hysteria'">
          <el-form-item label="服务器">
            <el-input v-model="form.options.server" placeholder="server address" />
          </el-form-item>
          <el-form-item label="服务器端口">
            <el-input-number v-model="form.options.server_port" :min="1" :max="65535" />
          </el-form-item>
          <el-form-item label="认证字符串">
            <el-input v-model="form.options.auth" placeholder="auth" />
          </el-form-item>
          <el-form-item label="上行 Mbps">
            <el-input-number v-model="form.options.up_mbps" :min="1" />
          </el-form-item>
          <el-form-item label="下行 Mbps">
            <el-input-number v-model="form.options.down_mbps" :min="1" />
          </el-form-item>
        </template>

        <template v-if="form.type === 'fallback'">
          <el-divider content-position="left">Fallback 出站</el-divider>
          <el-form-item label="出站列表">
            <div class="outbound-list-container">
              <div class="selected-outbounds" v-if="form.options.outbounds && form.options.outbounds.length > 0">
                <div
                  v-for="(tag, index) in form.options.outbounds"
                  :key="tag"
                  class="outbound-item"
                  draggable="true"
                  @dragstart="onDragStart(index)"
                  @dragover.prevent="onDragOver(index)"
                  @drop="onDrop(index)"
                  @dragend="onDragEnd"
                  :class="{ 'dragging': dragIndex === index }"
                >
                  <span class="drag-handle">⋮⋮</span>
                  <span class="outbound-tag">{{ tag }}</span>
                  <el-button type="danger" link size="small" @click="removeOutbound(index)">
                    <el-icon><Close /></el-icon>
                  </el-button>
                </div>
              </div>
              <div v-else class="empty-outbounds">暂无出站，请添加</div>
              <el-select
                v-model="addOutboundValue"
                placeholder="添加出站"
                @change="addOutbound"
                style="width: 100%; margin-top: 8px;"
                clearable
              >
                <el-option
                  v-for="ob in availableOutboundsToAdd"
                  :key="ob.tag"
                  :label="ob.tag"
                  :value="ob.tag"
                />
              </el-select>
            </div>
          </el-form-item>
          <div class="form-tip">按顺序尝试列表中的出站，定制化功能关闭时该出站会被禁用。</div>
        </template>

        <template v-if="form.type === 'loadbalance'">
          <el-divider content-position="left">LoadBalance 出站</el-divider>
          <el-form-item label="出站列表">
            <div class="outbound-list-container">
              <div class="selected-outbounds" v-if="form.options.outbounds && form.options.outbounds.length > 0">
                <div
                  v-for="(tag, index) in form.options.outbounds"
                  :key="tag"
                  class="outbound-item"
                  draggable="true"
                  @dragstart="onDragStart(index)"
                  @dragover.prevent="onDragOver(index)"
                  @drop="onDrop(index)"
                  @dragend="onDragEnd"
                  :class="{ 'dragging': dragIndex === index }"
                >
                  <span class="drag-handle">⋮⋮</span>
                  <span class="outbound-tag">{{ tag }}</span>
                  <el-input-number
                    :model-value="getLoadBalanceWeight(tag)"
                    :min="1"
                    :max="65535"
                    :precision="0"
                    controls-position="right"
                    size="small"
                    class="loadbalance-weight"
                    @update:model-value="setLoadBalanceWeight(tag, $event)"
                  />
                  <el-button type="danger" link size="small" @click="removeOutbound(index)">
                    <el-icon><Close /></el-icon>
                  </el-button>
                </div>
              </div>
              <div v-else class="empty-outbounds">暂无出站，请添加</div>
              <el-select
                v-model="addOutboundValue"
                placeholder="添加出站"
                @change="addOutbound"
                style="width: 100%; margin-top: 8px;"
                clearable
              >
                <el-option
                  v-for="ob in availableOutboundsToAdd"
                  :key="ob.tag"
                  :label="ob.tag"
                  :value="ob.tag"
                />
              </el-select>
            </div>
          </el-form-item>

          <el-form-item label="选择策略">
            <el-select v-model="form.options.strategy" placeholder="round_robin">
              <el-option label="轮询 (round_robin)" value="round_robin" />
              <el-option label="随机 (random)" value="random" />
              <el-option label="加权轮询 (weighted_round_robin)" value="weighted_round_robin" />
              <el-option label="加权随机 (weighted_random)" value="weighted_random" />
              <el-option label="最少连接 (least_connections)" value="least_connections" />
              <el-option label="一致性哈希 (consistent_hash)" value="consistent_hash" />
            </el-select>
          </el-form-item>
          <div class="form-tip">按顺序拖动出站列表可调整配置顺序；权重用于加权策略，未填写时默认为 1。</div>
        </template>

        <!-- Selector/URLTest: select outbounds -->
        <template v-if="form.type === 'selector' || form.type === 'urltest'">
          <el-divider content-position="left">出站选择</el-divider>

          <el-form-item label="出站列表">
            <div class="outbound-list-container">
              <div class="selected-outbounds" v-if="form.options.outbounds && form.options.outbounds.length > 0">
                <div 
                  v-for="(tag, index) in form.options.outbounds" 
                  :key="tag"
                  class="outbound-item"
                  draggable="true"
                  @dragstart="onDragStart(index)"
                  @dragover.prevent="onDragOver(index)"
                  @drop="onDrop(index)"
                  @dragend="onDragEnd"
                  :class="{ 'dragging': dragIndex === index }"
                >
                  <span class="drag-handle">⋮⋮</span>
                  <span class="outbound-tag">{{ tag }}</span>
                  <el-button type="danger" link size="small" @click="removeOutbound(index)">
                    <el-icon><Close /></el-icon>
                  </el-button>
                </div>
              </div>
              <div v-else class="empty-outbounds">暂无出站，请添加</div>
              <el-select 
                v-model="addOutboundValue" 
                placeholder="添加出站" 
                @change="addOutbound"
                style="width: 100%; margin-top: 8px;"
                clearable
              >
                <el-option
                  v-for="ob in availableOutboundsToAdd"
                  :key="ob.tag"
                  :label="ob.tag"
                  :value="ob.tag"
                />
              </el-select>
            </div>
          </el-form-item>

          <el-form-item v-if="form.type === 'selector' || form.type === 'urltest'" label="默认出站">
            <el-select v-model="form.options.default" clearable placeholder="留空使用第一个">
              <el-option
                v-for="tag in form.options.outbounds"
                :key="tag"
                :label="tag"
                :value="tag"
              />
            </el-select>
          </el-form-item>

          <el-form-item v-if="form.type === 'urltest'" label="探测地址">
            <el-input v-model="form.options.url" placeholder="http://www.gstatic.com/generate_204" />
          </el-form-item>

          <el-form-item v-if="form.type === 'urltest'" label="探测间隔">
            <el-input v-model="form.options.interval" placeholder="3m" />
          </el-form-item>

          <el-form-item v-if="form.type === 'urltest'" label="容差 (ms)">
            <el-input-number v-model="form.options.tolerance" :min="0" :max="65535" />
          </el-form-item>
        </template>

        <!-- Bind Interface - only for direct -->
        <template v-if="form.type === 'direct'">
          <el-divider content-position="left">网络设置</el-divider>
          <el-form-item label="绑定网卡">
            <el-select v-model="form.options.bind_interface" placeholder="不指定 (使用系统默认)" clearable>
              <el-option
                v-for="iface in networkInterfaces"
                :key="iface.name"
                :label="iface.name"
                :value="iface.name"
              />
            </el-select>
          </el-form-item>
        </template>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveOutbound" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <!-- Import Dialog -->
    <el-dialog
      v-model="importDialogVisible"
      title="导入链接"
      width="600px"
    >
      <el-alert
        title="支持的链接格式"
        type="info"
        :closable="false"
        class="import-alert"
      >
        <template #default>
          <p>VMess: vmess://base64编码的配置</p>
          <p>VLESS: vless://uuid@server:port?params#名称</p>
        </template>
      </el-alert>

      <el-form label-width="80px">
        <el-form-item label="链接">
          <el-input
            v-model="importLink"
            type="textarea"
            :rows="4"
            placeholder="粘贴 vmess:// 或 vless:// 链接"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="importDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="doImport" :loading="importing">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { singboxApi } from '../../api/singbox'
import { configApi } from '../../api/config'
import { Download, Plus, Upload, Close } from '@element-plus/icons-vue'

const outbounds = ref([])
const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const importDialogVisible = ref(false)
const importLink = ref('')
const importing = ref(false)
const editingOutbound = ref(null)
const outboundTypes = ref([])
const networkInterfaces = ref([])
const formRef = ref(null)
const customizedEnabled = ref(false)
const customizedOutboundTypes = new Set(['fallback', 'loadbalance'])

// VLESS helper fields
const vlessSecurity = ref('')
const vlessFingerprint = ref('')
const vlessAlpn = ref('')
const vlessTransport = ref('tcp')

const form = ref({
  type: 'direct',
  tag: '',
  enabled: true,
  options: {}
})

const isCustomizedOutbound = (outbound) => customizedOutboundTypes.has(outbound?.type)

// Watch for outbounds list changes to sync default outbound
watch(() => form.value.options.outbounds, (newOutbounds, oldOutbounds) => {
  if (!newOutbounds || !form.value.options.default) return
  // If the default outbound is no longer in the list, clear it
  if (!newOutbounds.includes(form.value.options.default)) {
    form.value.options.default = ''
  }
}, { deep: true })

// Available outbounds for selector/urltest (exclude current editing outbound)
const availableOutbounds = computed(() => {
  return outbounds.value.filter(ob => {
    if (editingOutbound.value && ob.id === editingOutbound.value.id) return false
    return true
  })
})

// Available outbounds to add (exclude current editing outbound and already selected)
const availableOutboundsToAdd = computed(() => {
  const selected = form.value.options.outbounds || []
  return availableOutbounds.value.filter(ob => !selected.includes(ob.tag))
})

// Drag and drop state
const dragIndex = ref(null)
const dragOverIndex = ref(null)

// Drag and drop handlers
const onDragStart = (index) => {
  dragIndex.value = index
}

const onDragOver = (index) => {
  dragOverIndex.value = index
}

const onDrop = (index) => {
  if (dragIndex.value === null || dragIndex.value === index) return
  
  const outbounds = [...(form.value.options.outbounds || [])]
  const draggedItem = outbounds.splice(dragIndex.value, 1)[0]
  outbounds.splice(index, 0, draggedItem)
  form.value.options.outbounds = outbounds
  
  dragIndex.value = null
  dragOverIndex.value = null
}

const onDragEnd = () => {
  dragIndex.value = null
  dragOverIndex.value = null
}

// Add outbound to list
const addOutboundValue = ref('')
const addOutbound = (tag) => {
  if (!tag) return
  if (!form.value.options.outbounds) {
    form.value.options.outbounds = []
  }
  if (!form.value.options.outbounds.includes(tag)) {
    form.value.options.outbounds.push(tag)
    if (form.value.type === 'loadbalance') {
      if (!form.value.options.weights) form.value.options.weights = {}
      form.value.options.weights[tag] = 1
    }
  }
  addOutboundValue.value = ''
}

// Remove outbound from list
const removeOutbound = (index) => {
  if (form.value.options.outbounds) {
    const tag = form.value.options.outbounds[index]
    form.value.options.outbounds.splice(index, 1)
    if (form.value.type === 'loadbalance' && form.value.options.weights) {
      delete form.value.options.weights[tag]
    }
  }
}

const getLoadBalanceWeight = (tag) => {
  return form.value.options.weights?.[tag] || 1
}

const setLoadBalanceWeight = (tag, value) => {
  if (!form.value.options.weights) form.value.options.weights = {}
  form.value.options.weights[tag] = Number(value) > 0 ? Number(value) : 1
}

const rules = {
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  tag: [
    { required: true, message: '请输入标签', trigger: 'blur' }
  ]
}

const getTypeTag = (type) => {
  const map = {
    direct: 'success',
    block: 'danger',
    selector: 'warning',
    urltest: 'warning',
    shadowsocks: 'info',
    vmess: 'info',
    vless: 'info',
    trojan: 'info',
    hysteria: 'info'
  }
  return map[type] || 'info'
}

// VLESS security change handler
const onVlessSecurityChange = (val) => {
  if (!form.value.options.tls) {
    form.value.options.tls = {}
  }
  if (val === 'reality') {
    form.value.options.tls.enabled = true
    form.value.options.tls.reality = { enabled: true }
    // Reset transport to empty for Reality
    vlessTransport.value = ''
    delete form.value.options.transport
  } else if (val === 'tls') {
    form.value.options.tls.enabled = true
    delete form.value.options.tls.reality
  } else {
    delete form.value.options.tls
  }
}

// VLESS transport change handler
const onVlessTransportChange = (val) => {
  if (!val || val === 'tcp') {
    delete form.value.options.transport
  } else {
    form.value.options.transport = { type: val }
  }
}

// Sync helper fields to options on save
const syncVlessOptions = () => {
  if (form.value.type !== 'vless') return

  // Sync fingerprint
  if (vlessFingerprint.value && form.value.options.tls) {
    form.value.options.tls.utls = { enabled: true, fingerprint: vlessFingerprint.value }
  }

  // Sync ALPN
  if (vlessAlpn.value && form.value.options.tls) {
    form.value.options.tls.alpn = vlessAlpn.value.split(',')
  }

  // Sync transport host/path as string (for h2)
  if (form.value.options.transport && ['h2', 'http'].includes(vlessTransport.value)) {
    if (form.value.options.transport.host && Array.isArray(form.value.options.transport.host)) {
      form.value.options.transport.host = form.value.options.transport.host[0]
    }
  }
}

// Initialize VLESS helper fields from form
const initVlessHelpers = () => {
  if (form.value.type !== 'vless') return

  const tls = form.value.options.tls
  if (tls) {
    if (tls.reality) {
      vlessSecurity.value = 'reality'
    } else if (tls.enabled) {
      vlessSecurity.value = 'tls'
    } else {
      vlessSecurity.value = ''
    }
    if (tls.utls) {
      vlessFingerprint.value = tls.utls.fingerprint || ''
    }
    if (tls.alpn) {
      vlessAlpn.value = Array.isArray(tls.alpn) ? tls.alpn.join(',') : tls.alpn
    }
  } else {
    vlessSecurity.value = ''
    vlessFingerprint.value = ''
    vlessAlpn.value = ''
  }

  const transport = form.value.options.transport
  if (transport) {
    vlessTransport.value = transport.type || ''
  } else {
    // For Reality, default to empty; otherwise default to tcp
    vlessTransport.value = vlessSecurity.value === 'reality' ? '' : 'tcp'
  }
}

const loadOutbounds = async () => {
  loading.value = true
  try {
    const res = await singboxApi.getOutbounds()
    if (res.data.success) {
      outbounds.value = res.data.data || []
    }
  } catch (err) {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

const loadExperimental = async () => {
  try {
    const res = await configApi.get()
    if (res.data.success) {
      customizedEnabled.value = res.data.data?.customizedFeaturesEnabled === true
    }
  } catch (err) {
    console.error('Failed to load experimental config:', err)
  }
}

const loadTypes = async () => {
  try {
    const res = await singboxApi.getOutboundTypes()
    if (res.data.success) {
      outboundTypes.value = res.data.data || []
    }
  } catch (err) {
    console.error('Failed to load types:', err)
  }
}

const loadNetworkInterfaces = async () => {
  try {
    const res = await singboxApi.getNetworkInterfaces()
    if (res.data.success) {
      networkInterfaces.value = res.data.data || []
    }
  } catch (err) {
    console.error('Failed to load network interfaces:', err)
  }
}

const showAddDialog = () => {
  editingOutbound.value = null
  form.value = {
    type: 'direct',
    tag: '',
    enabled: true,
    options: {}
  }
  vlessSecurity.value = ''
  vlessFingerprint.value = ''
  vlessAlpn.value = ''
  vlessTransport.value = 'tcp'
  dialogVisible.value = true
}

// Fields that each outbound type supports in the form
const outboundTypeFields = {
  shadowsocks: ['server', 'server_port', 'password', 'method'],
  vmess: ['server', 'server_port', 'uuid', 'transport'],
  vless: ['server', 'server_port', 'uuid', 'flow', 'tls', 'transport'],
  trojan: ['server', 'server_port', 'password'],
  hysteria: ['server', 'server_port', 'auth', 'up_mbps', 'down_mbps'],
  direct: ['bind_interface'],
  block: [],
  selector: ['outbounds', 'default'],
  urltest: ['outbounds', 'default', 'url', 'interval', 'tolerance'],
  fallback: ['outbounds'],
  loadbalance: ['outbounds', 'strategy', 'weights']
}

const editOutbound = (outbound) => {
  editingOutbound.value = outbound
  // Only copy fields that the form supports for this type
  const allowedFields = outboundTypeFields[outbound.type] || []
  const cleanOptions = {}
  for (const key of allowedFields) {
    if (outbound.options && outbound.options[key] !== undefined) {
      cleanOptions[key] = outbound.options[key]
    }
  }
  form.value = {
    id: outbound.id,
    type: outbound.type,
    tag: outbound.tag,
    enabled: outbound.enabled,
    options: cleanOptions
  }
  initVlessHelpers()
  dialogVisible.value = true
}

const saveOutbound = async () => {
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  // Sync VLESS helper fields to options
  syncVlessOptions()

  form.value.tag = String(form.value.tag || '').trim()
  let overwrite = false
  const duplicate = outbounds.value.find(item =>
    item.id !== form.value.id && String(item.tag || '').trim() === form.value.tag
  )
  if (duplicate) {
    try {
      await ElMessageBox.confirm(
        `已存在名为「${form.value.tag}」的 Outbound，是否覆盖？`,
        '名称重复',
        { confirmButtonText: '覆盖', cancelButtonText: '取消', type: 'warning' }
      )
    } catch (err) {
      if (err === 'cancel' || err === 'close') return
      throw err
    }
    form.value.id = duplicate.id
    overwrite = true
  }

  saving.value = true
  try {
    if (editingOutbound.value || overwrite) {
      await singboxApi.updateOutbound(form.value)
      ElMessage.success('更新成功')
    } else {
      await singboxApi.addOutbound(form.value)
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    loadOutbounds()
  } catch (err) {
    ElMessage.error('保存失败: ' + (err.response?.data?.error || err.message))
  } finally {
    saving.value = false
  }
}

const toggleEnabled = async (outbound) => {
  try {
    await singboxApi.updateOutbound(outbound)
    ElMessage.success('状态已更新')
  } catch (err) {
    outbound.enabled = !outbound.enabled
    ElMessage.error('更新失败')
  }
}

const deleteOutbound = async (outbound) => {
  try {
    await ElMessageBox.confirm(`确定要删除 "${outbound.tag}" 吗？`, '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await singboxApi.deleteOutbound(outbound.id)
    ElMessage.success('删除成功')
    loadOutbounds()
  } catch (err) {
    if (err !== 'cancel') {
      const message = err.response?.data?.error || ''
      if (message.includes('used') || message.includes('引用') || message.includes('route')) {
        ElMessage.error('该出站已被路由规则使用，请先移除引用')
      } else {
        ElMessage.error('删除失败: ' + (message || err.message || '未知错误'))
      }
    }
  }
}

const showImportDialog = () => {
  importLink.value = ''
  importDialogVisible.value = true
}

// Parse vmess link
const parseVMess = (link) => {
  const b64 = link.slice(8) // remove vmess://
  const decoded = atob(b64)
  const info = JSON.parse(decoded)

  const options = {
    server: info.add,
    server_port: parseInt(info.port) || 443,
    uuid: info.id,
  }

  if (info.aid) options.alter_id = parseInt(info.aid)

  // Transport
  const net = info.net || 'tcp'
  if (net !== 'tcp') {
    const transport = { type: net }
    if (net === 'ws') {
      if (info.host) transport.host = [info.host]
      if (info.path) transport.path = info.path
    }
    if (net === 'grpc' && info.path) {
      transport.service_name = info.path
    }
    options.transport = transport
  }

  // TLS
  if (info.tls === 'tls') {
    options.tls = { enabled: true, server_name: info.sni || '' }
    if (info.fp) {
      options.tls.utls = { enabled: true, fingerprint: info.fp }
    }
  }

  return {
    type: 'vmess',
    tag: info.ps || `vmess-${info.add}:${info.port}`,
    enabled: true,
    options
  }
}

// Parse vless link
const parseVLESS = (link) => {
  const rest = link.slice(8) // remove vless://
  const parts = rest.split('#')
  const remark = parts[1] || ''

  const hostParams = parts[0].split('?')
  const query = {}
  if (hostParams[1]) {
    hostParams[1].split('&').forEach(p => {
      const kv = p.split('=')
      if (kv.length === 2) query[kv[0]] = decodeURIComponent(kv[1])
    })
  }

  const addrParts = hostParams[0].split('@')
  const userID = addrParts[0]
  const serverParts = addrParts[1].split(':')
  const server = serverParts[0]
  const port = parseInt(serverParts[1]) || 443

  const options = {
    server,
    server_port: port,
    uuid: userID,
    flow: query.flow || '',
  }

  // TLS / Reality
  const security = query.security
  if (security === 'reality') {
    options.tls = {
      enabled: true,
      server_name: query.sni || '',
      reality: { enabled: true, public_key: query.pbk || '' }
    }
    if (query.sid) options.tls.reality.short_id = query.sid
    if (query.fp) options.tls.utls = { enabled: true, fingerprint: query.fp }
    if (query.alpn) options.tls.alpn = query.alpn.split(',')
  } else if (security === 'tls') {
    options.tls = { enabled: true, server_name: query.sni || '' }
    if (query.fp) options.tls.utls = { enabled: true, fingerprint: query.fp }
    if (query.alpn) options.tls.alpn = query.alpn.split(',')
  }

  // Transport
  const type_ = query.type
  if (type_ && type_ !== 'tcp') {
    const transport = { type: type_ }
    if (type_ === 'ws') {
      if (query.host) transport.host = [query.host]
      if (query.path) transport.path = query.path
    }
    if (type_ === 'grpc') {
      if (query.serviceName) transport.service_name = query.serviceName
    }
    if (type_ === 'h2') {
      if (query.host) transport.host = query.host
      if (query.path) transport.path = query.path
    }
    options.transport = transport
  }

  return {
    type: 'vless',
    tag: remark || `vless-${server}:${port}`,
    enabled: true,
    options
  }
}

const doImport = () => {
  const link = importLink.value.trim()
  if (!link) {
    ElMessage.warning('请输入链接')
    return
  }

  try {
    let parsed
    if (link.startsWith('vmess://')) {
      parsed = parseVMess(link)
    } else if (link.startsWith('vless://')) {
      parsed = parseVLESS(link)
    } else {
      ElMessage.error('不支持的链接格式，仅支持 vmess:// 和 vless://')
      return
    }

    // Fill form with parsed data
    editingOutbound.value = null
    form.value = parsed
    importDialogVisible.value = false

    // Init helpers for vless
    if (parsed.type === 'vless') {
      initVlessHelpers()
    }

    // Open the add/edit dialog
    dialogVisible.value = true
    ElMessage.success('链接已解析，请确认后保存')
  } catch (err) {
    ElMessage.error('解析链接失败: ' + err.message)
  }
}

onMounted(() => {
  loadOutbounds()
  loadTypes()
  loadNetworkInterfaces()
  loadExperimental()
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

.import-alert {
  margin-bottom: 16px;
}

.import-alert p {
  margin: 4px 0;
  font-size: 12px;
}

.outbound-list-container {
  width: 100%;
}

.selected-outbounds {
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  padding: 8px;
  min-height: 40px;
  background: var(--el-fill-color-lighter);
}

.empty-outbounds {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  padding: 12px;
  text-align: center;
}

.outbound-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  margin-bottom: 4px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  cursor: move;
  transition: all 0.2s;
}

.outbound-item:last-child {
  margin-bottom: 0;
}

.outbound-item:hover {
  border-color: var(--el-color-primary);
}

.outbound-item.dragging {
  opacity: 0.5;
  border-color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}

.drag-handle {
  color: var(--el-text-color-secondary);
  margin-right: 8px;
  cursor: move;
  user-select: none;
}

.outbound-tag {
  flex: 1;
  font-size: 14px;
}
</style>
