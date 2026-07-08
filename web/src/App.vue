<script setup>
import { ref, reactive, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { message } from 'ant-design-vue'
import QRCode from 'qrcode'
import http from './api'

// 登录状态
const authenticated = ref(false)
const ready = ref(false)
const loginStatus = ref('')
const qrUrl = ref('')
const loginError = ref('')
const starting = ref(false)

// 两步验证
const twoFAVisible = ref(false)
const twoFAPassword = ref('')
const submitting2FA = ref(false)

// 下载
const url = ref('')
const submitting = ref(false)
const jobs = ref([])

// 下载预览弹窗
const downloadModalVisible = ref(false)
const downloadFilename = ref('')
const downloadError = ref('')
const previewing = ref(false)

// 多选与删除
const selectedIds = ref([])
const deleteModalVisible = ref(false)
const deleteModalIds = ref([])
const deleteModalFile = ref(false)
const deleting = ref(false)

// 单个任务操作 loading 状态，格式 "action:jobId"
const loadingAction = ref('')
function isLoading(action, id) {
  return loadingAction.value === action + ':' + id
}

// 轮询定时器
let loginTimer = null
let jobTimer = null
const qrCanvas = ref(null)

onMounted(() => {
  checkStatus()
})

onUnmounted(() => {
  stopLoginPoll()
  stopJobPoll()
})

watch(qrUrl, (val) => {
  if (val) renderQR(val)
})

watch(authenticated, (val) => {
  if (val) {
    stopLoginPoll()
    loadJobs()
    startJobPoll()
  } else {
    stopJobPoll()
  }
})

async function checkStatus() {
  try {
    const s = await http.get('/api/login/status')
    applyStatus(s)
  } catch (e) {
    loginError.value = e.message
  }
}

function applyStatus(s) {
  authenticated.value = s.authenticated
  ready.value = s.ready
  loginStatus.value = s.login_status
  qrUrl.value = s.qr_url || ''
  loginError.value = s.error || ''
  if (!authenticated.value && loginStatus.value === 'pending') {
    startLoginPoll()
  }
}

async function startLogin() {
  starting.value = true
  loginError.value = ''
  try {
    const s = await http.post('/api/login/qr/start')
    applyStatus(s)
  } catch (e) {
    loginError.value = e.message
  } finally {
    starting.value = false
  }
}

function startLoginPoll() {
  if (loginTimer) return
  loginTimer = setInterval(async () => {
    try {
      const s = await http.get('/api/login/status')
      applyStatus(s)
    } catch (e) {
      /* ignore */
    }
  }, 1500)
}

function stopLoginPoll() {
  if (loginTimer) clearInterval(loginTimer)
  loginTimer = null
}

async function submit2FA() {
  if (!twoFAPassword.value) return
  submitting2FA.value = true
  try {
    await http.post('/api/login/qr/2fa', { password: twoFAPassword.value })
    twoFAVisible.value = false
    twoFAPassword.value = ''
  } catch (e) {
    loginError.value = e.message
  } finally {
    submitting2FA.value = false
  }
}

function renderQR(val) {
  nextTick(() => {
    if (!qrCanvas.value) return
    QRCode.toCanvas(qrCanvas.value, val, { width: 220, margin: 1 }, (err) => {
      if (err) console.error(err)
    })
  })
}

async function createDownload() {
  if (!url.value) return
  previewing.value = true
  downloadError.value = ''
  try {
    const info = await http.post('/api/download/preview', { url: url.value })
    downloadFilename.value = info.filename || ''
    downloadModalVisible.value = true
  } catch (e) {
    message.error(e.message)
  } finally {
    previewing.value = false
  }
}

async function confirmDownload() {
  if (!downloadFilename.value.trim()) {
    downloadError.value = '文件名不能为空'
    return
  }
  submitting.value = true
  downloadError.value = ''
  try {
    await http.post('/api/download', { url: url.value, filename: downloadFilename.value.trim() })
    downloadModalVisible.value = false
    url.value = ''
    downloadFilename.value = ''
    await loadJobs()
    startJobPoll()
    message.success('已创建下载任务')
  } catch (e) {
    downloadError.value = e.message
  } finally {
    submitting.value = false
  }
}

async function loadJobs() {
  try {
    const r = await http.get('/api/jobs', { params: { page: 1, page_size: 50 } })
    jobs.value = r.list || []
  } catch (e) {
    /* ignore */
  }
}

function startJobPoll() {
  if (jobTimer) return
  jobTimer = setInterval(async () => {
    await loadJobs()
    const downloading = jobs.value.some((j) => j.status === 'downloading' || j.status === 'pending' || j.status === 'paused')
    if (!downloading) stopJobPoll()
  }, 2000)
}

function stopJobPoll() {
  if (jobTimer) clearInterval(jobTimer)
  jobTimer = null
}

function progressStatus(job) {
  if (job.status === 'failed') return 'exception'
  if (job.status === 'success') return 'success'
  if (job.status === 'paused') return 'normal'
  return 'active'
}

function formatSize(bytes) {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  while (bytes >= 1024 && i < units.length - 1) {
    bytes /= 1024
    i++
  }
  return bytes.toFixed(1) + ' ' + units[i]
}

async function pauseJob(id) {
  loadingAction.value = 'pause:' + id
  try {
    await http.post('/api/jobs/' + id + '/pause')
    await loadJobs()
    message.success('已暂停下载')
  } catch (e) {
    message.error(e.message)
  } finally {
    loadingAction.value = ''
  }
}

async function retryJob(id) {
  loadingAction.value = 'retry:' + id
  try {
    await http.post('/api/jobs/' + id + '/retry')
    await loadJobs()
    startJobPoll()
    message.success('已开始下载')
  } catch (e) {
    message.error(e.message)
  } finally {
    loadingAction.value = ''
  }
}

async function cancelJob(id) {
  loadingAction.value = 'cancel:' + id
  try {
    await http.post('/api/jobs/' + id + '/cancel')
    await loadJobs()
    message.success('已取消下载')
  } catch (e) {
    message.error(e.message)
  } finally {
    loadingAction.value = ''
  }
}

function confirmDelete(ids) {
  deleteModalIds.value = Array.isArray(ids) ? ids : [ids]
  deleteModalFile.value = false
  deleteModalVisible.value = true
}

async function doDelete() {
  deleting.value = true
  try {
    if (deleteModalIds.value.length === 1) {
      await http.delete('/api/jobs/' + deleteModalIds.value[0], { params: { delete_file: deleteModalFile.value } })
    } else {
      await http.delete('/api/jobs', { data: { ids: deleteModalIds.value, delete_file: deleteModalFile.value } })
    }
    deleteModalVisible.value = false
    selectedIds.value = selectedIds.value.filter((id) => !deleteModalIds.value.includes(id))
    await loadJobs()
    message.success('已删除任务')
  } catch (e) {
    message.error(e.message)
  } finally {
    deleting.value = false
  }
}

const allSelected = computed(() => jobs.value.length > 0 && selectedIds.value.length === jobs.value.length)

function toggleSelectAll(e) {
  selectedIds.value = e.target.checked ? jobs.value.map((j) => j.id) : []
}

function toggleSelect(id, checked) {
  if (checked) {
    if (!selectedIds.value.includes(id)) selectedIds.value.push(id)
  } else {
    selectedIds.value = selectedIds.value.filter((i) => i !== id)
  }
}

function viewFile(id) {
  window.open('/api/jobs/' + id + '/file', '_blank')
}

function formatSpeed(bytesPerSec) {
  if (!bytesPerSec) return '0 B/s'
  const units = ['B/s', 'KB/s', 'MB/s', 'GB/s']
  let i = 0
  let v = bytesPerSec
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024
    i++
  }
  return v.toFixed(1) + ' ' + units[i]
}

function formatEta(seconds) {
  if (seconds <= 0) return '--'
  if (seconds < 60) return seconds + ' 秒'
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  if (m < 60) return m + ' 分 ' + s + ' 秒'
  const h = Math.floor(m / 60)
  return h + ' 时 ' + (m % 60) + ' 分'
}
</script>

<template>
  <!-- 登录页 -->
  <div v-if="!authenticated" class="login-wrap">
    <a-card title="tdl-filegram 登录" style="width: 380px">
      <div class="qr-box">
        <canvas ref="qrCanvas"></canvas>
        <div class="footer-tip">使用 Telegram App 扫描二维码登录</div>
      </div>
      <a-alert
        v-if="!ready"
        message="Telegram 未就绪，请检查网络或在 config.yaml 配置代理（telegram.proxy）"
        type="warning"
        show-icon
        class="mt-16"
      />
      <a-alert v-if="loginError" :message="loginError" type="error" show-icon class="mt-16" />
      <a-alert
        v-if="loginStatus === 'success'"
        message="登录成功，正在加载..."
        type="success"
        show-icon
        class="mt-16"
      />
      <a-button
        v-if="loginStatus === 'need_2fa'"
        type="primary"
        block
        class="mt-16"
        @click="twoFAVisible = true"
      >
        输入两步验证密码
      </a-button>
      <a-button
        v-if="!qrUrl || loginStatus === 'error'"
        type="primary"
        block
        class="mt-16"
        :loading="starting"
        @click="startLogin"
      >
        {{ qrUrl ? '重新生成二维码' : '开始登录' }}
      </a-button>
    </a-card>
  </div>

  <!-- 主页 -->
  <div v-else class="page">
    <div class="header">
      <h1>tdl-filegram · Telegram 下载</h1>
      <a-tag color="green">已登录</a-tag>
    </div>

    <a-card title="新建下载" class="mt-16">
      <a-input-group compact>
        <a-input
          v-model:value="url"
          placeholder="粘贴 Telegram 消息链接，如 https://t.me/channel/123"
          style="width: calc(100% - 100px)"
          allow-clear
          @press-enter="createDownload"
        />
        <a-button type="primary" style="width: 100px" :loading="previewing" @click="createDownload">
          下载
        </a-button>
      </a-input-group>
    </a-card>

    <a-card title="下载任务" class="mt-16">
      <template #extra>
        <a-space v-if="jobs.length > 0">
          <a-checkbox :checked="allSelected" @change="toggleSelectAll">全选</a-checkbox>
          <a-button
            danger
            size="small"
            :disabled="selectedIds.length === 0"
            @click="confirmDelete(selectedIds)"
          >
            批量删除<template v-if="selectedIds.length > 0"> ({{ selectedIds.length }})</template>
          </a-button>
        </a-space>
      </template>
      <a-empty v-if="jobs.length === 0" description="暂无任务" />
      <a-list v-else :data-source="jobs" item-layout="horizontal">
        <template #renderItem="{ item }">
          <a-list-item>
            <a-list-item-meta>
              <template #title>
                <a-checkbox
                  :checked="selectedIds.includes(item.id)"
                  @change="(e) => toggleSelect(item.id, e.target.checked)"
                  style="margin-right: 8px"
                />
                <span>{{ item.file_name || '解析中...' }}</span>
                <a-tag v-if="item.status === 'cancelled'" color="default" style="margin-left: 8px">已取消</a-tag>
              </template>
              <template #description>
                <div style="font-size: 12px; color: #999">{{ item.url }}</div>
                <div style="margin-top: 6px">
                  <a-progress :percent="item.progress" :status="progressStatus(item)" size="small" />
                  <span style="font-size: 12px; color: #999; margin-left: 8px">
                    {{ formatSize(item.downloaded_bytes) }} / {{ formatSize(item.file_size) }}
                  </span>
                  <span
                    v-if="item.status === 'downloading' && item.speed > 0"
                    style="font-size: 12px; color: #999; margin-left: 8px"
                  >
                    {{ formatSpeed(item.speed) }}
                  </span>
                  <span
                    v-if="item.status === 'downloading' && item.eta_seconds > 0"
                    style="font-size: 12px; color: #999; margin-left: 8px"
                  >
                    剩余 {{ formatEta(item.eta_seconds) }}
                  </span>
                </div>
                <a-alert
                  v-if="item.error"
                  :message="item.error"
                  type="error"
                  show-icon
                  style="margin-top: 8px"
                />
              </template>
            </a-list-item-meta>
            <template #actions>
              <a-button
                v-if="item.status === 'downloading' || item.status === 'pending'"
                type="link"
                :loading="isLoading('pause', item.id)"
                @click="pauseJob(item.id)"
              >
                暂停下载
              </a-button>
              <a-button
                v-if="item.status === 'downloading' || item.status === 'pending'"
                type="link"
                danger
                :loading="isLoading('cancel', item.id)"
                @click="cancelJob(item.id)"
              >
                取消
              </a-button>
              <a-button
                v-if="item.status === 'paused'"
                type="link"
                :loading="isLoading('retry', item.id)"
                @click="retryJob(item.id)"
              >
                继续下载
              </a-button>
              <template v-if="item.status === 'success'">
                <a-button type="link" @click="viewFile(item.id)">查看文件</a-button>
                <a-button
                  type="link"
                  :loading="isLoading('retry', item.id)"
                  @click="retryJob(item.id)"
                >
                  重新下载
                </a-button>
              </template>
              <a-button
                v-if="item.status === 'failed' || item.status === 'cancelled'"
                type="link"
                :loading="isLoading('retry', item.id)"
                @click="retryJob(item.id)"
              >
                重试
              </a-button>
              <a-button type="link" danger @click="confirmDelete(item.id)">删除</a-button>
            </template>
          </a-list-item>
        </template>
      </a-list>
    </a-card>
  </div>

  <!-- 两步验证弹窗 -->
  <a-modal
    v-model:open="twoFAVisible"
    title="两步验证"
    :confirm-loading="submitting2FA"
    @ok="submit2FA"
  >
    <a-input-password v-model:value="twoFAPassword" placeholder="请输入两步验证密码" />
  </a-modal>

  <!-- 删除确认弹窗 -->
  <a-modal
    v-model:open="deleteModalVisible"
    title="删除任务"
    :confirm-loading="deleting"
    @ok="doDelete"
  >
    <p>确认删除 {{ deleteModalIds.length }} 个任务？</p>
    <a-checkbox v-model:checked="deleteModalFile">同时删除本地文件</a-checkbox>
  </a-modal>

  <!-- 下载预览弹窗 -->
  <a-modal
    v-model:open="downloadModalVisible"
    title="确认下载"
    :confirm-loading="submitting"
    @ok="confirmDownload"
  >
    <a-form layout="vertical">
      <a-form-item label="文件名">
        <a-input v-model:value="downloadFilename" placeholder="请输入文件名" allow-clear />
      </a-form-item>
      <a-alert
        v-if="downloadError"
        :message="downloadError"
        type="error"
        show-icon
      />
    </a-form>
  </a-modal>
</template>
