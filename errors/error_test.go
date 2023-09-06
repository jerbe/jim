package errors

import (
	"fmt"
	"testing"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/20 23:05
  @describe :
*/

var T = New("this is test")

func TestWrap(t *testing.T) {
	e := T.Wrap(New("aa"))
	e = e.(*Error).Wrap(e)
	e = Wrap(e)
	e = Wrap(e)
	e = Wrap(e)

	e2 := New("c")
	e = e.(*Error).Wrap(e2)
	e = Wrap(e)
	e = Wrap(e)
	fmt.Printf("%+v", e)
}
