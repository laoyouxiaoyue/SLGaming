pipeline {

    // Jenkins 自己运行在 jenkins/jenkins:lts-jdk17 容器中，已通过 docker-compose 预装并挂载了 Go：
    // - /home/ws/go           -> /usr/local/go
    // - /home/ws/go-workspace -> /var/jenkins_home/go-workspace
    // 并在环境变量中设置了 GOROOT / GOPATH / PATH
    // 所以这里直接使用节点自带的 go，不再额外启动 golang:1.23 容器
    agent any

    environment {
        // Go 版本（标记用）
        GO_VERSION   = '1.25.5'
        // 项目路径
        PROJECT_PATH = 'back'
        // 构建输出目录
        BUILD_DIR    = 'build'
        // Go 模块缓存目录（持久化到 Jenkins 挂载目录，避免每次重新下载）
        GOMODCACHE   = "${GOPATH}/pkg/mod"

        // 使用国内 Go 模块代理，避免 proxy.golang.org 超时
        GOPROXY      = 'https://goproxy.cn,direct'
        // 关闭 Go 官方校验（国内环境下常用配置）
        GOSUMDB      = 'off'
    }

    stages {
        stage('Checkout') {
            steps {
                echo 'Checkout code from repository...'
                checkout scm
            }
        }

        stage('Setup Go Environment') {
            steps {
                echo 'Setting up Go environment (from Jenkins node)...'
                sh '''
                    echo "GOROOT = $GOROOT"
                    echo "GOPATH = $GOPATH"
                    echo "GOMODCACHE = $GOMODCACHE"
                    echo "GOPROXY = $GOPROXY"
                    echo "GOSUMDB = $GOSUMDB"
                    echo "PATH   = $PATH"
                    echo "Go version:"
                    go version
                    mkdir -p ${GOMODCACHE}
                '''
            }
        }



        stage('Build Gateway') {
            steps {
                echo 'Building Gateway service...'
                dir("${PROJECT_PATH}") {
                    sh '''
                        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${WORKSPACE}/${BUILD_DIR}/gateway ./services/gateway/gateway.go
                    '''
                }
            }
        }

        stage('Build Code Service') {
            steps {
                echo 'Building Code service...'
                dir("${PROJECT_PATH}") {
                    sh '''
                        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${WORKSPACE}/${BUILD_DIR}/code ./services/code/code.go
                    '''
                }
            }
        }

        stage('Build User Service') {
            steps {
                echo 'Building User service...'
                dir("${PROJECT_PATH}") {
                    sh '''
                        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${WORKSPACE}/${BUILD_DIR}/user ./services/user/user.go
                    '''
                }
            }
        }

        stage('Package') {
            steps {
                echo 'Packaging services...'
                script {
                    sh '''
                        mkdir -p ${BUILD_DIR}/deploy
                        
                        # 复制可执行文件
                        cp ${BUILD_DIR}/gateway ${BUILD_DIR}/deploy/
                        cp ${BUILD_DIR}/code ${BUILD_DIR}/deploy/
                        cp ${BUILD_DIR}/user ${BUILD_DIR}/deploy/
                        
                        # 复制配置文件
                        cp -r ${PROJECT_PATH}/services/gateway/etc ${BUILD_DIR}/deploy/gateway-etc
                        cp -r ${PROJECT_PATH}/services/code/etc ${BUILD_DIR}/deploy/code-etc
                        cp -r ${PROJECT_PATH}/services/user/etc ${BUILD_DIR}/deploy/user-etc
                        
                        # 创建启动脚本
                        cat > ${BUILD_DIR}/deploy/start.sh << 'EOF'
#!/bin/bash
cd $(dirname $0)

# 启动 Code Service
nohup ./code -f code-etc/code.yaml > code.log 2>&1 &
echo $! > code.pid
echo "Code service started, PID: $(cat code.pid)"

# 启动 User Service
nohup ./user -f user-etc/user.yaml > user.log 2>&1 &
echo $! > user.pid
echo "User service started, PID: $(cat user.pid)"

# 启动 Gateway
nohup ./gateway -f gateway-etc/gateway.yaml > gateway.log 2>&1 &
echo $! > gateway.pid
echo "Gateway started, PID: $(cat gateway.pid)"

echo "All services started"
EOF

                        # 创建停止脚本
                        cat > ${BUILD_DIR}/deploy/stop.sh << 'EOF'
#!/bin/bash
cd $(dirname $0)

if [ -f gateway.pid ]; then
    kill $(cat gateway.pid) 2>/dev/null
    rm gateway.pid
    echo "Gateway stopped"
fi

if [ -f code.pid ]; then
    kill $(cat code.pid) 2>/dev/null
    rm code.pid
    echo "Code service stopped"
fi

if [ -f user.pid ]; then
    kill $(cat user.pid) 2>/dev/null
    rm user.pid
    echo "User service stopped"
fi

echo "All services stopped"
EOF

                        chmod +x ${BUILD_DIR}/deploy/*.sh
                    '''
                }
            }
        }

    }

    post {
        success {
            echo 'Build and deploy completed successfully!'
        }
        failure {
            echo 'Build or deploy failed!'
        }
        always {
            echo 'Pipeline finished (post always).'
            // 如需清理工作空间可开启：
            // cleanWs()
        }
    }
}