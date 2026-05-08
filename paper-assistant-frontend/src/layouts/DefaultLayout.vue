<script setup lang="ts">
import { useUserStore } from '@/stores/user'
import { useRouter } from 'vue-router'
import { Document, Reading } from '@element-plus/icons-vue'

const userStore = useUserStore()
const router = useRouter()

const handleLogout = () => {
  userStore.logout()
  router.push('/login')
}
</script>

<template>
  <el-container class="h-screen bg-gray-50">
    <!-- Header -->
    <el-header class="bg-white border-b flex items-center justify-between px-6">
      <div class="flex items-center gap-2 cursor-pointer" @click="router.push('/')">
        <el-icon :size="24" color="#409EFC"><Reading /></el-icon>
        <span class="text-xl font-bold text-gray-800">论文阅读助手</span>
      </div>
      <div class="flex items-center gap-4">
        <span class="text-sm text-gray-600">欢迎, {{ userStore.userInfo?.username || '用户' }}</span>
        <el-button type="danger" link @click="handleLogout">退出登录</el-button>
      </div>
    </el-header>

    <el-container class="overflow-hidden">
      <!-- Sidebar -->
      <el-aside width="200px" class="bg-white border-r">
        <el-menu
          :default-active="$route.path"
          class="h-full border-none"
          router
        >
          <el-menu-item index="/papers">
            <el-icon><Document /></el-icon>
            <span>我的论文</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- Main Content -->
      <el-main class="p-6">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>