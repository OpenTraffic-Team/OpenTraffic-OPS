import { login, logout, getInfo, getPublicKey } from '@/api/login'
import { getToken, setToken, removeToken } from '@/utils/auth'
import defAva from '@/assets/images/profile.png'
import { encrypt } from "../../utils/jsencrypt";

const useUserStore = defineStore(
  'user',
  {
    state: () => ({
      token: getToken(),
      id: '',
      name: '',
      avatar: '',
      roles: [],
      permissions: []
    }),
    actions: {
      getPublicKey() {
        return new Promise((resolve, reject) => {
          getPublicKey()
            .then(res => {
              resolve(res)
            })
            .catch(error => {
              reject(error)
            })
        })
      },
      // 登录
      login(userInfo) {
        const username = userInfo.username.trim()
        const code = userInfo.code
        const uuid = userInfo.uuid
        return new Promise((resolve, reject) => {
          getPublicKey().then(res => {
            // res 是统一响应包装 {code, msg, data}
            const keyData = res.data || res
            let publicKey = keyData.publicKey
            if (!publicKey) {
              reject(new Error('获取公钥失败'))
              return
            }
            //调用加密方法(传密码和公钥)
            const password = encrypt(userInfo.password, publicKey)
            if (password === false) {
              reject(new Error('密码加密失败'))
              return
            }
            login(username, password, code, uuid).then(res => {
              // res 是统一响应包装 {code, msg, data}
              const data = res.data || res
              setToken(data.token)
              this.token = data.token
              resolve()
            }).catch(error => {
              reject(error)
            })
          }).catch(error => {
            reject(error)
          })
        })
      },
      // 获取用户信息
      getInfo() {
        return new Promise((resolve, reject) => {
          getInfo().then(res => {
            // res 是统一响应包装 {code, msg, data}
            const data = res.data || res
            const user = data.user
            const avatar = (user.avatar == "" || user.avatar == null) ? defAva : import.meta.env.VITE_APP_BASE_API + user.avatar;

            if (data.roles && data.roles.length > 0) {
              this.roles = data.roles
              this.permissions = data.permissions
            } else {
              this.roles = ['ROLE_DEFAULT']
            }
            this.id = user.userId
            this.name = user.userName
            this.avatar = avatar
            resolve(data)
          }).catch(error => {
            reject(error)
          })
        })
      },
      // 退出系统
      logOut() {
        return new Promise((resolve, reject) => {
          logout(this.token).then(() => {
            this.token = ''
            this.roles = []
            this.permissions = []
            removeToken()
            resolve()
          }).catch(error => {
            reject(error)
          })
        })
      }
    }
  })

export default useUserStore
