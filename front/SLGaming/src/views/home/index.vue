<script setup>
import { onMounted, ref } from "vue";
import UserCard from "./component/UserCard.vue";
import { getgameskillapi } from "@/api/home/gameskill";
import { getcompanionlist } from "@/api/home/companions";

const skills = ref([]);
const companions = ref([]);
const page = ref(1);
const pageSize = ref(12);
const total = ref(0);
const loading = ref(false);
const finished = ref(false);
const skill = ref("");
const minPrice = ref(undefined);
const maxPrice = ref(undefined);
const status = ref("");

const statusOptions = [
  { label: "全部", value: "" },
  { label: "在线", value: 1 },
  { label: "离线", value: 0 },
  { label: "忙碌", value: 2 },
];

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
  const res = await getgameskillapi();
  skills.value = res.data;
};

const loadCompanions = async () => {
  if (loading.value || finished.value) return;
  loading.value = true;
  try {
    const res = await getcompanionlist({
      gameSkill: skill.value || undefined,
      page: page.value,
      pageSize: pageSize.value,
      minPrice: minPrice.value,
      maxPrice: maxPrice.value,
      status: status.value === "" ? undefined : status.value,
    });
    // 假设 http 已拦截 response.data，res 为后端统一返回结构 { code, data, msg }
    const payload = res?.data || {};
    const list = Array.isArray(payload.companions) ? payload.companions : [];
    const normalized = normalizeCompanions(list);

    companions.value = companions.value.concat(normalized);
    total.value = payload.total ?? 0;
    page.value = (payload.page ?? page.value) + 1;

    if (companions.value.length >= total.value || list.length < pageSize.value) {
      console.log("chufala");
      finished.value = true;
    }
  } catch (error) {
    console.error("加载陪玩列表失败:", error);
  } finally {
    loading.value = false;
  }
};

const loadMore = () => {
  // Element Plus 的 el-scrollbar @end-reached 事件在触底时触发
  // 内部已有 loading 和 finished 状态保护，直接调用即可
  loadCompanions();
};

onMounted(() => {
  loadSkills();
  loadCompanions();
});

const changecom = (name) => {
  skill.value = name;
  handleFilter();
};

const handleFilter = () => {
  page.value = 1;
  total.value = 0;
  companions.value = [];
  finished.value = false;
  loadCompanions();
};
</script>

<template>
  <div class="home">
    <section class="home__section">
      <div class="skills-box">
        <template v-if="skills?.length">
          <el-button
            :type="skill === '' ? 'primary' : ''"
            round
            class="skills-button"
            @click="() => changecom('')"
            >全部</el-button
          >
          <el-button
            v-for="(item, index) in skills"
            :key="index"
            :type="skill === item.name ? 'primary' : ''"
            round
            class="skills-button"
            @click="() => changecom(item.name)"
          >
            {{ item.name }}
          </el-button>
        </template>
        <span v-else class="skill-empty">暂无分类</span>
      </div>
    </section>

    <section class="home__section">
      <div class="filter-box">
        <span class="filter-label">价格区间:</span>
        <el-input-number
          v-model="minPrice"
          :min="0"
          placeholder="最低价格"
          class="filter-input"
          :controls="false"
        />
        <span class="filter-separator">-</span>
        <el-input-number
          v-model="maxPrice"
          :min="0"
          placeholder="最高价格"
          class="filter-input"
          :controls="false"
        />
        <span class="filter-label" style="margin-left: 10px">状态:</span>
        <el-select
          v-model="status"
          placeholder="默认全部"
          style="width: 120px"
          @change="handleFilter"
        >
          <el-option
            v-for="item in statusOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <el-button type="primary" plain class="filter-button" @click="handleFilter">筛选</el-button>
        <el-button
          plain
          @click="
            () => {
              minPrice = undefined;
              maxPrice = undefined;
              status = '';
              handleFilter();
            }
          "
          >重置</el-button
        >
      </div>
    </section>

    <section class="home__section">
      <el-scrollbar height="calc(100vh - 200px)" @end-reached="loadMore">
        <div class="companions">
          <UserCard v-for="item in companions" :key="item.userId" :user="item" />
        </div>
        <div v-if="loading" class="companions__state">加载中...</div>
        <div v-else-if="!companions.length" class="companions__state">暂无陪玩</div>
        <div v-else-if="finished" class="companions__state">已加载全部</div>
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
.skills-button {
  font-size: 18px;
  font-weight: 400;
}
.skill-empty {
  color: #999;
  font-size: 13px;
}

.filter-box {
  margin-left: 10px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 0;
}
.filter-label {
  font-size: 16px;
  color: #606266;
}
.filter-input {
  width: 120px;
}
.filter-separator {
  color: #dcdfe6;
}
.filter-button {
  margin-left: 10px;
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
