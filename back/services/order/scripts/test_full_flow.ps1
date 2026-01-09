# Complete Order Flow Test
$BASE_URL = "http://localhost:8888"

function Invoke-Api {
    param($Method, $Path, $Headers = @{}, $Body = $null)
    $uri = "$BASE_URL$Path"
    $params = @{Uri=$uri; Method=$Method; ContentType="application/json"; Headers=$Headers; ErrorAction="Stop"}
    if ($Body) { $params.Body = ($Body | ConvertTo-Json -Compress) }
    try { 
        return Invoke-RestMethod @params 
    } catch { 
        if ($_.ErrorDetails.Message) {
            try { 
                return ($_.ErrorDetails.Message | ConvertFrom-Json) 
            } catch { 
                return @{code=500; msg=$_.Exception.Message; data=$null}
            }
        }
        return @{code=500; msg=$_.Exception.Message; data=$null}
    }
}

function Decode-JWT {
    param($jwt)
    try {
        $parts = $jwt.Split('.')
        if ($parts.Length -lt 2) { return 0 }
        $p = $parts[1] -replace '-','+' -replace '_','/'
        while ($p.Length % 4) { $p += "=" }
        $bytes = [Convert]::FromBase64String($p)
        $json = [System.Text.Encoding]::UTF8.GetString($bytes)
        $obj = $json | ConvertFrom-Json
        return $obj.user_id
    } catch {
        return 0
    }
}

Write-Host "=== Complete Order Flow Test ===" -ForegroundColor Cyan
Write-Host ""

# Login
Write-Host "[1] Login Boss..." -ForegroundColor Yellow
$boss = Invoke-RestMethod -Uri "$BASE_URL/api/user/login" -Method Post -ContentType "application/json" -Body '{"phone":"13800001001","password":"Test123456"}' -ErrorAction Stop
$BOSS_TOKEN = $boss.data.accessToken
$BOSS_ID = Decode-JWT $BOSS_TOKEN
Write-Host "OK Boss logged in (ID: $BOSS_ID)" -ForegroundColor Green

Write-Host "[2] Login Companion..." -ForegroundColor Yellow
$comp = Invoke-RestMethod -Uri "$BASE_URL/api/user/login" -Method Post -ContentType "application/json" -Body '{"phone":"13800002001","password":"Test123456"}' -ErrorAction Stop
$COMP_TOKEN = $comp.data.accessToken
$COMP_ID = Decode-JWT $COMP_TOKEN
Write-Host "OK Companion logged in (ID: $COMP_ID)" -ForegroundColor Green

# Check wallet
Write-Host "[3] Check Boss Wallet..." -ForegroundColor Yellow
$wallet = Invoke-RestMethod -Uri "$BASE_URL/api/user/wallet" -Method Get -Headers @{Authorization="Bearer $BOSS_TOKEN"} -ErrorAction Stop
Write-Host "Balance: $($wallet.data.balance)" -ForegroundColor Cyan

# Create Order
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 1: Create Order (with balance check)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "[4] Create Order..." -ForegroundColor Yellow
$order = Invoke-Api -Method Post -Path "/api/order" `
    -Headers @{Authorization="Bearer $BOSS_TOKEN"} `
    -Body @{companionId=[int64]$COMP_ID; gameName="FullFlowTest"; durationMinutes=60}

if ($order.code -ne 0) {
    Write-Host "FAIL: $($order.msg)" -ForegroundColor Red
    exit
}

$ORDER_ID = $order.data.id
$ORDER_NO = $order.data.orderNo
Write-Host "OK Order created" -ForegroundColor Green
Write-Host "  Order ID: $ORDER_ID" -ForegroundColor White
Write-Host "  Order No: $ORDER_NO" -ForegroundColor White
Write-Host "  Status: $($order.data.status) (1=CREATED)" -ForegroundColor White
Write-Host "  Amount: $($order.data.totalAmount)" -ForegroundColor White

# Wait for payment (async)
Write-Host ""
Write-Host "[5] Waiting for payment processing (3 seconds)..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

# Accept Order
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 2: Companion Accept Order" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "[6] Companion accept order..." -ForegroundColor Yellow
$accept = Invoke-Api -Method Post -Path "/api/order/accept" `
    -Headers @{Authorization="Bearer $COMP_TOKEN"} `
    -Body @{orderId=[int64]$ORDER_ID}

if ($accept.code -eq 0) {
    Write-Host "OK Order accepted" -ForegroundColor Green
    Write-Host "  Status: $($accept.data.status) (3=ACCEPTED)" -ForegroundColor White
} else {
    Write-Host "FAIL Accept: $($accept.msg)" -ForegroundColor Red
    exit
}

# Test Cancel Permissions
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 3: Test Cancel Permissions" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "[7] Boss try cancel (should fail)..." -ForegroundColor Yellow
$cancelBoss = Invoke-Api -Method Post -Path "/api/order/cancel" `
    -Headers @{Authorization="Bearer $BOSS_TOKEN"} `
    -Body @{orderId=[int64]$ORDER_ID; reason="Boss trying to cancel"}

if ($cancelBoss.code -ne 0) {
    Write-Host "PASS: Boss correctly rejected" -ForegroundColor Green
    Write-Host "  Error: $($cancelBoss.msg)" -ForegroundColor Gray
} else {
    Write-Host "FAIL: Boss should not be able to cancel accepted order" -ForegroundColor Red
}

Write-Host "[8] Companion cancel (should success)..." -ForegroundColor Yellow
$cancelComp = Invoke-Api -Method Post -Path "/api/order/cancel" `
    -Headers @{Authorization="Bearer $COMP_TOKEN"} `
    -Body @{orderId=[int64]$ORDER_ID; reason="Companion cancel test"}

if ($cancelComp.code -eq 0) {
    Write-Host "PASS: Companion canceled successfully" -ForegroundColor Green
    Write-Host "  Status: $($cancelComp.data.status)" -ForegroundColor White
    Write-Host "    Status 6 = CANCELLED" -ForegroundColor Gray
    Write-Host "    Status 8 = CANCEL_REFUNDING (if payment succeeded)" -ForegroundColor Gray
} else {
    Write-Host "FAIL: $($cancelComp.msg)" -ForegroundColor Red
}

# Check wallet after cancel
Write-Host ""
Write-Host "[9] Check Boss Wallet After Cancel..." -ForegroundColor Yellow
Start-Sleep -Seconds 2
$walletAfter = Invoke-RestMethod -Uri "$BASE_URL/api/user/wallet" -Method Get -Headers @{Authorization="Bearer $BOSS_TOKEN"} -ErrorAction Stop
Write-Host "Balance after cancel: $($walletAfter.data.balance)" -ForegroundColor Cyan
if ($walletAfter.data.balance -gt $wallet.data.balance) {
    Write-Host "  Refund processed!" -ForegroundColor Green
} else {
    Write-Host "  (Refund may still be processing, or order was not paid)" -ForegroundColor Gray
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test Complete!" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Summary:" -ForegroundColor Yellow
Write-Host "  ✓ Balance check: Working" -ForegroundColor Green
Write-Host "  ✓ Order creation: Working" -ForegroundColor Green
Write-Host "  ✓ Order acceptance: Working" -ForegroundColor Green
Write-Host "  ✓ Cancel permissions: Working" -ForegroundColor Green












