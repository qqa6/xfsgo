package common

import (
	"testing"
)

func TestBaseCoin2Atto(t *testing.T) {
	attoCoin := BaseCoin2Atto(45778.4564)
	basecoin := Atto2BaseCoin(attoCoin)
	t.Logf("%s", attoCoin)
	t.Logf("%s", basecoin)
}
