// Global Clash API WebSocket streams - persists across route changes
import { ref } from 'vue'

const trafficWs = ref(null)
const connectionsWs = ref(null)

const currentSpeed = ref({ up: 0, down: 0 })
const cumulativeTraffic = ref({ up: 0, down: 0 })
const speedHistory = ref([])
const totalHistory = ref([])
const connected = ref({ traffic: false, connections: false })
const speedInterval = ref(3)
const totalInterval = ref(3)

const MAX_HISTORY = 300

let trafficReconnectTimer = null
let connectionsReconnectTimer = null

const startTrafficStream = () => {
  if (trafficWs.value) trafficWs.value.close()

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/clash_api/traffic?interval=${speedInterval.value}`

  const ws = new WebSocket(wsUrl)
  trafficWs.value = ws

  ws.onopen = () => { connected.value.traffic = true }

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      const timeStr = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })

      currentSpeed.value = { up: data.up || 0, down: data.down || 0 }

      speedHistory.value.push({
        time: timeStr,
        up: data.up || 0,
        down: data.down || 0,
        totalUp: 0,
        totalDown: 0
      })
      if (speedHistory.value.length > MAX_HISTORY) speedHistory.value.shift()
    } catch (err) {}
  }

  ws.onerror = () => {}
  ws.onclose = () => {
    trafficWs.value = null
    connected.value.traffic = false
    trafficReconnectTimer = setTimeout(startTrafficStream, 3000)
  }
}

const startConnectionsStream = () => {
  if (connectionsWs.value) connectionsWs.value.close()

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/clash_api/connections?interval=${totalInterval.value}`

  const ws = new WebSocket(wsUrl)
  connectionsWs.value = ws

  ws.onopen = () => { connected.value.connections = true }

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      const timeStr = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })

      cumulativeTraffic.value = { up: data.uploadTotal || 0, down: data.downloadTotal || 0 }

      totalHistory.value.push({
        time: timeStr,
        up: 0,
        down: 0,
        totalUp: data.uploadTotal || 0,
        totalDown: data.downloadTotal || 0
      })
      if (totalHistory.value.length > MAX_HISTORY) totalHistory.value.shift()
    } catch (err) {}
  }

  ws.onerror = () => {}
  ws.onclose = () => {
    connectionsWs.value = null
    connected.value.connections = false
    connectionsReconnectTimer = setTimeout(startConnectionsStream, 3000)
  }
}

const setSpeedInterval = (seconds) => {
  speedInterval.value = seconds
  speedHistory.value = []
  startTrafficStream()
}

const setTotalInterval = (seconds) => {
  totalInterval.value = seconds
  totalHistory.value = []
  startConnectionsStream()
}

const stopStreams = () => {
  if (trafficReconnectTimer) { clearTimeout(trafficReconnectTimer); trafficReconnectTimer = null }
  if (connectionsReconnectTimer) { clearTimeout(connectionsReconnectTimer); connectionsReconnectTimer = null }
  if (trafficWs.value) { trafficWs.value.close(); trafficWs.value = null }
  if (connectionsWs.value) { connectionsWs.value.close(); connectionsWs.value = null }
  connected.value = { traffic: false, connections: false }
}

export function useClashStream() {
  return {
    currentSpeed,
    cumulativeTraffic,
    speedHistory,
    totalHistory,
    connected,
    speedInterval,
    totalInterval,
    startTrafficStream,
    startConnectionsStream,
    stopStreams,
    setSpeedInterval,
    setTotalInterval
  }
}
