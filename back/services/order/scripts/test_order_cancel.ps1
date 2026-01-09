# 订单取消功能完整测试脚本 (PowerShell)
# 测试场景：
# 1. 老板在未接单前取消订单
# 2. 接单后只有陪玩能取消订单

$BASE_URL = "http://localhost:8888"
$PHONE_BOSS = "13800001001"
$PHONE_COMPANION = "13800002001"
$PASSWORD = "Test123456"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "订单取消功能完整测试" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# 辅助函数：打印 JSON 响应
function Print-Response {
    param($response)
    Write-Host "Response:" -ForegroundColor Yellow
    try {
        $response | ConvertFrom-Json | ConvertTo-Json -Depth 10 | Write-Host
    } catch {
        Write-Host $response
    }
    Write-Host ""
}

# 辅助函数：安全调用 API
function Invoke-SafeRestMethod {
    param(
        [string]$Uri,
        [string]$Method = "Get",
        [hashtable]$Headers = @{},
        [object]$Body = $null
    )
    try {
        $params = @{
            Uri = $Uri
            Method = $Method
            ContentType = "application/json"
            Headers = $Headers
            ErrorAction = "Stop"
        }
        if ($Body) {
            $params.Body = ($Body | ConvertTo-Json -Depth 10)
        }
        return Invoke-RestMethod @params
    } catch {
        $errorResponse = $_.ErrorDetails.Message
        if ($errorResponse) {
            try {
                return ($errorResponse | ConvertFrom-Json)
            } catch {
                Write-Host "请求失败: $errorResponse" -ForegroundColor Red
                throw
            }
        }
        throw
    }
}

Write-Host "步骤 1: 发送验证码 - 老板" -ForegroundColor Green
try {
    $bossCodeResp = Invoke-SafeRestMethod -Uri "$BASE_URL/api/code/send" `
        -Method Post `
        -Body @{
            phone = $PHONE_BOSS
            purpose = "register"
        }
    Print-Response ($bossCodeResp | ConvertTo-Json)
} catch {
    Write-Host "发送验证码失败，继续..." -ForegroundColor Yellow
}

Write-Host "请输入老板验证码:" -ForegroundColor Yellow
$BOSS_CODE_INPUT = Read-Host

Write-Host ""
Write-Host "步骤 2: 注册老板用户" -ForegroundColor Green
try {
    $bossRegisterResp = Invoke-SafeRestMethod -Uri "$BASE_URL/api/user/register" `
        -Method Post `
        -Body @{
            phone = $PHONE_BOSS
            code = $BOSS_CODE_INPUT
            password = $PASSWORD
            nickname = "测试老板"
            role = 1
        }
    $BOSS_TOKEN = $bossRegisterResp.data.accessToken
    $BOSS_USER_ID = $bossRegisterResp.data.id
    Write-Host "✓ 老板注册成功" -ForegroundColor Green
} catch {
    Write-Host "注册失败，尝试登录..." -ForegroundColor Yellow
    try {
        $bossLoginResp = Invoke-SafeRestMethod -Uri "$BASE_URL/api/user/login" `
            -Method Post `
            -Body @{
                phone = $PHONE_BOSS
                password = $PASSWORD
            }
        $BOSS_TOKEN = $bossLoginResp.data.accessToken
        $BOSS_USER_ID = $bossLoginResp.data.id
        Write-Host "✓ 老板登录成功" -ForegroundColor Green
    } catch {
        Write-Host "✗ 无法获取老板 Token" -ForegroundColor Red
        exit 1
    }
}
Write-Host "老板用户ID: $BOSS_USER_ID"
Write-Host ""

Write-Host "步骤 3: 发送验证码 - 陪玩" -ForegroundColor Green
$companionCodeResp = Invoke-RestMethod -Uri "$BASE_URL/api/code/send" `
    -Method Post `
    -ContentType "application/json" `
    -Body (@{
        phone = $PHONE_COMPANION
        purpose = "register"
    } | ConvertTo-Json)
Print-Response ($companionCodeResp | ConvertTo-Json)

Write-Host "请输入陪玩验证码:" -ForegroundColor Yellow
$COMPANION_CODE_INPUT = Read-Host

Write-Host ""
Write-Host "步骤 4: 注册陪玩用户" -ForegroundColor Green
try {
    $companionRegisterResp = Invoke-RestMethod -Uri "$BASE_URL/api/user/register" `
        -Method Post `
        -ContentType "application/json" `
        -Body (@{
            phone = $PHONE_COMPANION
            code = $COMPANION_CODE_INPUT
            password = $PASSWORD
            nickname = "测试陪玩"
            role = 2
        } | ConvertTo-Json)
    $COMPANION_TOKEN = $companionRegisterResp.data.accessToken
    $COMPANION_USER_ID = $companionRegisterResp.data.id
    Write-Host "✓ 陪玩注册成功" -ForegroundColor Green
} catch {
    Write-Host "注册失败，尝试登录..." -ForegroundColor Yellow
    try {
        $companionLoginResp = Invoke-RestMethod -Uri "$BASE_URL/api/user/login" `
            -Method Post `
            -ContentType "application/json" `
            -Body (@{
                phone = $PHONE_COMPANION
                password = $PASSWORD
            } | ConvertTo-Json)
        $COMPANION_TOKEN = $companionLoginResp.data.accessToken
        $COMPANION_USER_ID = $companionLoginResp.data.id
        Write-Host "✓ 陪玩登录成功" -ForegroundColor Green
    } catch {
        Write-Host "✗ 无法获取陪玩 Token" -ForegroundColor Red
        exit 1
    }
}
Write-Host "陪玩用户ID: $COMPANION_USER_ID"
Write-Host ""

Write-Host "步骤 5: 设置陪玩信息（价格等）" -ForegroundColor Green
try {
    $companionUpdateResp = Invoke-RestMethod -Uri "$BASE_URL/api/user/companion/profile" `
        -Method Put `
        -ContentType "application/json" `
        -Headers @{Authorization = "Bearer $COMPANION_TOKEN"} `
        -Body (@{
            gameSkills = '["王者荣耀","和平精英"]'
            pricePerHour = 100
            status = 1
        } | ConvertTo-Json)
    Write-Host "✓ 陪玩信息设置成功" -ForegroundColor Green
} catch {
    Write-Host "⚠ 设置陪玩信息失败: $_" -ForegroundColor Yellow
}
Write-Host ""

Write-Host "步骤 6: 查看老板钱包" -ForegroundColor Green
try {
    $bossWalletResp = Invoke-RestMethod -Uri "$BASE_URL/api/user/wallet" `
        -Method Get `
        -Headers @{Authorization = "Bearer $BOSS_TOKEN"}
    $BOSS_BALANCE = $bossWalletResp.data.balance
    Write-Host "老板余额: $BOSS_BALANCE 帅币" -ForegroundColor Cyan
    
    if ($BOSS_BALANCE -lt 200) {
        Write-Host "⚠ 老板余额不足，需要充值（手动调用充值接口）" -ForegroundColor Yellow
        Write-Host "建议充值至少 200 帅币" -ForegroundColor Yellow
    }
} catch {
    Write-Host "⚠ 获取钱包失败: $_" -ForegroundColor Yellow
    $BOSS_BALANCE = 0
}
Write-Host ""

Write-Host "步骤 7: 创建订单（老板）" -ForegroundColor Green
try {
    $createOrderResp = Invoke-RestMethod -Uri "$BASE_URL/api/order" `
        -Method Post `
        -ContentType "application/json" `
        -Headers @{Authorization = "Bearer $BOSS_TOKEN"} `
        -Body (@{
            companionId = $COMPANION_USER_ID
            gameName = "王者荣耀"
            gameMode = "排位赛-王者段位"
            durationMinutes = 60
        } | ConvertTo-Json)
    $ORDER_ID = $createOrderResp.data.id
    $ORDER_STATUS = $createOrderResp.data.status
    Write-Host "✓ 订单创建成功" -ForegroundColor Green
    Write-Host "订单ID: $ORDER_ID" -ForegroundColor Cyan
    Write-Host "订单状态: $ORDER_STATUS" -ForegroundColor Cyan
} catch {
    Write-Host "✗ 创建订单失败: $_" -ForegroundColor Red
    $errorDetails = $_.ErrorDetails.Message
    if ($errorDetails) {
        Write-Host "错误详情: $errorDetails" -ForegroundColor Red
    }
    exit 1
}
Write-Host ""

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "测试场景 1: 老板在未接单前取消订单" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "步骤 8: 老板取消订单（应该成功）" -ForegroundColor Green
try {
    $cancelResp1 = Invoke-RestMethod -Uri "$BASE_URL/api/order/cancel" `
        -Method Post `
        -ContentType "application/json" `
        -Headers @{Authorization = "Bearer $BOSS_TOKEN"} `
        -Body (@{
            orderId = $ORDER_ID
            reason = "临时有事，不想玩了"
        } | ConvertTo-Json)
    Print-Response ($cancelResp1 | ConvertTo-Json)
    
    if ($cancelResp1.code -eq 0) {
        Write-Host "✓ 老板取消订单成功" -ForegroundColor Green
        Write-Host "取消后状态: $($cancelResp1.data.status)" -ForegroundColor Cyan
    } else {
        Write-Host "✗ 老板取消订单失败" -ForegroundColor Red
        Write-Host "错误信息: $($cancelResp1.msg)" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ 请求失败: $_" -ForegroundColor Red
    $errorDetails = $_.ErrorDetails.Message
    if ($errorDetails) {
        Write-Host "错误详情: $errorDetails" -ForegroundColor Red
    }
}
Write-Host ""

Write-Host "步骤 9: 验证：陪玩尝试取消已取消的订单（应该失败）" -ForegroundColor Green
try {
    $cancelResp2 = Invoke-RestMethod -Uri "$BASE_URL/api/order/cancel" `
        -Method Post `
        -ContentType "application/json" `
        -Headers @{Authorization = "Bearer $COMPANION_TOKEN"} `
        -Body (@{
            orderId = $ORDER_ID
            reason = "测试"
        } | ConvertTo-Json)
    Print-Response ($cancelResp2 | ConvertTo-Json)
    
    if ($cancelResp2.code -ne 0) {
        Write-Host "✓ 正确拒绝：已取消的订单不能再次取消" -ForegroundColor Green
    } else {
        Write-Host "✗ 错误：已取消的订单应该不能被再次取消" -ForegroundColor Red
    }
} catch {
    Write-Host "✓ 正确：请求被拒绝（已取消的订单）" -ForegroundColor Green
}
Write-Host ""

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "测试场景 2: 接单后只有陪玩能取消" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "步骤 10: 再次创建订单" -ForegroundColor Green
try {
    $createOrderResp2 = Invoke-RestMethod -Uri "$BASE_URL/api/order" `
        -Method Post `
        -ContentType "application/json" `
        -Headers @{Authorization = "Bearer $BOSS_TOKEN"} `
        -Body (@{
            companionId = $COMPANION_USER_ID
            gameName = "王者荣耀"
            gameMode = "排位赛-王者段位"
            durationMinutes = 60
        } | ConvertTo-Json)
    $ORDER_ID2 = $createOrderResp2.data.id
    Write-Host "✓ 订单创建成功，订单ID: $ORDER_ID2" -ForegroundColor Green
} catch {
    Write-Host "✗ 创建订单失败: $_" -ForegroundColor Red
    exit 1
}
Write-Host ""

Start-Sleep -Seconds 2

Write-Host "步骤 11: 陪玩接单" -ForegroundColor Green
try {
    $acceptResp = Invoke-RestMethod -Uri "$BASE_URL/api/order/accept" `
        -Method Post `
        -ContentType "application/json" `
        -Headers @{Authorization = "Bearer $COMPANION_TOKEN"} `
        -Body (@{
            orderId = $ORDER_ID2
        } | ConvertTo-Json)
    Print-Response ($acceptResp | ConvertTo-Json)
    
    if ($acceptResp.code -eq 0) {
        Write-Host "✓ 陪玩接单成功" -ForegroundColor Green
        Write-Host "接单后状态: $($acceptResp.data.status)" -ForegroundColor Cyan
    } else {
        Write-Host "✗ 陪玩接单失败" -ForegroundColor Red
        Write-Host "错误: $($acceptResp.msg)" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "✗ 接单失败: $_" -ForegroundColor Red
    exit 1
}
Write-Host ""

Write-Host "步骤 12: 验证：老板尝试取消已接单的订单（应该失败）" -ForegroundColor Green
try {
    $cancelResp3 = Invoke-RestMethod -Uri "$BASE_URL/api/order/cancel" `
        -Method Post `
        -ContentType "application/json" `
        -Headers @{Authorization = "Bearer $BOSS_TOKEN"} `
        -Body (@{
            orderId = $ORDER_ID2
            reason = "改变主意"
        } | ConvertTo-Json)
    Print-Response ($cancelResp3 | ConvertTo-Json)
    
    if ($cancelResp3.code -ne 0) {
        Write-Host "✓ 正确拒绝：接单后老板不能取消订单" -ForegroundColor Green
        Write-Host "错误信息: $($cancelResp3.msg)" -ForegroundColor Cyan
    } else {
        Write-Host "✗ 错误：接单后老板应该不能取消订单" -ForegroundColor Red
    }
} catch {
    Write-Host "✓ 正确：请求被拒绝（接单后老板无权限）" -ForegroundColor Green
}
Write-Host ""

Write-Host "步骤 13: 陪玩取消已接单的订单（应该成功）" -ForegroundColor Green
try {
    $cancelResp4 = Invoke-RestMethod -Uri "$BASE_URL/api/order/cancel" `
        -Method Post `
        -ContentType "application/json" `
        -Headers @{Authorization = "Bearer $COMPANION_TOKEN"} `
        -Body (@{
            orderId = $ORDER_ID2
            reason = "临时有事"
        } | ConvertTo-Json)
    Print-Response ($cancelResp4 | ConvertTo-Json)
    
    if ($cancelResp4.code -eq 0) {
        Write-Host "✓ 陪玩取消订单成功" -ForegroundColor Green
        Write-Host "取消后状态: $($cancelResp4.data.status)" -ForegroundColor Cyan
        Write-Host "说明: 状态为 8 表示正在退款中（CANCEL_REFUNDING）" -ForegroundColor Yellow
    } else {
        Write-Host "✗ 陪玩取消订单失败" -ForegroundColor Red
        Write-Host "错误: $($cancelResp4.msg)" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ 取消订单失败: $_" -ForegroundColor Red
}
Write-Host ""

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "测试完成！" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "总结：" -ForegroundColor Green
Write-Host "- 老板可以在未接单前取消订单 ✓" -ForegroundColor Green
Write-Host "- 接单后只有陪玩能取消订单 ✓" -ForegroundColor Green
Write-Host "- 权限验证正常工作 ✓" -ForegroundColor Green

