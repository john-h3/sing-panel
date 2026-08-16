<template>
  <div class="diff-view">
    <div v-if="view === 'split'" class="diff-view-header">
      <div class="dv-col-title">{{ localTitle }}</div>
      <div class="dv-col-title">{{ remoteTitle }}</div>
    </div>
    <div class="dv-body">
      <template v-if="view === 'split'">
        <div
          v-for="(row, i) in rows"
          :key="i"
          class="dv-row"
          :class="row.kind === 'pair' ? 'dv-row-pair' : ''"
        >
          <div v-if="row.kind === 'skip'" class="dv-skip">
            跳过 {{ row.count }} 行未变化内容
          </div>
          <template v-else>
            <div class="dv-cell" :class="leftCellClass(row)">
              <span class="dv-lineno">{{ row.leftNo || '' }}</span>
              <span class="dv-marker">{{ leftMarker(row) }}</span>
              <span class="dv-text">
                <template v-if="row.kind === 'pair'">
                  <span
                    v-for="(seg, si) in row.leftSegs"
                    :key="si"
                    :class="seg.cls"
                  >{{ seg.text }}</span>
                </template>
                <template v-else>{{ row.left }}</template>
              </span>
            </div>
            <div class="dv-cell" :class="rightCellClass(row)">
              <span class="dv-lineno">{{ row.rightNo || '' }}</span>
              <span class="dv-marker">{{ rightMarker(row) }}</span>
              <span class="dv-text">
                <template v-if="row.kind === 'pair'">
                  <span
                    v-for="(seg, si) in row.rightSegs"
                    :key="si"
                    :class="seg.cls"
                  >{{ seg.text }}</span>
                </template>
                <template v-else>{{ row.right }}</template>
              </span>
            </div>
          </template>
        </div>
      </template>
      <template v-else>
        <div
          v-for="(row, i) in unifiedRows"
          :key="i"
          class="dv-row dv-row-unified"
        >
          <div v-if="row.kind === 'skip'" class="dv-skip">
            跳过 {{ row.count }} 行未变化内容
          </div>
          <div v-else class="dv-cell" :class="unifiedCellClass(row)">
            <span class="dv-lineno">{{ row.leftNo || row.rightNo || '' }}</span>
            <span class="dv-marker">{{ unifiedMarker(row) }}</span>
            <span class="dv-text">
              <span
                v-for="(seg, si) in row.segs"
                :key="si"
                :class="seg.cls"
              >{{ seg.text }}</span>
            </span>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { diffLines, diffChars } from 'diff'

const props = defineProps({
  local: { type: String, default: '' },
  remote: { type: String, default: '' },
  localTitle: { type: String, default: '本机' },
  remoteTitle: { type: String, default: '远端' },
  view: { type: String, default: 'split' } // "split" | "unified"
})

const toLines = (v) => {
  const lines = (v || '').split('\n')
  if (lines[lines.length - 1] === '') lines.pop()
  return lines
}

const rows = computed(() => {
  const chunks = diffLines(props.local || '', props.remote || '')
  const raw = []
  for (const chunk of chunks) {
    for (const line of toLines(chunk.value)) {
      if (chunk.removed) raw.push({ kind: 'removed', line })
      else if (chunk.added) raw.push({ kind: 'added', line })
      else raw.push({ kind: 'context', line })
    }
  }

  // Collapse long unchanged runs, keep 3 lines of context around changes
  const collapsed = []
  for (let i = 0; i < raw.length;) {
    if (raw[i].kind !== 'context') {
      collapsed.push(raw[i])
      i++
      continue
    }
    let j = i
    while (j < raw.length && raw[j].kind === 'context') j++
    const n = j - i
    if (n <= 8) {
      for (let k = i; k < j; k++) collapsed.push(raw[k])
    } else {
      for (let k = i; k < i + 3; k++) collapsed.push(raw[k])
      collapsed.push({ kind: 'skip', count: n - 6 })
      for (let k = j - 3; k < j; k++) collapsed.push(raw[k])
    }
    i = j
  }

  // Pair removed lines with following added lines for inline highlighting
  const result = []
  let leftNo = 1
  let rightNo = 1
  for (let i = 0; i < collapsed.length; i++) {
    const row = collapsed[i]
    if (row.kind === 'skip') {
      result.push(row)
      continue
    }
    if (row.kind === 'removed') {
      const removed = []
      while (i < collapsed.length && collapsed[i].kind === 'removed') removed.push(collapsed[i++])
      const added = []
      while (i < collapsed.length && collapsed[i].kind === 'added') added.push(collapsed[i++])
      const pairCount = Math.min(removed.length, added.length)
      for (let k = 0; k < removed.length; k++) {
        if (k < pairCount) {
          const segs = diffChars(removed[k].line, added[k].line)
          result.push({
            kind: 'pair',
            leftNo: leftNo++,
            rightNo: rightNo++,
            leftSegs: segs
              .filter(s => !s.added)
              .map(s => ({ text: s.value, cls: s.removed ? 'seg-del' : 'seg-same' })),
            rightSegs: segs
              .filter(s => !s.removed)
              .map(s => ({ text: s.value, cls: s.added ? 'seg-ins' : 'seg-same' }))
          })
        } else {
          result.push({ kind: 'removed', left: removed[k].line, leftNo: leftNo++, rightNo: null })
        }
      }
      for (let k = pairCount; k < added.length; k++) {
        result.push({ kind: 'added', right: added[k].line, leftNo: null, rightNo: rightNo++ })
      }
      i--
    } else if (row.kind === 'added') {
      result.push({ kind: 'added', right: row.line, leftNo: null, rightNo: rightNo++ })
    } else {
      result.push({ kind: 'context', left: row.line, right: row.line, leftNo: leftNo++, rightNo: rightNo++ })
    }
  }
  return result
})

const leftMarker = (row) => {
  if (row.kind === 'pair' || row.kind === 'removed') return '-'
  return ' '
}

const rightMarker = (row) => {
  if (row.kind === 'pair' || row.kind === 'added') return '+'
  return ' '
}

const leftCellClass = (row) => {
  if (row.kind === 'pair' || row.kind === 'removed') return 'cell-del'
  return ''
}

const rightCellClass = (row) => {
  if (row.kind === 'pair' || row.kind === 'added') return 'cell-ins'
  return ''
}

// Unified view flattens pair rows into a removed line and an added line
const unifiedRows = computed(() => {
  const result = []
  for (const row of rows.value) {
    if (row.kind === 'pair') {
      result.push({
        kind: 'removed',
        leftNo: row.leftNo,
        rightNo: null,
        segs: row.leftSegs
      })
      result.push({
        kind: 'added',
        leftNo: null,
        rightNo: row.rightNo,
        segs: row.rightSegs
      })
    } else if (row.kind === 'skip') {
      result.push(row)
    } else {
      result.push({
        kind: row.kind,
        leftNo: row.leftNo,
        rightNo: row.rightNo,
        segs: [{ text: row.left, cls: 'seg-same' }]
      })
    }
  }
  return result
})

const unifiedMarker = (row) => {
  if (row.kind === 'removed') return '-'
  if (row.kind === 'added') return '+'
  return ' '
}

const unifiedCellClass = (row) => {
  if (row.kind === 'removed') return 'cell-del'
  if (row.kind === 'added') return 'cell-ins'
  return ''
}
</script>

<style scoped>
.diff-view {
  width: 100%;
}

.diff-view-header {
  display: flex;
  border: 1px solid var(--border-color);
  border-bottom: none;
  border-radius: 4px 4px 0 0;
  overflow: hidden;
}

.dv-col-title {
  flex: 1;
  padding: 4px 10px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
}

.dv-col-title:first-child {
  background: rgba(64, 158, 255, 0.15);
  color: #409eff;
  border-right: 1px solid var(--border-color);
}

.dv-col-title:last-child {
  background: rgba(103, 194, 58, 0.15);
  color: #67c23a;
}

.dv-body {
  border: 1px solid var(--border-color);
  border-radius: 0 0 4px 4px;
  overflow-x: auto;
  font-family: monospace;
  font-size: 12px;
  line-height: 1.6;
}

.dv-row {
  display: flex;
  min-height: 20px;
}

.dv-row-unified .dv-cell {
  width: 100%;
}

.dv-row-unified .dv-cell + .dv-cell {
  border-left: none;
}

.dv-cell {
  width: 50%;
  flex-shrink: 0;
  display: flex;
  align-items: stretch;
}

.dv-cell + .dv-cell {
  border-left: 1px solid var(--border-color);
}

.dv-lineno {
  width: 44px;
  flex-shrink: 0;
  text-align: right;
  padding-right: 8px;
  color: var(--text-secondary);
  user-select: none;
  background: rgba(128, 128, 128, 0.08);
  border-right: 1px solid var(--border-color);
}

.dv-marker {
  width: 18px;
  flex-shrink: 0;
  text-align: center;
  user-select: none;
  color: var(--text-secondary);
}

.dv-text {
  flex: 1;
  padding: 0 8px;
  white-space: pre-wrap;
  word-break: break-all;
}

.cell-del {
  background: rgba(248, 81, 73, 0.12);
}

.cell-del .dv-marker {
  color: #f56c6c;
}

.cell-ins {
  background: rgba(48, 176, 72, 0.12);
}

.cell-ins .dv-marker {
  color: #67c23a;
}

.seg-same {
  background: transparent;
}

.seg-del {
  background: rgba(248, 81, 73, 0.4);
}

.seg-ins {
  background: rgba(48, 176, 72, 0.4);
}

.dv-skip {
  width: 100%;
  padding: 2px 12px;
  text-align: center;
  color: var(--text-secondary);
  background: var(--bg-card-hover);
  font-size: 11px;
}

html.dark .cell-del {
  background: rgba(248, 81, 73, 0.18);
}

html.dark .cell-ins {
  background: rgba(48, 176, 72, 0.18);
}

html.dark .seg-del {
  background: rgba(248, 81, 73, 0.5);
}

html.dark .seg-ins {
  background: rgba(48, 176, 72, 0.5);
}
</style>
