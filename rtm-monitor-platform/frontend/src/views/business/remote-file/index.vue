<template>
  <div class="app-container">
    <div class="file-header">
      <el-page-header @back="goBack" :content="`文件管理 - ${hostIp}`" />
      <div class="header-actions">
        <el-button type="primary" size="small" icon="Upload" @click="handleUploadClick">上传</el-button>
        <el-button type="success" size="small" icon="FolderAdd" @click="handleMkdir">新建文件夹</el-button>
        <el-button size="small" icon="Refresh" @click="loadFileList">刷新</el-button>
      </div>
    </div>

    <!-- 面包屑导航 -->
    <div class="breadcrumb-bar">
      <el-breadcrumb separator="/">
        <el-breadcrumb-item>
          <el-link @click="navigateTo('/')" :underline="false">根目录</el-link>
        </el-breadcrumb-item>
        <el-breadcrumb-item v-for="(part, index) in breadcrumbParts" :key="index">
          <el-link @click="navigateToBreadcrumb(index)" :underline="false">{{ part }}</el-link>
        </el-breadcrumb-item>
      </el-breadcrumb>
    </div>

    <!-- 文件列表 -->
    <div class="file-table-wrapper">
      <el-table
        v-loading="loading"
        :data="fileList"
        style="width: 100%; height: 100%; padding-top: 15px;"
        @row-dblclick="handleRowDblclick"
        @row-contextmenu="handleRowContextmenu"
      >
      <el-table-column prop="name" label="名称" min-width="200">
        <template #default="{ row }">
          <div class="file-name-cell" @click="handleFileClick(row)">
            <el-icon v-if="row.isDir" class="file-icon" :size="20"><Folder /></el-icon>
            <el-icon v-else class="file-icon" :size="20"><Document /></el-icon>
            <span class="file-name">{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="size" label="大小" width="120">
        <template #default="{ row }">
          <span v-if="!row.isDir">{{ formatFileSize(row.size) }}</span>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column prop="mode" label="权限" width="100" />
      <el-table-column prop="modTime" label="修改时间" width="180">
        <template #default="{ row }">
          {{ formatTime(row.modTime) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button v-if="!row.isDir" type="primary" size="small" text icon="View" @click="handlePreview(row)">预览</el-button>
          <el-button v-if="!row.isDir" type="primary" size="small" text icon="Download" @click="handleDownload(row)">下载</el-button>
        </template>
      </el-table-column>
    </el-table>
    </div>

    <!-- 文件预览/编辑对话框 -->
    <el-dialog v-model="previewVisible" :title="previewTitle" width="70%" destroy-on-close>
      <el-input
        v-model="fileContent"
        type="textarea"
        :rows="20"
        placeholder="文件内容"
        v-if="isTextFile"
      />
      <div v-else class="binary-file">二进制文件，无法预览</div>
      <template #footer>
        <el-button @click="previewVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleSaveFile" v-if="isTextFile">保存</el-button>
      </template>
    </el-dialog>

    <!-- 上传对话框 -->
    <el-dialog v-model="uploadVisible" title="上传文件" width="400px">
      <el-upload
        ref="uploadRef"
        action="#"
        :auto-upload="false"
        :on-change="handleUploadChange"
        :limit="1"
      >
        <el-button type="primary">选择文件</el-button>
      </el-upload>
      <template #footer>
        <el-button @click="uploadVisible = false">取消</el-button>
        <el-button type="primary" @click="handleUploadConfirm">上传</el-button>
      </template>
    </el-dialog>

    <!-- 新建文件夹对话框 -->
    <el-dialog v-model="mkdirVisible" title="新建文件夹" width="400px">
      <el-input v-model="newFolderName" placeholder="请输入文件夹名称" />
      <template #footer>
        <el-button @click="mkdirVisible = false">取消</el-button>
        <el-button type="primary" @click="handleMkdirConfirm">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup name="RemoteFile">
import { ref, computed, onMounted, onBeforeUnmount, onActivated, onDeactivated, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { listFiles, readFile, writeFile, uploadFile, downloadFile, mkdir as createDir } from '@/api/remote/file'

const route = useRoute()
const router = useRouter()

const hostIp = ref(route.query.ip || '')
const currentPath = ref('/')
const fileList = ref([])
const loading = ref(false)

// 预览相关
const previewVisible = ref(false)
const previewTitle = ref('')
const previewPath = ref('')
const fileContent = ref('')
const isTextFile = ref(true)
const uploadFileData = ref(null)

// 上传相关
const uploadVisible = ref(false)
const uploadRef = ref(null)

// 新建文件夹
const mkdirVisible = ref(false)
const newFolderName = ref('')

// 请求取消控制器（切换主机时取消未完成的请求）
let abortController = null
let loadDebounceTimer = null

/** 取消未完成的请求 */
function cancelPendingRequests() {
  if (abortController) {
    abortController.abort()
    abortController = null
  }
  if (loadDebounceTimer) {
    clearTimeout(loadDebounceTimer)
    loadDebounceTimer = null
  }
}

/** 面包屑路径 */
const breadcrumbParts = computed(() => {
  if (currentPath.value === '/') return []
  const parts = currentPath.value.split('/').filter(p => p)
  return parts
})

/** 返回 */
function goBack() {
  router.push('/business/remote-ops')
}

/** 初始化/切换主机
 * @param {boolean} force - 是否强制刷新（首次加载或手动刷新时使用）
 */
function setupHost(force = false) {
  const newIp = route.query.ip || ''
  if (newIp && (force || newIp !== hostIp.value)) {
    cancelPendingRequests()  // 取消之前未完成的请求
    hostIp.value = newIp
    currentPath.value = '/'
    fileList.value = []
    loadFileList()
  }
}

/** 加载文件列表（200ms防抖，防止快速切换时重复请求） */
function loadFileList() {
  if (!hostIp.value || loading.value) return
  cancelPendingRequests()  // 取消之前未完成的请求和延迟
  loadDebounceTimer = setTimeout(() => {
    loadDebounceTimer = null
    if (!hostIp.value) return
    abortController = new AbortController()
    loading.value = true
    listFiles({ hostIp: hostIp.value, path: currentPath.value }, { signal: abortController.signal })
      .then(res => {
        const data = res.data || res
        fileList.value = data.files || []
      })
      .catch(err => {
        // 请求被取消或组件不活跃时不显示错误
        if (err.code === 'ERR_CANCELED' || err.message === 'canceled') {
          return
        }
        ElMessage.error(err.message || '加载文件列表失败')
      })
      .finally(() => {
        loading.value = false
        abortController = null
      })
  }, 200)
}

/** 点击文件/目录 */
function handleFileClick(row) {
  if (row.isDir) {
    navigateTo(row.path)
  }
}

/** 双击行 */
function handleRowDblclick(row) {
  if (row.isDir) {
    navigateTo(row.path)
  } else {
    handlePreview(row)
  }
}

/** 右键菜单 */
function handleRowContextmenu(row, column, event) {
  event.preventDefault()
}

/** 导航到路径 */
function navigateTo(path) {
  currentPath.value = path
  loadFileList()
}

/** 面包屑导航 */
function navigateToBreadcrumb(index) {
  const parts = currentPath.value.split('/').filter(p => p)
  const newPath = '/' + parts.slice(0, index + 1).join('/')
  navigateTo(newPath)
}

/** 预览文件 */
function handlePreview(row) {
  previewPath.value = row.path
  previewTitle.value = row.name
  loading.value = true
  readFile({ hostIp: hostIp.value, path: row.path })
    .then(res => {
      const data = res.data || res
      isTextFile.value = isTextFileType(row.name)
      if (isTextFile.value && data.content) {
        try {
          fileContent.value = atob(data.content)
        } catch (e) {
          fileContent.value = '[无法解码的文件内容]'
        }
      } else {
        fileContent.value = ''
      }
      previewVisible.value = true
    })
    .catch(err => {
      ElMessage.error(err.message || '读取文件失败')
    })
    .finally(() => {
      loading.value = false
    })
}

/** 判断是否为文本文件 */
function isTextFileType(name) {
  const textExts = ['.txt', '.log', '.md', '.json', '.xml', '.yaml', '.yml', '.ini', '.conf', '.sh', '.bat', '.cmd', '.ps1', '.py', '.js', '.ts', '.html', '.css', '.go', '.java', '.c', '.cpp', '.h', '.vue', '.sql', '.properties', '.env', '.gitignore']
  const ext = name.substring(name.lastIndexOf('.')).toLowerCase()
  return textExts.includes(ext)
}

/** 保存文件 */
function handleSaveFile() {
  const content = btoa(unescape(encodeURIComponent(fileContent.value)))
  writeFile({
    hostIp: hostIp.value,
    path: previewPath.value,
    content: content
  })
    .then(() => {
      ElMessage.success('保存成功')
      previewVisible.value = false
    })
    .catch(err => {
      ElMessage.error(err.message || '保存失败')
    })
}

/** 下载文件 */
function handleDownload(row) {
  loading.value = true
  downloadFile({ hostIp: hostIp.value, path: row.path })
    .then(res => {
      const data = res.data || res
      if (data.content) {
        const blob = base64ToBlob(data.content)
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = row.name
        a.click()
        URL.revokeObjectURL(url)
        ElMessage.success('下载成功')
      }
    })
    .catch(err => {
      ElMessage.error(err.message || '下载失败')
    })
    .finally(() => {
      loading.value = false
    })
}

/** base64转Blob */
function base64ToBlob(base64) {
  const byteCharacters = atob(base64)
  const byteNumbers = new Array(byteCharacters.length)
  for (let i = 0; i < byteCharacters.length; i++) {
    byteNumbers[i] = byteCharacters.charCodeAt(i)
  }
  const byteArray = new Uint8Array(byteNumbers)
  return new Blob([byteArray])
}

/** 上传按钮点击 */
function handleUploadClick() {
  uploadVisible.value = true
  uploadFileData.value = null
}

/** 文件选择变化 */
function handleUploadChange(file) {
  uploadFileData.value = file
}

/** 上传确认 */
function handleUploadConfirm() {
  if (!uploadFileData.value) {
    ElMessage.warning('请选择文件')
    return
  }
  const reader = new FileReader()
  reader.onload = (e) => {
    const base64 = btoa(
      new Uint8Array(e.target.result)
        .reduce((data, byte) => data + String.fromCharCode(byte), '')
    )
    uploadFile({
      hostIp: hostIp.value,
      path: currentPath.value,
      fileName: uploadFileData.value.name,
      content: base64
    })
      .then(() => {
        ElMessage.success('上传成功')
        uploadVisible.value = false
        loadFileList()
      })
      .catch(err => {
        ElMessage.error(err.message || '上传失败')
      })
  }
  reader.readAsArrayBuffer(uploadFileData.value.raw)
}

/** 新建文件夹 */
function handleMkdir() {
  mkdirVisible.value = true
  newFolderName.value = ''
}

/** 新建文件夹确认 */
function handleMkdirConfirm() {
  if (!newFolderName.value.trim()) {
    ElMessage.warning('请输入文件夹名称')
    return
  }
  const path = currentPath.value === '/' ? '/' + newFolderName.value : currentPath.value + '/' + newFolderName.value
  createDir({ hostIp: hostIp.value, path: path })
    .then(() => {
      ElMessage.success('创建成功')
      mkdirVisible.value = false
      loadFileList()
    })
    .catch(err => {
      ElMessage.error(err.message || '创建失败')
    })
}

/** 格式化文件大小 */
function formatFileSize(size) {
  if (size === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const k = 1024
  const i = Math.floor(Math.log(size) / Math.log(k))
  return parseFloat((size / Math.pow(k, i)).toFixed(2)) + ' ' + units[i]
}

/** 格式化时间 */
function formatTime(timestamp) {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('zh-CN')
}

// 监听路由 query 变化（同一组件实例内导航）
watch(() => route.query.ip, (newIp) => {
  if (newIp && newIp !== hostIp.value) {
    setupHost()
  }
})

onMounted(() => {
  // 首次挂载强制加载，避免 hostIp 初始值与 route.query.ip 相同导致条件不满足
  setupHost(true)
})

// keep-alive 缓存后重新激活
onActivated(() => {
  // IP 变化时才重新加载，避免不必要的重复请求
  setupHost()
})

// keep-alive 缓存时取消未完成请求
onDeactivated(() => {
  cancelPendingRequests()
})

// 组件卸载前清理
onBeforeUnmount(() => {
  cancelPendingRequests()
  loading.value = false
})
</script>

<style scoped lang="scss">
.app-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 84px);
  padding: 24px;
}

.file-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  padding: 16px 24px;
  border-radius: 12px;
  margin-bottom: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 1px 2px rgba(0, 0, 0, 0.03);
  border: 1px solid #E2E8F0;
}

.header-actions {
  display: flex;
  gap: 10px;

  .el-button {
    border-radius: 8px;
    font-weight: 600;
  }
}

.breadcrumb-bar {
  background: #fff;
  padding: 14px 24px;
  border-radius: 12px;
  margin-bottom: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 1px 2px rgba(0, 0, 0, 0.03);
  border: 1px solid #E2E8F0;
}

/* 文件列表容器：占满剩余高度，内部滚动 */
.file-table-wrapper {
  flex: 1;
  overflow: hidden;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 1px 2px rgba(0, 0, 0, 0.03);
  border: 1px solid #E2E8F0;
  padding: 0 16px 16px 16px;
}

.file-name-cell {
  display: flex;
  align-items: center;
  cursor: pointer;

  .file-icon {
    margin-right: 10px;
    color: #2563EB;
  }

  .file-name {
    color: #334155;
    font-weight: 500;

    &:hover {
      color: #2563EB;
    }
  }
}

.binary-file {
  text-align: center;
  padding: 40px;
  color: #94A3B8;
}
</style>
