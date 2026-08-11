import { createI18n } from 'vue-i18n';
import zhNested from '../i18n/lang/zh'
import zhTw from '../i18n/lang/zh-tw'
import en from '../i18n/lang/en'
import ja from '../i18n/lang/ja'
import ko from '../i18n/lang/ko'
import vi from '../i18n/lang/vi'
// lang('中文文案') 以中文为 key；与嵌套英文 key 并存
import zhFlat from '../i18n/language/zh.json'
import enFlat from '../i18n/language/en.json'

// 最少改动：把已有 flat 表注入各语言。
// 非英文语言暂用 enFlat 兜底，避免切语言后仍显示中文。
export default createI18n({
    legacy: false,
    locale: localStorage.getItem("lan") || "zh",
    missingWarn: false,
    fallbackWarn: false,
    messages: {
        zh: { ...zhFlat, ...zhNested },
        'zh-tw': { ...zhFlat, ...zhTw },
        en: { ...enFlat, ...en },
        ja: { ...enFlat, ...ja },
        ko: { ...enFlat, ...ko },
        vi: { ...enFlat, ...vi },
    },
})
