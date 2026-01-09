$BASE_URL = "http://localhost:8888"

Write-Host "Order Cancel Test" -ForegroundColor Cyan
Write-Host ""

# Step 1: Send code for boss
Write-Host "[1] Send code for boss (13800001001)..." -ForegroundColor Yellow
try {
    $code1 = Invoke-RestMethod -Uri "$BASE_URL/api/code/send" -Method Post -ContentType "application/json" -Body '{"phone":"13800001001","purpose":"register"}'
    Write-Host "Code sent: OK"
} catch {
    Write-Host "Code send: $_"
}

$codeBoss = Read-Host "Enter boss verification code"

# Step 2: Register boss
Write-Host "[2] Register boss..." -ForegroundColor Yellow
try {
    $regBoss = Invoke-RestMethod -Uri "$BASE_URL/api/user/register" -Method Post -ContentType "application/json" -Body "{\"phone\":\"13800001001\",\"code\":\"$codeBoss\",\"password\":\"Test123456\",\"nickname\":\"Boss\",\"role\":1}"
    $BOSS_TOKEN = $regBoss.data.accessToken
    $BOSS_ID = $regBoss.data.id
    Write-Host "Boss registered: ID=$BOSS_ID"
} catch {
    Write-Host "Boss register failed, try login..."
    try {
        $loginBoss = Invoke-RestMethod -Uri "$BASE_URL/api/user/login" -Method Post -ContentType "application/json" -Body '{"phone":"13800001001","password":"Test123456"}'
        $BOSS_TOKEN = $loginBoss.data.accessToken
        $BOSS_ID = $loginBoss.data.id
        Write-Host "Boss logged in: ID=$BOSS_ID"
    } catch {
        Write-Host "Boss login failed: $_"
        exit
    }
}

# Step 3: Send code for companion
Write-Host "[3] Send code for companion (13800002001)..." -ForegroundColor Yellow
try {
    $code2 = Invoke-RestMethod -Uri "$BASE_URL/api/code/send" -Method Post -ContentType "application/json" -Body '{"phone":"13800002001","purpose":"register"}'
    Write-Host "Code sent: OK"
} catch {
    Write-Host "Code send: $_"
}

$codeComp = Read-Host "Enter companion verification code"

# Step 4: Register companion
Write-Host "[4] Register companion..." -ForegroundColor Yellow
try {
    $regComp = Invoke-RestMethod -Uri "$BASE_URL/api/user/register" -Method Post -ContentType "application/json" -Body "{\"phone\":\"13800002001\",\"code\":\"$codeComp\",\"password\":\"Test123456\",\"nickname\":\"Companion\",\"role\":2}"
    $COMP_TOKEN = $regComp.data.accessToken
    $COMP_ID = $regComp.data.id
    Write-Host "Companion registered: ID=$COMP_ID"
} catch {
    Write-Host "Companion register failed, try login..."
    try {
        $loginComp = Invoke-RestMethod -Uri "$BASE_URL/api/user/login" -Method Post -ContentType "application/json" -Body '{"phone":"13800002001","password":"Test123456"}'
        $COMP_TOKEN = $loginComp.data.accessToken
        $COMP_ID = $loginComp.data.id
        Write-Host "Companion logged in: ID=$COMP_ID"
    } catch {
        Write-Host "Companion login failed: $_"
        exit
    }
}

# Step 5: Set companion profile
Write-Host "[5] Set companion profile..." -ForegroundColor Yellow
try {
    $headers = @{Authorization = "Bearer $COMP_TOKEN"}
    $profile = Invoke-RestMethod -Uri "$BASE_URL/api/user/companion/profile" -Method Put -ContentType "application/json" -Headers $headers -Body '{"gameSkills":"[\"Game1\"]","pricePerHour":100,"status":1}'
    Write-Host "Profile set: OK"
} catch {
    Write-Host "Profile set failed (may already exist): $_"
}

# Step 6: Create order
Write-Host "[6] Create order..." -ForegroundColor Yellow
$headersBoss = @{Authorization = "Bearer $BOSS_TOKEN"}
try {
    $order = Invoke-RestMethod -Uri "$BASE_URL/api/order" -Method Post -ContentType "application/json" -Headers $headersBoss -Body "{\"companionId\":$COMP_ID,\"gameName\":\"Game1\",\"gameMode\":\"Mode1\",\"durationMinutes\":60}"
    $ORDER_ID = $order.data.id
    $ORDER_STATUS = $order.data.status
    Write-Host "Order created: ID=$ORDER_ID, Status=$ORDER_STATUS"
} catch {
    Write-Host "Create order failed: $_"
    exit
}

# Test 1: Boss cancel before accept
Write-Host ""
Write-Host "=== Test 1: Boss cancel before accept ===" -ForegroundColor Cyan
Write-Host "[7] Boss cancel order..." -ForegroundColor Yellow
try {
    $cancel1 = Invoke-RestMethod -Uri "$BASE_URL/api/order/cancel" -Method Post -ContentType "application/json" -Headers $headersBoss -Body "{\"orderId\":$ORDER_ID,\"reason\":\"Test cancel\"}"
    if ($cancel1.code -eq 0) {
        Write-Host "OK: Boss canceled order (Status=$($cancel1.data.status))" -ForegroundColor Green
    } else {
        Write-Host "FAIL: $($cancel1.msg)" -ForegroundColor Red
    }
    Write-Host ($cancel1 | ConvertTo-Json)
} catch {
    Write-Host "FAIL: $_" -ForegroundColor Red
}

# Test 2: After accept
Write-Host ""
Write-Host "=== Test 2: After accept, only companion can cancel ===" -ForegroundColor Cyan
Write-Host "[8] Create 2nd order..." -ForegroundColor Yellow
try {
    $order2 = Invoke-RestMethod -Uri "$BASE_URL/api/order" -Method Post -ContentType "application/json" -Headers $headersBoss -Body "{\"companionId\":$COMP_ID,\"gameName\":\"Game2\",\"durationMinutes\":60}"
    $ORDER_ID2 = $order2.data.id
    Write-Host "Order created: ID=$ORDER_ID2"
} catch {
    Write-Host "Create order failed: $_"
    exit
}

Start-Sleep -Seconds 2

Write-Host "[9] Companion accept order..." -ForegroundColor Yellow
$headersComp = @{Authorization = "Bearer $COMP_TOKEN"}
try {
    $accept = Invoke-RestMethod -Uri "$BASE_URL/api/order/accept" -Method Post -ContentType "application/json" -Headers $headersComp -Body "{\"orderId\":$ORDER_ID2}"
    if ($accept.code -eq 0) {
        Write-Host "OK: Order accepted (Status=$($accept.data.status))" -ForegroundColor Green
    } else {
        Write-Host "FAIL: $($accept.msg)" -ForegroundColor Red
        exit
    }
} catch {
    Write-Host "FAIL: $_" -ForegroundColor Red
    exit
}

Write-Host "[10] Boss try cancel (should fail)..." -ForegroundColor Yellow
try {
    $cancel2 = Invoke-RestMethod -Uri "$BASE_URL/api/order/cancel" -Method Post -ContentType "application/json" -Headers $headersBoss -Body "{\"orderId\":$ORDER_ID2,\"reason\":\"Test\"}"
    if ($cancel2.code -ne 0) {
        Write-Host "OK: Correctly rejected - Boss cannot cancel after accept" -ForegroundColor Green
        Write-Host "Error: $($cancel2.msg)"
    } else {
        Write-Host "FAIL: Boss should not be able to cancel" -ForegroundColor Red
    }
} catch {
    Write-Host "OK: Correctly rejected" -ForegroundColor Green
}

Write-Host "[11] Companion cancel (should success)..." -ForegroundColor Yellow
try {
    $cancel3 = Invoke-RestMethod -Uri "$BASE_URL/api/order/cancel" -Method Post -ContentType "application/json" -Headers $headersComp -Body "{\"orderId\":$ORDER_ID2,\"reason\":\"Test\"}"
    if ($cancel3.code -eq 0) {
        Write-Host "OK: Companion canceled (Status=$($cancel3.data.status))" -ForegroundColor Green
    } else {
        Write-Host "FAIL: $($cancel3.msg)" -ForegroundColor Red
    }
    Write-Host ($cancel3 | ConvertTo-Json)
} catch {
    Write-Host "FAIL: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== Test Complete ===" -ForegroundColor Cyan

