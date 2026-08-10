/** 生成可打开的邀请链接（跟随当前协议/主机，避免写死 https） */
export function buildInviteLink(address?: string) {
  const addr = (address || '').trim()
  if (!addr) return ''
  return `${window.location.origin}/#/?inviteCode=-inviteTdh-${addr}-inviteTdh-`
}
