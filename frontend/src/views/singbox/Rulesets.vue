<template>
  <div class="rulesets-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <el-icon><Document /></el-icon>
          <span>Ruleset 配置</span>
          <el-button type="success" @click="showGeoDialog" :loading="loadingGeo">
            <el-icon><Download /></el-icon>
            导入 GeoIP
          </el-button>
          <el-button type="warning" @click="showCommonDialog" :loading="loadingCommon">
            <el-icon><Download /></el-icon>
            导入常用规则
          </el-button>
          <el-button type="primary" @click="showAddDialog">
            <el-icon><Plus /></el-icon>
            添加
          </el-button>
        </div>
      </template>

      <el-table
        :data="rulesets"
        v-loading="loading"
        stripe
        @selection-change="onSelectionChange"
      >
        <el-table-column type="selection" width="45" />
        <el-table-column prop="tag" label="标签" width="180" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag size="small" :type="getTypeTag(row.type)">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="详情" min-width="300">
          <template #default="{ row }">
            <span v-if="row.type === 'inline'" class="detail-text">
              内联规则 ({{ getInlineRuleCount(row) }} 条)
            </span>
            <span v-else-if="row.type === 'local'" class="detail-text">
              {{ row.options?.path || '-' }}
            </span>
            <span v-else-if="row.type === 'remote'" class="detail-text">
              {{ row.options?.url || '-' }}
            </span>
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
            <el-button type="primary" link size="small" @click="editRuleset(row)">
              编辑
            </el-button>
            <el-button type="danger" link size="small" @click="deleteRuleset(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="selectedRulesets.length > 0" class="batch-actions">
        <el-button type="danger" @click="batchDelete">
          <el-icon><Delete /></el-icon>
          删除选中 ({{ selectedRulesets.length }})
        </el-button>
      </div>
    </el-card>

    <!-- Add/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="editingRuleset ? '编辑 Ruleset' : '添加 Ruleset'"
      width="780px"
    >
      <el-form :model="form" label-width="100px" ref="formRef" :rules="rules">
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" placeholder="选择类型">
            <el-option label="内联 (inline)" value="inline" />
            <el-option label="本地文件 (local)" value="local" />
            <el-option label="远程文件 (remote)" value="remote" />
          </el-select>
        </el-form-item>

        <el-form-item label="标签" prop="tag">
          <el-input v-model="form.tag" placeholder="例如: geosite-cn" />
        </el-form-item>

        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>

        <el-divider content-position="left">规则集选项</el-divider>

        <el-form-item label="格式 (format)">
          <el-select v-model="form.options.format" placeholder="选择格式">
            <el-option label="源文件 (source)" value="source" />
            <el-option label="二进制 (binary)" value="binary" />
          </el-select>
        </el-form-item>

        <!-- Inline type -->
        <template v-if="form.type === 'inline'">
          <el-form-item label="编辑方式">
            <el-radio-group v-model="inlineEditMode" size="small" @change="onInlineModeChange">
              <el-radio-button value="form">表单编辑</el-radio-button>
              <el-radio-button value="json">JSON 编辑</el-radio-button>
            </el-radio-group>
            <div class="form-tip">表单模式可可视化编辑规则组；JSON 模式可直接粘贴或修改规则数组</div>
          </el-form-item>

          <el-form-item v-if="inlineEditMode === 'json'" label="规则 (JSON)">
            <el-input
              v-model="form.options.rules"
              type="textarea"
              :rows="12"
              placeholder='[
  {
    "domain_suffix": ["dns.google", "github.com"]
  },
  {
    "invert": true,
    "source_ip_cidr": ["198.51.100.10/32"]
  }
]'
            />
            <div class="form-tip">JSON 格式的规则数组，参考 sing-box rule-set rules 格式</div>
          </el-form-item>

          <el-form-item v-else label="规则列表">
            <div class="rule-groups">
              <div v-for="(group, gi) in ruleGroups" :key="gi" class="rule-group">
                <div class="rule-group-header">
                  <span class="rule-group-title">规则组 {{ gi + 1 }}</span>
                  <div class="rule-group-actions">
                    <span class="invert-label">反转匹配</span>
                    <el-switch v-model="group.invert" size="small" />
                    <el-button
                      type="danger"
                      link
                      size="small"
                      :disabled="ruleGroups.length <= 1"
                      @click="removeRuleGroup(gi)"
                    >删除组</el-button>
                  </div>
                </div>

                <div v-for="(field, fi) in group.fields" :key="fi" class="rule-field-row">
                  <el-select v-model="field.key" class="rule-field-key" size="small" placeholder="字段类型">
                    <el-option
                      v-for="opt in RULE_FIELD_OPTIONS"
                      :key="opt.value"
                      :label="opt.label"
                      :value="opt.value"
                    />
                  </el-select>
                  <div class="rule-field-values">
                    <div v-for="(val, vi) in field.values" :key="vi" class="rule-value-row">
                      <el-input
                        v-model="field.values[vi]"
                        size="small"
                        :placeholder="getFieldPlaceholder(field.key)"
                      />
                      <el-button link size="small" class="rule-remove-btn" title="删除该条" @click="removeValue(field, vi)">
                        <el-icon><Close /></el-icon>
                      </el-button>
                    </div>
                    <el-button type="primary" link size="small" class="rule-add-value" @click="addValue(field)">
                      <el-icon><Plus /></el-icon>
                      添加值
                    </el-button>
                  </div>
                  <el-button type="danger" link size="small" title="删除该字段" @click="removeField(group, fi)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>

                <el-button type="primary" link size="small" class="rule-add-field" @click="addField(group)">
                  <el-icon><Plus /></el-icon>
                  添加匹配字段
                </el-button>
              </div>

              <el-button type="primary" plain class="rule-add-group" @click="addRuleGroup">
                <el-icon><Plus /></el-icon>
                添加规则组
              </el-button>
            </div>
          </el-form-item>
        </template>

        <!-- Local type -->
        <template v-if="form.type === 'local'">
          <el-form-item label="文件路径">
            <el-input v-model="form.options.path" placeholder="/path/to/ruleset.db" />
            <div class="form-tip">本地规则集文件路径</div>
          </el-form-item>
        </template>

        <!-- Remote type -->
        <template v-if="form.type === 'remote'">
          <el-form-item label="下载地址">
            <el-input v-model="form.options.url" placeholder="https://example.com/ruleset.db" />
          </el-form-item>
          <el-form-item label="HTTP 客户端">
            <el-select v-model="form.options.http_client" placeholder="选择 HTTP 客户端" filterable clearable>
              <el-option
                v-for="hc in httpClientList"
                :key="hc.tag"
                :label="hc.tag"
                :value="hc.tag"
              />
            </el-select>
            <div class="form-tip">用于下载规则集的 HTTP 客户端，留空使用默认</div>
          </el-form-item>

          <el-form-item label="更新间隔">
            <el-input v-model="form.options.update_interval" placeholder="例如: 1d" />
            <div class="form-tip">自动更新间隔，如 1d、12h、1w</div>
          </el-form-item>
        </template>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveRuleset" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <!-- GeoIP Import Dialog -->
    <el-dialog
      v-model="geoDialogVisible"
      title="导入 GeoIP 规则集"
      width="600px"
    >
      <div v-loading="loadingGeo">
        <el-input
          v-model="geoSearch"
          placeholder="搜索国家..."
          clearable
          class="geo-search"
        />
        <el-checkbox
          v-model="geoSelectAll"
          :indeterminate="geoIndeterminate"
          @change="onGeoSelectAll"
          class="geo-select-all"
        >全选</el-checkbox>
        <el-scrollbar height="400px" class="geo-scrollbar">
          <div v-for="item in filteredGeoList" :key="item.name" class="geo-item">
            <el-checkbox v-model="item.selected">
              <span class="geo-label">{{ item.label }}</span>
              <span class="geo-filename">{{ item.name }}</span>
            </el-checkbox>
          </div>
          <el-empty v-if="filteredGeoList.length === 0" description="无匹配结果" />
        </el-scrollbar>
      </div>
      <template #footer>
        <el-button @click="geoDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="importGeoRulesets" :loading="importingGeo" :disabled="geoSelectedCount === 0">
          导入 ({{ geoSelectedCount }})
        </el-button>
      </template>
    </el-dialog>

    <!-- Common Rulesets Import Dialog -->
    <el-dialog
      v-model="commonDialogVisible"
      title="导入常用规则集"
      width="600px"
    >
      <div v-loading="loadingCommon">
        <el-input
          v-model="commonSearch"
          placeholder="搜索规则集..."
          clearable
          class="geo-search"
        />
        <el-checkbox
          v-model="commonSelectAll"
          :indeterminate="commonIndeterminate"
          @change="onCommonSelectAll"
          class="geo-select-all"
        >全选</el-checkbox>
        <el-scrollbar height="400px" class="geo-scrollbar">
          <div v-for="item in filteredCommonList" :key="item.name" class="geo-item">
            <el-checkbox v-model="item.selected">
              <span class="geo-label">{{ item.label }}</span>
              <span class="geo-filename">{{ item.name }}</span>
            </el-checkbox>
          </div>
          <el-empty v-if="filteredCommonList.length === 0" description="无匹配结果" />
        </el-scrollbar>
      </div>
      <template #footer>
        <el-button @click="commonDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="importCommonRulesets" :loading="importingCommon" :disabled="commonSelectedCount === 0">
          导入 ({{ commonSelectedCount }})
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { singboxApi } from '../../api/singbox'
import { Document, Plus, Download, Delete, Close } from '@element-plus/icons-vue'

const rulesets = ref([])
const httpClientList = ref([])
const selectedRulesets = ref([])
const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const editingRuleset = ref(null)
const formRef = ref(null)

// GeoIP import state
const geoDialogVisible = ref(false)
const loadingGeo = ref(false)
const importingGeo = ref(false)
const geoSearch = ref('')
const geoList = ref([])
const geoSelectAll = ref(false)

// Common rulesets import state
const commonDialogVisible = ref(false)
const loadingCommon = ref(false)
const importingCommon = ref(false)
const commonSearch = ref('')
const commonList = ref([])
const commonSelectAll = ref(false)

const GEO_RAW_BASE = 'https://github.com/SagerNet/sing-geoip/raw/refs/heads/rule-set'

const GEO_NAMES = {
  ac: '阿森松岛', ad: '安道尔', ae: '阿联酋', af: '阿富汗', ag: '安提瓜和巴布达',
  ai: '安圭拉', al: '阿尔巴尼亚', am: '亚美尼亚', ao: '安哥拉', aq: '南极洲',
  ar: '阿根廷', as: '美属萨摩亚', at: '奥地利', au: '澳大利亚', aw: '阿鲁巴',
  ax: '奥兰群岛', az: '阿塞拜疆', ba: '波黑', bb: '巴巴多斯', bd: '孟加拉国',
  be: '比利时', bf: '布基纳法索', bg: '保加利亚', bh: '巴林', bi: '布隆迪',
  bj: '贝宁', bl: '圣巴泰勒米', bm: '百慕大', bn: '文莱', bo: '玻利维亚',
  bq: '荷属加勒比', br: '巴西', bs: '巴哈马', bt: '不丹', bw: '博茨瓦纳',
  by: '白俄罗斯', bz: '伯利兹', ca: '加拿大', cc: '科科斯群岛', cd: '刚果(金)',
  cf: '中非共和国', cg: '刚果(布)', ch: '瑞士', ci: '科特迪瓦', ck: '库克群岛',
  cl: '智利', cm: '喀麦隆', cn: '中国', co: '哥伦比亚', cr: '哥斯达黎加',
  cu: '古巴', cv: '佛得角', cw: '库拉索', cx: '圣诞岛', cy: '塞浦路斯',
  cz: '捷克', de: '德国', dj: '吉布提', dk: '丹麦', dm: '多米尼克',
  do: '多米尼加', dz: '阿尔及利亚', ec: '厄瓜多尔', ee: '爱沙尼亚', eg: '埃及',
  er: '厄立特里亚', es: '西班牙', et: '埃塞俄比亚', fi: '芬兰', fj: '斐济',
  fk: '福克兰群岛', fm: '密克罗尼西亚', fo: '法罗群岛', fr: '法国', ga: '加蓬',
  gb: '英国', gd: '格林纳达', ge: '格鲁吉亚', gf: '法属圭亚那', gg: '根西岛',
  gh: '加纳', gi: '直布罗陀', gl: '格陵兰', gm: '冈比亚', gn: '几内亚',
  gp: '瓜德罗普', gq: '赤道几内亚', gr: '希腊', gs: '南乔治亚', gt: '危地马拉',
  gu: '关岛', gw: '几内亚比绍', gy: '圭亚那', hk: '中国香港', hn: '洪都拉斯',
  hr: '克罗地亚', ht: '海地', hu: '匈牙利', id: '印度尼西亚', ie: '爱尔兰',
  il: '以色列', im: '马恩岛', in: '印度', io: '英属印度洋领地', iq: '伊拉克',
  ir: '伊朗', is: '冰岛', it: '意大利', je: '泽西岛', jm: '牙买加',
  jo: '约旦', jp: '日本', ke: '肯尼亚', kg: '吉尔吉斯斯坦', kh: '柬埔寨',
  ki: '基里巴斯', km: '科摩罗', kn: '圣基茨和尼维斯', kp: '朝鲜', kr: '韩国',
  kw: '科威特', ky: '开曼群岛', kz: '哈萨克斯坦', la: '老挝', lb: '黎巴嫩',
  lc: '圣卢西亚', li: '列支敦士登', lk: '斯里兰卡', lr: '利比里亚', ls: '莱索托',
  lt: '立陶宛', lu: '卢森堡', lv: '拉脱维亚', ly: '利比亚', ma: '摩洛哥',
  mc: '摩纳哥', md: '摩尔多瓦', me: '黑山', mf: '法属圣马丁', mg: '马达加斯加',
  mh: '马绍尔群岛', mk: '北马其顿', ml: '马里', mm: '缅甸', mn: '蒙古',
  mo: '中国澳门', mp: '北马里亚纳群岛', mq: '马提尼克', mr: '毛里塔尼亚',
  ms: '蒙特塞拉特', mt: '马耳他', mu: '毛里求斯', mv: '马尔代夫', mw: '马拉维',
  mx: '墨西哥', my: '马来西亚', mz: '莫桑比克', na: '纳米比亚', nc: '新喀里多尼亚',
  ne: '尼日尔', nf: '诺福克岛', ng: '尼日利亚', ni: '尼加拉瓜', nl: '荷兰',
  no: '挪威', np: '尼泊尔', nr: '瑙鲁', nu: '纽埃', nz: '新西兰',
  om: '阿曼', pa: '巴拿马', pe: '秘鲁', pf: '法属波利尼西亚', pg: '巴布亚新几内亚',
  ph: '菲律宾', pk: '巴基斯坦', pl: '波兰', pm: '圣皮埃尔和密克隆', pn: '皮特凯恩群岛',
  pr: '波多黎各', ps: '巴勒斯坦', pt: '葡萄牙', pw: '帕劳', py: '巴拉圭',
  qa: '卡塔尔', re: '留尼汪', ro: '罗马尼亚', rs: '塞尔维亚', ru: '俄罗斯',
  rw: '卢旺达', sa: '沙特阿拉伯', sb: '所罗门群岛', sc: '塞舌尔', sd: '苏丹',
  se: '瑞典', sg: '新加坡', sh: '圣赫勒拿', si: '斯洛文尼亚', sj: '斯瓦尔巴和扬马延',
  sk: '斯洛伐克', sl: '塞拉利昂', sm: '圣马力诺', sn: '塞内加尔', so: '索马里',
  sr: '苏里南', ss: '南苏丹', st: '圣多美和普林西比', sv: '萨尔瓦多', sx: '荷属圣马丁',
  sy: '叙利亚', sz: '斯威士兰', tc: '特克斯和凯科斯群岛', td: '乍得', tf: '法属南部领地',
  tg: '多哥', th: '泰国', tj: '塔吉克斯坦', tk: '托克劳', tl: '东帝汶',
  tm: '土库曼斯坦', tn: '突尼斯', to: '汤加', tr: '土耳其', tt: '特立尼达和多巴哥',
  tv: '图瓦卢', tw: '中国台湾', tz: '坦桑尼亚', ua: '乌克兰', ug: '乌干达',
  um: '美国本土外小岛屿', us: '美国', uy: '乌拉圭', uz: '乌兹别克斯坦',
  va: '梵蒂冈', vc: '圣文森特和格林纳丁斯', ve: '委内瑞拉', vg: '英属维尔京群岛',
  vi: '美属维尔京群岛', vn: '越南', vu: '瓦努阿图', wf: '瓦利斯和富图纳',
  ws: '萨摩亚', xk: '科索沃', ye: '也门', yt: '马约特', za: '南非',
  zm: '赞比亚', zw: '津巴布韦'
}

const COMMON_RAW_BASE = 'https://github.com/DustinWin/ruleset_geodata/raw/refs/heads/sing-box-ruleset'

const COMMON_RULES_DESC = {
  'ads': '广告拦截',
  'ai': 'AI 服务',
  'apple-cn': 'Apple 中国区服务',
  'appletv': 'Apple TV',
  'applications': '常用应用',
  'bilibili': '哔哩哔哩',
  'cn': '中国域名',
  'cn-lite': '中国域名（精简版）',
  'cnip': '中国 IP 段',
  'disney': 'Disney+ 流媒体',
  'fakeip-filter': 'FakeIP 过滤列表',
  'fakeip-filter-lite': 'FakeIP 过滤列表（精简版）',
  'games': '游戏平台',
  'games-cn': '中国游戏平台',
  'gfw': 'GFW 列表',
  'google-cn': 'Google 中国区服务',
  'max': 'Max (HBO) 流媒体',
  'media': '流媒体服务',
  'mediaip': '流媒体 IP 段',
  'microsoft-cn': '微软中国区服务',
  'netflix': 'Netflix 流媒体',
  'netflixip': 'Netflix IP 段',
  'networktest': '网络测试',
  'primevideo': 'Amazon Prime Video',
  'private': '私有网络域名',
  'privateip': '私有网络 IP 段',
  'proxy': '代理域名',
  'spotify': 'Spotify 流媒体',
  'telegramip': 'Telegram IP 段',
  'tiktok': 'TikTok',
  'tld-proxy': '顶级域名代理',
  'trackerslist': 'BT 追踪器列表',
  'youtube': 'YouTube'
}

const form = ref({
  type: 'inline',
  tag: '',
  enabled: true,
  options: {
    format: 'source',
    rules: '',
    path: '',
    url: '',
    update_interval: '',
    http_client: '',
  }
})

// ---- Inline rules form editor ----
const inlineEditMode = ref('form')
const ruleGroups = ref([])

const RULE_FIELD_OPTIONS = [
  { value: 'domain', label: 'domain（完整域名）', placeholder: '例如: example.com' },
  { value: 'domain_suffix', label: 'domain_suffix（域名后缀）', placeholder: '例如: github.com' },
  { value: 'domain_keyword', label: 'domain_keyword（域名关键字）', placeholder: '例如: google' },
  { value: 'domain_regex', label: 'domain_regex（域名正则）', placeholder: '例如: (^|\\.)example\\.com$' },
  { value: 'ip_cidr', label: 'ip_cidr（目标 IP）', placeholder: '例如: 1.2.3.0/24' },
  { value: 'source_ip_cidr', label: 'source_ip_cidr（来源 IP）', placeholder: '例如: 192.168.1.0/24' },
  { value: 'port', label: 'port（目标端口）', placeholder: '例如: 80', number: true },
  { value: 'source_port', label: 'source_port（来源端口）', placeholder: '例如: 443', number: true },
  { value: 'port_range', label: 'port_range（端口范围）', placeholder: '例如: 8000-9000' },
  { value: 'source_port_range', label: 'source_port_range（来源端口范围）', placeholder: '例如: 8000-9000' },
  { value: 'process_name', label: 'process_name（进程名）', placeholder: '例如: curl' },
  { value: 'process_path', label: 'process_path（进程路径）', placeholder: '例如: /usr/bin/curl' },
  { value: 'process_path_regex', label: 'process_path_regex（进程路径正则）', placeholder: '正则表达式' },
  { value: 'package_name', label: 'package_name（应用包名）', placeholder: '例如: com.android.chrome' },
  { value: 'network', label: 'network（网络协议）', placeholder: '例如: tcp' },
  { value: 'source_mac_address', label: 'source_mac_address（来源 MAC）', placeholder: '例如: 00:11:22:33:44:55' },
  { value: 'source_hostname', label: 'source_hostname（来源主机名）', placeholder: '例如: mypc.local' },
  { value: 'query_type', label: 'query_type（DNS 查询类型）', placeholder: '例如: A' }
]

const RULE_FIELD_KEY_SET = new Set(RULE_FIELD_OPTIONS.map(opt => opt.value))
const RULE_NUMBER_FIELD_SET = new Set(
  RULE_FIELD_OPTIONS.filter(opt => opt.number).map(opt => opt.value)
)

const getFieldPlaceholder = (key) => {
  const opt = RULE_FIELD_OPTIONS.find(item => item.value === key)
  return opt ? opt.placeholder : '输入值后回车确认'
}

// 将规则 JSON 数组转换为表单模型（无法识别的字段保留到 unknown，序列化时原样合并）
const parseRulesToGroups = (rules) => {
  if (!Array.isArray(rules)) return []
  return rules.map(item => {
    if (typeof item !== 'object' || item === null || Array.isArray(item)) {
      return { invert: false, fields: [], unknown: { __raw: item } }
    }
    const { invert, ...rest } = item
    const fields = []
    const unknown = {}
    for (const [key, value] of Object.entries(rest)) {
      if (RULE_FIELD_KEY_SET.has(key)) {
        const arr = Array.isArray(value) ? value : [value]
        fields.push({ key, values: arr.map(v => String(v)) })
      } else {
        unknown[key] = value
      }
    }
    return { invert: !!invert, fields, unknown }
  })
}

// 将表单模型序列化为规则 JSON 数组
const groupsToRules = (groups) => {
  return (groups || []).map(group => {
    const rule = {}
    if (group.invert) rule.invert = true
    for (const field of group.fields || []) {
      if (!field.key) continue
      const values = (field.values || []).filter(v => v !== '' && v !== null && v !== undefined)
      if (values.length === 0) continue
      rule[field.key] = RULE_NUMBER_FIELD_SET.has(field.key)
        ? values.map(v => Number(v))
        : values.map(v => String(v))
    }
    if (group.unknown && typeof group.unknown === 'object') {
      Object.assign(rule, group.unknown)
    }
    return rule
  })
}

// 从 form.options.rules 同步表单模型（解析失败时置空）
const syncRuleGroupsFromRules = () => {
  try {
    ruleGroups.value = parseRulesToGroups(JSON.parse(form.value.options.rules || '[]'))
  } catch {
    ruleGroups.value = []
  }
}

const onInlineModeChange = (mode) => {
  if (mode === 'form') {
    try {
      ruleGroups.value = parseRulesToGroups(JSON.parse(form.value.options.rules || '[]'))
    } catch {
      ElMessage.error('当前规则 JSON 无法解析，已停留在 JSON 编辑模式')
      inlineEditMode.value = 'json'
    }
  } else {
    form.value.options.rules = JSON.stringify(groupsToRules(ruleGroups.value), null, 2)
  }
}

const newRuleField = () => ({ key: 'domain_suffix', values: [] })

const addRuleGroup = () => {
  ruleGroups.value.push({ invert: false, fields: [newRuleField()], unknown: {} })
}

const removeRuleGroup = (index) => {
  if (ruleGroups.value.length <= 1) return
  ruleGroups.value.splice(index, 1)
}

const addField = (group) => {
  group.fields.push(newRuleField())
}

const removeField = (group, index) => {
  group.fields.splice(index, 1)
}

const addValue = (field) => {
  field.values.push('')
}

const removeValue = (field, index) => {
  field.values.splice(index, 1)
}

const rules = {
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  tag: [{ required: true, message: '请输入标签', trigger: 'blur' }]
}

const getTypeTag = (type) => {
  const map = { inline: 'success', local: 'info', remote: 'warning' }
  return map[type] || 'info'
}

const getInlineRuleCount = (ruleset) => {
  const rules = ruleset.options?.rules
  if (!rules) return 0
  if (Array.isArray(rules)) return rules.length
  try {
    return JSON.parse(rules).length
  } catch {
    return 0
  }
}

const loadRulesets = async () => {
  loading.value = true
  try {
    const [rsRes, hcRes] = await Promise.all([
      singboxApi.getRulesets(),
      singboxApi.getHTTPClients()
    ])
    if (rsRes.data.success) {
      rulesets.value = rsRes.data.data || []
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
  editingRuleset.value = null
  form.value = {
    type: 'inline',
    tag: '',
    enabled: true,
    options: { format: 'source', rules: '', path: '', url: '', update_interval: '', http_client: '' }
  }
  inlineEditMode.value = 'form'
  syncRuleGroupsFromRules()
  dialogVisible.value = true
}

const editRuleset = (ruleset) => {
  editingRuleset.value = ruleset
  const opts = { ...ruleset.options }
  form.value = {
    id: ruleset.id,
    type: ruleset.type,
    tag: ruleset.tag,
    enabled: ruleset.enabled,
    options: {
      format: opts.format || 'source',
      rules: typeof opts.rules === 'string' ? opts.rules : JSON.stringify(opts.rules || [], null, 2),
      path: opts.path || '',
      url: opts.url || '',
      update_interval: opts.update_interval || '',
      http_client: opts.http_client || ''
    }
  }
  inlineEditMode.value = 'form'
  syncRuleGroupsFromRules()
  dialogVisible.value = true
}

const saveRuleset = async () => {
  try {
    await formRef.value.validate()
  } catch { return }

  const cleanOptions = { format: form.value.options.format }
  if (form.value.type === 'inline') {
    // 表单模式下先从表单同步到 JSON 文本
    if (inlineEditMode.value === 'form') {
      form.value.options.rules = JSON.stringify(groupsToRules(ruleGroups.value), null, 2)
    }
    // Parse JSON rules
    try {
      cleanOptions.rules = JSON.parse(form.value.options.rules || '[]')
    } catch {
      ElMessage.error('规则 JSON 格式错误')
      return
    }
  } else if (form.value.type === 'local') {
    cleanOptions.path = form.value.options.path
  } else if (form.value.type === 'remote') {
    cleanOptions.url = form.value.options.url
    if (form.value.options.update_interval) cleanOptions.update_interval = form.value.options.update_interval
    if (form.value.options.http_client) cleanOptions.http_client = form.value.options.http_client
  }

  const data = { id: form.value.id, type: form.value.type, tag: form.value.tag, enabled: form.value.enabled, options: cleanOptions }

  saving.value = true
  try {
    if (editingRuleset.value) {
      await singboxApi.updateRuleset(data)
      ElMessage.success('更新成功')
    } else {
      await singboxApi.addRuleset(data)
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    loadRulesets()
  } catch (err) {
    ElMessage.error('保存失败: ' + (err.response?.data?.error || err.message))
  } finally {
    saving.value = false
  }
}

const toggleEnabled = async (ruleset) => {
  try {
    await singboxApi.updateRuleset(ruleset)
    ElMessage.success('状态已更新')
  } catch {
    ruleset.enabled = !ruleset.enabled
    ElMessage.error('更新失败')
  }
}

const deleteRuleset = async (ruleset) => {
  try {
    await ElMessageBox.confirm(`确定要删除 "${ruleset.tag}" 吗？`, '确认删除', {
      confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
    })
    await singboxApi.deleteRuleset(ruleset.id)
    ElMessage.success('删除成功')
    loadRulesets()
  } catch (err) {
    if (err !== 'cancel') ElMessage.error('删除失败')
  }
}

const onSelectionChange = (val) => {
  selectedRulesets.value = val
}

const batchDelete = async () => {
  if (selectedRulesets.value.length === 0) return
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedRulesets.value.length} 个规则集吗？`, '批量删除', {
      confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
    })
    const ids = selectedRulesets.value.map(rs => rs.id)
    await singboxApi.deleteRulesets(ids)
    ElMessage.success(`成功删除 ${ids.length} 个规则集`)
    loadRulesets()
  } catch (err) {
    if (err !== 'cancel') ElMessage.error('删除失败')
  }
}

// GeoIP Import
const filteredGeoList = computed(() => {
  const q = geoSearch.value.toLowerCase()
  return geoList.value.filter(item =>
    !q || item.label.toLowerCase().includes(q) || item.name.toLowerCase().includes(q)
  )
})

const geoSelectedCount = computed(() => geoList.value.filter(i => i.selected).length)
const geoIndeterminate = computed(() => {
  const count = geoSelectedCount.value
  return count > 0 && count < geoList.value.length
})

const onGeoSelectAll = (val) => {
  filteredGeoList.value.forEach(item => { item.selected = val })
}

const showGeoDialog = async () => {
  geoDialogVisible.value = true
  geoSearch.value = ''
  geoSelectAll.value = false

  if (geoList.value.length > 0) return

  loadingGeo.value = true
  try {
    const res = await singboxApi.fetchGeoTree()
    if (!res.data.success) {
      ElMessage.error('获取 GeoIP 列表失败: ' + (res.data.error || '未知错误'))
      return
    }
    const data = res.data.data
    const items = (data.tree || [])
      .filter(f => f.type === 'blob' && f.path.endsWith('.srs'))
      .map(f => {
        const filename = f.path.replace('rule-set/', '').replace('.srs', '')
        const parts = filename.split('-')
        const code = parts[parts.length - 1]
        return {
          name: filename,
          label: GEO_NAMES[code] || filename,
          selected: false
        }
      })
      .sort((a, b) => a.label.localeCompare(b.label, 'zh-CN'))
    geoList.value = items
  } catch (err) {
    ElMessage.error('获取 GeoIP 列表失败: ' + (err.response?.data?.error || err.message || '网络错误'))
  } finally {
    loadingGeo.value = false
  }
}

const importGeoRulesets = async () => {
  const selected = geoList.value.filter(i => i.selected)
  if (selected.length === 0) return

  importingGeo.value = true
  try {
    const rulesets = selected.map(item => ({
      type: 'remote',
      tag: item.label,
      enabled: true,
      options: {
        format: 'binary',
        url: `${GEO_RAW_BASE}/${item.name}.srs`,
        update_interval: '1d'
      }
    }))
    await singboxApi.addRulesets({ rulesets })
    ElMessage.success(`成功导入 ${rulesets.length} 个规则集`)
    geoDialogVisible.value = false
    geoList.value = []
    loadRulesets()
  } catch (err) {
    ElMessage.error('导入失败: ' + (err.response?.data?.error || err.message))
  } finally {
    importingGeo.value = false
  }
}

// Common rulesets
const filteredCommonList = computed(() => {
  const q = commonSearch.value.toLowerCase()
  return commonList.value.filter(item =>
    !q || item.label.toLowerCase().includes(q) || item.name.toLowerCase().includes(q) || (item.desc && item.desc.toLowerCase().includes(q))
  )
})

const commonSelectedCount = computed(() => commonList.value.filter(i => i.selected).length)
const commonIndeterminate = computed(() => {
  const count = commonSelectedCount.value
  return count > 0 && count < commonList.value.length
})

const onCommonSelectAll = (val) => {
  filteredCommonList.value.forEach(item => { item.selected = val })
}

const showCommonDialog = async () => {
  commonDialogVisible.value = true
  commonSearch.value = ''
  commonSelectAll.value = false

  if (commonList.value.length > 0) return

  loadingCommon.value = true
  try {
    const res = await singboxApi.fetchCommonRulesetTree()
    if (!res.data.success) {
      ElMessage.error('获取常用规则集列表失败: ' + (res.data.error || '未知错误'))
      return
    }
    const data = res.data.data
    const items = (data.tree || [])
      .filter(f => f.type === 'blob' && f.path.endsWith('.srs'))
      .map(f => {
        const filename = f.path.replace('sing-box-ruleset/', '').replace('.srs', '')
        return {
          name: filename,
          label: COMMON_RULES_DESC[filename] || filename,
          selected: false
        }
      })
      .sort((a, b) => {
        // Sort: items with description first, then alphabetically
        if (a.desc && !b.desc) return -1
        if (!a.desc && b.desc) return 1
        return a.label.localeCompare(b.label, 'zh-CN')
      })
    commonList.value = items
  } catch (err) {
    ElMessage.error('获取常用规则集列表失败: ' + (err.response?.data?.error || err.message || '网络错误'))
  } finally {
    loadingCommon.value = false
  }
}

const importCommonRulesets = async () => {
  const selected = commonList.value.filter(i => i.selected)
  if (selected.length === 0) return

  importingCommon.value = true
  try {
    const rulesets = selected.map(item => ({
      type: 'remote',
      tag: item.label,
      enabled: true,
      options: {
        format: 'binary',
        url: `${COMMON_RAW_BASE}/${item.name}.srs`,
        update_interval: '1d'
      }
    }))
    await singboxApi.addRulesets({ rulesets })
    ElMessage.success(`成功导入 ${rulesets.length} 个规则集`)
    commonDialogVisible.value = false
    commonList.value = []
    loadRulesets()
  } catch (err) {
    ElMessage.error('导入失败: ' + (err.response?.data?.error || err.message))
  } finally {
    importingCommon.value = false
  }
}

onMounted(() => {
  loadRulesets()
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

.card-header .el-button + .el-button {
  margin-left: 8px;
}

.batch-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}

.batch-actions .el-button {
  margin-left: 0;
}

.form-tip {
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-secondary);
}

.rule-groups {
  width: 100%;
  max-height: 400px;
  overflow-y: auto;
  padding-right: 4px;
}

.rule-group {
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 10px 12px;
  margin-bottom: 10px;
  background: var(--bg-page);
}

.rule-group-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.rule-group-title {
  font-weight: 600;
  font-size: 13px;
}

.rule-group-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.invert-label {
  font-size: 12px;
  color: var(--text-secondary);
}

.rule-field-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 8px;
}

.rule-field-key {
  width: 220px;
  flex-shrink: 0;
}

.rule-field-values {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.rule-value-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.rule-value-row .el-input {
  flex: 1;
}

.rule-add-value {
  margin-left: 0;
  align-self: flex-start;
}

.rule-remove-btn {
  color: var(--text-secondary);
}

.rule-remove-btn:hover {
  color: var(--el-color-danger);
}

.rule-add-field {
  margin-left: 0;
}

.rule-add-group {
  width: 100%;
}

.detail-text {
  font-size: 13px;
  color: var(--text-secondary);
  word-break: break-all;
}

.geo-search {
  margin-bottom: 12px;
}

.geo-select-all {
  margin-bottom: 12px;
}

.geo-scrollbar {
  border: 1px solid var(--border-color);
  border-radius: 6px;
}

.geo-item {
  padding: 6px 12px;
}

.geo-item:hover {
  background: var(--bg-page);
}

.geo-label {
  margin-right: 8px;
  font-weight: 500;
}

.geo-filename {
  color: var(--text-secondary);
  font-size: 12px;
}
</style>
