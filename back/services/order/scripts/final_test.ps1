# Complete Order Cancel Test - Fixed Version
$BASE_URL = "http://localhost:8888"
$ErrorActionPreference = "Continue"

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
                throw 
            }
        }
        throw
    }
}

function Get-UserIdFromToken {
    param($token)
    try {
        $parts = $token.Split('.')
        $payload = $parts[1]
        $payload = $payload -replace '-', '+' -replace '_', '/'
        while ($payload.Length % 4) { $payload += "=" }
        $bytes = [Convert]::FromBase64String($payload)
        $json = [System.Text.Encoding]::UTF8.GetString($bytes)
        $obj = $json | ConvertFrom-Json
        return $obj.user_id
    } catch {
        return $null
    }
}

Write-Host "=== Order Cancel Full Test ===" -ForegroundColor Cyan
Write-Host ""

# Login
Write-Host "[1] Login Boss..." -ForegroundColor Yellow
$boss = Invoke-Api -Method Post -Path "/api/user/login" -Body @{phone="13800001001"; password="Test123456"}
$BOSS_TOKEN = $boss.data.accessToken
$BOSS_ID = Get-UserIdFromToken -token $BOSS_TOKEN
Write-Host "OK Boss logged in (ID: $BOSS_ID)" -ForegroundColor Green

Write-Host "[2] Login Companion..." -ForegroundColor Yellow
$comp = Invoke-Api -Method Post -Path "/api/user/login" -Body @{phone="13800002001"; password="Test123456"}
$COMP_TOKEN = $comp.data.accessToken
$COMP_ID = Get-UserIdFromToken -token $COMP_TOKEN
Write-Host "OK Companion logged in (ID: $COMP_ID)" -ForegroundColor Green

Write-Host "[3] Set Companion Profile..." -ForegroundColor Yellow
try {
    Invoke-Api -Method Put -Path "/api/user/companion/profile" -Headers @{Authorization="Bearer $COMP_TOKEN"} -Body @{gameSkills='["Game1"]'; pricePerHour=100; status=1} | Out-Null
    Write-Host "OK Profile set" -ForegroundColor Green
} catch {
    Write-Host "Profile may already be set" -ForegroundColor Yellow
}

# Test 1: Boss cancel before accept
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test 1: Boss cancel before accept" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "[4] Create Order..." -ForegroundColor Yellow
$order1 = Invoke-Api -Method Post -Path "/api/order" -Headers @{Authorization="Bearer $BOSS_TOKEN"} -Body @{companionId=[int64]$COMP_ID; gameName="Game1"; durationMinutes=60}
$ORDER_ID1 = $order1.data.id
Write-Host "OK Order created (ID: $ORDER_ID1, Status: $($order1.data.status))" -ForegroundColor Green
Write-Host ($order1 | ConvertTo-Json)

Write-Host "[5] Boss cancel order..." -ForegroundColor Yellow
$cancel1 = Invoke-Api -Method Post -Path "/api/order/cancel" -Headers @{Authorization="Bearer $BOSS_TOKEN"} -Body @{orderId=[int64]$ORDER_ID1; reason="Test cancel"}
Write-Host ($cancel1 | ConvertTo-Json)
if ($cancel1.code -eq 0) {
    Write-Host "PASS: Boss canceled order (Status: $($cancel1.data.status))" -ForegroundColor Green
} else {
    Write-Host "FAIL: $($cancel1.msg)" -ForegroundColor Red
}

# Test 2: After accept
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test 2: After accept, only companion can cancel" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "[6] Create 2nd Order..." -ForegroundColor Yellow
$order2 = Invoke-Api -Method Post -Path "/api/order" -Headers @{Authorization="Bearer $BOSS_TOKEN"} -Body @{companionId=[int64]$COMP_ID; gameName="Game2"; durationMinutes=60}
$ORDER_ID2 = $order2.data.id
Write-Host "OK Order created (ID: $ORDER_ID2)" -ForegroundColor Green

Start-Sleep -Seconds 2

Write-Host "[7] Companion accept order..." -ForegroundColor Yellow
$accept = Invoke-Api -Method Post -Path "/api/order/accept" -Headers @{Authorization="Bearer $COMP_TOKEN"} -Body @{orderId=[int64]$ORDER_ID2}
Write-Host ($accept | ConvertTo-Json)
if ($accept.code -eq 0) {
    Write-Host "OK Order accepted (Status: $($accept.data.status))" -ForegroundColor Green
} else {
    Write-Host "FAIL Accept: $($accept.msg)" -ForegroundColor Red
    exit
}

Write-Host "[8] Boss try cancel (should fail)..." -ForegroundColor Yellow
try {
    $cancel2 = Invoke-Api -Method Post -Path "/api/order/cancel" -Headers @{Authorization="Bearer $BOSS_TOKEN"} -Body @{orderId=[int64]$ORDER_ID2; reason="Test"}
    Write-Host ($cancel2 | ConvertTo-Json)
    if ($cancel2.code -ne 0) {
        Write-Host "PASS: Boss correctly rejected - $($cancel2.msg)" -ForegroundColor Green
    } else {
        Write-Host "FAIL: Boss should not be able to cancel" -ForegroundColor Red
    }
} catch {
    Write-Host "PASS: Boss correctly rejected (exception)" -ForegroundColor Green
}

Write-Host "[9] Companion cancel (should success)..." -ForegroundColor Yellow
$cancel3 = Invoke-Api -Method Post -Path "/api/order/cancel" -Headers @{Authorization="Bearer $COMP_TOKEN"} -Body @{orderId=[int64]$ORDER_ID2; reason="Test"}
Write-Host ($cancel3 | ConvertTo-Json)
if ($cancel3.code -eq 0) {
    Write-Host "PASS: Companion canceled (Status: $($cancel3.data.status))" -ForegroundColor Green
    Write-Host "Note: Status 8 = CANCEL_REFUNDING" -ForegroundColor Yellow
} else {
    Write-Host "FAIL: $($cancel3.msg)" -ForegroundColor Red
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test Complete!" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan












