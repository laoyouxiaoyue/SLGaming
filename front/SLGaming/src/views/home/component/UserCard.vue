<script setup>
import { computed } from "vue";
const props = defineProps({
  user: {
    type: Object,
    default: () => ({}),
  },
});

const map = {
  王者荣耀: "https://game.gtimg.cn/images/yxzj/img201606/skin/hero-info/545/545-bigskin-2.jpg",
  三角洲行动:
    "https://image.baidu.com/search/down?url=https://tvax1.sinaimg.cn/large/e2737f34gy1i56ca9fh0hj22yo1o0x6p.jpg",
  无畏契约:
    "https://image.baidu.com/search/down?url=https://tvax1.sinaimg.cn/large/ea98bce0gy1ha20l29adnj21hc0u0wpc.jpg",
  英雄联盟: "https://www.bizhigq.com/pc-img/2023-05/g1931.jpg",
};

const mediaStyle = computed(() => {
  let url = props.user.avatarUrl;

  // 适配头像地址: 如果是远程的具体IP地址，转为相对路径走代理
  if (url && url.includes("http://120.26.29.242")) {
    url = url.replace("http://120.26.29.242", "");
  }

  if (!url) {
    url = map[props.user.gameSkill] || "";
  }
  return { backgroundImage: `url(${url})` };
});

const statusText = computed(() => {
  const status = Number(props.user.status);
  if (status === 0) return "离线";
  if (status === 2) return "忙碌";
  return "在线";
});

const statusClass = computed(() => {
  const status = Number(props.user.status);
  if (status === 0) return "is-offline";
  if (status === 2) return "is-busy";
  return "is-online";
});

const gameSkillsText = computed(() => {
  let skills = props.user.gameSkill;

  // 如果是字符串，尝试解析为数组
  if (typeof skills === "string") {
    try {
      skills = JSON.parse(skills);
    } catch {
      return skills;
    }
  }

  if (Array.isArray(skills) && skills.length) return skills.join("/");
  return "暂无技能";
});

const priceText = computed(() => {
  const price = props.user.pricePerHour;
  if (price === 0 || price) return `￥${price}/小时`;
  if (props.user.minPrice || props.user.maxPrice) {
    const min = props.user.minPrice ? `￥${props.user.minPrice}` : "";
    const max = props.user.maxPrice ? `￥${props.user.maxPrice}` : "";
    return [min, max].filter(Boolean).join(" - ");
  }
  return "价格待定";
});

const totalOrdersText = computed(() => {
  const value = props.user.totalOrders;
  return value === 0 || value ? value : "--";
});

const ratingText = computed(() => {
  const value = props.user.rating;
  return value === 0 || value ? `${value}%` : "--";
});
</script>

<template>
  <div class="user-card">
    <div class="user-card__media" :style="mediaStyle">
      <span class="user-card__status" :class="statusClass">{{ statusText }}</span>
      <div class="user-card__stats">
        <span class="user-card__stat">订单 {{ totalOrdersText }}</span>
        <span class="user-card__stat">好评率 {{ ratingText }}</span>
      </div>
    </div>

    <div class="user-card__info">
      <div class="user-card__name">
        {{ user.nickname || "未命名" }}
        <sl-icon name="icon-guanfangrenzheng" v-if="user?.isVerified === true"></sl-icon>
      </div>
      <div class="user-card__meta">
        <span class="user-card__skills">{{ gameSkillsText }}</span>
        <span class="user-card__price">{{ priceText }}</span>
      </div>
    </div>

    <div class="user-card__action">
      <button
        class="user-card__button"
        type="button"
        @click="$router.push(`/detail/${user.userId}`)"
      >
        去下单
      </button>
    </div>
  </div>
</template>

<style scoped>
.user-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 12px;
  align-items: stretch;
}

.user-card__media {
  position: relative;
  min-height: 150px;
  border-radius: 10px;
  background-size: cover;
  background-position: center;
  overflow: hidden;
}

.user-card__status {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 12px;
  color: #fff;
  background: rgba(0, 0, 0, 0.6);
}

.user-card__status.is-online {
  background: rgba(46, 204, 113, 0.9);
}

.user-card__status.is-busy {
  background: rgba(241, 196, 15, 0.9);
}

.user-card__status.is-offline {
  background: rgba(149, 165, 166, 0.9);
}

.user-card__stats {
  position: absolute;
  left: 8px;
  bottom: 8px;
  display: flex;
  gap: 8px;
  padding: 4px 6px;
  background: rgba(0, 0, 0, 0.55);
  border-radius: 8px;
  color: #fff;
  font-size: 12px;
}

.user-card__stat {
  white-space: nowrap;
}

.user-card__info {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 8px;
}

.user-card__name {
  font-size: 18px;
  font-weight: 600;
  color: #222;
}

.user-card__meta {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  gap: 12px;
  color: #666;
  font-size: 14px;
}

.user-card__skills {
  max-width: 220px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-card__price {
  color: #ff6b35;
  font-weight: 600;
}

.user-card__action {
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

.user-card__button {
  width: 100%;
  padding: 10px 18px;
  border: none;
  border-radius: 999px;
  background: linear-gradient(135deg, #fc8d61, #f1787a);
  color: #fff;
  font-weight: 600;
  font-size: 16px;
  cursor: pointer;
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease;
}

.user-card__button:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 14px rgba(255, 77, 79, 0.35);
}

.user-card__button:active {
  transform: translateY(0);
  box-shadow: none;
}
</style>
