/** 收益明细 Tab：与 backendAdapter reward_list reqType 对齐 */
export const REWARD_TABS = [
  ['1', '认购'],
  ['2', '静态收益'],
  ['3', '直推收益'],
  ['4', '直推加速收益'],
  ['5', '团队收益'],
  ['6', '平级收益'],
  ['7', '全网收益'],
] as const

export function rewardMenuType(lang: (k: string) => string) {
  return REWARD_TABS.map(([id, label]) => [id, lang(label)] as [string, string])
}
