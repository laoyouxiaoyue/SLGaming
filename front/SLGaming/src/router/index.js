import { createRouter, createWebHistory } from "vue-router";
import Login from "@/views/login/UserLogin.vue";
import Home from "@/views/home/index.vue";
import Layout from "@/views/layout/index.vue";
import Register from "@/views/login/RegisterUser.vue";
import Forgot from "@/views/login/ForgotPassword.vue";
import Account from "@/views/account/index.vue";
import Companion from "@/views/account/component/CompanionInfo.vue";
import Order from "@/views/account/component/OrderInfo.vue";
import Setting from "@/views/account/component/SettingInfo.vue";
import Wallet from "@/views/account/component/WalletInfo.vue";
import Scion from "@/views/recharge/index.vue";
import ScionRecharge from "@/views/recharge/component/Scionrecharge.vue";
import ScionRecord from "@/views/recharge/component/ScionRecord.vue";
import Pay from "@/views/pay/index.vue";
import ApplyCompanion from "@/views/account/component/ApplyCompanion.vue";
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      component: Layout,
      children: [
        {
          path: "",
          component: Home,
        },
        {
          path: "scion",
          component: Scion,
          children: [
            {
              path: "recharge",
              component: ScionRecharge,
            },
            {
              path: "recond",
              component: ScionRecord,
            },
          ],
        },
        {
          path: "pay",
          component: Pay,
        },
        {
          path: "account",
          component: Account,
          children: [
            {
              path: "companion",
              component: Companion,
            },
            {
              path: "order",
              component: Order,
            },
            {
              path: "setting",
              component: Setting,
            },
            {
              path: "wallet",
              component: Wallet,
            },
            {
              path: "apply_companion",
              component: ApplyCompanion,
            },
          ],
        },
      ],
    },
    {
      path: "/login",
      component: Login,
    },
    {
      path: "/register",
      component: Register,
    },
    {
      path: "/forgot",
      component: Forgot,
    },
  ],
});

export default router;
