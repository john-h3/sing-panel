<template>
  <div class="traffic-chart">
    <canvas ref="chartCanvas"></canvas>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { Chart, registerables } from 'chart.js'

Chart.register(...registerables)

const props = defineProps({
  data: {
    type: Array,
    default: () => []
  },
  mode: {
    type: String,
    default: 'speed' // 'speed' or 'total'
  }
})

const chartCanvas = ref(null)
let chart = null

const formatBytes = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytes
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }
  return `${size.toFixed(1)} ${units[unitIndex]}`
}

const getLabel = () => {
  return props.mode === 'speed' ? '实时速率' : '累计流量'
}

const createChart = () => {
  if (!chartCanvas.value) return

  const ctx = chartCanvas.value.getContext('2d')
  chart = new Chart(ctx, {
    type: 'line',
    data: {
      labels: props.data.map(d => d.time),
      datasets: [
        {
          label: '上行',
          data: props.data.map(d => props.mode === 'speed' ? d.up : d.totalUp),
          borderColor: '#409eff',
          backgroundColor: 'rgba(64, 158, 255, 0.1)',
          fill: true,
          tension: 0.3,
          pointRadius: 0
        },
        {
          label: '下行',
          data: props.data.map(d => props.mode === 'speed' ? d.down : d.totalDown),
          borderColor: '#67c23a',
          backgroundColor: 'rgba(103, 194, 58, 0.1)',
          fill: true,
          tension: 0.3,
          pointRadius: 0
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: {
        duration: 300
      },
      interaction: {
        intersect: false,
        mode: 'index'
      },
      plugins: {
        legend: {
          position: 'top',
          labels: {
            usePointStyle: true,
            padding: 16
          }
        },
        tooltip: {
          callbacks: {
            label: (ctx) => {
              return `${ctx.dataset.label}: ${formatBytes(ctx.raw)}`
            }
          }
        }
      },
      scales: {
        x: {
          display: true,
          grid: {
            display: false
          },
          ticks: {
            maxTicksLimit: 10,
            font: { size: 10 }
          }
        },
        y: {
          display: true,
          beginAtZero: true,
          grid: {
            color: 'rgba(0,0,0,0.05)'
          },
          ticks: {
            callback: (value) => formatBytes(value),
            font: { size: 10 }
          }
        }
      }
    }
  })
}

watch(() => props.data, (newData) => {
  if (chart) {
    chart.data.labels = newData.map(d => d.time)
    chart.data.datasets[0].data = newData.map(d => props.mode === 'speed' ? d.up : d.totalUp)
    chart.data.datasets[1].data = newData.map(d => props.mode === 'speed' ? d.down : d.totalDown)
    chart.update('none')
  }
}, { deep: true })

onMounted(() => {
  createChart()
})

onUnmounted(() => {
  if (chart) {
    chart.destroy()
  }
})
</script>

<style scoped>
.traffic-chart {
  width: 100%;
  height: 280px;
}
</style>
