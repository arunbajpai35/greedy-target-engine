#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"

echo -e "${YELLOW}🧪 Testing Targeting Engine API${NC}"
echo "=================================="

# Test health endpoint
echo -e "\n${YELLOW}1. Testing Health Endpoint${NC}"
response=$(curl -s -w "%{http_code}" "$BASE_URL/healthz")
http_code="${response: -3}"
body="${response%???}"

if [ "$http_code" -eq 200 ]; then
    echo -e "${GREEN}✅ Health check passed${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Health check failed${NC}"
    echo "HTTP Code: $http_code"
    echo "Response: $body"
fi

# Test successful delivery request
echo -e "\n${YELLOW}2. Testing Successful Delivery Request${NC}"
response=$(curl -s -w "%{http_code}" "$BASE_URL/v1/delivery?app=com.gametion.ludokinggame&country=us&os=android")
http_code="${response: -3}"
body="${response%???}"

if [ "$http_code" -eq 200 ]; then
    echo -e "${GREEN}✅ Delivery request successful${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Delivery request failed${NC}"
    echo "HTTP Code: $http_code"
    echo "Response: $body"
fi

# Test delivery request with no matches
echo -e "\n${YELLOW}3. Testing Delivery Request with No Matches${NC}"
response=$(curl -s -w "%{http_code}" "$BASE_URL/v1/delivery?app=com.test&country=us&os=web")
http_code="${response: -3}"

if [ "$http_code" -eq 204 ]; then
    echo -e "${GREEN}✅ No matches response correct (204)${NC}"
else
    echo -e "${RED}❌ Expected 204, got $http_code${NC}"
fi

# Test missing parameters
echo -e "\n${YELLOW}4. Testing Missing Parameters${NC}"
response=$(curl -s -w "%{http_code}" "$BASE_URL/v1/delivery?country=us&os=android")
http_code="${response: -3}"
body="${response%???}"

if [ "$http_code" -eq 400 ]; then
    echo -e "${GREEN}✅ Missing parameter handled correctly${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Expected 400, got $http_code${NC}"
    echo "Response: $body"
fi

# Test case insensitive matching
echo -e "\n${YELLOW}5. Testing Case Insensitive Matching${NC}"
response=$(curl -s -w "%{http_code}" "$BASE_URL/v1/delivery?app=COM.GAMETION.LUDOKINGGAME&country=US&os=ANDROID")
http_code="${response: -3}"
body="${response%???}"

if [ "$http_code" -eq 200 ]; then
    echo -e "${GREEN}✅ Case insensitive matching works${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Case insensitive matching failed${NC}"
    echo "HTTP Code: $http_code"
    echo "Response: $body"
fi

# Test duolingo campaign
echo -e "\n${YELLOW}6. Testing Duolingo Campaign${NC}"
response=$(curl -s -w "%{http_code}" "$BASE_URL/v1/delivery?app=com.test&country=germany&os=android")
http_code="${response: -3}"
body="${response%???}"

if [ "$http_code" -eq 200 ]; then
    echo -e "${GREEN}✅ Duolingo campaign found${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Duolingo campaign not found${NC}"
    echo "HTTP Code: $http_code"
    echo "Response: $body"
fi

echo -e "\n${GREEN}🎉 API Testing Complete!${NC}" 