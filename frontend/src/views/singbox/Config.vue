<template>
  <div class="singbox-config">
    <el-row :gutter="24">
      <el-col :span="24">
        <h2 class="page-title">Sing-Box 配置</h2>
      </el-col>
    </el-row>

    <el-row :gutter="24">
      <el-col :span="6">
        <el-card shadow="never" class="menu-card">
          <el-menu
            :default-active="activeMenu"
            router
            class="config-menu"
          >
            <el-menu-item index="/singbox/inbounds">
              <el-icon><Upload /></el-icon>
              <span>Inbound 配置</span>
            </el-menu-item>
            <el-menu-item index="/singbox/outbounds">
              <el-icon><Download /></el-icon>
              <span>Outbound 配置</span>
            </el-menu-item>
          </el-menu>
        </el-card>
      </el-col>

      <el-col :span="18">
        <router-view />
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Upload, Download } from '@element-plus/icons-vue'

const route = useRoute()
const activeMenu = computed(() => route.path)
</script>

<style scoped>
.singbox-config {
  max-width: 1200px;
  margin: 0 auto;
}

.page-title {
  margin: 0 0 24px 0;
  color: #303133;
  font-size: 24px;
  font-weight: 600;
}

.menu-card {
  min-height: 200px;
}

.config-menu {
  border-right: none;
}

.config-menu .el-menu-item {
  height: 48px;
  line-height: 48px;
}

.config-menu .el-menu-item.is-active {
  background: #ecf5ff;
  color: #409eff;
}
</style>
