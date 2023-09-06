package errors

import (
	"errors"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/20 11:51
  @describe :
*/

// Is 报告错误链中的任何错误是否与目标匹配。
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// InIs 报告错误链中的任何错误是否与目标组中其中一个目标匹配。
func InIs(err error, targets ...error) bool {
	for i := 0; i < len(targets); i++ {
		if Is(err, targets[i]) {
			return true
		}
	}
	return false
}

// As 查找 err 链中与 target 匹配的第一个错误，如果找到，则将 target 设置为该错误值并返回 true。否则，它将返回 false。
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Unwrap 如果 err 的类型包含返回错误的解包方法，则返回在 err 上调用解包方法的结果。否则，解包将返回 nil。
func Unwrap(err error) error {
	return errors.Unwrap(err)
}
