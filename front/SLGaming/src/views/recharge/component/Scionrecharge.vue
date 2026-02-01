<script setup>
import { ref, computed, watch } from "vue";
import { ElMessage } from "element-plus";

// 预设充值档位
const presetAmounts = [6, 18, 68, 233, 648, 998];
const amount = ref(6); // 默认选中第一个
const customAmount = ref(""); // 自定义金额
const isCustom = ref(false); // 是否是自定义模式

const paymentMethod = ref("wechat"); // 支付方式：wechat | alipay

// 监听预设金额选择，选中时清空自定义状态
const selectPreset = (val) => {
  amount.value = val;
  isCustom.value = false;
  customAmount.value = "";
};

// 监听自定义金额输入，输入时激活自定义状态
watch(customAmount, (val) => {
  if (val !== "") {
    isCustom.value = true;
    amount.value = Number(val);
  }
});

// 点击自定义输入框时
const handleCustomFocus = () => {
  isCustom.value = true;
  if (customAmount.value) {
    amount.value = Number(customAmount.value);
  } else {
    amount.value = 0;
  }
};

// 校验自定义金额范围 1-5000
const handleCustomInput = (val) => {
  let num = Number(val);
  if (num > 5000) customAmount.value = 5000;
  // 小于0的情况已包含在type="number"中，但需注意
};

// 最终需支付金额（1:1）
const payAmount = computed(() => {
  return amount.value || 0;
});

const handlePay = () => {
  if (payAmount.value <= 0) {
    ElMessage.warning("请选择或输入充值金额");
    return;
  }
  if (isCustom.value && (payAmount.value < 1 || payAmount.value > 5000)) {
    ElMessage.warning("自定义金额范围为 1-5000 元");
    return;
  }

  ElMessage.success(
    `发起支付：充值 ${amount.value} 帅币，需支付 ¥${payAmount.value}，方式：${paymentMethod.value === "wechat" ? "微信" : "支付宝"}`,
  );
  // TODO: 调用后端下单接口，跳转收银台或弹出二维码
};
</script>

<template>
  <div class="recharge-panel">
    <h2 class="panel-title">充值中心</h2>

    <!-- 版块一：充值档位 -->
    <div class="section">
      <div class="section-title">充值金额 <span class="rate-tip">（1元 = 1帅币）</span></div>
      <div class="amount-grid">
        <div
          v-for="item in presetAmounts"
          :key="item"
          class="amount-item"
          :class="{ active: !isCustom && amount === item }"
          @click="selectPreset(item)"
        >
          <div class="coin-val">{{ item }} 帅币</div>
          <div class="cny-val">¥ {{ item }}</div>
        </div>

        <!-- 自定义金额项 -->
        <div class="amount-item custom-item" :class="{ active: isCustom }">
          <div class="custom-label">自定义金额</div>
          <el-input
            v-model="customAmount"
            placeholder="1-5000"
            type="number"
            class="custom-input"
            @focus="handleCustomFocus"
            @input="handleCustomInput"
          >
            <template #suffix>元</template>
          </el-input>
        </div>
      </div>
    </div>

    <!-- 版块二：支付方式 -->
    <div class="section">
      <div class="section-title">选择支付方式</div>
      <div class="payment-methods">
        <div
          class="pay-item wechat"
          :class="{ active: paymentMethod === 'wechat' }"
          @click="paymentMethod = 'wechat'"
        >
          <div class="icon-wrapper">
            <!-- 这里使用 sl-icon 或者 svg，暂时用文字代替或者模拟图标 -->
            <span class="pay-icon-text wechat-icon">微</span>
          </div>
          <span class="pay-name">微信支付</span>
          <div class="radio-check" v-if="paymentMethod === 'wechat'">✔</div>
        </div>

        <div
          class="pay-item alipay"
          :class="{ active: paymentMethod === 'alipay' }"
          @click="paymentMethod = 'alipay'"
        >
          <div class="icon-wrapper">
            <span class="pay-icon-text alipay-icon">支</span>
          </div>
          <span class="pay-name">支付宝</span>
          <div class="radio-check" v-if="paymentMethod === 'alipay'">✔</div>
        </div>
      </div>
    </div>

    <!-- 版块三：操作区域 -->
    <div class="action-bar">
      <div class="total-price">
        应付金额：<span class="price-num">¥ {{ payAmount }}</span>
      </div>
      <el-button type="primary" size="large" class="pay-btn" @click="handlePay"
        >前往收银台</el-button
      >
    </div>
  </div>
</template>

<style scoped lang="scss">
.recharge-panel {
  padding: 0 10px;

  .panel-title {
    font-size: 20px;
    font-weight: 600;
    margin-bottom: 25px;
    color: #333;
    border-left: 4px solid #ff6b35;
    padding-left: 12px;
  }

  .section {
    margin-bottom: 35px;

    .section-title {
      font-size: 16px;
      font-weight: 500;
      margin-bottom: 15px;
      color: #333;

      .rate-tip {
        font-size: 12px;
        color: #999;
        font-weight: normal;
        margin-left: 8px;
      }
    }

    .amount-grid {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 15px;

      .amount-item {
        border: 1px solid #e0e0e0;
        border-radius: 8px;
        height: 90px;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        transition: all 0.2s;
        position: relative;
        background: #fff;

        &:hover {
          border-color: #ffccb8;
          box-shadow: 0 2px 8px rgba(255, 107, 53, 0.1);
        }

        &.active {
          border-color: #ff6b35;
          background-color: #fff6f2;
          color: #ff6b35;

          .coin-val,
          .cny-val,
          .custom-label {
            color: #ff6b35;
          }

          &::after {
            content: "✔";
            position: absolute;
            bottom: -1px;
            right: -1px;
            background: #ff6b35;
            color: #fff;
            font-size: 10px;
            padding: 0 4px;
            border-top-left-radius: 6px;
            border-bottom-right-radius: 6px;
          }
        }

        .coin-val {
          font-size: 18px;
          font-weight: 600;
          color: #333;
          margin-bottom: 4px;
        }

        .cny-val {
          font-size: 13px;
          color: #999;
        }

        &.custom-item {
          padding: 0 15px;

          .custom-label {
            font-size: 14px;
            color: #333;
            margin-bottom: 8px;
          }

          .custom-input {
            :deep(.el-input__wrapper) {
              box-shadow: none;
              border-bottom: 1px solid #ddd;
              border-radius: 0;
              padding: 0;
              background: transparent;
            }
            :deep(.el-input__inner) {
              text-align: center;
              font-size: 16px;
              color: inherit;
            }
            :deep(.el-input__inner::-webkit-inner-spin-button) {
              -webkit-appearance: none;
            }
          }
        }
      }
    }

    .payment-methods {
      display: flex;
      gap: 20px;

      .pay-item {
        flex: 1;
        max-width: 220px;
        height: 60px;
        border: 1px solid #e0e0e0;
        border-radius: 6px;
        display: flex;
        align-items: center;
        padding: 0 20px;
        cursor: pointer;
        transition: all 0.2s;
        position: relative;

        &:hover {
          border-color: #bbb;
        }

        &.active {
          border-color: #ff6b35;
          background-color: #fffbf9;
        }

        .icon-wrapper {
          margin-right: 12px;
          display: flex;
          align-items: center;

          .pay-icon-text {
            width: 24px;
            height: 24px;
            border-radius: 4px;
            color: #fff;
            font-size: 14px;
            display: flex;
            justify-content: center;
            align-items: center;
            font-weight: bold;

            &.wechat-icon {
              background-color: #09bb07;
            }
            &.alipay-icon {
              background-color: #1677ff;
            }
          }
        }

        .pay-name {
          font-size: 15px;
          color: #333;
        }

        .radio-check {
          position: absolute;
          right: 15px;
          color: #ff6b35;
          font-size: 16px;
          font-weight: bold;
        }
      }
    }
  }

  .action-bar {
    margin-top: 50px;
    padding-top: 20px;
    border-top: 1px solid #f0f0f0;
    display: flex;
    justify-content: flex-end;
    align-items: center;
    gap: 30px;

    .total-price {
      font-size: 16px;
      color: #333;

      .price-num {
        font-size: 28px;
        color: #ff6b35;
        font-weight: bold;
        margin-left: 5px;
      }
    }

    .pay-btn {
      width: 180px;
      font-size: 16px;
      font-weight: 500;
      background: linear-gradient(135deg, #ff8e61, #ff6b35);
      border: none;

      &:hover {
        background: linear-gradient(135deg, #ff9ca4, #ff7a45);
        opacity: 0.9;
      }
    }
  }
}
</style>
