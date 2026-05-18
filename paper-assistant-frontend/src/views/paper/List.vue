<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { UploadFilled, Document, Reading } from '@element-plus/icons-vue'
import { getPaperList, uploadPaper } from '@/api/paper'

const router = useRouter()
const paperList = ref<any[]>([])
const loading = ref(false)

const dialogVisible = ref(false)
const uploading = ref(false)
const fileList = ref<any[]>([])

const fetchPapers = async () => {
  loading.value = true
  try {
    const res: any = await getPaperList()
    paperList.value = res.items || []
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const handleUploadChange = (_file: any, fileListConfig: any) => {
  fileList.value = fileListConfig.slice(-1)
}

const submitUpload = async () => {
  if (fileList.value.length === 0) {
    ElMessage.warning('请选择文件')
    return
  }
  const file = fileList.value[0].raw
  const formData = new FormData()
  formData.append('file', file)
  formData.append('title', file.name)

  uploading.value = true
  try {
    const res: any = await uploadPaper(formData)
    ElMessage.success('上传成功')
    dialogVisible.value = false
    fileList.value = []
    fetchPapers()
    
    // 如果需要自动跳转到阅读页
    if (res.paper?.id) {
      router.push(`/reader/${res.paper.id}`)
    }
  } catch (error) {
    console.error(error)
  } finally {
    uploading.value = false
  }
}

const goToReader = (id: number) => {
  router.push(`/reader/${id}`)
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleString()
}

onMounted(() => {
  fetchPapers()
})
</script>

<template>
  <div class="bg-white rounded-lg shadow p-6 min-h-full">
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-2xl font-bold text-gray-800">我的论文</h2>
      <el-button type="primary" :icon="UploadFilled" @click="dialogVisible = true">
        上传论文
      </el-button>
    </div>

    <el-table :data="paperList" v-loading="loading" style="width: 100%" border>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip>
        <template #default="{ row }">
          <div class="flex items-center gap-2">
            <el-icon color="#409EFC"><Document /></el-icon>
            <span class="font-medium text-gray-700">{{ row.title }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="parse_status" label="解析状态" width="120">
        <template #default="{ row }">
          <el-tag :type="row.parse_status === 'completed' ? 'success' : (row.parse_status === 'pending' ? 'info' : 'warning')">
            {{ row.parse_status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="上传时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link :icon="Reading" @click="goToReader(row.id)">
            阅读
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 上传弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      title="上传论文"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-upload
        class="upload-demo"
        drag
        action=""
        :auto-upload="false"
        :on-change="handleUploadChange"
        :file-list="fileList"
        accept="application/pdf"
      >
        <el-icon class="el-icon--upload"><upload-filled /></el-icon>
        <div class="el-upload__text">
          拖拽 PDF 文件到此处或 <em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip text-gray-500 mt-2">
            目前仅支持 PDF 格式文件
          </div>
        </template>
      </el-upload>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitUpload" :loading="uploading">
            确认上传
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>
