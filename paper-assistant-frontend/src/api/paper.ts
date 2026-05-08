import request from '@/utils/request'

export function getPaperList() {
  return request({
    url: '/papers',
    method: 'get'
  })
}

export function getPaperDetail(id: string | number) {
  return request({
    url: `/papers/${id}`,
    method: 'get'
  })
}

export function uploadPaper(data: FormData) {
  return request({
    url: '/papers/upload',
    method: 'post',
    data,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export function getParseJobLatest(id: string | number) {
  return request({
    url: `/papers/${id}/parse-jobs/latest`,
    method: 'get'
  })
}