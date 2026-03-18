
import axios from 'axios'
import {isMobileV2} from "@/utils/libs";

axios.defaults.timeout = 180000;
axios.defaults.baseURL = "";
// axios.defaults.baseURL = process.env.VUE_APP_API_HOST
axios.defaults.withCredentials = true;
//axios.defaults.headers.post['Content-Type'] = 'application/json'

// HTTP拦截器
axios.interceptors.request.use(
    config => {
        return config
    }, error => {
        return Promise.reject(error)
    })
axios.interceptors.response.use(
    response => {
        return response
    }, error => {
        if (error.response.status === 401) {
            error.message = "未登录"
            if (isMobileV2()) {
                window.location.href = "/mobile/login"
            } else {
                window.location.href = "/login"
            }
            return Promise.reject(error.response.data)
        }
        if (error.response.status === 400) {
            return Promise.reject(new Error(error.response.data.message))
        } else {
            return Promise.reject(error)
        }
    })


// send a http get request
export function httpGet(url, params = {}) {
    return new Promise((resolve, reject) => {
        axios.get(url, {
            params: params
        }).then(response => {
            resolve(response.data)
        }).catch(err => {
            reject(err)
        })
    })
}


// send a http post request
export function httpPost(url, data = {}, options = {}) {
    return new Promise((resolve, reject) => {
        axios.post(url, data, options).then(response => {
            resolve(response.data)
        }).catch(err => {
            reject(err)
        })
    })
}

export function httpDownload(url) {
    return new Promise((resolve, reject) => {
        axios({
            method: 'GET',
            url: url,
            responseType: 'blob' // 将响应类型设置为 `blob`
        }).then(response => {
            resolve(response)
        }).catch(err => {
            reject(err)
        })
    })
}

export function httpPostDownload(url, data) {
    return new Promise((resolve, reject) => {
        axios({
            method: 'POST',
            url: url,
            data: data,
            responseType: 'blob' // 将响应类型设置为 `blob`
        }).then(response => {
            resolve(response)
        }).catch(err => {
            reject(err)
        })
    })
}