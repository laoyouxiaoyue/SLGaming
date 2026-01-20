import { createRouter, createWebHistory } from "vue-router";
import Login from "@/views/login/login.vue";
import Home from "@/views/home/index.vue";
import Layout from "@/views/layout/index.vue";
import Register from "@/views/login/Register.vue";
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
  ],
});

export default router;
