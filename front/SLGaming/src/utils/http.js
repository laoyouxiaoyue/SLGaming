import axios from "axios";
import { useUserStore } from "@/stores/userStore";

import { ElMessage } from "element-plus";
import "element-plus/theme-chalk/el-message.css";
import router from "@/router";

// 创建axios实例（相当于造了一个专属的"请求工具"）
const http = axios.create({
  baseURL: "http://120.26.29.194:8888/api",
  timeout: 10000, // 请求超过5秒没响应就报错
});

// axios请求拦截器：请求"发出去之前"会经过这里
http.interceptors.request.use(
  // 第一个函数：请求成功准备发送时执行
  (config) => {
    // config就是你的请求配置（比如请求地址、请求头、参数等）
    // 这里可以修改config，比如给所有请求加token、加请求头
    // console.log("请求要发出去啦！我可以在这里加东西", config);
    // 1. 从pinia获取token数据
    const userStore = useUserStore();
    // 2. 按照后端的要求拼接token数据
    if (userStore.userInfo.accessToken) {
      config.headers.Authorization = `Bearer ${userStore.userInfo.accessToken}`;
    }
    return config; // 必须返回config，请求才能继续发出去
  },
  // 第二个函数：请求准备阶段出错时执行（比如参数格式错）
  (e) => {
    console.log("请求还没发出去就出错了", e);
    return Promise.reject(e); // 把错误抛出去，让外面能捕获
  },
);

// axios响应拦截器：服务器"返回结果后"会经过这里
http.interceptors.response.use(
  // 第一个函数：响应成功（状态码2xx）时执行
  (res) => {
    // res是服务器返回的完整数据（包含状态码、响应头、数据体等）
    // 这里我们只返回res.data，外面用的时候就不用每次都写.data了
    // console.log("服务器返回数据啦！我帮你把核心数据提出来了", res);
    return res.data;
  },
  // 第二个函数：响应失败（比如404、500、超时）时执行
  (e) => {
    const userStore = useUserStore();
    // 可以在这里统一处理错误，比如401跳登录、500提示服务器错误
    const errorMsg = e.response?.data.msg || "请求失败，请稍后重试";
    ElMessage.warning(errorMsg);
    //401 token 失效处理
    // 1、清楚本地用户数据
    // 2、跳转到登录页
    if (e.response && e.response.status === 401) {
      userStore.clearUserInfo();
      router.push("/login");
    }
    return Promise.reject(e); // 把错误抛出去，让外面能捕获
  },
);

export default http;
