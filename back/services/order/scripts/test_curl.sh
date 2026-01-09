#!/bin/bash
# 订单取消功能 cURL 测试脚本

BASE_URL="http://localhost:8888"
PHONE_BOSS="13800001001"
PHONE_COMPANION="13800002001"
PASSWORD="Test123456"

echo "=========================================="
echo "订单取消功能完整测试 (cURL)"
echo "=========================================="
echo ""

# 步骤 1: 老板登录
echo "[1] 老板登录..."
BOSS_LOGIN_RESP=$(curl -s -X POST "$BASE_URL/api/user/login" \
  -H "Content-Type: application/json" \
  -d "{\"phone\":\"$PHONE_BOSS\",\"password\":\"$PASSWORD\"}")

BOSS_TOKEN=$(echo "$BOSS_LOGIN_RESP" | grep -o '"accessToken":"[^"]*' | cut -d'"' -f4)
BOSS_ID=$(echo "$BOSS_LOGIN_RESP" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

if [ -z "$BOSS_TOKEN" ]; then
    echo "✗ 老板登录失败，请先注册用户"
    echo "响应: $BOSS_LOGIN_RESP"
    exit 1
fi

echo "✓ 老板登录成功 (ID: $BOSS_ID)"
echo ""

# 步骤 2: 陪玩登录
echo "[2] 陪玩登录..."
COMPANION_LOGIN_RESP=$(curl -s -X POST "$BASE_URL/api/user/login" \
  -H "Content-Type: application/json" \
  -d "{\"phone\":\"$PHONE_COMPANION\",\"password\":\"$PASSWORD\"}")

COMPANION_TOKEN=$(echo "$COMPANION_LOGIN_RESP" | grep -o '"accessToken":"[^"]*' | cut -d'"' -f4)
COMPANION_ID=$(echo "$COMPANION_LOGIN_RESP" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

if [ -z "$COMPANION_TOKEN" ]; then
    echo "✗ 陪玩登录失败，请先注册用户"
    echo "响应: $COMPANION_LOGIN_RESP"
    exit 1
fi

echo "✓ 陪玩登录成功 (ID: $COMPANION_ID)"
echo ""

# 步骤 3: 创建订单
echo "[3] 创建订单（老板）..."
CREATE_ORDER_RESP=$(curl -s -X POST "$BASE_URL/api/order" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BOSS_TOKEN" \
  -d "{\"companionId\":$COMPANION_ID,\"gameName\":\"王者荣耀\",\"gameMode\":\"排位赛\",\"durationMinutes\":60}")

ORDER_ID=$(echo "$CREATE_ORDER_RESP" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
ORDER_STATUS=$(echo "$CREATE_ORDER_RESP" | grep -o '"status":[0-9]*' | head -1 | cut -d':' -f2)

if [ -z "$ORDER_ID" ]; then
    echo "✗ 创建订单失败"
    echo "响应: $CREATE_ORDER_RESP"
    exit 1
fi

echo "✓ 订单创建成功 (ID: $ORDER_ID, 状态: $ORDER_STATUS)"
echo "完整响应: $CREATE_ORDER_RESP"
echo ""

# 测试场景 1: 老板取消未接单订单
echo "=========================================="
echo "测试场景 1: 老板取消未接单订单"
echo "=========================================="
echo ""

echo "[4] 老板取消订单..."
CANCEL_RESP1=$(curl -s -X POST "$BASE_URL/api/order/cancel" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BOSS_TOKEN" \
  -d "{\"orderId\":$ORDER_ID,\"reason\":\"临时有事\"}")

echo "响应: $CANCEL_RESP1"
CANCEL_CODE1=$(echo "$CANCEL_RESP1" | grep -o '"code":[0-9]*' | cut -d':' -f2)
CANCEL_STATUS1=$(echo "$CANCEL_RESP1" | grep -o '"status":[0-9]*' | head -1 | cut -d':' -f2)

if [ "$CANCEL_CODE1" = "0" ]; then
    echo "✓ 老板取消订单成功 (状态: $CANCEL_STATUS1)"
else
    echo "✗ 老板取消订单失败 (code: $CANCEL_CODE1)"
fi
echo ""

# 测试场景 2: 接单后取消权限测试
echo "=========================================="
echo "测试场景 2: 接单后取消权限测试"
echo "=========================================="
echo ""

echo "[5] 创建第二个订单..."
CREATE_ORDER_RESP2=$(curl -s -X POST "$BASE_URL/api/order" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BOSS_TOKEN" \
  -d "{\"companionId\":$COMPANION_ID,\"gameName\":\"王者荣耀\",\"gameMode\":\"排位赛\",\"durationMinutes\":60}")

ORDER_ID2=$(echo "$CREATE_ORDER_RESP2" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
echo "✓ 订单创建成功 (ID: $ORDER_ID2)"
echo ""

sleep 2

echo "[6] 陪玩接单..."
ACCEPT_RESP=$(curl -s -X POST "$BASE_URL/api/order/accept" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $COMPANION_TOKEN" \
  -d "{\"orderId\":$ORDER_ID2}")

ACCEPT_CODE=$(echo "$ACCEPT_RESP" | grep -o '"code":[0-9]*' | cut -d':' -f2)
if [ "$ACCEPT_CODE" = "0" ]; then
    echo "✓ 接单成功"
else
    echo "✗ 接单失败"
    echo "响应: $ACCEPT_RESP"
    exit 1
fi
echo ""

echo "[7] 老板尝试取消已接单订单（应该失败）..."
CANCEL_RESP2=$(curl -s -X POST "$BASE_URL/api/order/cancel" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BOSS_TOKEN" \
  -d "{\"orderId\":$ORDER_ID2,\"reason\":\"改变主意\"}")

CANCEL_CODE2=$(echo "$CANCEL_RESP2" | grep -o '"code":[0-9]*' | cut -d':' -f2)
echo "响应: $CANCEL_RESP2"

if [ "$CANCEL_CODE2" != "0" ]; then
    echo "✓ 正确拒绝：接单后老板不能取消订单"
else
    echo "✗ 错误：接单后老板应该不能取消订单"
fi
echo ""

echo "[8] 陪玩取消已接单订单（应该成功）..."
CANCEL_RESP3=$(curl -s -X POST "$BASE_URL/api/order/cancel" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $COMPANION_TOKEN" \
  -d "{\"orderId\":$ORDER_ID2,\"reason\":\"临时有事\"}")

CANCEL_CODE3=$(echo "$CANCEL_RESP3" | grep -o '"code":[0-9]*' | cut -d':' -f2)
CANCEL_STATUS3=$(echo "$CANCEL_RESP3" | grep -o '"status":[0-9]*' | head -1 | cut -d':' -f2)
echo "响应: $CANCEL_RESP3"

if [ "$CANCEL_CODE3" = "0" ]; then
    echo "✓ 陪玩取消订单成功 (状态: $CANCEL_STATUS3)"
else
    echo "✗ 陪玩取消订单失败 (code: $CANCEL_CODE3)"
fi
echo ""

echo "=========================================="
echo "测试完成！"
echo "=========================================="












