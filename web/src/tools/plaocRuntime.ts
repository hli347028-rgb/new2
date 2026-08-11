const PLAOC_EXTERNAL_URL = 'X-Plaoc-External-Url'

/** 只有 PLaOC/dweb 宿主会注入外部通信地址。 */
export function isPlaocRuntime(): boolean {
  const queryValue = new URLSearchParams(window.location.search).get(PLAOC_EXTERNAL_URL)
  const cachedValue = localStorage.getItem(`url:${PLAOC_EXTERNAL_URL}`)
  return Boolean(queryValue || cachedValue)
}

export async function restartCurrentApp(): Promise<void> {
  if (isPlaocRuntime()) {
    const { dwebServiceWorker } = await import('@plaoc/plugins')
    await dwebServiceWorker.restart()
    return
  }
  window.location.reload()
}
