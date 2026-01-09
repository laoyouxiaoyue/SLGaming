# Order Cancel Test Script
$BASE_URL = "http://localhost:8888"
$PHONE_BOSS = "13800001001"
$PHONE_COMPANION = "13800002001"
$PASSWORD = "Test123456"

Write-Host "=== Order Cancel Test ===" -ForegroundColor Cyan
Write-Host ""

function Invoke-Api {
    param($Method, $Path, $Headers = @{}, $Body = $null)
    $uri = "$BASE_URL$Path"
    $params = @{Uri=$uri; Method=$Method; ContentType="application/json"; Headers=$Headers; ErrorAction="Stop"}
    if ($Body) { $params.Body = ($Body | ConvertTo-Json -Compress) }
    try { return Invoke-RestMethod @params } catch { 
        if ($_.ErrorDetails.Message) {
            try { return ($_.ErrorDetails.Message | ConvertFrom-Json) } catch { throw }
        }
        throw
    }
}

# 1. Boss Login
Write-Host "[1] Boss Login..." -ForegroundColor Yellow
try {
    $bossResp = Invoke-Api -Method Post -Path "/api/user/login" -Body @{phone=$PHONE_BOSS; password=$PASSWORD}
    $BOSS_TOKEN = $bossResp.data.accessToken
    $BOSS_ID = $bossResp.data.id
    Write-Host "OK Boss logged in (ID: $BOSS_ID)" -ForegroundColor Green
} catch {
    Write-Host "FAIL Boss login: $_" -ForegroundColor Red
    exit
}

# 2. Companion Login
Write-Host "[2] Companion Login..." -ForegroundColor Yellow
try {
    $compResp = Invoke-Api -Method Post -Path "/api/user/login" -Body @{phone=$PHONE_COMPANION; password=$PASSWORD}
    $COMP_TOKEN = $compResp.data.accessToken
    $COMP_ID = $compResp.data.id
    Write-Host "OK Companion logged in (ID: $COMP_ID)" -ForegroundColor Green
} catch {
    Write-Host "FAIL Companion login: $_" -ForegroundColor Red
    exit
}

# 3. Create Order
Write-Host "[3] Create Order..." -ForegroundColor Yellow
try {
    $orderResp = Invoke-Api -Method Post -Path "/api/order" `
        -Headers @{Authorization="Bearer $BOSS_TOKEN"} `
        -Body @{companionId=[int64]$COMP_ID; gameName="Game1"; gameMode="Mode1"; durationMinutes=60}
    $ORDER_ID = $orderResp.data.id
    Write-Host "OK Order created (ID: $ORDER_ID, Status: $($orderResp.data.status))" -ForegroundColor Green
} catch {
    Write-Host "FAIL Create order: $_" -ForegroundColor Red
    exit
}

# Test 1: Boss cancel before accept
Write-Host ""
Write-Host "=== Test 1: Boss cancel before accept ===" -ForegroundColor Cyan
Write-Host "[4] Boss cancel order..." -ForegroundColor Yellow
try {
    $cancelResp1 = Invoke-Api -Method Post -Path "/api/order/cancel" `
        -Headers @{Authorization="Bearer $BOSS_TOKEN"} `
        -Body @{orderId=[int64]$ORDER_ID; reason="Test cancel"}
    if ($cancelResp1.code -eq 0) {
        Write-Host "OK Boss canceled order (Status: $($cancelResp1.data.status))" -ForegroundColor Green
    } else {
        Write-Host "FAIL Boss cancel failed: $($cancelResp1.msg)" -ForegroundColor Red
    }
    Write-Host ($cancelResp1 | ConvertTo-Json)
} catch {
    Write-Host "FAIL Cancel order: $_" -ForegroundColor Red
}

# Test 2: After accept
Write-Host ""
Write-Host "=== Test 2: After accept, only companion can cancel ===" -ForegroundColor Cyan
Write-Host "[5] Create 2nd order..." -ForegroundColor Yellow
try {
    $orderResp2 = Invoke-Api -Method Post -Path "/api/order" `
        -Headers @{Authorization="Bearer $BOSS_TOKEN"} `
        -Body @{companionId=[int64]$COMP_ID; gameName="Game2"; durationMinutes=60}
    $ORDER_ID2 = $orderResp2.data.id
    Write-Host "OK Order created (ID: $ORDER_ID2)" -ForegroundColor Green
} catch {
    Write-Host "FAIL Create order: $_" -ForegroundColor Red
    exit
}

Start-Sleep -Seconds 2

Write-Host "[6] Companion accept order..." -ForegroundColor Yellow
try {
    $acceptResp = Invoke-Api -Method Post -Path "/api/order/accept" `
        -Headers @{Authorization="Bearer $COMP_TOKEN"} `
        -Body @{orderId=[int64]$ORDER_ID2}
    if ($acceptResp.code -eq 0) {
        Write-Host "OK Order accepted (Status: $($acceptResp.data.status))" -ForegroundColor Green
    } else {
        Write-Host "FAIL Accept failed: $($acceptResp.msg)" -ForegroundColor Red
        exit
    }
} catch {
    Write-Host "FAIL Accept: $_" -ForegroundColor Red
    exit
}

Write-Host "[7] Boss try cancel (should fail)..." -ForegroundColor Yellow
try {
    $cancelResp2 = Invoke-Api -Method Post -Path "/api/order/cancel" `
        -Headers @{Authorization="Bearer $BOSS_TOKEN"} `
        -Body @{orderId=[int64]$ORDER_ID2; reason="Test"}
    if ($cancelResp2.code -ne 0) {
        Write-Host "OK Correctly rejected: Boss cannot cancel after accept" -ForegroundColor Green
        Write-Host "Error: $($cancelResp2.msg)" -ForegroundColor Cyan
    } else {
        Write-Host "FAIL Boss should not be able to cancel" -ForegroundColor Red
    }
} catch {
    Write-Host "OK Correctly rejected" -ForegroundColor Green
}

Write-Host "[8] Companion cancel (should success)..." -ForegroundColor Yellow
try {
    $cancelResp3 = Invoke-Api -Method Post -Path "/api/order/cancel" `
        -Headers @{Authorization="Bearer $COMP_TOKEN"} `
        -Body @{orderId=[int64]$ORDER_ID2; reason="Test"}
    if ($cancelResp3.code -eq 0) {
        Write-Host "OK Companion canceled (Status: $($cancelResp3.data.status))" -ForegroundColor Green
    } else {
        Write-Host "FAIL Companion cancel: $($cancelResp3.msg)" -ForegroundColor Red
    }
} catch {
    Write-Host "FAIL Cancel: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== Test Complete ===" -ForegroundColor Cyan












