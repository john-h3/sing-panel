import axios from 'axios'

const api = axios.create({
  baseURL: '/api/singbox',
  timeout: 10000
})

export const singboxApi = {
  // Config
  getConfig() {
    return api.get('/')
  },
  exportConfig() {
    return api.get('/export')
  },

  // Inbounds
  getInbounds() {
    return api.get('/inbounds')
  },
  addInbound(data) {
    return api.post('/inbounds', data)
  },
  updateInbound(data) {
    return api.put('/inbounds', data)
  },
  deleteInbound(id) {
    return api.delete(`/inbounds/${id}`)
  },

  // Outbounds
  getOutbounds() {
    return api.get('/outbounds')
  },
  addOutbound(data) {
    return api.post('/outbounds', data)
  },
  updateOutbound(data) {
    return api.put('/outbounds', data)
  },
  deleteOutbound(id) {
    return api.delete(`/outbounds/${id}`)
  },

  // Types
  getInboundTypes() {
    return api.get('/types/inbound')
  },
  getOutboundTypes() {
    return api.get('/types/outbound')
  },

  // Import
  importLink(link) {
    return api.post('/import', { link })
  }
}

export default api
