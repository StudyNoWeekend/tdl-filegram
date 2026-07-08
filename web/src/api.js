import axios from 'axios'

const http = axios.create({
  baseURL: '/',
  timeout: 30000,
})

// 统一解析后端响应：{ data, code, msg, trace_id }
http.interceptors.response.use(
  (resp) => {
    const body = resp.data
    if (body.code !== 0) {
      return Promise.reject(new Error(body.msg || '请求失败'))
    }
    return body.data
  },
  (err) => {
    // 后端返回非 2xx 状态码时，axios 走错误拦截器，需从 response.data.msg 提取后端错误信息
    if (err.response && err.response.data && err.response.data.msg) {
      return Promise.reject(new Error(err.response.data.msg))
    }
    return Promise.reject(new Error(err.message || '网络错误'))
  },
)

export default http
