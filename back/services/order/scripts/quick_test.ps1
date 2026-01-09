# 快速测试脚本 - 订单取消功能
$BASE_URL = "http://localhost:8888"
$PHONE_BOSS = "13800001001"
$PHONE_COMPANION = "13800002001"
$PASSWORD = "Test123456"

Write-Host "=== 订单取消功能测试 ===" -ForegroundColor Cyan
Write-Host ""

# 步骤 1: 老板登录（假设已注册）
Write-Host "1. 老板登录..." -ForegroundColor Yellow
$bossLogin = @{
    phone = $PHONE_BOSS
    password = $PASSWORD
} | ConvertTo-Json

try {
    $bossResp = Invoke-RestMethod -Uri "$BASE_URL/api/user/login" -Method Post -ContentType "application/json" -Body $bossLogin
    $BOSS_TOKEN = $bossResp.data.accessToken
    $BOSS_ID = $bossResp.data.id
    Write-Host "✓ 老板登录成功 (ID: $BOSS_ID)" -ForegroundColor Green
} catch {
    Write-Host "✗ 老板登录失败，请先注册" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit
}

# 步骤 2: 陪玩登录
Write-Host "2. 陪玩登录..." -ForegroundColor Yellow
$companionLogin = @{
    phone = $PHONE_COMPANION
    password = $PASSWORD
} | ConvertTo-Json

try {
    $companionResp = Invoke-RestMethod -Uri "$BASE_URL/api/user/login" -Method Post -ContentType "application/json" -Body $companionLogin
    $COMPANION_TOKEN = $companionResp.data.accessToken
    $COMPANION_ID = $companionResp.data.id
    Write-Host "✓ 陪玩登录成功 (ID: $COMPANION_ID)" -ForegroundColor Green
} catch {
    Write-Host "✗ 陪玩登录失败，请先注册" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit
}

# 步骤 3: 创建订单
Write-Host ""
Write-Host "3. 创建订单..." -ForegroundColor Yellow
$createOrder = @{
    companionId = [int64]$COMPANION_ID
    gameName = "王者荣耀"
    gameMode = "排位赛"
    durationMinutes = 60
} | ConvertTo-Json

try {
    $headers = @{ Authorization = "Bearer $BOSS_TOKEN" }
    $orderResp = Invoke-RestMethod -Uri "$BASE_URL/api/order" -Method Post -ContentType "application/json" -Headers $headers -Body $createOrder
    $ORDER_ID = $orderResp.data.id
    $ORDER_STATUS = $orderResp.data.status
    Write-Host "✓ 订单创建成功 (ID: $ORDER_ID, 状态: $ORDER_STATUS)" -ForegroundColor Green
} catch {
    Write-Host "✗ 创建订单失败" -ForegroundColor Red
    Write-Host $_.Exception.Message
    if ($_.ErrorDetails.Message) {
        Write-Host $_.ErrorDetails.Message
    }
    exit
}

# 步骤 4: 测试场景1 - 老板取消订单（应该成功）
Write-Host ""
Write-Host "=== 测试场景1: 老板取消未接单订单 ===" -ForegroundColor Cyan
$cancelOrder1 = @{
    orderId = [int64]$ORDER_ID
    reason = "不想玩了"
} | ConvertTo-Json

try {
    $cancelResp = Invoke-RestMethod -Uri "$BASE_URL/api/order/cancel" -Method Post -ContentType "application/json" -Headers $headers -Body $cancelOrder1
    if ($cancelResp.code -eq 0) {
        Write-Host "✓ 老板取消订单成功 (状态: $($cancelResp.data.status))" -ForegroundColor Green
    } else {
        Write-Host "✗ 取消失败: $($cancelResp.msg)" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ 取消订单失败" -ForegroundColor Red
    Write-Host $_.ErrorDetails.Message
}

# 步骤 5: 创建第二个订单用于测试场景2
Write-Host ""
Write-Host "4. 创建第二个订单..." -ForegroundColor Yellow
try {
    $orderResp2 = Invoke-RestMethod -Uri "$BASE_URL/api/order" -Method Post -ContentType "application/json" -Headers $headers -Body $createOrder
    $ORDER_ID2 = $orderResp2.data.id
    Write-Host "✓ 订单创建成功 (ID: $ORDER_ID2)" -ForegroundColor Green
} catch {
    Write-Host "✗ 创建订单失败" -ForegroundColor Red
    exit
}

Start-Sleep -Seconds 1

# 步骤 6: 陪玩接单
Write-Host ""
Write-Host "5. 陪玩接单..." -ForegroundColor Yellow
$acceptOrder = @{
    orderId = [int64]$ORDER_ID2
} | ConvertTo-Json

$companionHeaders = @{ Authorization = "Bearer $COMPANION_TOKEN" }
try {
    $acceptResp = Invoke-RestMethod -Uri "$BASE_URL/api/order/accept" -Method Post -ContentType "application/json" -Headers $companionHeaders -Body $acceptOrder
    if ($acceptResp.code -eq 0) {
        Write-Host "✓ 接单成功 (状态: $($acceptResp.data.status))" -ForegroundColor Green
    } else {
        Write-Host "✗ 接单失败: $($acceptResp.msg)" -ForegroundColor Red
        exit
    }
} catch {
    Write-Host "✗ 接单失败" -ForegroundColor Red
    Write-Host $_.ErrorDetails.Message
    exit
}

# 步骤 7: 测试场景2 - 老板尝试取消已接单订单（应该失败）
Write-Host ""
Write-Host "=== 测试场景2: 接单后老板取消订单 ===" -ForegroundColor Cyan
$cancelOrder2 = @{
    orderId = [int64]$ORDER_ID2
    reason = "改变主意"
} | ConvertTo-Json

try {
    $cancelResp2 = Invoke-RestMethod -Uri "$BASE_URL/api/order/cancel" -Method Post -ContentType "application/json" -Headers $headers -Body $cancelOrder2
    if ($cancelResp2.code -eq 0) {
        Write-Host "✗ 错误：老板应该不能取消已接单的订单" -ForegroundColor Red
    } else {
        Write-Host "✓ 正确拒绝：接单后老板不能取消 ($($cancelResp2.msg))" -ForegroundColor Green
    }
} catch {
    Write-Host "✓ 正确拒绝：接单后老板不能取消" -ForegroundColor Green
}

# 步骤 8: 陪玩取消订单（应该成功）
Write-Host ""
Write-Host "6. 陪玩取消已接单订单..." -ForegroundColor Yellow
try {
    $cancelResp3 = Invoke-RestMethod -Uri "$BASE_URL/api/order/cancel" -Method Post -ContentType "application/json" -Headers $companionHeaders -Body $cancelOrder2
    if ($cancelResp3.code -eq 0) {
        Write-Host "✓ 陪玩取消订单成功 (状态: $($cancelResp3.data.status))" -ForegroundColor Green
    } else {
        Write-Host "✗ 取消失败: $($cancelResp3.msg)" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ 取消订单失败" -ForegroundColor Red
    Write-Host $_.ErrorDetails.Message
}

Write-Host ""
Write-Host "=== 测试完成 ===" -ForegroundColor Cyan












