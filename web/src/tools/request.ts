import axios from "axios";
import { showLoadingToast, showFailToast, showSuccessToast, closeToast } from "vant";
import userPerson from "../pinia/person";
import lang from '@/i18n/index'
import { adaptRequest } from "./backendAdapter";

const instance = axios.create({
    baseURL: "",
});

function isLegacyApi(url: string = "") {
    return url.startsWith("app_server/");
}

async function dispatch(method: "get" | "post" | "put" | "delete", url: string, data?: any, config?: any) {
    if (isLegacyApi(url)) {
        return adaptRequest(method, url, data, config?.params);
    }
    const token = localStorage.getItem("token");
    const headers = { ...(config?.headers || {}) };
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }
    const res = await instance.request({ method, url, data, params: config?.params, headers });
    return res.data;
}

const request = {
    get(url: string, config?: any) {
        return dispatch("get", url, undefined, config);
    },
    post(url: string, data?: any, config?: any) {
        return dispatch("post", url, data, config);
    },
    put(url: string, data?: any, config?: any) {
        return dispatch("put", url, data, config);
    },
    delete(url: string, config?: any) {
        return dispatch("delete", url, undefined, config);
    },
};

// 请求拦截（仅用于直连 axios 的请求）
instance.interceptors.request.use((config: any) => {
    let token = localStorage.getItem("token")
    if (token) {
        config.headers['Authorization'] = `Bearer ${token}`;
    }
    if (config.data?.loading || config.params?.loading) {
        config.data?.loading ? delete config.data.loading : delete config.params.loading;
        showLoadingToast({
            message: lang("加载中"), duration: 0, overlay: true, overlayStyle: {
                background: "transparent"
            }
        });
    }
    return config;
}, err => {
    showFailToast(err.response?.status);
    return Promise.reject(err);
})

instance.interceptors.response.use(res => {
    closeToast();
    if (res.config.method === "post") {
        if (res.data?.status === "fail") {
            showFailToast(lang("操作失败"));
            return Promise.reject(lang("操作失败"));
        } else {
            if (res.config?.data && !(JSON.parse(res.config?.data)?.nosuccess) && res.config?.data.status === 'OK') {
                showSuccessToast(lang("操作成功"))
            }
        }
    }
    return res.data;
}, err => {
    closeToast();
    const message = err.response?.data?.message;
    const reason = err.response?.data?.reason;
    const status = err.response?.status;
    const noMsg = err.config?.method === "post" && err.config?.data && JSON.parse(err.config.data)?.noMsg;
    const language = localStorage.getItem("language") || "zh";

    if (!noMsg) {
        if (message) {
            showFailToast(language === "zh" ? message : reason);
        } else {
            showFailToast(err.response?.statusText);
        }
    }
    if (status === 401 || message === "user not found") {
        showFailToast(lang("登录过期"));
        const person = userPerson();
        person.outLogin();
    }
    return Promise.reject(err);
})

export default request as typeof axios;
