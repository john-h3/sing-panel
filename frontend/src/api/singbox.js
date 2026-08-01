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

  // Rulesets
  getRulesets() {
    return api.get('/rulesets')
  },
  addRuleset(data) {
    return api.post('/rulesets', data)
  },
  addRulesets(data) {
    return api.post('/rulesets/batch', data)
  },
  updateRuleset(data) {
    return api.put('/rulesets', data)
  },
  deleteRuleset(id) {
    return api.delete(`/rulesets/${id}`)
  },
  deleteRulesets(ids) {
    return api.post('/rulesets/delete', { ids })
  },

  // Route Rules
  getRouteRules() {
    return api.get('/route-rules')
  },
  addRouteRule(data) {
    return api.post('/route-rules', data)
  },
  updateRouteRule(data) {
    return api.put('/route-rules', data)
  },
  deleteRouteRule(id) {
    return api.delete(`/route-rules/${id}`)
  },
  reorderRouteRules(ids) {
    return api.post('/route-rules/reorder', { ids })
  },

  // Route Config
  getRouteConfig() {
    return api.get('/route-config')
  },
  updateRouteConfig(data) {
    return api.put('/route-config', data)
  },

  // DNS
  getDNS() {
    return api.get('/dns')
  },
  updateDNS(data) {
    return api.put('/dns', data)
  },

  // Services
  getServices() {
    return api.get('/services')
  },
  addService(data) {
    return api.post('/services', data)
  },
  updateService(data) {
    return api.put('/services', data)
  },
  deleteService(id) {
    return api.delete(`/services/${id}`)
  },

  // HTTP Clients
  getHTTPClients() {
    return api.get('/http-clients')
  },
  addHTTPClient(data) {
    return api.post('/http-clients', data)
  },
  updateHTTPClient(data) {
    return api.put('/http-clients', data)
  },
  deleteHTTPClient(id) {
    return api.delete(`/http-clients/${id}`)
  },

  // Experimental
  getExperimental() {
    return api.get('/experimental')
  },
  updateExperimental(data) {
    return api.put('/experimental', data)
  },

  // Geo
  fetchGeoTree() {
    return api.get('/geo-tree')
  },
  fetchCommonRulesetTree() {
    return api.get('/common-ruleset-tree')
  },

  // Types
  getInboundTypes() {
    return api.get('/types/inbound')
  },
  getOutboundTypes() {
    return api.get('/types/outbound')
  },
  getNetworkInterfaces() {
    return api.get('/network-interfaces')
  },

  // Import
  importLink(link) {
    return api.post('/import', { link })
  }
}

export default api
