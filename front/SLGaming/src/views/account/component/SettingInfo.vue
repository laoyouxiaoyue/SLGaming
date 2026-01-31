<script setup>
import { ref, onMounted } from "vue";
import { useInfoStore } from "@/stores/infoStore";
import { storeToRefs } from "pinia";
import { ElMessage } from "element-plus";
import { Camera } from "@element-plus/icons-vue";

const infoStore = useInfoStore();
const { info } = storeToRefs(infoStore);

const formRef = ref(null);
const fileInputRef = ref(null);
const form = ref({
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

  // 模拟预览 (此处应调用后端上传接口)
  const reader = new FileReader();
  reader.readAsDataURL(file);
  reader.onload = () => {
    form.value.avatarUrl = reader.result;
    ElMessage.success("图片已选择 (仅为本地预览，需后端接口支持)");
  };
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
            <div class="avatar-click-area" @click="triggerFileUpload">
              <el-avatar :size="60" :src="form.avatarUrl" v-if="form.avatarUrl">
                <img src="https://cube.elemecdn.com/e/fd/0fc7d20532fdaf769a25683617711png.png" />
              </el-avatar>
              <el-avatar :size="60" v-else>
                <img src="https://cube.elemecdn.com/e/fd/0fc7d20532fdaf769a25683617711png.png" />
              </el-avatar>
              <div class="avatar-mask">
                <el-icon><Camera /></el-icon>
              </div>
            </div>
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

        .avatar-click-area {
          position: relative;
          cursor: pointer;
          width: 60px;
          height: 60px;
          border-radius: 50%;
          overflow: hidden;

          &:hover .avatar-mask {
            opacity: 1;
          }

          .avatar-mask {
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.5);
            display: flex;
            justify-content: center;
            align-items: center;
            opacity: 0;
            transition: opacity 0.3s;
            color: #fff;

            i {
              font-size: 20px;
            }
          }
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
