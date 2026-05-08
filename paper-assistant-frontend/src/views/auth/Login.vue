<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { login, getUserInfo } from '@/api/auth'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const loginForm = reactive({
  email: '',
  password: ''
})

const handleLogin = async () => {
  if (!loginForm.email || !loginForm.password) {
    ElMessage.warning('请输入邮箱和密码')
    return
  }
  
  loading.value = true
  try {
    const res: any = await login(loginForm)
    userStore.setToken(res.token)
    userStore.setUserInfo(res.user)
    ElMessage.success('登录成功')
    router.push('/')
  } catch (error) {
    // 错误在拦截器已处理
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-gray-100 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full bg-white rounded-lg shadow-md p-8">
      <div class="text-center mb-8">
        <h2 class="text-3xl font-extrabold text-gray-900">论文阅读助手</h2>
        <p class="mt-2 text-sm text-gray-600">登录您的账号继续</p>
      </div>

      <el-form :model="loginForm" @keyup.enter="handleLogin" size="large">
        <el-form-item>
          <el-input
            v-model="loginForm.email"
            placeholder="邮箱"
            :prefix-icon="User"
          />
        </el-form-item>
        
        <el-form-item>
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="密码"
            :prefix-icon="Lock"
            show-password
          />
        </el-form-item>

        <el-button
          type="primary"
          class="w-full"
          :loading="loading"
          @click="handleLogin"
        >
          登 录
        </el-button>
      </el-form>

      <div class="mt-6 text-center text-sm">
        <span class="text-gray-600">没有账号？</span>
        <router-link to="/register" class="text-blue-600 hover:text-blue-500 font-medium">
          立即注册
        </router-link>
      </div>
    </div>
  </div>
</template>