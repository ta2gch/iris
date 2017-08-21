// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package runtime

import (
	"fmt"
	"reflect"

	"github.com/ta2gch/iris/runtime/environment"
	"github.com/ta2gch/iris/runtime/ilos"
	"github.com/ta2gch/iris/runtime/ilos/class"
	"github.com/ta2gch/iris/runtime/ilos/instance"
)

func Class(local, global environment.Environment, className ilos.Instance) (ilos.Class, ilos.Instance) {
	if v, ok := global.Class.Get(className); ok {
		return v.(ilos.Class), nil
	}
	return nil, instance.NewUndefinedClass(className)
}

func checkSuperClass(a, b ilos.Class) bool {
	if reflect.DeepEqual(a, class.StandardObject) || reflect.DeepEqual(b, class.StandardObject) {
		return false
	}
	if ilos.SubclassOf(a, b) || ilos.SubclassOf(b, a) {
		return true
	}
	for _, c := range a.Supers() {
		if checkSuperClass(c, b) {
			return true
		}
	}
	for _, c := range b.Supers() {
		if checkSuperClass(a, c) {
			return true
		}
	}
	return false
}

func Defclass(local, global environment.Environment, className, scNames, slotSpecs ilos.Instance, classOpts ...ilos.Instance) (ilos.Instance, ilos.Instance) {
	if err := ensure(class.Symbol, className); err != nil {
		return nil, err
	}
	if err := ensure(class.List, scNames, slotSpecs); err != nil {
		return nil, err
	}
	supers := []ilos.Class{class.StandardObject}
	for _, scName := range scNames.(instance.List).Slice() {
		super, err := Class(local, global, scName)
		if err != nil {
			return nil, err
		}
		for _, before := range supers {
			if checkSuperClass(before, super) {
				return nil, instance.NewArityError()
			}
		}
		supers = append(supers, super.(ilos.Class))
	}
	slots := []ilos.Instance{}
	initforms := map[ilos.Instance]ilos.Instance{}
	initargs := map[ilos.Instance]ilos.Instance{}
	for _, slotSpec := range slotSpecs.(instance.List).Slice() {
		if ilos.InstanceOf(class.Symbol, slotSpec) {
			slotName := slotSpec
			slots = append(slots, slotName)
			continue
		}
		slotName := slotSpec.(*instance.Cons).Car
		slots = append(slots, slotName)
		slotOpts := slotSpec.(*instance.Cons).Cdr.(instance.List).Slice()
		for i := 0; i < len(slotOpts); i += 2 {
			switch slotOpts[i] {
			case instance.NewSymbol(":INITFORM"):
				initforms[slotName] = slotOpts[i+1]
			case instance.NewSymbol(":INITARG"):
				initargs[slotOpts[i+1]] = slotName
			}
		}
	}
	metaclass := class.StandardClass
	abstractp := Nil
	for _, classOpt := range classOpts {
		var err ilos.Instance
		switch classOpt.(*instance.Cons).Car {
		case instance.NewSymbol(":METACLASS"):
			if metaclass, err = Class(local, global, classOpt.(instance.List).Nth(1)); err != nil {
				return nil, err
			}
		case instance.NewSymbol(":ABSTRACTP"):
			if abstractp, err = Eval(local, global, classOpt.(instance.List).Nth(1)); err != nil {
				return nil, err
			}
		}
	}
	classObject := instance.NewStandardClass(className, supers, slots, initforms, initargs, metaclass, abstractp)
	global.Class.Define(className, classObject)
	for _, slotSpec := range slotSpecs.(instance.List).Slice() {
		if ilos.InstanceOf(class.Symbol, slotSpec) {
			continue
		}
		slotName := slotSpec.(*instance.Cons).Car
		slotOpts := slotSpec.(*instance.Cons).Cdr.(instance.List).Slice()
		var readerFunctionName, writerFunctionName, boundpFunctionName ilos.Instance
		for i := 0; i < len(slotOpts); i += 2 {
			switch slotOpts[i] {
			case instance.NewSymbol(":READER"):
				readerFunctionName = slotOpts[i+1]
			case instance.NewSymbol(":WRITER"):
				writerFunctionName = instance.NewSymbol(fmt.Sprintf("(SETF %v)", slotOpts[i+1]))
			case instance.NewSymbol(":ACCESSOR"):
				readerFunctionName = slotOpts[i+1]
				writerFunctionName = instance.NewSymbol(fmt.Sprintf("(SETF %v)", slotOpts[i+1]))
			case instance.NewSymbol(":BOUNDP"):
				boundpFunctionName = slotOpts[i+1]
			}
		}
		if readerFunctionName != nil {
			lambdaList, err := List(local, global, instance.NewSymbol("INSTANCE"))
			if err != nil {
				return nil, err
			}
			if _, ok := global.Function.Get(readerFunctionName); !ok {
				Defgeneric(local, global, readerFunctionName, lambdaList)
			}
			fun, _ := global.Function.Get(readerFunctionName)
			fun.(*instance.GenericFunction).AddMethod(nil, lambdaList, []ilos.Class{classObject}, instance.NewFunction(readerFunctionName, func(local, global environment.Environment, instance ilos.Instance) (ilos.Instance, ilos.Instance) {
				slot, ok := instance.GetSlotValue(slotName, classObject)
				if ok {
					return slot, nil
				}
				return Nil, nil
			}))
		}
		if writerFunctionName != nil {
			lambdaList, err := List(local, global, instance.NewSymbol("INSTANCE"))
			if err != nil {
				return nil, err
			}
			if _, ok := global.Function.Get(writerFunctionName); !ok {
				Defgeneric(local, global, writerFunctionName, lambdaList)

			}
			fun, _ := global.Function.Get(writerFunctionName)
			fun.(*instance.GenericFunction).AddMethod(nil, lambdaList, []ilos.Class{classObject}, instance.NewFunction(writerFunctionName, func(local, global environment.Environment, obj, instance ilos.Instance) (ilos.Instance, ilos.Instance) {
				ok := instance.SetSlotValue(obj, slotName, classObject)
				if ok {
					return obj, nil
				}
				return Nil, nil
			}))
		}
		if boundpFunctionName != nil {
			lambdaList, err := List(local, global, instance.NewSymbol("INSTANCE"))
			if err != nil {
				return nil, err
			}
			if _, ok := global.Function.Get(boundpFunctionName); !ok {
				Defgeneric(local, global, boundpFunctionName, lambdaList)
			}
			fun, _ := global.Function.Get(boundpFunctionName)
			fun.(*instance.GenericFunction).AddMethod(nil, lambdaList, []ilos.Class{classObject}, instance.NewFunction(boundpFunctionName, func(local, global environment.Environment, instance ilos.Instance) (ilos.Instance, ilos.Instance) {
				_, ok := instance.GetSlotValue(slotName, classObject)
				if ok {
					return T, nil
				}
				return Nil, nil
			}))
		}
	}
	return className, nil
}

func create(local, global environment.Environment, c ilos.Instance, i ...ilos.Instance) (ilos.Instance, ilos.Instance) {
	p := []ilos.Instance{}
	for _, q := range c.(ilos.Class).Supers() {
		s, err := create(local, global, q, i...)
		if err != nil {
			return nil, err
		}
		p = append(p, s)
	}
	return initializeObject(local, global, instance.NewInstance(c.(ilos.Class), p), i...)
}

func Create(local, global environment.Environment, c ilos.Instance, i ...ilos.Instance) (ilos.Instance, ilos.Instance) {
	if err := ensure(class.StandardClass, c); err != nil {
		return nil, err
	}
	//pp.Print(create(local, global, c, i...))
	return create(local, global, c, i...)
}

func initializeObject(local, global environment.Environment, object ilos.Instance, inits ...ilos.Instance) (ilos.Instance, ilos.Instance) {
	for _, super := range object.(instance.Instance).Supers {
		initializeObject(local, global, super, inits...)
	}
	for i := 0; i < len(inits); i += 2 {
		argName := inits[i]
		argValue := inits[i+1]
		if slotName, ok := object.Class().Initarg(argName); ok {
			object.SetSlotValue(slotName, argValue, object.Class())
		}
	}
	for _, slotName := range object.Class().Slots() {
		if _, ok := object.GetSlotValue(slotName, object.Class()); !ok {
			form, _ := object.Class().Initform(slotName)
			value, _ := Eval(local, global, form)
			object.SetSlotValue(slotName, value, object.Class())
		}
	}
	return object, nil
}

func InitializeObject(local, global environment.Environment, object ilos.Instance, inits ...ilos.Instance) (ilos.Instance, ilos.Instance) {
	if err := ensure(class.StandardObject, object); err != nil {
		return nil, err
	}
	return initializeObject(local, global, object, inits...)
}