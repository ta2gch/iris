package core

import (
	"errors"
	"fmt"

	"github.com/ta2gch/gazelle/core/class"
	"github.com/ta2gch/gazelle/core/class/cons"
	"github.com/ta2gch/gazelle/core/class/function"
	env "github.com/ta2gch/gazelle/core/environment"
)

func evalArguments(args *class.Instance, local *env.Environment, dynamic *env.Environment, global *env.Environment) (*class.Instance, error) {
	if args.Class() == class.Null {
		return class.Null.New(nil), nil
	}
	car, err := cons.Car(args)
	if err != nil {
		return nil, err
	}
	cdr, err := cons.Cdr(args)
	if err != nil {
		return nil, err
	}
	a, err := Eval(car, local, dynamic, global)
	if err != nil {
		return nil, err
	}
	b, err := evalArguments(cdr, local, dynamic, global)
	if err != nil {
		return nil, err
	}
	return cons.New(a, b), nil
}

func evalFunction(obj *class.Instance, local *env.Environment, dynamic *env.Environment, global *env.Environment) (*class.Instance, error) {
	// get function symbol
	car, err := cons.Car(obj)
	if err != nil {
		return nil, err
	}
	if car.Class() != class.Symbol {
		return nil, fmt.Errorf("%v is not a symbol", obj.Value())
	}
	// get function arguments
	cdr, err := cons.Cdr(obj)
	if err != nil {
		return nil, err
	}
	// get function instance has value of Function interface
	var fun *class.Instance
	if f, ok := local.Function[car.Value().(string)]; ok {
		fun = f
	}
	if f, ok := global.Function[car.Value().(string)]; ok {
		fun = f
	}
	if fun == nil {
		return nil, fmt.Errorf("%v is not defined", obj.Value())
	}
	// evaluate each arguments
	a, err := evalArguments(cdr, local, dynamic, global)
	if err != nil {
		return nil, err
	}
	// keep what dynamic envrionment has.
	ks := []string{}
	for k := range dynamic.Variable {
		ks = append(ks, k)
	}
	// apply function to arguments
	r, err := function.Apply(fun, a, env.New(), dynamic, global)
	if err != nil {
		return nil, err
	}
	// remove dynamic variables defined by function called in this time
	for k := range dynamic.Variable {
		v := true
		for _, l := range ks {
			if k == l {
				v = false
			}
		}
		if v {
			delete(dynamic.Variable, k)
		}
	}
	return r, nil
}

// Eval evaluates any classs
func Eval(obj *class.Instance, local *env.Environment, dynamic *env.Environment, global *env.Environment) (*class.Instance, error) {
	if obj.Class() == class.Null {
		return class.Null.New(nil), nil
	}
	switch obj.Class() {
	case class.Symbol:
		if val, ok := local.Variable[obj.Value().(string)]; ok {
			return val, nil
		}
		if val, ok := global.Variable[obj.Value().(string)]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("%v is not defined", obj.Value())
	case class.List:
		ret, err := evalFunction(obj, local, dynamic, global)
		if err != nil {
			return nil, err
		}
		return ret, nil
	case class.Integer, class.Float, class.Character, class.String:
		return obj, nil
	}
	return nil, errors.New("I have no ideas")
}
