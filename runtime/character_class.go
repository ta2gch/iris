// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package runtime

import (
	"github.com/ta2gch/iris/runtime/environment"
	"github.com/ta2gch/iris/runtime/ilos"
	"github.com/ta2gch/iris/runtime/ilos/class"
	"github.com/ta2gch/iris/runtime/ilos/instance"
)

// Characterp returns t if obj is a character (instance of class character);
// otherwise, returns nil. obj may be any ISLISP object.
func Characterp(local, global environment.Environment, obj ilos.Instance) (ilos.Instance, ilos.Instance) {
	if instance.Of(class.Character, obj) {
		return T, nil
	}
	return Nil, nil
}

// CharEqual tests whether char1 is the same character as char2.
func CharEqual(local, global environment.Environment, char1, char2 ilos.Instance) (ilos.Instance, ilos.Instance) {
	if err := ensure(class.Character, char1, char2); err != nil {
		return nil, err
	}
	if char1 == char2 {
		return T, nil
	}
	return Nil, nil
}

// CharNotEqual if and only if they are not char=.
func CharNotEqual(local, global environment.Environment, char1, char2 ilos.Instance) (ilos.Instance, ilos.Instance) {
	ret, err := CharEqual(local, global, char1, char2)
	if err != nil {
		return nil, err
	}
	return Not(local, global, ret)
}

// CharGreaterThan tests whether char1 is greater than char2.
// An error shall be signaled if either char1 or char2 is not a character (error-id. domain-error).
func CharGreaterThan(local, global environment.Environment, char1, char2 ilos.Instance) (ilos.Instance, ilos.Instance) {
	if err := ensure(class.Character, char1, char2); err != nil {
		return nil, err
	}
	if char1.(instance.Character) > char2.(instance.Character) {
		return T, nil
	}
	return Nil, nil
}

// CharGreaterThanOrEqual tests whether char1 is greater than or equal to char2.
// An error shall be signaled if either char1 or char2 is not a character (error-id. domain-error).
func CharGreaterThanOrEqual(local, global environment.Environment, char1, char2 ilos.Instance) (ilos.Instance, ilos.Instance) {
	gt, err := CharGreaterThan(local, global, char1, char2)
	if err != nil {
		return nil, err
	}
	eq, err := CharEqual(local, global, char1, char2)
	if err != nil {
		return nil, err
	}
	if gt == Nil && eq == Nil {
		return Nil, nil
	}
	return T, nil
}

// CharLessThan tests whether char1 is less than char2.
// An error shall be signaled if either char1 or char2 is not a character (error-id. domain-error).
func CharLessThan(local, global environment.Environment, char1, char2 ilos.Instance) (ilos.Instance, ilos.Instance) {
	gt, err := CharGreaterThanOrEqual(local, global, char1, char2)
	if err != nil {
		return nil, err
	}
	return Not(local, global, gt)
}

// CharLessThanOrEqual tests whether char1 is less than or equal to char2.
// An error shall be signaled if either char1 or char2 is not a character (error-id. domain-error).
func CharLessThanOrEqual(local, global environment.Environment, char1, char2 ilos.Instance) (ilos.Instance, ilos.Instance) {
	gt, err := CharGreaterThan(local, global, char1, char2)
	if err != nil {
		return nil, err
	}
	return Not(local, global, gt)
}
