# Test script after users are registered
# Prerequisites: 
# 1. Boss user registered (phone: 13800001001, password: Test123456, role: 1)
# 2. Companion user registered (phone: 13800002001, password: Test123456, role: 2)
# 3. Companion profile set (pricePerHour: 100, status: 1)

$BASE_URL = "http://localhost:8888"
$PHONE_BOSS = "13800001001"
$PHONE_COMPANION = "13800002001"
$PASSWORD = "Test123456"

Write-Host "=== Order Cancel Function Test ===" -ForegroundColor Cyan
Write-Host ""

# Helper function
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
    Write-Host "Please register boss user first (phone: $PHONE_BOSS)" -ForegroundColor Yellow
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
    Write-Host "Please register companion user first (phone: $PHONE_COMPANION)" -ForegroundColor Yellow
    exit
}

# 3. Check companion profile
Write-Host "[3] Check companion profile..." -ForegroundColor Yellow
try {
    $profile = Invoke-Api -Method Get -Path "/api/user/companion/profile" -Headers @{Authorization="Bearer $COMP_TOKEN"}
    Write-Host "OK Companion profile exists (Price: $($profile.data.pricePerHour))" -ForegroundColor Green
} catch {
    Write-Host "Setting companion profile..." -ForegroundColor Yellow
    try {
        $setProfile = Invoke-Api -Method Put -Path "/api/user/companion/profile" `
            -Headers @{Authorization="Bearer $COMP_TOKEN"} `
            -Body @{gameSkills='["Game1"]'; pricePerHour=100; status=1}
        Write-Host "OK Profile set" -ForegroundColor Green
    } catch {
        Write-Host "WARN Profile set failed: $_" -ForegroundColor Yellow
    }
}

# 4. Create Order
Write-Host "[4] Create Order..." -ForegroundColor Yellow
try {
    $orderResp = Invoke-Api -Method Post -Path "/api/order" `
        -Headers @{Authorization="Bearer $BOSS_TOKEN"} `
        -Body @{companionId=[int64]$COMP_ID; gameName="Game1"; gameMode="Mode1"; durationMinutes=60}
    $ORDER_ID = $orderResp.data.id
    $ORDER_STATUS = $orderResp.data.status
    Write-Host "OK Order created (ID: $ORDER_ID, Status: $ORDER_STATUS)" -ForegroundColor Green
    Write-Host ($orderResp | ConvertTo-Json)
} catch {
    Write-Host "FAIL Create order: $_" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        $err = $_.ErrorDetails.Message | ConvertFrom-Json
        Write-Host "Error details: $($err.msg)" -ForegroundColor Red
    }
    exit
}

# Test 1: Boss cancel before accept
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test 1: Boss cancel before accept" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "[5] Boss cancel order..." -ForegroundColor Yellow
try {
    $cancelResp1 = Invoke-Api -Method Post -Path "/api/order/cancel" `
        -Headers @{Authorization="Bearer $BOSS_TOKEN"} `
        -Body @{orderId=[int64]$ORDER_ID; reason="Test cancel"}
    Write-Host "Response:" -ForegroundColor Gray
    Write-Host ($cancelResp1 | ConvertTo-Json)
    if ($cancelResp1.code -eq 0) {
        Write-Host "OK Boss canceled order (Status: $($cancelResp1.data.status))" -ForegroundColor Green
    } else {
        Write-Host "FAIL Boss cancel failed: $($cancelResp1.msg)" -ForegroundColor Red
    }
} catch {
    Write-Host "FAIL Cancel order: $_" -ForegroundColor Red
}

# Test 2: After accept
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test 2: After accept, only companion can cancel" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "[6] Create 2nd order..." -ForegroundColor Yellow
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

Write-Host "[7] Companion accept order..." -ForegroundColor Yellow
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

Write-Host "[8] Boss try cancel (should fail)..." -ForegroundColor Yellow
try {
    $cancelResp2 = Invoke-Api -Method Post -Path "/api/order/cancel" `
        -Headers @{Authorization="Bearer $BOSS_TOKEN"} `
        -Body @{orderId=[int64]$ORDER_ID2; reason="Test"}
    Write-Host "Response:" -ForegroundColor Gray
    Write-Host ($cancelResp2 | ConvertTo-Json)
    if ($cancelResp2.code -ne 0) {
        Write-Host "OK Correctly rejected: Boss cannot cancel after accept" -ForegroundColor Green
        Write-Host "Error message: $($cancelResp2.msg)" -ForegroundColor Cyan
    } else {
        Write-Host "FAIL Boss should not be able to cancel" -ForegroundColor Red
    }
} catch {
    Write-Host "OK Correctly rejected (exception)" -ForegroundColor Green
}

Write-Host "[9] Companion cancel (should success)..." -ForegroundColor Yellow
try {
    $cancelResp3 = Invoke-Api -Method Post -Path "/api/order/cancel" `
        -Headers @{Authorization="Bearer $COMP_TOKEN"} `
        -Body @{orderId=[int64]$ORDER_ID2; reason="Test"}
    Write-Host "Response:" -ForegroundColor Gray
    Write-Host ($cancelResp3 | ConvertTo-Json)
    if ($cancelResp3.code -eq 0) {
        Write-Host "OK Companion canceled (Status: $($cancelResp3.data.status))" -ForegroundColor Green
        Write-Host "Note: Status 8 = CANCEL_REFUNDING" -ForegroundColor Yellow
    } else {
        Write-Host "FAIL Companion cancel: $($cancelResp3.msg)" -ForegroundColor Red
    }
} catch {
    Write-Host "FAIL Cancel: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test Complete!" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan












