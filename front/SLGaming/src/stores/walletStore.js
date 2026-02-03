import { defineStore } from "pinia";
import { ref } from "vue";
import { getwalletapi } from "@/api/money/wallet";

export const useWalletStore = defineStore(
  "wallet",
  () => {
    const walletInfo = ref({});

    // 获取钱包详情
    const getWallet = async () => {
      const res = await getwalletapi();
      walletInfo.value = res.data;
    };

    // 清除钱包信息
    const clearWallet = () => {
      walletInfo.value = {};
    };

    return {
      walletInfo,
      getWallet,
      clearWallet,
    };
  },
  {
    persist: true,
  },
);
