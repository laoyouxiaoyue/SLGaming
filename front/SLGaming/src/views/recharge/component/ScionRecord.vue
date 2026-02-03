<script setup>
import { ref, onMounted } from "vue";
import { getrechargelistapi } from "@/api/money/wallet";
import zhCn from "element-plus/es/locale/lang/zh-cn";
import dayjs from "dayjs";
import "dayjs/locale/zh-cn";

dayjs.locale("zh-cn");

const locale = zhCn;

// 查询参数
const queryParams = ref({
  page: 1,
  pageSize: 20,
  status: 1,
});

const total = ref(0);
const loading = ref(false);
const rechargeList = ref([]);

// 状态映射
const statusMap = {
  0: { text: "待支付", type: "warning" },
  1: { text: "成功", type: "success" },
  2: { text: "失败", type: "danger" },
  3: { text: "关闭", type: "info" },
};

// 获取数据
const fetchRecords = async () => {
  loading.value = true;
  try {
    const res = await getrechargelistapi(queryParams.value);
    if (res.code === 0) {
      rechargeList.value = res.data.orders || [];
      total.value = res.data.total || 0;
    }
  } finally {
    loading.value = false;
  }
};

// 分页处理
const handleCurrentChange = (val) => {
  queryParams.value.page = val;
  fetchRecords();
};

onMounted(() => {
  fetchRecords();
});
</script>

<template>
  <el-config-provider :locale="locale">
    <div class="setting-info">
      <div class="panel-title">消费记录</div>
      <div class="setting-content">
        <el-table :data="rechargeList" v-loading="loading" style="width: 100%" border stripe>
          <el-table-column prop="orderNo" label="订单号" min-width="180" />
          <el-table-column label="金额" width="120">
            <template #default="{ row }">
              <span class="amount">{{ row.amount.toFixed(2) }} 帅币</span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="statusMap[row.status]?.type">
                {{ statusMap[row.status]?.text || "未知" }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="创建时间" width="180">
            <template #default="{ row }">
              {{ dayjs(row.createdAt).format("YYYY-MM-DD HH:mm:ss") }}
            </template>
          </el-table-column>
        </el-table>

        <!-- 分页 -->
        <div class="pagination-container">
          <el-pagination
            v-model:current-page="queryParams.page"
            v-model:page-size="queryParams.pageSize"
            layout="total, prev, pager, next, jumper"
            :total="total"
            @current-change="handleCurrentChange"
          />
        </div>
      </div>
    </div>
  </el-config-provider>
</template>

<style scoped lang="scss">
.setting-info {
  padding: 0 10px;

  .panel-title {
    font-size: 20px;
    font-weight: 600;
    margin-bottom: 25px;
    color: #333;
    border-left: 4px solid #ff6b35;
    padding-left: 12px;
  }

  .setting-content {
    padding-top: 10px;

    .amount {
      color: #ff6b35;
      font-weight: bold;
    }

    .pagination-container {
      margin-top: 20px;
      display: flex;
      justify-content: flex-end;
    }
  }
}
</style>
