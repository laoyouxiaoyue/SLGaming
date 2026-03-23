import axios from "axios";
import JSONBig from "json-bigint";
import { useUserStore } from "@/stores/userStore";
import { refreshTokenAPI } from "@/api/user/auth"; // 1. 引入刷新token的API
import router from "@/router";

import { ElMessage } from "element-plus";
import "element-plus/theme-chalk/el-message.css";

const JSONBigInt = JSONBig({ storeAsString: true });

// 2. 定义状态变量和队列
let isRefreshing = false; // 是否正在刷新token
let requestQueue = []; // 请求队列，存储刷新token期间失败的请求

// 创建axios实例
const http = axios.create({
  baseURL: "/api",
  timeout: 1000000,
  // 显式设置 responseType 为 text，防止 axios/浏览器自动尝试解析 JSON 导致精度丢失
  responseType: "text",
  // 处理服务器返回的大整数精度问题
  transformResponse: [
    function (data) {
      try {
        // 使用 json-bigint 解析数据，超长数字会被自动转为字符串
        return JSONBigInt.parse(data);
      } catch {
        // 如果解析失败（比如不是 JSON 格式），返回原始数据
        return data;
      }
    },
  ],
});

// axios请求拦截器
http.interceptors.request.use(
  (config) => {
    const userStore = useUserStore();

    // 1. 添加 Token
    if (userStore.userInfo.accessToken) {
      config.headers.Authorization = `Bearer ${userStore.userInfo.accessToken}`; // BUG修复：这里应该是 accessToken
    }

    // 2. 处理发送给后端的 int64 引号问题
    // 仅对普通对象进行处理（排除 FormData、Blob 等）
    if (config.data && typeof config.data === "object" && !(config.data instanceof FormData)) {
      // 使用 JSONBigInt 序列化
      const jsonStr = JSONBigInt.stringify(config.data);
      // 替换长数字字符串为数字类型，支持可能存在的空格
      config.data = jsonStr.replace(/:\s*"(\d{16,})"/g, ":$1");
      config.headers["Content-Type"] = "application/json";
    }

    return config;
  },
  (e) => {
    return Promise.reject(e);
  },
);

// axios响应拦截器
http.interceptors.response.use(
  (res) => {
    // 返回 data 核心数据
    return res.data;
  },
  async (e) => {
    const userStore = useUserStore();
    const { config, response } = e;

    // 3. 如果不是401错误，直接提示错误并返回
    if (!response || response.status !== 401) {
      const errorMsg = e.response?.data?.msg || "请求失败，请稍后重试";
      ElMessage.warning(errorMsg);
      return Promise.reject(e);
    }

    // 4. 如果正在刷新token，将当前失败的请求加入队列
    if (isRefreshing) {
      return new Promise((resolve) => {
        requestQueue.push((newToken) => {
          config.headers.Authorization = `Bearer ${newToken}`;
          resolve(http(config));
        });
      });
    }

    // 5. 开始刷新token
    isRefreshing = true;
    try {
      const res = await refreshTokenAPI();
      const { accessToken, refreshToken } = res.data;

      // 5.1 更新 Pinia 和请求头
      userStore.setTokens({ accessToken, refreshToken });
      config.headers.Authorization = `Bearer ${accessToken}`;

      // 5.2 重新执行队列中的所有请求
      requestQueue.forEach((cb) => cb(accessToken));
      requestQueue = []; // 清空队列

      // 5.3 重新执行本次失败的请求
      return http(config);
    } catch (error) {
      // 6. 如果刷新token也失败了，清除用户信息并跳转到登录页
      userStore.clearUserInfo();
      ElMessage.error("登录已过期，请重新登录");
      router.push("/login");
      return Promise.reject(error);
    } finally {
      isRefreshing = false;
    }
  },
);

export default http;
