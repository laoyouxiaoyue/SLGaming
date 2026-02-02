<script setup>
import { ref, onMounted } from "vue";
import { useInfoStore } from "@/stores/infoStore";
import { storeToRefs } from "pinia";
import { ElMessage } from "element-plus";
import { upavatarUrlapi } from "@/api/user/info";
import { CameraFilled } from "@element-plus/icons-vue";

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
    getUserInfo();
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
    <div class="panel-title">我的信息</div>

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
              <el-avatar :size="60" v-else class="no-avatar-trigger" @click="triggerFileUpload">
                <el-icon :size="24"><CameraFilled /></el-icon>
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
    flex: 1;

    .user-form {
      max-width: 600px;
      margin-top: 10px;

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
          border: 2px solid #fff;
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);

          .no-avatar-trigger {
            background-color: #f2f3f5;
            cursor: pointer;
            color: #909399;
            transition: all 0.3s;

            &:hover {
              color: #ff6b35;
              background-color: #fff6f2;
            }
          }
        }
      }

      .save-btn {
        width: 120px;
        margin-top: 10px;
        margin-left: 0;
        background: linear-gradient(135deg, #ff8e61, #ff6b35);
        border: none;
        font-weight: 500;

        &:hover {
          background: linear-gradient(135deg, #ff9ca4, #ff7a45);
          opacity: 0.9;
        }
      }

      :deep(.el-form-item__label) {
        font-weight: 500;
      }
    }
  }
}
</style>
