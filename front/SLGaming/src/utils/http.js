import axios from "axios";
import JSONBig from "json-bigint";
import { useUserStore } from "@/stores/userStore";

import { ElMessage } from "element-plus";
import "element-plus/theme-chalk/el-message.css";

const JSONBigInt = JSONBig({ storeAsString: true });

// 创建axios实例
const http = axios.create({
  baseURL: "/api",
  timeout: 100000,
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
      config.headers.Authorization = `Bearer ${userStore.userInfo.refreshToken}`;
    }

    // 2. 处理发送给后端的 int64 引号问题
    // 仅对普通对象进行处理（排除 FormData、Blob 等）
    if (config.data && typeof config.data === "object" && !(config.data instanceof FormData)) {
      // 强制手动序列化并替换掉长数字两侧的引号
      const jsonStr = JSON.stringify(config.data);
      config.data = jsonStr.replace(/"(\d{16,})"/g, "$1");
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
  (e) => {
    const userStore = useUserStore();
    const errorMsg = e.response?.data?.msg || "请求失败，请稍后重试";
    ElMessage.warning(errorMsg);

    // 401 token 失效处理
    if (e.response && e.response.status === 401) {
      userStore.clearUserInfo();
    }
    return Promise.reject(e);
  },
);

export default http;
