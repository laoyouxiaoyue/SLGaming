# 启动前端项目的脚本
# 假设镜像已推送并存放在 ${DEPLOY_PATH}/vue/slgaming-vue.tar.gz

# 1. 加载Docker镜像
docker load < docker-images/slgaming-vue.tar.gz

# 2. 运行容器（映射端口80到宿主机的9090端口）
docker run -d --name slgaming-vue -p 9090:80 slgaming-vue:latest

# 3. 检查容器状态
docker ps | grep slgaming-vue

# 停止容器：docker stop slgaming-vue
# 删除容器：docker rm slgaming-vue
# 删除镜像：docker rmi slgaming-vue:latest