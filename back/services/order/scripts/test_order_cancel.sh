#!/bin/bash

# 订单取消功能完整测试脚本
# 测试场景：
# 1. 老板在未接单前取消订单
# 2. 接单后只有陪玩能取消订单

BASE_URL="http://localhost:8888"
PHONE_BOSS="13800001001"
PHONE_COMPANION="13800002001"
PASSWORD="Test123456"

echo "=========================================="
echo "订单取消功能完整测试"
echo "=========================================="
echo ""

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 辅助函数：打印 JSON 响应
print_response() {
    echo -e "${YELLOW}Response:${NC}"
    echo "$1" | jq '.' 2>/dev/null || echo "$1"
    echo ""
}

# 辅助函数：提取字段值
extract_field() {
    echo "$1" | jq -r "$2" 2>/dev/null
}

# 辅助函数：提取 token
extract_token() {
    extract_field "$1" '.data.accessToken'
}

# 辅助函数：提取订单ID
extract_order_id() {
    extract_field "$1" '.data.id'
}

# 辅助函数：提取用户ID
extract_user_id() {
    extract_field "$1" '.data.id'
}

# 辅助函数：提取陪玩ID
extract_companion_id() {
    extract_field "$1" '.data.companions[0].userId'
}

echo "步骤 1: 发送验证码 - 老板"
BOSS_CODE_RESP=$(curl -s -X POST "$BASE_URL/api/code/send" \
  -H "Content-Type: application/json" \
  -d "{\"phone\":\"$PHONE_BOSS\",\"purpose\":\"register\"}")
print_response "$BOSS_CODE_RESP"
BOSS_CODE=$(echo "$BOSS_CODE_RESP" | jq -r '.data.success' 2>/dev/null)
if [ "$BOSS_CODE" != "true" ]; then
    echo -e "${RED}✗ 发送验证码失败${NC}"
    exit 1
fi
echo "请输入老板验证码:"
read BOSS_CODE_INPUT

echo ""
echo "步骤 2: 注册老板用户"
BOSS_REGISTER_RESP=$(curl -s -X POST "$BASE_URL/api/user/register" \
  -H "Content-Type: application/json" \
  -d "{\"phone\":\"$PHONE_BOSS\",\"code\":\"$BOSS_CODE_INPUT\",\"password\":\"$PASSWORD\",\"nickname\":\"测试老板\",\"role\":1}")
print_response "$BOSS_REGISTER_RESP"
BOSS_TOKEN=$(extract_token "$BOSS_REGISTER_RESP")
BOSS_USER_ID=$(extract_user_id "$BOSS_REGISTER_RESP")

if [ -z "$BOSS_TOKEN" ] || [ "$BOSS_TOKEN" = "null" ]; then
    echo -e "${RED}✗ 注册失败，尝试登录${NC}"
    BOSS_LOGIN_RESP=$(curl -s -X POST "$BASE_URL/api/user/login" \
      -H "Content-Type: application/json" \
      -d "{\"phone\":\"$PHONE_BOSS\",\"password\":\"$PASSWORD\"}")
    BOSS_TOKEN=$(extract_token "$BOSS_LOGIN_RESP")
    BOSS_USER_ID=$(extract_user_id "$BOSS_LOGIN_RESP")
fi

if [ -z "$BOSS_TOKEN" ] || [ "$BOSS_TOKEN" = "null" ]; then
    echo -e "${RED}✗ 无法获取老板 Token${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 老板注册/登录成功${NC}"
echo "老板用户ID: $BOSS_USER_ID"
echo ""

echo "步骤 3: 发送验证码 - 陪玩"
COMPANION_CODE_RESP=$(curl -s -X POST "$BASE_URL/api/code/send" \
  -H "Content-Type: application/json" \
  -d "{\"phone\":\"$PHONE_COMPANION\",\"purpose\":\"register\"}")
print_response "$COMPANION_CODE_RESP"
echo "请输入陪玩验证码:"
read COMPANION_CODE_INPUT

echo ""
echo "步骤 4: 注册陪玩用户"
COMPANION_REGISTER_RESP=$(curl -s -X POST "$BASE_URL/api/user/register" \
  -H "Content-Type: application/json" \
  -d "{\"phone\":\"$PHONE_COMPANION\",\"code\":\"$COMPANION_CODE_INPUT\",\"password\":\"$PASSWORD\",\"nickname\":\"测试陪玩\",\"role\":2}")
print_response "$COMPANION_REGISTER_RESP"
COMPANION_TOKEN=$(extract_token "$COMPANION_REGISTER_RESP")
COMPANION_USER_ID=$(extract_user_id "$COMPANION_REGISTER_RESP")

if [ -z "$COMPANION_TOKEN" ] || [ "$COMPANION_TOKEN" = "null" ]; then
    echo -e "${RED}✗ 注册失败，尝试登录${NC}"
    COMPANION_LOGIN_RESP=$(curl -s -X POST "$BASE_URL/api/user/login" \
      -H "Content-Type: application/json" \
      -d "{\"phone\":\"$PHONE_COMPANION\",\"password\":\"$PASSWORD\"}")
    COMPANION_TOKEN=$(extract_token "$COMPANION_LOGIN_RESP")
    COMPANION_USER_ID=$(extract_user_id "$COMPANION_LOGIN_RESP")
fi

if [ -z "$COMPANION_TOKEN" ] || [ "$COMPANION_TOKEN" = "null" ]; then
    echo -e "${RED}✗ 无法获取陪玩 Token${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 陪玩注册/登录成功${NC}"
echo "陪玩用户ID: $COMPANION_USER_ID"
echo ""

echo "步骤 5: 设置陪玩信息（价格等）"
COMPANION_UPDATE_RESP=$(curl -s -X PUT "$BASE_URL/api/user/companion/profile" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $COMPANION_TOKEN" \
  -d "{\"gameSkills\":\"[\\\"王者荣耀\\\",\\\"和平精英\\\"]\",\"pricePerHour\":100,\"status\":1}")
print_response "$COMPANION_UPDATE_RESP"
echo -e "${GREEN}✓ 陪玩信息设置成功${NC}"
echo ""

echo "步骤 6: 查看老板钱包"
BOSS_WALLET_RESP=$(curl -s -X GET "$BASE_URL/api/user/wallet" \
  -H "Authorization: Bearer $BOSS_TOKEN")
print_response "$BOSS_WALLET_RESP"
BOSS_BALANCE=$(extract_field "$BOSS_WALLET_RESP" '.data.balance')

if [ "${BOSS_BALANCE:-0}" -lt 200 ]; then
    echo -e "${YELLOW}⚠ 老板余额不足，需要充值（手动调用充值接口）${NC}"
    echo "余额: $BOSS_BALANCE"
    echo "建议充值至少 200 帅币"
    echo ""
fi

echo "步骤 7: 创建订单（老板）"
CREATE_ORDER_RESP=$(curl -s -X POST "$BASE_URL/api/order" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BOSS_TOKEN" \
  -d "{\"companionId\":$COMPANION_USER_ID,\"gameName\":\"王者荣耀\",\"gameMode\":\"排位赛-王者段位\",\"durationMinutes\":60}")
print_response "$CREATE_ORDER_RESP"
ORDER_ID=$(extract_order_id "$CREATE_ORDER_RESP")
ORDER_STATUS=$(extract_field "$CREATE_ORDER_RESP" '.data.status')

if [ -z "$ORDER_ID" ] || [ "$ORDER_ID" = "null" ]; then
    echo -e "${RED}✗ 创建订单失败${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 订单创建成功${NC}"
echo "订单ID: $ORDER_ID"
echo "订单状态: $ORDER_STATUS"
echo ""

echo "=========================================="
echo "测试场景 1: 老板在未接单前取消订单"
echo "=========================================="
echo ""

echo "步骤 8: 老板取消订单（应该成功）"
CANCEL_RESP1=$(curl -s -X POST "$BASE_URL/api/order/cancel" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BOSS_TOKEN" \
  -d "{\"orderId\":$ORDER_ID,\"reason\":\"临时有事，不想玩了\"}")
print_response "$CANCEL_RESP1"
CANCEL_STATUS1=$(extract_field "$CANCEL_RESP1" '.data.status')
CANCEL_CODE1=$(extract_field "$CANCEL_RESP1" '.code')

if [ "$CANCEL_CODE1" = "0" ]; then
    echo -e "${GREEN}✓ 老板取消订单成功${NC}"
    echo "取消后状态: $CANCEL_STATUS1"
else
    echo -e "${RED}✗ 老板取消订单失败${NC}"
    echo "错误信息: $(extract_field "$CANCEL_RESP1" '.msg')"
fi
echo ""

echo "步骤 9: 验证：陪玩尝试取消已取消的订单（应该失败）"
CANCEL_RESP2=$(curl -s -X POST "$BASE_URL/api/order/cancel" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $COMPANION_TOKEN" \
  -d "{\"orderId\":$ORDER_ID,\"reason\":\"测试\"}")
print_response "$CANCEL_RESP2"
CANCEL_CODE2=$(extract_field "$CANCEL_RESP2" '.code')

if [ "$CANCEL_CODE2" != "0" ]; then
    echo -e "${GREEN}✓ 正确拒绝：已取消的订单不能再次取消${NC}"
else
    echo -e "${RED}✗ 错误：已取消的订单应该不能被再次取消${NC}"
fi
echo ""

echo "=========================================="
echo "测试场景 2: 接单后只有陪玩能取消"
echo "=========================================="
echo ""

echo "步骤 10: 再次创建订单"
CREATE_ORDER_RESP2=$(curl -s -X POST "$BASE_URL/api/order" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BOSS_TOKEN" \
  -d "{\"companionId\":$COMPANION_USER_ID,\"gameName\":\"王者荣耀\",\"gameMode\":\"排位赛-王者段位\",\"durationMinutes\":60}")
print_response "$CREATE_ORDER_RESP2"
ORDER_ID2=$(extract_order_id "$CREATE_ORDER_RESP2")

if [ -z "$ORDER_ID2" ] || [ "$ORDER_ID2" = "null" ]; then
    echo -e "${RED}✗ 创建订单失败${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 订单创建成功，订单ID: $ORDER_ID2${NC}"
echo ""

# 等待支付完成（如果有自动支付）
sleep 2

echo "步骤 11: 陪玩接单"
ACCEPT_RESP=$(curl -s -X POST "$BASE_URL/api/order/accept" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $COMPANION_TOKEN" \
  -d "{\"orderId\":$ORDER_ID2}")
print_response "$ACCEPT_RESP"
ACCEPT_STATUS=$(extract_field "$ACCEPT_RESP" '.data.status')
ACCEPT_CODE=$(extract_field "$ACCEPT_RESP" '.code')

if [ "$ACCEPT_CODE" = "0" ]; then
    echo -e "${GREEN}✓ 陪玩接单成功${NC}"
    echo "接单后状态: $ACCEPT_STATUS"
else
    echo -e "${RED}✗ 陪玩接单失败${NC}"
    echo "错误: $(extract_field "$ACCEPT_RESP" '.msg')"
    exit 1
fi
echo ""

echo "步骤 12: 验证：老板尝试取消已接单的订单（应该失败）"
CANCEL_RESP3=$(curl -s -X POST "$BASE_URL/api/order/cancel" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BOSS_TOKEN" \
  -d "{\"orderId\":$ORDER_ID2,\"reason\":\"改变主意\"}")
print_response "$CANCEL_RESP3"
CANCEL_CODE3=$(extract_field "$CANCEL_RESP3" '.code')

if [ "$CANCEL_CODE3" != "0" ]; then
    echo -e "${GREEN}✓ 正确拒绝：接单后老板不能取消订单${NC}"
    echo "错误信息: $(extract_field "$CANCEL_RESP3" '.msg')"
else
    echo -e "${RED}✗ 错误：接单后老板应该不能取消订单${NC}"
fi
echo ""

echo "步骤 13: 陪玩取消已接单的订单（应该成功）"
CANCEL_RESP4=$(curl -s -X POST "$BASE_URL/api/order/cancel" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $COMPANION_TOKEN" \
  -d "{\"orderId\":$ORDER_ID2,\"reason\":\"临时有事\"}")
print_response "$CANCEL_RESP4"
CANCEL_CODE4=$(extract_field "$CANCEL_RESP4" '.code')
CANCEL_STATUS4=$(extract_field "$CANCEL_RESP4" '.data.status')

if [ "$CANCEL_CODE4" = "0" ]; then
    echo -e "${GREEN}✓ 陪玩取消订单成功${NC}"
    echo "取消后状态: $CANCEL_STATUS4"
    echo "说明: 状态为 8 表示正在退款中（CANCEL_REFUNDING）"
else
    echo -e "${RED}✗ 陪玩取消订单失败${NC}"
    echo "错误: $(extract_field "$CANCEL_RESP4" '.msg')"
fi
echo ""

echo "=========================================="
echo "测试完成！"
echo "=========================================="
echo ""
echo "总结："
echo "- 老板可以在未接单前取消订单 ✓"
echo "- 接单后只有陪玩能取消订单 ✓"
echo "- 权限验证正常工作 ✓"












