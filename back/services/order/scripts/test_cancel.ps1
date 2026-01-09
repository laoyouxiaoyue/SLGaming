# 订单取消功能测试
$BASE_URL = "http://localhost:8888"
$PHONE_BOSS = "13800001001"
$PHONE_COMPANION = "13800002001"
$PASSWORD = "Test123456"

Write-Host "=== 订单取消功能测试 ===" -ForegroundColor Cyan
Write-Host ""

# 辅助函数
function Invoke-Api {
    param(
        [string]$Method,
        [string]$Path,
        [hashtable]$Headers = @{},
        [object]$Body = $null
    )
    $uri = "$BASE_URL$Path"
    $params = @{
        Uri = $uri
        Method = $Method
        ContentType = "application/json"
        Headers = $Headers
        ErrorAction = "Stop"
    }
    if ($Body) {
        $params.Body = ($Body | ConvertTo-Json -Compress)
    }
    try {
        return Invoke-RestMethod @params
    } catch {
        if ($_.ErrorDetails.Message) {
            try {
                return ($_.ErrorDetails.Message | ConvertFrom-Json)
            } catch {
                throw
            }
        }
        throw
    }
}

# 1. 老板登录
Write-Host "[1] 老板登录..." -ForegroundColor Yellow
try {
    $bossResp = Invoke-Api -Method Post -Path "/api/user/login" -Body @{
        phone = $PHONE_BOSS
        password = $PASSWORD
    }
    $BOSS_TOKEN = $bossResp.data.accessToken
    $BOSS_ID = $bossResp.data.id
    Write-Host "✓ 老板登录成功 (ID: $BOSS_ID)" -ForegroundColor Green
} catch {
    Write-Host "✗ 老板登录失败: $_" -ForegroundColor Red
    Write-Host "请先注册用户" -ForegroundColor Yellow
    exit
}
Write-Host ""

# 2. 陪玩登录
Write-Host "[2] 陪玩登录..." -ForegroundColor Yellow
try {
    $companionResp = Invoke-Api -Method Post -Path "/api/user/login" -Body @{
        phone = $PHONE_COMPANION
        password = $PASSWORD
    }
    $COMPANION_TOKEN = $companionResp.data.accessToken
    $COMPANION_ID = $companionResp.data.id
    Write-Host "✓ 陪玩登录成功 (ID: $COMPANION_ID)" -ForegroundColor Green
} catch {
    Write-Host "✗ 陪玩登录失败: $_" -ForegroundColor Red
    Write-Host "请先注册用户" -ForegroundColor Yellow
    exit
}
Write-Host ""

# 3. 创建订单
Write-Host "[3] 创建订单..." -ForegroundColor Yellow
try {
    $orderResp = Invoke-Api -Method Post -Path "/api/order" `
        -Headers @{Authorization = "Bearer $BOSS_TOKEN"} `
        -Body @{
            companionId = [int64]$COMPANION_ID
            gameName = "王者荣耀"
            gameMode = "排位赛"
            durationMinutes = 60
        }
    $ORDER_ID = $orderResp.data.id
    $ORDER_STATUS = $orderResp.data.status
    Write-Host "✓ 订单创建成功 (ID: $ORDER_ID, 状态: $ORDER_STATUS)" -ForegroundColor Green
} catch {
    Write-Host "✗ 创建订单失败: $_" -ForegroundColor Red
    exit
}
Write-Host ""

# 测试场景1
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "测试场景1: 老板取消未接单订单" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "[4] 老板取消订单..." -ForegroundColor Yellow
try {
    $cancelResp1 = Invoke-Api -Method Post -Path "/api/order/cancel" `
        -Headers @{Authorization = "Bearer $BOSS_TOKEN"} `
        -Body @{
            orderId = [int64]$ORDER_ID
            reason = "临时有事"
        }
    Write-Host "响应: $($cancelResp1 | ConvertTo-Json)" -ForegroundColor Gray
    if ($cancelResp1.code -eq 0) {
        Write-Host "✓ 老板取消订单成功 (状态: $($cancelResp1.data.status))" -ForegroundColor Green
    } else {
        Write-Host "✗ 老板取消订单失败: $($cancelResp1.msg)" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ 取消订单失败: $_" -ForegroundColor Red
}
Write-Host ""

# 测试场景2
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "测试场景2: 接单后只有陪玩能取消" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "[5] 创建第二个订单..." -ForegroundColor Yellow
try {
    $orderResp2 = Invoke-Api -Method Post -Path "/api/order" `
        -Headers @{Authorization = "Bearer $BOSS_TOKEN"} `
        -Body @{
            companionId = [int64]$COMPANION_ID
            gameName = "王者荣耀"
            gameMode = "排位赛"
            durationMinutes = 60
        }
    $ORDER_ID2 = $orderResp2.data.id
    Write-Host "✓ 订单创建成功 (ID: $ORDER_ID2)" -ForegroundColor Green
} catch {
    Write-Host "✗ 创建订单失败: $_" -ForegroundColor Red
    exit
}
Write-Host ""

Start-Sleep -Seconds 2

Write-Host "[6] 陪玩接单..." -ForegroundColor Yellow
try {
    $acceptResp = Invoke-Api -Method Post -Path "/api/order/accept" `
        -Headers @{Authorization = "Bearer $COMPANION_TOKEN"} `
        -Body @{
            orderId = [int64]$ORDER_ID2
        }
    if ($acceptResp.code -eq 0) {
        Write-Host "✓ 接单成功 (状态: $($acceptResp.data.status))" -ForegroundColor Green
    } else {
        Write-Host "✗ 接单失败: $($acceptResp.msg)" -ForegroundColor Red
        exit
    }
} catch {
    Write-Host "✗ 接单失败: $_" -ForegroundColor Red
    exit
}
Write-Host ""

Write-Host "[7] 老板尝试取消已接单订单（应该失败）..." -ForegroundColor Yellow
try {
    $cancelResp2 = Invoke-Api -Method Post -Path "/api/order/cancel" `
        -Headers @{Authorization = "Bearer $BOSS_TOKEN"} `
        -Body @{
            orderId = [int64]$ORDER_ID2
            reason = "改变主意"
        }
    Write-Host "响应: $($cancelResp2 | ConvertTo-Json)" -ForegroundColor Gray
    if ($cancelResp2.code -ne 0) {
        Write-Host "✓ 正确拒绝：接单后老板不能取消订单" -ForegroundColor Green
        Write-Host "错误信息: $($cancelResp2.msg)" -ForegroundColor Cyan
    } else {
        Write-Host "✗ 错误：接单后老板应该不能取消订单" -ForegroundColor Red
    }
} catch {
    Write-Host "✓ 正确拒绝：接单后老板不能取消订单" -ForegroundColor Green
}
Write-Host ""

Write-Host "[8] 陪玩取消已接单订单（应该成功）..." -ForegroundColor Yellow
try {
    $cancelResp3 = Invoke-Api -Method Post -Path "/api/order/cancel" `
        -Headers @{Authorization = "Bearer $COMPANION_TOKEN"} `
        -Body @{
            orderId = [int64]$ORDER_ID2
            reason = "临时有事"
        }
    Write-Host "响应: $($cancelResp3 | ConvertTo-Json)" -ForegroundColor Gray
    if ($cancelResp3.code -eq 0) {
        Write-Host "✓ 陪玩取消订单成功 (状态: $($cancelResp3.data.status))" -ForegroundColor Green
        Write-Host "说明: 状态为 8 表示正在退款中（CANCEL_REFUNDING）" -ForegroundColor Yellow
    } else {
        Write-Host "✗ 陪玩取消订单失败: $($cancelResp3.msg)" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ 取消订单失败: $_" -ForegroundColor Red
}
Write-Host ""

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "测试完成！" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan












