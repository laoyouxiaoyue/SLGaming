# 订单取消功能测试脚本
$BASE_URL = "http://localhost:8888"
$PHONE_BOSS = "13800001001"
$PHONE_COMPANION = "13800002001"
$PASSWORD = "Test123456"

Write-Host "=== 订单取消功能测试 ===" -ForegroundColor Cyan
Write-Host ""

# 1. 老板登录
Write-Host "[1] 老板登录..." -ForegroundColor Yellow
$bossLoginResp = curl.exe -s -X POST "$BASE_URL/api/user/login" `
    -H "Content-Type: application/json" `
    -d "{\"phone\":\"$PHONE_BOSS\",\"password\":\"$PASSWORD\"}"

$bossLoginObj = $bossLoginResp | ConvertFrom-Json
if ($bossLoginObj.code -ne 0) {
    Write-Host "老板登录失败: $($bossLoginObj.msg)" -ForegroundColor Red
    Write-Host "请先注册用户或检查账号密码" -ForegroundColor Yellow
    exit
}

$BOSS_TOKEN = $bossLoginObj.data.accessToken
$BOSS_ID = $bossLoginObj.data.id
Write-Host "✓ 老板登录成功 (ID: $BOSS_ID)" -ForegroundColor Green
Write-Host ""

# 2. 陪玩登录
Write-Host "[2] 陪玩登录..." -ForegroundColor Yellow
$companionLoginResp = curl.exe -s -X POST "$BASE_URL/api/user/login" `
    -H "Content-Type: application/json" `
    -d "{\"phone\":\"$PHONE_COMPANION\",\"password\":\"$PASSWORD\"}"

$companionLoginObj = $companionLoginResp | ConvertFrom-Json
if ($companionLoginObj.code -ne 0) {
    Write-Host "陪玩登录失败: $($companionLoginObj.msg)" -ForegroundColor Red
    Write-Host "请先注册用户或检查账号密码" -ForegroundColor Yellow
    exit
}

$COMPANION_TOKEN = $companionLoginObj.data.accessToken
$COMPANION_ID = $companionLoginObj.data.id
Write-Host "✓ 陪玩登录成功 (ID: $COMPANION_ID)" -ForegroundColor Green
Write-Host ""

# 3. 创建订单
Write-Host "[3] 创建订单..." -ForegroundColor Yellow
$createOrderResp = curl.exe -s -X POST "$BASE_URL/api/order" `
    -H "Content-Type: application/json" `
    -H "Authorization: Bearer $BOSS_TOKEN" `
    -d "{\"companionId\":$COMPANION_ID,\"gameName\":\"王者荣耀\",\"gameMode\":\"排位赛\",\"durationMinutes\":60}"

$createOrderObj = $createOrderResp | ConvertFrom-Json
if ($createOrderObj.code -ne 0) {
    Write-Host "创建订单失败: $($createOrderObj.msg)" -ForegroundColor Red
    Write-Host $createOrderResp
    exit
}

$ORDER_ID = $createOrderObj.data.id
$ORDER_STATUS = $createOrderObj.data.status
Write-Host "✓ 订单创建成功 (ID: $ORDER_ID, 状态: $ORDER_STATUS)" -ForegroundColor Green
Write-Host ""

# 测试场景1: 老板取消未接单订单
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "测试场景1: 老板取消未接单订单" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "[4] 老板取消订单..." -ForegroundColor Yellow
$cancelResp1 = curl.exe -s -X POST "$BASE_URL/api/order/cancel" `
    -H "Content-Type: application/json" `
    -H "Authorization: Bearer $BOSS_TOKEN" `
    -d "{\"orderId\":$ORDER_ID,\"reason\":\"临时有事\"}"

$cancelObj1 = $cancelResp1 | ConvertFrom-Json
Write-Host "响应: $cancelResp1" -ForegroundColor Gray
if ($cancelObj1.code -eq 0) {
    Write-Host "✓ 老板取消订单成功 (状态: $($cancelObj1.data.status))" -ForegroundColor Green
} else {
    Write-Host "✗ 老板取消订单失败: $($cancelObj1.msg)" -ForegroundColor Red
}
Write-Host ""

# 测试场景2: 接单后权限测试
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "测试场景2: 接单后只有陪玩能取消" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "[5] 创建第二个订单..." -ForegroundColor Yellow
$createOrderResp2 = curl.exe -s -X POST "$BASE_URL/api/order" `
    -H "Content-Type: application/json" `
    -H "Authorization: Bearer $BOSS_TOKEN" `
    -d "{\"companionId\":$COMPANION_ID,\"gameName\":\"王者荣耀\",\"gameMode\":\"排位赛\",\"durationMinutes\":60}"

$createOrderObj2 = $createOrderResp2 | ConvertFrom-Json
$ORDER_ID2 = $createOrderObj2.data.id
Write-Host "✓ 订单创建成功 (ID: $ORDER_ID2)" -ForegroundColor Green
Write-Host ""

Start-Sleep -Seconds 2

Write-Host "[6] 陪玩接单..." -ForegroundColor Yellow
$acceptResp = curl.exe -s -X POST "$BASE_URL/api/order/accept" `
    -H "Content-Type: application/json" `
    -H "Authorization: Bearer $COMPANION_TOKEN" `
    -d "{\"orderId\":$ORDER_ID2}"

$acceptObj = $acceptResp | ConvertFrom-Json
if ($acceptObj.code -eq 0) {
    Write-Host "✓ 接单成功 (状态: $($acceptObj.data.status))" -ForegroundColor Green
} else {
    Write-Host "✗ 接单失败: $($acceptObj.msg)" -ForegroundColor Red
    Write-Host $acceptResp
    exit
}
Write-Host ""

Write-Host "[7] 老板尝试取消已接单订单（应该失败）..." -ForegroundColor Yellow
$cancelResp2 = curl.exe -s -X POST "$BASE_URL/api/order/cancel" `
    -H "Content-Type: application/json" `
    -H "Authorization: Bearer $BOSS_TOKEN" `
    -d "{\"orderId\":$ORDER_ID2,\"reason\":\"改变主意\"}"

$cancelObj2 = $cancelResp2 | ConvertFrom-Json
Write-Host "响应: $cancelResp2" -ForegroundColor Gray
if ($cancelObj2.code -ne 0) {
    Write-Host "✓ 正确拒绝：接单后老板不能取消订单" -ForegroundColor Green
    Write-Host "错误信息: $($cancelObj2.msg)" -ForegroundColor Cyan
} else {
    Write-Host "✗ 错误：接单后老板应该不能取消订单" -ForegroundColor Red
}
Write-Host ""

Write-Host "[8] 陪玩取消已接单订单（应该成功）..." -ForegroundColor Yellow
$cancelResp3 = curl.exe -s -X POST "$BASE_URL/api/order/cancel" `
    -H "Content-Type: application/json" `
    -H "Authorization: Bearer $COMPANION_TOKEN" `
    -d "{\"orderId\":$ORDER_ID2,\"reason\":\"临时有事\"}"

$cancelObj3 = $cancelResp3 | ConvertFrom-Json
Write-Host "响应: $cancelResp3" -ForegroundColor Gray
if ($cancelObj3.code -eq 0) {
    Write-Host "✓ 陪玩取消订单成功 (状态: $($cancelObj3.data.status))" -ForegroundColor Green
    Write-Host "说明: 状态为 8 表示正在退款中（CANCEL_REFUNDING）" -ForegroundColor Yellow
} else {
    Write-Host "✗ 陪玩取消订单失败: $($cancelObj3.msg)" -ForegroundColor Red
}
Write-Host ""

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "测试完成！" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan












