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
                $errorObj = $_.ErrorDetails.Message | ConvertFrom-Json
                return $errorObj
            } catch {
                # If JSON parse fails, create error response object
                return @{code=500; msg=$_.ErrorDetails.Message; data=$null}
            }
        }
        throw
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
        Write-Host "JWT decode error: $_" -ForegroundColor Red
        return 0
    }
}

Write-Host "=== Complete Order Flow Test ===" -ForegroundColor Cyan
Write-Host ""

# Login
Write-Host "[1] Login Boss..." -ForegroundColor Yellow
$boss = Invoke-Api -Method Post -Path "/api/user/login" -Body @{phone="13800001001"; password="Test123456"}
$BOSS_TOKEN = $boss.data.accessToken
$BOSS_ID = Decode-JWT $BOSS_TOKEN
Write-Host "OK Boss logged in (ID: $BOSS_ID)" -ForegroundColor Green

Write-Host "[2] Login Companion..." -ForegroundColor Yellow
$comp = Invoke-Api -Method Post -Path "/api/user/login" -Body @{phone="13800002001"; password="Test123456"}
$COMP_TOKEN = $comp.data.accessToken
$COMP_ID = Decode-JWT $COMP_TOKEN
Write-Host "OK Companion logged in (ID: $COMP_ID)" -ForegroundColor Green

# Check wallet
Write-Host "[3] Check Boss Wallet..." -ForegroundColor Yellow
$wallet = Invoke-Api -Method Get -Path "/api/user/wallet" -Headers @{Authorization="Bearer $BOSS_TOKEN"}
Write-Host "Balance: $($wallet.data.balance)" -ForegroundColor Cyan
if ($wallet.data.balance -lt 100) {
    Write-Host "WARNING: Balance is less than 100, may fail to create order" -ForegroundColor Yellow
}

# Check companion profile
Write-Host "[4] Check Companion Profile..." -ForegroundColor Yellow
$profile = Invoke-Api -Method Get -Path "/api/user/companion/profile" -Headers @{Authorization="Bearer $COMP_TOKEN"}
Write-Host "PricePerHour: $($profile.data.pricePerHour), Status: $($profile.data.status)" -ForegroundColor Cyan

# Create Order
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 1: Create Order" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "[5] Create Order..." -ForegroundColor Yellow
try {
    $order = Invoke-Api -Method Post -Path "/api/order" `
        -Headers @{Authorization="Bearer $BOSS_TOKEN"} `
        -Body @{companionId=[int64]$COMP_ID; gameName="Game1"; durationMinutes=60}
    
    if ($order.code -ne 0) {
        Write-Host "FAIL: $($order.msg)" -ForegroundColor Red
        Write-Host ($order | ConvertTo-Json)
        exit
    }
    
    $ORDER_ID = $order.data.id
    $ORDER_STATUS = $order.data.status
    $TOTAL_AMOUNT = $order.data.totalAmount
    
    Write-Host "OK Order created" -ForegroundColor Green
    Write-Host "  Order ID: $ORDER_ID" -ForegroundColor White
    Write-Host "  Order No: $($order.data.orderNo)" -ForegroundColor White
    Write-Host "  Status: $ORDER_STATUS (1=CREATED)" -ForegroundColor White
    Write-Host "  Total Amount: $TOTAL_AMOUNT" -ForegroundColor White
} catch {
    Write-Host "FAIL: $_" -ForegroundColor Red
    exit
}

# Wait for payment processing (async)
Write-Host ""
Write-Host "[6] Waiting for payment processing (5 seconds)..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Check order status after payment
Write-Host "[7] Check Order Status..." -ForegroundColor Yellow
$orderCheck = Invoke-Api -Method Get -Path "/api/order/$ORDER_ID" -Headers @{Authorization="Bearer $BOSS_TOKEN"}
Write-Host "Order Status: $($orderCheck.data.status)" -ForegroundColor Cyan
Write-Host "  Status 1 = CREATED (payment pending)" -ForegroundColor Gray
Write-Host "  Status 2 = PAID (payment succeeded)" -ForegroundColor Gray
Write-Host "  Status 6 = CANCELLED (payment failed)" -ForegroundColor Gray

$CURRENT_STATUS = $orderCheck.data.status

if ($CURRENT_STATUS -eq 2) {
    Write-Host "Payment succeeded!" -ForegroundColor Green
} elseif ($CURRENT_STATUS -eq 6) {
    Write-Host "Payment failed! Cannot continue." -ForegroundColor Red
    exit
} else {
    Write-Host "Payment still pending or processing..." -ForegroundColor Yellow
    Write-Host "Waiting additional 3 seconds..." -ForegroundColor Yellow
    Start-Sleep -Seconds 3
    $orderCheck = Invoke-Api -Method Get -Path "/api/order/$ORDER_ID" -Headers @{Authorization="Bearer $BOSS_TOKEN"}
    $CURRENT_STATUS = $orderCheck.data.status
    Write-Host "Order Status after wait: $CURRENT_STATUS" -ForegroundColor Cyan
    
    if ($CURRENT_STATUS -ne 2) {
        Write-Host "Payment may have failed or still pending. Status: $CURRENT_STATUS" -ForegroundColor Yellow
        Write-Host "Continuing with acceptance test anyway..." -ForegroundColor Yellow
    }
}

# Accept Order
if ($CURRENT_STATUS -eq 2 -or $CURRENT_STATUS -eq 1) {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "Step 2: Companion Accept Order" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "[8] Companion accept order..." -ForegroundColor Yellow
    try {
        $accept = Invoke-Api -Method Post -Path "/api/order/accept" `
            -Headers @{Authorization="Bearer $COMP_TOKEN"} `
            -Body @{orderId=[int64]$ORDER_ID}
        
        if ($accept.code -eq 0) {
            Write-Host "OK Order accepted" -ForegroundColor Green
            Write-Host "  Status: $($accept.data.status) (3=ACCEPTED)" -ForegroundColor White
        } else {
            Write-Host "FAIL Accept: $($accept.msg)" -ForegroundColor Red
            Write-Host ($accept | ConvertTo-Json)
        }
    } catch {
        Write-Host "FAIL Accept: $_" -ForegroundColor Red
    }
}

# Test Cancel (after accept)
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 3: Test Cancel After Accept" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "[9] Boss try cancel (should fail)..." -ForegroundColor Yellow
try {
    $cancelBoss = Invoke-Api -Method Post -Path "/api/order/cancel" `
        -Headers @{Authorization="Bearer $BOSS_TOKEN"} `
        -Body @{orderId=[int64]$ORDER_ID; reason="Test"}
    
    if ($cancelBoss.code -ne 0) {
        Write-Host "PASS: Boss correctly rejected - $($cancelBoss.msg)" -ForegroundColor Green
    } else {
        Write-Host "FAIL: Boss should not be able to cancel" -ForegroundColor Red
    }
} catch {
    Write-Host "PASS: Boss correctly rejected (exception)" -ForegroundColor Green
}

Write-Host "[10] Companion cancel (should success)..." -ForegroundColor Yellow
try {
    $cancelComp = Invoke-Api -Method Post -Path "/api/order/cancel" `
        -Headers @{Authorization="Bearer $COMP_TOKEN"} `
        -Body @{orderId=[int64]$ORDER_ID; reason="Test cancel"}
    
    if ($cancelComp.code -eq 0) {
        Write-Host "PASS: Companion canceled" -ForegroundColor Green
        Write-Host "  Status: $($cancelComp.data.status) (8=CANCEL_REFUNDING)" -ForegroundColor White
        Write-Host "  Refund will be processed asynchronously" -ForegroundColor Cyan
    } else {
        Write-Host "FAIL: $($cancelComp.msg)" -ForegroundColor Red
    }
} catch {
    Write-Host "FAIL: $_" -ForegroundColor Red
}

# Check wallet after cancellation
Write-Host ""
Write-Host "[11] Check Boss Wallet After Cancel..." -ForegroundColor Yellow
Start-Sleep -Seconds 2
$walletAfter = Invoke-Api -Method Get -Path "/api/user/wallet" -Headers @{Authorization="Bearer $BOSS_TOKEN"}
Write-Host "Balance: $($walletAfter.data.balance)" -ForegroundColor Cyan
Write-Host "  (Refund may take a few seconds to process)" -ForegroundColor Gray

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test Complete!" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

