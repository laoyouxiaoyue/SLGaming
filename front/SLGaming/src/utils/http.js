import axios from "axios";
import JSONBig from "json-bigint";
import { useUserStore } from "@/stores/userStore";
import { ElMessage } from "element-plus";

const JSONBigInt = JSONBig({ storeAsString: true });

// 创建axios实例
const http = axios.create({
  baseURL: "/api",
  timeout: 100000,
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
      config.headers.Authorization = `Bearer ${userStore.userInfo.refreshToken}`;
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
