package snowflake

import "github.com/bwmarrin/snowflake"

var node, _ = snowflake.NewNode(1)

func GenID() uint64 {
	return uint64(node.Generate())
}

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// EncodeBase36 把雪花ID编码成 Base36
func EncodeBase36(id uint64) string {
	if id == 0 {
		return "0"
	}

	var buf [13]byte // 36 进制下 uint64 最多 13 位
	i := len(buf)

	for id > 0 {
		i--
		buf[i] = alphabet[id%36]
		id /= 36
	}

	return string(buf[i:])
}

// MakeUserUID 生成对外 UID，例如 U3F8K9ZQ
func MakeUserUID(id uint64) string {
	return "U" + EncodeBase36(id)
}
