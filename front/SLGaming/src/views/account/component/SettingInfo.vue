<script setup>
import { ref, onMounted } from "vue";
import { getInfoAPI, updateInfoAPI } from "@/api/user/info";
import { ElMessage } from "element-plus";

const formRef = ref(null);
const form = ref({
  id: "",
  nickname: "",
  phone: "",
  role: 1, // 默认老板
  avatarUrl: "",
  bio: "",
});

const rules = {
  nickname: [{ required: true, message: "请输入昵称", trigger: "blur" }],
  role: [{ required: true, message: "请选择角色", trigger: "change" }],
};

const getUserInfo = async () => {
  const res = await getInfoAPI();
  // 回显数据，注意处理空值情况
  if (res.data) {
    form.value = {
      id: res.data.id,
      nickname: res.data.nickname || "",
      phone: res.data.phone || "",
      role: res.data.role || 1,
      avatarUrl: res.data.avatarUrl || "",
      bio: res.data.bio || "",
    };
  }
};

const onSave = async () => {
  if (!formRef.value) return;
  await formRef.value.validate(async (valid) => {
    if (valid) {
      updateInfoAPI(form.value);
    }
  });
};

onMounted(() => {
  getUserInfo();
});
</script>

<template>
  <div class="setting-info">
    <!-- 标题栏 -->
    <div class="setting-title">我的信息</div>

    <!-- 表单内容 -->
    <div class="setting-content">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" class="user-form">
        <el-form-item label="头像" prop="avatarUrl">
          <div class="avatar-wrapper">
            <el-avatar :size="60" :src="form.avatarUrl" v-if="form.avatarUrl">
              <img src="https://cube.elemecdn.com/e/fd/0fc7d20532fdaf769a25683617711png.png" />
            </el-avatar>
            <el-avatar :size="60" v-else>
              <img src="https://cube.elemecdn.com/e/fd/0fc7d20532fdaf769a25683617711png.png" />
            </el-avatar>
            <el-input v-model="form.avatarUrl" placeholder="请输入头像URL链接" style="flex: 1" />
          </div>
        </el-form-item>

        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="form.nickname" placeholder="请输入昵称" />
        </el-form-item>

        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" placeholder="请输入手机号" />
        </el-form-item>

        <el-form-item label="用户角色" prop="role">
          <el-radio-group v-model="form.role">
            <el-radio :value="1">老板</el-radio>
            <el-radio :value="2">陪玩</el-radio>
            <el-radio :value="3">管理员</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="个人简介" prop="bio">
          <el-input type="textarea" v-model="form.bio" :rows="4" placeholder="介绍一下自己吧..." />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="onSave" class="save-btn">保存修改</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<style scoped lang="scss">
.setting-info {
  height: 100%;
  display: flex;
  flex-direction: column;

  .setting-title {
    font-size: 18px;
    padding: 0 0 15px 0; // 上下边距调整以对齐
    border-bottom: 1px solid #f0f0f0;
    margin-bottom: 24px;
    color: #409eff; // 蓝色
    font-weight: 500;
  }

  .setting-content {
    flex: 1;

    .user-form {
      max-width: 600px;

      .avatar-wrapper {
        display: flex;
        align-items: center;
        gap: 16px;
        width: 100%;
      }

      .save-btn {
        width: 120px;
        margin-top: 10px;
        margin-left: 183px;
      }
    }
  }
}
</style>
