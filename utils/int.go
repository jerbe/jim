package utils

import "strconv"

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/18 00:35
  @describe :
*/

// Int 整型的一个泛型
// 包括 ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
type Int interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func IntLen[T Int](val T) int {
	var s = int64(val)
	return len(strconv.FormatInt(int64(s), 10))
}

// SortInt 排序整型
// 返回的数据总是 a < b
func SortInt[T Int](a T, b T) (T, T) {
	if a > b {
		a, b = b, a
	}
	return a, b
}

// IntBetween dest在指定范围内
// 是一个闭区间 from <= dest <= to
func IntBetween[T Int](dest, from, to T) bool {
	if from <= dest && dest <= to {
		return true
	}
	return false
}

// BitSet 填充设置某一位的数值为1,从右到左方向
func BitSet[T Int](val T, position int) T {
	return val | (1 << position)
}

// BitClear 清理设置某一位的数值为0,从右到左方向
func BitClear[T Int](val T, position int) T {
	return val &^ (1 << position)
}
