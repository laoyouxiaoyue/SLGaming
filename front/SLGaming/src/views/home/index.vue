<script setup>
import { onMounted, ref } from "vue";
import UserCard from "./component/user.vue";
import { getgameskillapi } from "@/api/home/gameskill";
import { getcompanionlist } from "@/api/home/companions";

const skills = ref([]);
const companions = ref([]);
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);
const loading = ref(false);
const finished = ref(false);

const normalizeCompanions = (list) =>
  list.map((item) => {
    const skillArray = Array.isArray(item.gameSkills)
      ? item.gameSkills
      : item.gameSkill
        ? [item.gameSkill]
        : [];
    return {
      ...item,
      gameSkills: skillArray,
    };
  });

const loadSkills = async () => {
  try {
    const res = await getgameskillapi();
    const payload = res?.data?.data ?? res?.data ?? {};
    const list = payload.list || payload.skills || payload || [];
    skills.value = Array.isArray(list) ? list : [];
  } catch (error) {
    skills.value = ["王者荣耀"];
  }
};

const loadCompanions = async () => {
  if (loading.value || finished.value) return;
  loading.value = true;
  try {
    const res = await getcompanionlist({
      page: page.value,
      pageSize: pageSize.value,
    });
    const payload = res?.data?.data ?? res?.data ?? {};
    const list = Array.isArray(payload.companions) ? payload.companions : [];
    const normalized = normalizeCompanions(list);
    companions.value = companions.value.concat(normalized);
    total.value = payload.total ?? total.value;
    if (companions.value.length >= total.value && total.value > 0) {
      finished.value = true;
    } else {
      page.value += 1;
    }
  } finally {
    loading.value = false;
  }
};

const loadMore = (direction) => {
  if (direction === "bottom") {
    loadCompanions();
  }
};

onMounted(() => {
  loadSkills();
  loadCompanions();
});
</script>

<template>
  <div class="home">
    <section class="home__section">
      <div class="skills-box">
        <template v-if="skills.length">
          <span class="skill-chip">全部</span>
          <span v-for="(item, index) in skills" :key="index" class="skill-chip">
            {{ item.name || item.label || item }}
          </span>
        </template>
        <span v-else class="skill-empty">暂无分类</span>
      </div>
    </section>

    <section class="home__section">
      <el-scrollbar @end-reached="loadMore">
        <div class="companions">
          <UserCard v-for="item in companions" :key="item.userId" :user="item" />
          <div v-if="loading" class="companions__state">加载中...</div>
          <div v-else-if="!companions.length" class="companions__state">暂无陪玩</div>
          <div v-else-if="finished" class="companions__state">已加载全部</div>
        </div>
      </el-scrollbar>
    </section>
  </div>
</template>

<style scoped>
:global(:root) {
  --layout-padding: 60px;
}

.home {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 0 var(--layout-padding);
  width: 100%;
  max-width: calc(1980px + 2 * var(--layout-padding));
  margin: 0 auto;
}

.skills-box {
  margin-top: 15px;
  margin-left: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
  min-height: 36px;
}

.skill-chip {
  font-size: 18px;
  font-weight: 400;
  padding: 6px 12px;
  border-radius: 999px;
  background: rgb(248, 200, 208);
  color: #ffffff;
}

.skill-empty {
  color: #999;
  font-size: 13px;
}

.companions {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  padding-right: 8px;
}

.companions__state {
  padding: 12px 0;
  text-align: center;
  color: #999;
  font-size: 14px;
}
</style>
