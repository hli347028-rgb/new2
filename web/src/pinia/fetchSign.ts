import { showFailToast } from "vant"; // 导入用于显示提示信息的组件
import { ETH } from "@/tools/contract";
import { fetchChallenge } from '@/tools/backendAdapter'
import lang from '@/i18n/index'

const fetchSign = async () => {
  const account = await ETH.getAccount();
  try {
    // 挑战消息包含随机数和有效期，不能复用其他域名或上次登录缓存的签名。
    const challenge = await fetchChallenge(account)
    return await ETH.signMessage(challenge.message)
  } catch (error: any) {
    showFailToast(lang('签名失败'));
    throw error
  }
}

export default fetchSign
