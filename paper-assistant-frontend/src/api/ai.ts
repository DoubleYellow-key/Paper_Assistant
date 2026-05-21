import request from '@/utils/request'

export function askQuestion(paperId: string | number, query: string) {
  return request({
    url: `/papers/${paperId}/qa`,
    method: 'post',
    data: { query }
  })
}

export function getSummary(paperId: string | number, query: string = '请总结这篇论文的核心内容') {
  return request({
    url: `/papers/${paperId}/summary`,
    method: 'post',
    data: { query }
  })
}

export function explainTerm(paperId: string | number, query: string) {
  return request({
    url: `/papers/${paperId}/term-explain`,
    method: 'post',
    data: { query }
  })
}

export function translatePaper(
  paperId: string | number,
  targetLanguage: string = 'zh-CN',
  forceRegenerate: boolean = false
) {
  return request({
    url: `/papers/${paperId}/translate`,
    method: 'post',
    data: {
      target_language: targetLanguage,
      force_regenerate: forceRegenerate
    }
  })
}

export function getLatestTranslation(paperId: string | number, targetLanguage: string = 'zh-CN') {
  return request({
    url: `/papers/${paperId}/translations/latest`,
    method: 'get',
    params: {
      target_language: targetLanguage
    }
  })
}
