import { defineStore } from "pinia";
import { ref } from "vue";
import { getwalletapi } from "@/api/money/wallet";

export const useWalletStore = defineStore(
  "wallet",
  () => {
    /**
     * walletInfo存放内容形式:
     * {
     *   userId: string,       // 用户ID，例如 "1996080936390758400"
     *   balance: number,      // 余额，例如 628
     *   frozenBalance: number // 冻结余额，例如 0
     * }
     */
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
