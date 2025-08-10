# PowerShell script to test the Targeting Engine API

$BaseUrl = "http://localhost:8080"

Write-Host "🧪 Testing Targeting Engine API" -ForegroundColor Yellow
Write-Host "==================================" -ForegroundColor Yellow

# Test health endpoint
Write-Host "`n1. Testing Health Endpoint" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/healthz" -Method Get
    Write-Host "✅ Health check passed" -ForegroundColor Green
    Write-Host "Response: $($response | ConvertTo-Json)"
} catch {
    Write-Host "❌ Health check failed" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)"
}

# Test successful delivery request
Write-Host "`n2. Testing Successful Delivery Request" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/v1/delivery?app=com.gametion.ludokinggame&country=us&os=android" -Method Get
    Write-Host "✅ Delivery request successful" -ForegroundColor Green
    Write-Host "Response: $($response | ConvertTo-Json -Depth 3)"
} catch {
    Write-Host "❌ Delivery request failed" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)"
}

# Test delivery request with no matches
Write-Host "`n3. Testing Delivery Request with No Matches" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/v1/delivery?app=com.test&country=us&os=web" -Method Get
    Write-Host "❌ Expected 204, but got response" -ForegroundColor Red
} catch {
    if ($_.Exception.Response.StatusCode -eq 204) {
        Write-Host "✅ No matches response correct (204)" -ForegroundColor Green
    } else {
        Write-Host "❌ Expected 204, got $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    }
}

# Test missing parameters
Write-Host "`n4. Testing Missing Parameters" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/v1/delivery?country=us&os=android" -Method Get
    Write-Host "❌ Expected 400, but got response" -ForegroundColor Red
} catch {
    if ($_.Exception.Response.StatusCode -eq 400) {
        Write-Host "✅ Missing parameter handled correctly" -ForegroundColor Green
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $body = $reader.ReadToEnd()
        Write-Host "Response: $body"
    } else {
        Write-Host "❌ Expected 400, got $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    }
}

# Test case insensitive matching
Write-Host "`n5. Testing Case Insensitive Matching" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/v1/delivery?app=COM.GAMETION.LUDOKINGGAME&country=US&os=ANDROID" -Method Get
    Write-Host "✅ Case insensitive matching works" -ForegroundColor Green
    Write-Host "Response: $($response | ConvertTo-Json -Depth 3)"
} catch {
    Write-Host "❌ Case insensitive matching failed" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)"
}

# Test duolingo campaign
Write-Host "`n6. Testing Duolingo Campaign" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/v1/delivery?app=com.test&country=germany&os=android" -Method Get
    Write-Host "✅ Duolingo campaign found" -ForegroundColor Green
    Write-Host "Response: $($response | ConvertTo-Json -Depth 3)"
} catch {
    Write-Host "❌ Duolingo campaign not found" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)"
}

Write-Host "`n🎉 API Testing Complete!" -ForegroundColor Green 