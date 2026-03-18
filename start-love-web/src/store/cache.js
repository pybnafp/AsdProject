import {httpPost} from "@/utils/http";
import Storage from "good-storage";
import {randString} from "@/utils/libs";


export function checkSession() {
    return new Promise((resolve, reject) => {
        httpPost('/api/users/profile').then(res => {
            resolve(res.data)
        }).catch(e => {
            reject(e)
        })
    })
}



export function getClientId() {
    let clientId = Storage.get('client_id')
    if (clientId) {
        return clientId
    }
    clientId = randString(42)
    Storage.set('client_id', clientId)
    return clientId
}