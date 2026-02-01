<script setup>
import { ref, onMounted } from "vue";
import { useInfoStore } from "@/stores/infoStore";
import { storeToRefs } from "pinia";
import { ElMessage } from "element-plus";
import { upavatarUrlapi } from "@/api/user/info";

const infoStore = useInfoStore();
const { info } = storeToRefs(infoStore);

const formRef = ref(null);
const fileInputRef = ref(null);
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

const initForm = () => {
  if (info.value && Object.keys(info.value).length > 0) {
    form.value = {
      id: info.value.id || "",
      nickname: info.value.nickname || "",
      phone: info.value.phone || "",
      role: info.value.role || 1,
      avatarUrl: info.value.avatarUrl || "",
      bio: info.value.bio || "",
    };
  }
};

const getUserInfo = async () => {
  await infoStore.getUserDetail();
  initForm();
};

const triggerFileUpload = () => {
  fileInputRef.value.click();
};

const handleFileChange = async (e) => {
  const file = e.target.files[0];
  if (!file) return;

  const isJPGOrPNG = file.type === "image/jpeg" || file.type === "image/png";
  const isLt2M = file.size / 1024 / 1024 < 2;

  if (!isJPGOrPNG) {
    ElMessage.error("头像只能是 JPG 或 PNG 格式!");
    return;
  }
  if (!isLt2M) {
    ElMessage.error("头像大小不能超过 2MB!");
    return;
  }

  const res = await upavatarUrlapi(file);
  if (res.data && res.data.avatarUrl) {
    form.value.avatarUrl = res.data.avatarUrl;
    ElMessage.success("头像更换成功");
  }
  e.target.value = "";
};

const onSave = async () => {
  if (!formRef.value) return;
  await formRef.value.validate(async (valid) => {
    if (valid) {
      await infoStore.updateUserInfo(form.value);
      ElMessage.success("保存成功");
      // 更新成功后，Store中的数据已是最新的，这里可以再次同步一下表单（可选）
      initForm();
    }
  });
};

onMounted(() => {
  if (Object.keys(info.value).length > 0) {
    initForm();
  } else {
    getUserInfo();
  }
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
            <input
              type="file"
              ref="fileInputRef"
              style="display: none"
              accept="image/jpeg,image/png"
              @change="handleFileChange"
            />
            <div class="avatar-click-area">
              <el-avatar :size="60" :src="form.avatarUrl" v-if="form.avatarUrl" />
              <el-avatar :size="60" v-else>
                <img src="https://cube.elemecdn.com/e/fd/0fc7d20532fdaf769a25683617711png.png" />
              </el-avatar>
            </div>
            <div
              style="
                display: flex;
                flex-direction: column;
                justify-content: center;
                margin-left: 10px;
              "
            >
              <el-button type="primary" size="small" @click="triggerFileUpload"
                >点击更换头像</el-button
              >
              <div style="font-size: 12px; color: #999; margin-top: 5px">
                支持 JPG/PNG，小于 2MB
              </div>
            </div>
          </div>
        </el-form-item>

        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="form.nickname" placeholder="请输入昵称" />
        </el-form-item>

        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" placeholder="请输入手机号" />
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

        .avatar-click-area {
          position: relative;
          width: 60px;
          height: 60px;
          border-radius: 50%;
          overflow: hidden;
        }
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
