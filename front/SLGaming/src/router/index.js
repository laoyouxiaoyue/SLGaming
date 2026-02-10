import { createRouter, createWebHistory } from "vue-router";
import Login from "@/views/login/UserLogin.vue";
import Home from "@/views/home/index.vue";
import Layout from "@/views/layout/index.vue";
import Register from "@/views/login/RegisterUser.vue";
import Forgot from "@/views/login/ForgotPassword.vue";
import Account from "@/views/account/index.vue";
import Companion from "@/views/account/component/CompanionInfo.vue";
import Setting from "@/views/account/component/SettingInfo.vue";
import Wallet from "@/views/account/component/WalletInfo.vue";
import Scion from "@/views/recharge/index.vue";
import ScionRecharge from "@/views/recharge/component/Scionrecharge.vue";
import ScionRecord from "@/views/recharge/component/ScionRecord.vue";
import Pay from "@/views/pay/index.vue";
import ApplyCompanion from "@/views/account/component/ApplyCompanion.vue";
import SecuritySetting from "@/views/account/component/SecuritySetting.vue";
import Detail from "@/views/detail/index.vue";
import CompanionOrder from "@/views/order/component/CompanionOrder.vue";
import BossOrder from "@/views/order/component/BossOrder.vue";
import Order from "@/views/order/index.vue";
import Relation from "@/views/relation/index.vue";
import FansList from "@/views/relation/component/FansList.vue";
import FollowList from "@/views/relation/component/FollowList.vue";
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
          path: "detail/:id",
          component: Detail,
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
          path: "relation",
          component: Relation,
          children: [
            {
              path: "fans",
              component: FansList,
            },
            {
              path: "follow",
              component: FollowList,
            },
          ],
        },
        {
          path: "order",
          component: Order,
          children: [
            {
              path: "companion",
              component: CompanionOrder,
            },
            {
              path: "boss",
              component: BossOrder,
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
            {
              path: "security",
              component: SecuritySetting,
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
