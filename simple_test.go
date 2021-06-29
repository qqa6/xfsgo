package xblockchain

import "testing"

func TestSimple(t *testing.T) {
	i := 123
	p := &i
	t.Logf("i: %x\n", &i)
	t.Logf("i: %x\n", p)
}

func TestList(t *testing.T) {
	var list []*string
	t.Logf("list: %x\n", list)
	for i := 0; i < len(list); i++ {
		item := list[i]
		t.Logf("item(%d): %s\n",i, *item)
	}
}

func TestMap(t *testing.T) {
	var data map[string]*string
	a := "a"
	b := "b"
	c := "c"
	t.Logf("a: %x\n", &a)
	t.Logf("b: %x\n", &b)
	t.Logf("c: %x\n", &c)
	t.Logf("map: %x\n", &data)
	data = make(map[string]*string)
	data["a"] = &a
	data["b"] = &b
	data["c"] = &c
	for k := range data {
		v := data[k]
		t.Logf("item(%s): %x\n",k, v)
	}
	//for i := 0; i < len(data); i++ {
	//	item := data[i]
	//	t.Logf("item(%d): %s\n",i, *item)
	//}
}