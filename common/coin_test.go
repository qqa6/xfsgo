package common

import (
	"testing"
)

func TestBaseCoin2Atto(t *testing.T) {
	attoCoin := BaseCoin2Atto(50)
	attoCoin2 := BaseCoin2Atto(50)
	a := attoCoin.Add(attoCoin, attoCoin2)
	t.Logf("%s", a)
	basecoin := Atto2BaseCoin(a)
	// t.Logf("%s", attoCoin)
	t.Logf("%s", basecoin)
}
