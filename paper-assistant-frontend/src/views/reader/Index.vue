<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Position, ChatDotRound, Document, Refresh } from '@element-plus/icons-vue'
import { getPaperDetail, getParseJobLatest } from '@/api/paper'
import { askQuestion, getSummary } from '@/api/ai'

const route = useRoute()
const paperId = route.params.id as string

const paperInfo = ref<any>(null)
const loading = ref(true)
const parseStatus = ref('pending')
const parseProgress = ref(0)
const pollTimer = ref<any>(null)

// QA State
const chatList = ref<{role: 'user' | 'assistant', content: string}[]>([])
const inputQuery = ref('')
const asking = ref(false)

const fetchPaperInfo = async () => {
  try {
    const res: any = await getPaperDetail(paperId)
    paperInfo.value = res.paper
    parseStatus.value = res.paper.parse_status
    if (parseStatus.value === 'pending' || parseStatus.value === 'queued') {
      startPolling()
    } else if (parseStatus.value === 'completed') {
      if (chatList.value.length === 0) {
        generateSummary()
      }
    }
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const startPolling = () => {
  if (pollTimer.value) return
  pollTimer.value = setInterval(async () => {
    try {
      const res: any = await getParseJobLatest(paperId)
      if (res.parse_job) {
        parseProgress.value = res.parse_job.progress
        if (res.parse_job.status === 'completed') {
          parseStatus.value = 'completed'
          stopPolling()
          ElMessage.success('论文解析完成')
          generateSummary()
        } else if (res.parse_job.status === 'failed') {
          parseStatus.value = 'failed'
          stopPolling()
          ElMessage.error('论文解析失败')
        }
      }
    } catch (error) {
      stopPolling()
    }
  }, 3000)
}

const stopPolling = () => {
  if (pollTimer.value) {
    clearInterval(pollTimer.value)
    pollTimer.value = null
  }
}

const generateSummary = async () => {
  chatList.value.push({ role: 'assistant', content: '正在生成论文总结...' })
  try {
    const res: any = await getSummary(paperId)
    chatList.value.pop()
    chatList.value.push({ role: 'assistant', content: res.answer })
  } catch (error) {
    chatList.value.pop()
    chatList.value.push({ role: 'assistant', content: '总结生成失败，请重试。' })
  }
}

const handleAsk = async () => {
  if (!inputQuery.value.trim()) return
  const query = inputQuery.value
  inputQuery.value = ''
  chatList.value.push({ role: 'user', content: query })
  chatList.value.push({ role: 'assistant', content: '思考中...' })
  asking.value = true
  try {
    const res: any = await askQuestion(paperId, query)
    chatList.value.pop()
    chatList.value.push({ role: 'assistant', content: res.answer })
  } catch (error) {
    chatList.value.pop()
    chatList.value.push({ role: 'assistant', content: '回答生成失败，请检查模型服务是否正常。' })
  } finally {
    asking.value = false
  }
}

onMounted(() => {
  fetchPaperInfo()
})

onUnmounted(() => {
  stopPolling()
})
</script>

<template>
  <div class="h-full flex gap-4 bg-gray-50" v-loading="loading">
    <!-- Left: PDF Reader -->
    <div class="flex-1 bg-white rounded-lg shadow flex flex-col overflow-hidden">
      <div class="p-4 border-b flex items-center justify-between bg-gray-50">
        <div class="flex items-center gap-2 font-medium text-gray-800">
          <el-icon><Document /></el-icon>
          <span>{{ paperInfo?.title || '加载中...' }}</span>
        </div>
        <div>
          <el-tag v-if="parseStatus === 'completed'" type="success">解析完成</el-tag>
          <el-tag v-else-if="parseStatus === 'failed'" type="danger">解析失败</el-tag>
          <div v-else class="flex items-center gap-2">
            <el-tag type="warning">正在解析</el-tag>
            <el-progress :percentage="parseProgress" :show-text="false" style="width: 100px" />
          </div>
        </div>
      </div>
      <div class="flex-1 p-4 relative overflow-hidden bg-gray-100 flex items-center justify-center">
        <!-- 此处为PDF渲染占位，可用 iframe 或 pdf.js -->
        <iframe v-if="paperInfo?.file_path" :src="'/api/v1' + paperInfo.file_path" class="w-full h-full border-none"></iframe>
        <div v-else class="text-gray-400 flex flex-col items-center">
          <el-icon :size="48" class="mb-2"><Document /></el-icon>
          <p>暂无 PDF 预览</p>
        </div>
      </div>
    </div>

    <!-- Right: AI Assistant -->
    <div class="w-96 bg-white rounded-lg shadow flex flex-col overflow-hidden">
      <div class="p-4 border-b bg-gray-50 flex items-center gap-2">
        <el-icon color="#409EFC"><ChatDotRound /></el-icon>
        <span class="font-medium text-gray-800">智能助手</span>
      </div>
      
      <div class="flex-1 overflow-y-auto p-4 space-y-4 bg-gray-50">
        <div v-for="(msg, idx) in chatList" :key="idx" class="flex flex-col" :class="msg.role === 'user' ? 'items-end' : 'items-start'">
          <div 
            class="max-w-[85%] rounded-lg p-3 text-sm shadow-sm"
            :class="msg.role === 'user' ? 'bg-blue-500 text-white' : 'bg-white text-gray-800 border border-gray-100'"
          >
            <div class="whitespace-pre-wrap leading-relaxed">{{ msg.content }}</div>
          </div>
        </div>
        <div v-if="chatList.length === 0" class="text-center text-gray-400 mt-10 text-sm">
          解析完成后，助手将自动为您总结论文。
        </div>
      </div>

      <div class="p-4 border-t bg-white">
        <el-input
          v-model="inputQuery"
          type="textarea"
          :rows="3"
          placeholder="向助手提问关于这篇论文的内容... (按 Enter 发送)"
          resize="none"
          @keydown.enter.prevent="handleAsk"
          :disabled="parseStatus !== 'completed' || asking"
        />
        <div class="mt-3 flex justify-end">
          <el-button type="primary" :loading="asking" :disabled="parseStatus !== 'completed' || !inputQuery.trim()" @click="handleAsk">
            发送
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 滚动条美化 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
::-webkit-scrollbar-thumb {
  background: #dcdfe6;
  border-radius: 3px;
}
::-webkit-scrollbar-track {
  background: transparent;
}
</style>