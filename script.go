package xblockchain

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/robertkrimen/otto"
	"github.com/robertkrimen/otto/parser"
)

type ScriptVM struct {
	vm *otto.Otto
	runningStack *runningStack
}

type runningStack struct {
	elms [][]byte
	max int
	next int
}

func newRunningStack() *runningStack {
	s := &runningStack{}
	s.elms = make([][]byte,0)
	s.next = 0
	s.max = 32
	return s
}

func (s *runningStack) push(data []byte) {
	if s.next > s.max {
		return
	}
	s.elms = append(s.elms, data)
	s.next+=1
}

func (s *runningStack) pop() []byte {
	if len(s.elms) == 0 {
		return nil
	}
	elm := s.elms[s.next - 1]
	s.next -= 1
	s.elms = s.elms[0:s.next]
	return elm
}

func (s *runningStack) peek() []byte {
	if len(s.elms) == 0 {
		return nil
	}
	elm := s.elms[s.next - 1]
	return elm
}
func (s *runningStack) up() {
	if len(s.elms) != 0 {
		v := s.peek()
		s.push(v)
	}
}

func (s *runningStack) empty() bool {
	return s.next == 0
}

func (s *runningStack) top() int {
	return s.next - 1
}
func (s *runningStack) clear() {
	s.elms = make([][]byte,0)
	s.next = 0
}


func NewScriptVM() *ScriptVM {
	vm := otto.New()
	stack := newRunningStack()
	out := &ScriptVM{
		vm: vm,
		runningStack: stack,
	}
	_ = out.injectPushFunc()
	_ = out.injectPopFunc()
	_ = out.injectPeekFunc()
	_ = out.injectUpFunc()
	_ = out.injectPkhFunc()
	_ = out.injectEqFunc()
	return out
}




func BuildScriptBin(script string) (out []byte,err error){
	program, err := parser.ParseFile(nil, "", script, 0)
	if err != nil {
		return nil, err
	}
	_ = program
	out = make([]byte,base64.StdEncoding.EncodedLen(len(script)))
	base64.StdEncoding.Encode(out,[]byte(script))
	return
}

func ParseScriptBin2Str(bin []byte) (string,error) {
	dbuf := make([]byte, base64.StdEncoding.EncodedLen(len(bin)))
	n,err := base64.StdEncoding.Decode(dbuf,bin)
	if err != nil {
		return "",err
	}
	return string(dbuf[:n]),nil
}

func (vm *ScriptVM) execScriptBin(scriptBin []byte) (*otto.Value,error) {
	scriptStr,err := ParseScriptBin2Str(scriptBin)
	if err != nil {
		return nil,err
	}
	val,err := vm.vm.Run(scriptStr)
	if err != nil {
		return nil, err
	}
	err = checkValType(val)
	if err != nil {
		return nil, err
	}
	return &val, nil
}

func checkValType(val otto.Value) error {
	if val.IsUndefined() || val.IsFunction() {
		return fmt.Errorf("exec script result type is illegal")
	}
	return nil
}


func (vm *ScriptVM) ExecScriptBin(scriptBin []byte) error {
	_, err := vm.execScriptBin(scriptBin)
	if err != nil {
		return err
	}
	return nil
}

func (vm *ScriptVM) RunningStackIsEmpty() bool {
	return vm.runningStack.empty()
}

func (vm *ScriptVM) PeekRunningStack() string {
	if vm.RunningStackIsEmpty() {
		return ""
	}
	bs := vm.runningStack.peek()
	return string(bs)
}

func (vm *ScriptVM) PopRunningStack() string {
	if vm.RunningStackIsEmpty() {
		return ""
	}
	bs := vm.runningStack.pop()
	return string(bs)
}

func (vm *ScriptVM) ExecScriptBinCheckSign(scriptBin []byte,fn func([]byte,[]byte) bool) error {
	_ = vm.injectCheckSigFunc(fn)
	err := vm.ExecScriptBin(scriptBin)
	if err != nil {
		return err
	}
	_ = vm.vm.Set("checkSig", nil)
	return nil
}

func (vm *ScriptVM) injectPushFunc() error {
	return vm.vm.Set("push",func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) < 1 {
			return otto.UndefinedValue()
		}
		val := call.Argument(0)
		if !val.IsString() {
			return otto.UndefinedValue()
		}
		vm.runningStack.push([]byte(val.String()))
		return otto.TrueValue()
	})
}

func (vm *ScriptVM) injectPopFunc() error {
	return vm.vm.Set("pop",func(call otto.FunctionCall) otto.Value {
		valbs := vm.runningStack.pop()
		val, err := otto.ToValue(string(valbs))
		if err != nil {
			return otto.UndefinedValue()
		}
		return val
	})
}

func (vm *ScriptVM) injectPeekFunc() error {
	return vm.vm.Set("peek",func(call otto.FunctionCall) otto.Value {
		valbs := vm.runningStack.peek()
		val, err := otto.ToValue(string(valbs))
		if err != nil {
			return otto.UndefinedValue()
		}
		return val
	})
}

func (vm *ScriptVM) injectUpFunc() error {
	return vm.vm.Set("up",func(call otto.FunctionCall) otto.Value {
		vm.runningStack.up()
		return otto.TrueValue()
	})
}

func (vm *ScriptVM) injectPkhFunc() error {
	return vm.vm.Set("pkh",func(call otto.FunctionCall) otto.Value {
		val := vm.runningStack.pop()
		pubkey := base58.Decode(string(val))
		pkh := PubKeyHash(pubkey)
		pkhEnc := base58.Encode(pkh)
		vm.runningStack.push([]byte(pkhEnc))
		return otto.TrueValue()
	})
}

func (vm *ScriptVM) injectEqFunc() error {
	return vm.vm.Set("eq",func(call otto.FunctionCall) otto.Value {
		if vm.runningStack.top() < 1 {
			return otto.TrueValue()
		}
		val0 := vm.runningStack.pop()
		val1 := vm.runningStack.pop()
		if !bytes.Equal(val0,val1) {
			vm.runningStack.clear()
		}
		return otto.TrueValue()
	})
}


func (vm *ScriptVM) injectCheckSigFunc(fn func([]byte,[]byte) bool) error {
	return vm.vm.Set("checkSig",func(call otto.FunctionCall) otto.Value {
		if vm.runningStack.top() < 1 {
			return otto.TrueValue()
		}
		val0 := vm.runningStack.pop()
		val1 := vm.runningStack.pop()
		r := fn(val0, val1)
		if r != true {
			vm.runningStack.push([]byte(("0")))
		} else {
			vm.runningStack.push([]byte("1"))
		}
		return otto.TrueValue()
	})
}