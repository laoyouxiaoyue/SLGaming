# Test Balance Check in Create Order
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

Write-Host "=== Test Balance Check in Create Order ===" -ForegroundColor Cyan
Write-Host ""

# Login
Write-Host "[1] Login Boss..." -ForegroundColor Yellow
try {
    $boss = Invoke-RestMethod -Uri "$BASE_URL/api/user/login" -Method Post -ContentType "application/json" -Body '{"phone":"13800001001","password":"Test123456"}' -ErrorAction Stop
    $BOSS_TOKEN = $boss.data.accessToken
    $BOSS_ID = Decode-JWT $BOSS_TOKEN
    Write-Host "OK Boss logged in (ID: $BOSS_ID)" -ForegroundColor Green
} catch {
    Write-Host "FAIL Login: $_" -ForegroundColor Red
    exit
}

Write-Host "[2] Login Companion..." -ForegroundColor Yellow
try {
    $comp = Invoke-RestMethod -Uri "$BASE_URL/api/user/login" -Method Post -ContentType "application/json" -Body '{"phone":"13800002001","password":"Test123456"}' -ErrorAction Stop
    $COMP_TOKEN = $comp.data.accessToken
    $COMP_ID = Decode-JWT $COMP_TOKEN
    Write-Host "OK Companion logged in (ID: $COMP_ID)" -ForegroundColor Green
} catch {
    Write-Host "FAIL Login: $_" -ForegroundColor Red
    exit
}

# Check wallet
Write-Host "[3] Check Boss Wallet..." -ForegroundColor Yellow
try {
    $wallet = Invoke-RestMethod -Uri "$BASE_URL/api/user/wallet" -Method Get -Headers @{Authorization="Bearer $BOSS_TOKEN"} -ErrorAction Stop
    $BALANCE = $wallet.data.balance
    Write-Host "Balance: $BALANCE" -ForegroundColor Cyan
} catch {
    Write-Host "FAIL Get Wallet: $_" -ForegroundColor Red
    exit
}

# Check companion profile
Write-Host "[4] Check Companion Profile..." -ForegroundColor Yellow
try {
    $profile = Invoke-RestMethod -Uri "$BASE_URL/api/user/companion/profile" -Method Get -Headers @{Authorization="Bearer $COMP_TOKEN"} -ErrorAction Stop
    $PRICE = $profile.data.pricePerHour
    Write-Host "PricePerHour: $PRICE" -ForegroundColor Cyan
    
    if ($PRICE -eq 0) {
        Write-Host "Setting companion profile..." -ForegroundColor Yellow
        $setProfile = Invoke-RestMethod -Uri "$BASE_URL/api/user/companion/profile" -Method Put -ContentType "application/json" -Headers @{Authorization="Bearer $COMP_TOKEN"} -Body (@{gameSkills='["Game1"]'; pricePerHour=100; status=1} | ConvertTo-Json) -ErrorAction Stop
        $PRICE = $setProfile.data.pricePerHour
        Write-Host "Price set to: $PRICE" -ForegroundColor Green
    }
} catch {
    Write-Host "FAIL Profile: $_" -ForegroundColor Red
    exit
}

# Calculate expected amount (60 minutes = 1 hour = 100 coins)
$EXPECTED_AMOUNT = $PRICE * 1
Write-Host "Expected order amount: $EXPECTED_AMOUNT" -ForegroundColor Yellow

# Test create order
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test: Create Order with Balance Check" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "[5] Create Order..." -ForegroundColor Yellow
try {
    $order = Invoke-Api -Method Post -Path "/api/order" `
        -Headers @{Authorization="Bearer $BOSS_TOKEN"} `
        -Body @{companionId=[int64]$COMP_ID; gameName="TestGame"; durationMinutes=60}
    
    if ($order.code -eq 0) {
        Write-Host "OK Order created successfully!" -ForegroundColor Green
        Write-Host "  Order ID: $($order.data.id)" -ForegroundColor White
        Write-Host "  Status: $($order.data.status)" -ForegroundColor White
        Write-Host "  Amount: $($order.data.totalAmount)" -ForegroundColor White
    } else {
        Write-Host "Order creation failed:" -ForegroundColor Red
        Write-Host "  Code: $($order.code)" -ForegroundColor White
        Write-Host "  Message: $($order.msg)" -ForegroundColor White
        
        if ($order.msg -like "*insufficient*" -or $order.msg -like "*balance*") {
            Write-Host "`n✓ Balance check is working! Order rejected due to insufficient balance." -ForegroundColor Green
            Write-Host "  Current balance: $BALANCE" -ForegroundColor Cyan
            Write-Host "  Required amount: $EXPECTED_AMOUNT" -ForegroundColor Cyan
        } else {
            Write-Host "  Error might be from other causes" -ForegroundColor Yellow
        }
    }
} catch {
    Write-Host "FAIL: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test Complete!" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan












