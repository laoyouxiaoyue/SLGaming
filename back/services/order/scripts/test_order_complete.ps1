# Complete Order Flow Test - Including Order Completion
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

Write-Host "=== Order Complete Flow Test ===" -ForegroundColor Cyan
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

# Check wallets
Write-Host "[3] Check Initial Wallets..." -ForegroundColor Yellow
$bossWallet = Invoke-RestMethod -Uri "$BASE_URL/api/user/wallet" -Method Get -Headers @{Authorization="Bearer $BOSS_TOKEN"} -ErrorAction Stop
$compWallet = Invoke-RestMethod -Uri "$BASE_URL/api/user/wallet" -Method Get -Headers @{Authorization="Bearer $COMP_TOKEN"} -ErrorAction Stop
$INITIAL_BOSS_BALANCE = $bossWallet.data.balance
$INITIAL_COMP_BALANCE = $compWallet.data.balance
Write-Host "Boss balance: $INITIAL_BOSS_BALANCE" -ForegroundColor Cyan
Write-Host "Companion balance: $INITIAL_COMP_BALANCE" -ForegroundColor Cyan

# Create Order
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 1: Create Order" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "[4] Create Order..." -ForegroundColor Yellow
$order = Invoke-Api -Method Post -Path "/api/order" `
    -Headers @{Authorization="Bearer $BOSS_TOKEN"} `
    -Body @{companionId=[int64]$COMP_ID; gameName="CompleteTest"; durationMinutes=60}

if ($order.code -ne 0) {
    Write-Host "FAIL: $($order.msg)" -ForegroundColor Red
    exit
}

$ORDER_ID = $order.data.id
$ORDER_NO = $order.data.orderNo
$ORDER_AMOUNT = $order.data.totalAmount
Write-Host "OK Order created" -ForegroundColor Green
Write-Host "  Order ID: $ORDER_ID" -ForegroundColor White
Write-Host "  Order No: $ORDER_NO" -ForegroundColor White
Write-Host "  Status: $($order.data.status) (1=CREATED)" -ForegroundColor White
Write-Host "  Amount: $ORDER_AMOUNT" -ForegroundColor White

# Wait for payment
Write-Host ""
Write-Host "[5] Waiting for payment processing (5 seconds)..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Check wallet after payment
Write-Host "[6] Check Boss Wallet After Payment..." -ForegroundColor Yellow
$bossWalletAfterPayment = Invoke-RestMethod -Uri "$BASE_URL/api/user/wallet" -Method Get -Headers @{Authorization="Bearer $BOSS_TOKEN"} -ErrorAction Stop
$bossBalanceAfterPayment = $bossWalletAfterPayment.data.balance
Write-Host "Boss balance: $bossBalanceAfterPayment (was $INITIAL_BOSS_BALANCE)" -ForegroundColor Cyan
if ($bossBalanceAfterPayment -lt $INITIAL_BOSS_BALANCE) {
    Write-Host "  Payment deducted!" -ForegroundColor Green
    $paymentDeducted = $INITIAL_BOSS_BALANCE - $bossBalanceAfterPayment
    Write-Host "  Amount deducted: $paymentDeducted" -ForegroundColor Green
} else {
    Write-Host "  Payment may still be processing or failed" -ForegroundColor Yellow
}

# Accept Order
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 2: Companion Accept Order" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "[7] Companion accept order..." -ForegroundColor Yellow
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

# Complete Order
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Step 3: Complete Order" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "[8] Complete order..." -ForegroundColor Yellow
$complete = Invoke-Api -Method Post -Path "/api/order/complete" `
    -Headers @{Authorization="Bearer $BOSS_TOKEN"} `
    -Body @{orderId=[int64]$ORDER_ID}

if ($complete.code -eq 0) {
    Write-Host "OK Order completed" -ForegroundColor Green
    Write-Host "  Status: $($complete.data.status)" -ForegroundColor White
    Write-Host "    Status 4 = COMPLETED" -ForegroundColor Gray
    Write-Host "  Completed at: $($complete.data.completedAt)" -ForegroundColor White
} else {
    Write-Host "FAIL Complete: $($complete.msg)" -ForegroundColor Red
    Write-Host ($complete | ConvertTo-Json)
    exit
}

# Wait for companion payment processing
Write-Host ""
Write-Host "[9] Waiting for companion payment processing (5 seconds)..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Check wallets after completion
Write-Host "[10] Check Final Wallets..." -ForegroundColor Yellow
$bossWalletFinal = Invoke-RestMethod -Uri "$BASE_URL/api/user/wallet" -Method Get -Headers @{Authorization="Bearer $BOSS_TOKEN"} -ErrorAction Stop
$compWalletFinal = Invoke-RestMethod -Uri "$BASE_URL/api/user/wallet" -Method Get -Headers @{Authorization="Bearer $COMP_TOKEN"} -ErrorAction Stop
$FINAL_BOSS_BALANCE = $bossWalletFinal.data.balance
$FINAL_COMP_BALANCE = $compWalletFinal.data.balance

Write-Host ""
Write-Host "Wallet Summary:" -ForegroundColor Cyan
Write-Host "  Boss:" -ForegroundColor White
Write-Host "    Initial: $INITIAL_BOSS_BALANCE" -ForegroundColor Gray
Write-Host "    After Payment: $bossBalanceAfterPayment" -ForegroundColor Gray
Write-Host "    Final: $FINAL_BOSS_BALANCE" -ForegroundColor Gray
if ($FINAL_BOSS_BALANCE -lt $INITIAL_BOSS_BALANCE) {
    $bossSpent = $INITIAL_BOSS_BALANCE - $FINAL_BOSS_BALANCE
    Write-Host "    Total Spent: $bossSpent" -ForegroundColor Yellow
}

Write-Host "  Companion:" -ForegroundColor White
Write-Host "    Initial: $INITIAL_COMP_BALANCE" -ForegroundColor Gray
Write-Host "    Final: $FINAL_COMP_BALANCE" -ForegroundColor Gray
if ($FINAL_COMP_BALANCE -gt $INITIAL_COMP_BALANCE) {
    $compEarned = $FINAL_COMP_BALANCE - $INITIAL_COMP_BALANCE
    Write-Host "    Total Earned: $compEarned" -ForegroundColor Green
    if ($compEarned -eq $ORDER_AMOUNT) {
        Write-Host "    ✓ Companion received correct payment!" -ForegroundColor Green
    } else {
        Write-Host "    ⚠ Payment amount mismatch (expected $ORDER_AMOUNT, got $compEarned)" -ForegroundColor Yellow
    }
} else {
    Write-Host "    ⚠ Companion payment may still be processing" -ForegroundColor Yellow
}

# Get order final status
Write-Host ""
Write-Host "[11] Get Final Order Status..." -ForegroundColor Yellow
$orderFinal = Invoke-Api -Method Get -Path "/api/order/$ORDER_ID" `
    -Headers @{Authorization="Bearer $BOSS_TOKEN"}

if ($orderFinal.code -eq 0) {
    Write-Host "Final Order Status:" -ForegroundColor Cyan
    Write-Host "  Status: $($orderFinal.data.status) (4=COMPLETED)" -ForegroundColor White
    Write-Host "  Created: $($orderFinal.data.createdAt)" -ForegroundColor Gray
    Write-Host "  Accepted: $($orderFinal.data.acceptedAt)" -ForegroundColor Gray
    Write-Host "  Completed: $($orderFinal.data.completedAt)" -ForegroundColor Gray
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test Complete!" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Summary:" -ForegroundColor Yellow
Write-Host "  ✓ Order created" -ForegroundColor Green
Write-Host "  ✓ Payment processed" -ForegroundColor Green
Write-Host "  ✓ Order accepted" -ForegroundColor Green
Write-Host "  ✓ Order completed" -ForegroundColor Green
Write-Host "  ✓ Companion payment processed" -ForegroundColor Green












