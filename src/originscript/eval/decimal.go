package eval

import (
	"database/sql/driver"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

const decimal_name = "decimal"

func NewDecimalObj() *DecimalObj {
	ret := &DecimalObj{}
	SetGlobalObj(decimal_name, ret)

	return ret
}

const DECIMAL_OBJ = "DECIMAL_OBJ"

type DecimalObj struct {
	Number Decimal
	Valid  bool
}

func (d *DecimalObj) Inspect() string {
	return d.Number.String()
}

func (d *DecimalObj) Type() ObjectType { return DECIMAL_OBJ }

func (d *DecimalObj) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	/* create new decimal */
	case "new":
		return d.New(line, args...)
	case "fromString":
		return d.FromString(line, args...)
	case "fromFloat":
		return d.FromFloat(line, args...)
	case "fromFloatWithExponent":
		return d.FromFloatWithExponent(line, args...)

	// aggregate
	case "avg":
		return d.Avg(line, args...)
	case "max":
		return d.Max(line, args...)
	case "min":
		return d.Min(line, args...)
	case "sum":
		return d.Sum(line, args...)

	// arithmetic operation
	case "neg":
		return d.Neg(line, args...)
	case "abs":
		return d.Abs(line, args...)
	case "add":
		return d.Add(line, args...)
	case "sub":
		return d.Sub(line, args...)
	case "mul":
		return d.Mul(line, args...)
	case "div":
		return d.Div(line, args...)
	case "divRound":
		return d.DivRound(line, args...)
	case "mod":
		return d.Mod(line, args...)
	case "pow":
		return d.Pow(line, args...)

	//round, ceil, ...
	case "ceil":
		return d.Ceil(line, args...)
	case "round":
		return d.Round(line, args...)
	case "trunc", "truncate":
		return d.Truncate(line, args...)
	case "floor":
		return d.Floor(line, args...)

	//comparation
	case "cmp":
		return d.Cmp(line, args...)
	case "equal":
		return d.Equal(line, args...)
	case "greaterThan":
		return d.GreaterThan(line, args...)
	case "greaterThanOrEqual":
		return d.GreaterThanOrEqual(line, args...)
	case "lessThan":
		return d.LessThan(line, args...)
	case "lessThanOrEqual":
		return d.LessThanOrEqual(line, args...)

	//string conversion
	case "stringFixed":
		return d.StringFixed(line, args...)
	case "stringScaled":
		return d.StringScaled(line, args...)
	case "string":
		return d.String(line, args...)

	//other
	case "sign":
		return d.Sign(line, args...)
	case "exponent":
		return d.Exponent(line, args...)
	case "intPart":
		return d.IntPart(line, args...)
	case "float":
		return d.Float(line, args...)
	case "setDivisionPrecision":
		return d.SetDivisionPrecision(line, args...)
	case "getDivisionPrecision":
		return d.GetDivisionPrecision(line, args...)
	case "setMarshalJSONWithoutQuotes":
		return d.SetMarshalJSONWithoutQuotes(line, args...)
	case "getMarshalJSONWithoutQuotes":
		return d.GetMarshalJSONWithoutQuotes(line, args...)
	}
	return NewError(line, NOMETHODERROR, method, d.Type())
}

func (d *DecimalObj) New(line string, args ...Object) Object {
	if len(args) != 2 {
		return NewError(line, ARGUMENTERROR, "2", len(args))
	}

	value, ok := args[0].(*Integer)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "new", "*Integer", args[0].Type())
	}

	exp, ok := args[1].(*Integer)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "second", "new", "*Integer", args[1].Type())
	}

	return &DecimalObj{Number: NewDec(value.Int64, int32(exp.Int64)), Valid: true}
}

func (d *DecimalObj) FromString(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	value, ok := args[0].(*String)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "fromString", "*Integer", args[0].Type())
	}

	d1, err := NewFromString(value.String)
	if err != nil {
		return NewError(line, INVALIDARG)
	}

	return &DecimalObj{Number: d1, Valid: true}
}

func (d *DecimalObj) FromFloat(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	switch input := args[0].(type) {
	case *Integer:
		return &DecimalObj{Number: NewFromInt(input.Int64), Valid: true}
	case *UInteger:
		return &DecimalObj{Number: NewFromUInt(input.UInt64), Valid: true}
	case *Float:
		return &DecimalObj{Number: NewFromFloat(input.Float64), Valid: true}
	default:
		return NewError(line, PARAMTYPEERROR, "first", "fromFloat", "*Float|*Integer|*UInteger", args[0].Type())
	}
}

func (d *DecimalObj) FromFloatWithExponent(line string, args ...Object) Object {
	if len(args) != 2 {
		return NewError(line, ARGUMENTERROR, "2", len(args))
	}

	var value float64
	switch input := args[0].(type) {
	case *Integer:
		value = float64(input.Int64)
	case *UInteger:
		value = float64(input.UInt64)
	case *Float:
		value = input.Float64
	default:
		return NewError(line, PARAMTYPEERROR, "first", "fromFloatWithExponent", "*Float|*Integer|*UInteger", args[0].Type())
	}

	exp, ok := args[1].(*Integer)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "second", "fromFloatWithExponent", "*Integer", args[1].Type())
	}

	return &DecimalObj{Number: NewFromFloatWithExponent(value, int32(exp.Int64)), Valid: true}
}

func (d *DecimalObj) Avg(line string, args ...Object) Object {
	if len(args) < 1 {
		return NewError(line, ARGUMENTERROR, ">=1", len(args))
	}

	first, ok := args[0].(*DecimalObj)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "avg", "*DecimalObj", args[0].Type())
	}

	subArgs := args[1:]
	vals := make([]Decimal, len(subArgs))
	for i, v := range subArgs {
		vals[i] = v.(*DecimalObj).Number //if not DecimalObj object, panic
	}

	ret := Avg(first.Number, vals...)

	return &DecimalObj{Number: ret, Valid: true}
}

func (d *DecimalObj) Max(line string, args ...Object) Object {
	if len(args) < 1 {
		return NewError(line, ARGUMENTERROR, ">=1", len(args))
	}

	first, ok := args[0].(*DecimalObj)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "max", "*DecimalObj", args[0].Type())
	}

	subArgs := args[1:]
	vals := make([]Decimal, len(subArgs))
	for i, v := range subArgs {
		vals[i] = v.(*DecimalObj).Number //if not DecimalObj object, panic
	}

	ret := Max(first.Number, vals...)

	return &DecimalObj{Number: ret, Valid: true}
}

func (d *DecimalObj) Min(line string, args ...Object) Object {
	if len(args) < 1 {
		return NewError(line, ARGUMENTERROR, ">=1", len(args))
	}

	first, ok := args[0].(*DecimalObj)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "min", "*DecimalObj", args[0].Type())
	}

	subArgs := args[1:]
	vals := make([]Decimal, len(subArgs))
	for i, v := range subArgs {
		vals[i] = v.(*DecimalObj).Number //if not DecimalObj object, panic
	}

	ret := Min(first.Number, vals...)

	return &DecimalObj{Number: ret, Valid: true}
}

func (d *DecimalObj) Sum(line string, args ...Object) Object {
	if len(args) < 1 {
		return NewError(line, ARGUMENTERROR, ">=1", len(args))
	}

	first, ok := args[0].(*DecimalObj)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "sum", "*DecimalObj", args[0].Type())
	}

	subArgs := args[1:]
	vals := make([]Decimal, len(subArgs))
	for i, v := range subArgs {
		vals[i] = v.(*DecimalObj).Number //if not DecimalObj object, panic
	}

	ret := Sum(first.Number, vals...)

	return &DecimalObj{Number: ret, Valid: true}
}

func (d *DecimalObj) Neg(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	ret := d.Number.Neg()
	return &DecimalObj{Number: ret, Valid: true}
}

func (d *DecimalObj) Abs(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	ret := d.Number.Abs()

	return &DecimalObj{Number: ret, Valid: true}
}

func (d *DecimalObj) Add(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	var d2 Decimal
	switch input := args[0].(type) {
	case *Integer:
		d2 = NewFromInt(input.Int64)
	case *UInteger:
		d2 = NewFromUInt(input.UInt64)
	case *Float:
		d2 = NewFromFloat(input.Float64)
	case *String:
		var err error
		d2, err = NewFromString(input.String)
		if err != nil {
			return NewError(line, INVALIDARG)
		}
	case *DecimalObj:
		d2 = input.Number
	default:
		return NewError(line, INVALIDARG)
	}

	ret := d.Number.Add(d2)

	return &DecimalObj{Number: ret, Valid: true}
}

func (d *DecimalObj) Sub(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	var d2 Decimal
	switch input := args[0].(type) {
	case *Integer:
		d2 = NewFromInt(input.Int64)
	case *UInteger:
		d2 = NewFromUInt(input.UInt64)
	case *Float:
		d2 = NewFromFloat(input.Float64)
	case *String:
		var err error
		d2, err = NewFromString(input.String)
		if err != nil {
			return NewError(line, INVALIDARG)
		}
	case *DecimalObj:
		d2 = input.Number
	default:
		return NewError(line, INVALIDARG)
	}

	ret := d.Number.Sub(d2)

	return &DecimalObj{Number: ret, Valid: true}
}

func (d *DecimalObj) Mul(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	var d2 Decimal
	switch input := args[0].(type) {
	case *Integer:
		d2 = NewFromInt(input.Int64)
	case *UInteger:
		d2 = NewFromUInt(input.UInt64)
	case *Float:
		d2 = NewFromFloat(input.Float64)
	case *String:
		var err error
		d2, err = NewFromString(input.String)
		if err != nil {
			return NewError(line, INVALIDARG)
		}
	case *DecimalObj:
		d2 = input.Number
	default:
		return NewError(line, INVALIDARG)
	}

	ret := d.Number.Mul(d2)

	return &DecimalObj{Number: ret, Valid: true}
}

func (d *DecimalObj) Div(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	var d2 Decimal
	switch input := args[0].(type) {
	case *Integer:
		d2 = NewFromInt(input.Int64)
	case *UInteger:
		d2 = NewFromUInt(input.UInt64)
	case *Float:
		d2 = NewFromFloat(input.Float64)
	case *String:
		var err error
		d2, err = NewFromString(input.String)
		if err != nil {
			return NewError(line, INVALIDARG)
		}
	case *DecimalObj:
		d2 = input.Number
	default:
		return NewError(line, INVALIDARG)
	}

	ret := d.Number.Div(d2)

	return &DecimalObj{Number: ret, Valid: true}
}

func (d *DecimalObj) DivRound(line string, args ...Object) Object {
	if len(args) != 2 {
		return NewError(line, ARGUMENTERROR, "2", len(args))
	}

	var d2 Decimal
	switch input := args[0].(type) {
	case *Integer:
		d2 = NewFromInt(input.Int64)
	case *UInteger:
		d2 = NewFromUInt(input.UInt64)
	case *Float:
		d2 = NewFromFloat(input.Float64)
	case *String:
		var err error
		d2, err = NewFromString(input.String)
		if err != nil {
			return NewError(line, INVALIDARG)
		}
	case *DecimalObj:
		d2 = input.Number
	default:
		return NewError(line, INVALIDARG)
	}

	precision, ok := args[1].(*Integer)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "second", "divRound", "*Integer", args[1].Type())
	}

	ret := d.Number.DivRound(d2, int32(precision.Int64))

	return &DecimalObj{Number: ret, Valid: true}
}

func (d *DecimalObj) Mod(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	var d2 Decimal
	switch input := args[0].(type) {
	case *Integer:
		d2 = NewFromInt(input.Int64)
	case *UInteger:
		d2 = NewFromUInt(input.UInt64)
	case *Float:
		d2 = NewFromFloat(input.Float64)
	case *String:
		var err error
		d2, err = NewFromString(input.String)
		if err != nil {
			return NewError(line, INVALIDARG)
		}
	case *DecimalObj:
		d2 = input.Number
	default:
		return NewError(line, INVALIDARG)
	}

	ret := d.Number.Mod(d2)

	return &DecimalObj{Number: ret, Valid: true}
}

func (d *DecimalObj) Pow(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	var d2 Decimal
	switch input := args[0].(type) {
	case *Integer:
		d2 = NewFromInt(input.Int64)
	case *UInteger:
		d2 = NewFromUInt(input.UInt64)
	case *Float:
		d2 = NewFromFloat(input.Float64)
	case *String:
		var err error
		d2, err = NewFromString(input.String)
		if err != nil {
			return NewError(line, INVALIDARG)
		}
	case *DecimalObj:
		d2 = input.Number
	default:
		return NewError(line, INVALIDARG)
	}

	ret := d.Number.Pow(d2)

	return &DecimalObj{Number: ret, Valid: true}
}

func (d *DecimalObj) Ceil(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	ret := d.Number.Ceil()

	return &DecimalObj{Number: ret, Valid: true}
}

func (d *DecimalObj) Round(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	var places int64
	switch o := args[0].(type) {
	case *Integer:
		places = o.Int64
	case *UInteger:
		places = int64(o.UInt64)
	default:
		return NewError(line, PARAMTYPEERROR, "first", "round", "*Integer|*UInteger", args[0].Type())
	}

	ret := d.Number.Round(int32(places))

	return &DecimalObj{Number: ret, Valid: true}
}

func (d *DecimalObj) Truncate(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	var precision int64
	switch o := args[0].(type) {
	case *Integer:
		precision = o.Int64
	case *UInteger:
		precision = int64(o.UInt64)
	default:
		return NewError(line, PARAMTYPEERROR, "first", "trunc", "*Integer|*UInteger", args[0].Type())
	}

	ret := d.Number.Truncate(int32(precision))

	return &DecimalObj{Number: ret, Valid: true}
}

func (d *DecimalObj) Floor(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	ret := d.Number.Floor()

	return &DecimalObj{Number: ret, Valid: true}
}

func (d *DecimalObj) Cmp(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	d2, ok := args[0].(*DecimalObj)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "cmp", "*DecimalObj", args[0].Type())
	}

	ret := d.Number.Cmp(d2.Number)
	return NewInteger(int64(ret))
}

func (d *DecimalObj) Equal(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	d2, ok := args[0].(*DecimalObj)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "equal", "*DecimalObj", args[0].Type())
	}

	ret := d.Number.Equal(d2.Number)
	if ret {
		return TRUE
	}
	return FALSE
}

func (d *DecimalObj) GreaterThan(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	d2, ok := args[0].(*DecimalObj)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "greaterThan", "*DecimalObj", args[0].Type())
	}

	ret := d.Number.GreaterThan(d2.Number)
	if ret {
		return TRUE
	}
	return FALSE
}

func (d *DecimalObj) GreaterThanOrEqual(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	d2, ok := args[0].(*DecimalObj)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "greaterThanOrEqual", "*DecimalObj", args[0].Type())
	}

	ret := d.Number.GreaterThanOrEqual(d2.Number)
	if ret {
		return TRUE
	}
	return FALSE
}

func (d *DecimalObj) LessThan(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	d2, ok := args[0].(*DecimalObj)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "lessThan", "*DecimalObj", args[0].Type())
	}

	ret := d.Number.LessThan(d2.Number)
	if ret {
		return TRUE
	}
	return FALSE
}

func (d *DecimalObj) LessThanOrEqual(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	d2, ok := args[0].(*DecimalObj)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "lessThanOrEqual", "*DecimalObj", args[0].Type())
	}

	ret := d.Number.LessThanOrEqual(d2.Number)
	if ret {
		return TRUE
	}
	return FALSE
}

func (d *DecimalObj) StringFixed(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	var places int64
	switch o := args[0].(type) {
	case *Integer:
		places = o.Int64
	case *UInteger:
		places = int64(o.UInt64)
	default:
		return NewError(line, PARAMTYPEERROR, "first", "stringFixed", "*Integer|*UInteger", args[0].Type())
	}

	ret := d.Number.StringFixed(int32(places))

	return NewString(ret)
}

func (d *DecimalObj) StringScaled(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	var exp int64
	switch o := args[0].(type) {
	case *Integer:
		exp = o.Int64
	case *UInteger:
		exp = int64(o.UInt64)
	default:
		return NewError(line, PARAMTYPEERROR, "first", "stringScaled", "*Integer|*UInteger", args[0].Type())
	}

	ret := d.Number.StringScaled(int32(exp))

	return NewString(ret)
}

func (d *DecimalObj) String(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	ret := d.Number.String()

	return NewString(ret)
}

func (d *DecimalObj) Sign(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	ret := d.Number.Sign()

	return NewInteger(int64(ret))
}

func (d *DecimalObj) Exponent(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	ret := d.Number.Exponent()

	return NewInteger(int64(ret))
}

func (d *DecimalObj) IntPart(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	ret := d.Number.IntPart()

	return NewInteger(ret)
}

func (d *DecimalObj) Float(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	ret, _ := d.Number.Float64()

	return NewFloat(ret)
}

func (d *DecimalObj) SetDivisionPrecision(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	var value int64
	switch o := args[0].(type) {
	case *Integer:
		value = o.Int64
	case *UInteger:
		value = int64(o.UInt64)
	default:
		return NewError(line, PARAMTYPEERROR, "first", "setDivisionPrecision", "*Integer|*UInteger", args[0].Type())
	}

	DivisionPrecision = int(value)

	return NIL
}

func (d *DecimalObj) GetDivisionPrecision(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	return NewInteger(int64(DivisionPrecision))
}

func (d *DecimalObj) SetMarshalJSONWithoutQuotes(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	value, ok := args[0].(*Boolean)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "setMarshalJSONWithoutQuotes", "*Boolean", args[0].Type())
	}

	MarshalJSONWithoutQuotes = value.Bool

	return NIL
}

func (d *DecimalObj) GetMarshalJSONWithoutQuotes(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}
	if MarshalJSONWithoutQuotes {
		return TRUE
	}
	return FALSE
}

//Implements sql's Scanner Interface.
//So when calling sql.Rows.Scan(xxx), or sql.Row.Scan(xxx), we could pass this object to `Scan` method
func (d *DecimalObj) Scan(value interface{}) error {
	if value == nil {
		d.Valid = false
		return nil
	}
	d.Valid = true
	return d.Number.Scan(value)
}

//Implements driver's Valuer Interface.
//So when calling sql.Exec(xx), we could pass this object to `Exec` method
func (d DecimalObj) Value() (driver.Value, error) {
	if !d.Valid {
		return nil, nil
	}
	return d.Number.Value()
}

//Json marshal handling
func (d *DecimalObj) MarshalJSON() ([]byte, error) {
	if !d.Valid {
		return []byte("null"), nil
	}
	return d.Number.MarshalJSON()
}

func (d *DecimalObj) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		d.Valid = false
		return nil
	}

	d.Valid = true
	return d.Number.UnmarshalJSON(b)
}

/*============================================================================
	BELOW IS THE SOURCE OF DECIMAL(https://github.com/shopspring/decimal)
 ============================================================================*/
// Package decimal implements an arbitrary precision fixed-point decimal.
//
// To use as part of a struct:
//
//     type Struct struct {
//         Number Decimal
//     }
//
// The zero-value of a Decimal is 0, as you would expect.
//
// The best way to create a new Decimal is to use decimal.NewFromString, ex:
//
//     n, err := decimal.NewFromString("-123.4567")
//     n.String() // output: "-123.4567"
//
// NOTE: This can "only" represent numbers with a maximum of 2^31 digits
// after the decimal point.

// DivisionPrecision is the number of decimal places in the result when it
// doesn't divide exactly.
//
// Example:
//
//     d1 := decimal.NewFromFloat(2).Div(decimal.NewFromFloat(3)
//     d1.String() // output: "0.6666666666666667"
//     d2 := decimal.NewFromFloat(2).Div(decimal.NewFromFloat(30000)
//     d2.String() // output: "0.0000666666666667"
//     d3 := decimal.NewFromFloat(20000).Div(decimal.NewFromFloat(3)
//     d3.String() // output: "6666.6666666666666667"
//     decimal.DivisionPrecision = 3
//     d4 := decimal.NewFromFloat(2).Div(decimal.NewFromFloat(3)
//     d4.String() // output: "0.667"
//
var DivisionPrecision = 16

// MarshalJSONWithoutQuotes should be set to true if you want the decimal to
// be JSON marshaled as a number, instead of as a string.
// WARNING: this is dangerous for decimals with many digits, since many JSON
// unmarshallers (ex: Javascript's) will unmarshal JSON numbers to IEEE 754
// double-precision floating point numbers, which means you can potentially
// silently lose precision.
var MarshalJSONWithoutQuotes = false

// Zero constant, to make computations faster.
var Zero = NewDec(0, 1)

// fiveDec used in Cash Rounding
var fiveDec = NewDec(5, 0)

var zeroInt = big.NewInt(0)
var oneInt = big.NewInt(1)
var twoInt = big.NewInt(2)
var fourInt = big.NewInt(4)
var fiveInt = big.NewInt(5)
var tenInt = big.NewInt(10)
var twentyInt = big.NewInt(20)

// Decimal represents a fixed-point decimal. It is immutable.
// number = value * 10 ^ exp
type Decimal struct {
	value *big.Int

	// NOTE(vadim): this must be an int32, because we cast it to float64 during
	// calculations. If exp is 64 bit, we might lose precision.
	// If we cared about being able to represent every possible decimal, we
	// could make exp a *big.Int but it would hurt performance and numbers
	// like that are unrealistic.
	exp int32
}

// NewDec returns a new fixed-point decimal, value * 10 ^ exp.
func NewDec(value int64, exp int32) Decimal {
	return Decimal{
		value: big.NewInt(value),
		exp:   exp,
	}
}

// NewFromString returns a new Decimal from a string representation.
//
// Example:
//
//     d, err := NewFromString("-123.45")
//     d2, err := NewFromString(".0001")
//
func NewFromString(value string) (Decimal, error) {
	originalInput := value
	var intString string
	var exp int64

	// Check if number is using scientific notation
	eIndex := strings.IndexAny(value, "Ee")
	if eIndex != -1 {
		expInt, err := strconv.ParseInt(value[eIndex+1:], 10, 32)
		if err != nil {
			if e, ok := err.(*strconv.NumError); ok && e.Err == strconv.ErrRange {
				return Decimal{}, fmt.Errorf("can't convert %s to decimal: fractional part too long", value)
			}
			return Decimal{}, fmt.Errorf("can't convert %s to decimal: exponent is not numeric", value)
		}
		value = value[:eIndex]
		exp = expInt
	}

	parts := strings.Split(value, ".")
	if len(parts) == 1 {
		// There is no decimal point, we can just parse the original string as
		// an int
		intString = value
	} else if len(parts) == 2 {
		// strip the insignificant digits for more accurate comparisons.
		decimalPart := strings.TrimRight(parts[1], "0")
		intString = parts[0] + decimalPart
		expInt := -len(decimalPart)
		exp += int64(expInt)
	} else {
		return Decimal{}, fmt.Errorf("can't convert %s to decimal: too many .s", value)
	}

	dValue := new(big.Int)
	_, ok := dValue.SetString(intString, 10)
	if !ok {
		return Decimal{}, fmt.Errorf("can't convert %s to decimal", value)
	}

	if exp < math.MinInt32 || exp > math.MaxInt32 {
		// NOTE(vadim): I doubt a string could realistically be this long
		return Decimal{}, fmt.Errorf("can't convert %s to decimal: fractional part too long", originalInput)
	}

	return Decimal{
		value: dValue,
		exp:   int32(exp),
	}, nil
}

// NewFromFloat converts a float64 to Decimal.
//
// Example:
//
//     NewFromFloat(123.45678901234567).String() // output: "123.4567890123456"
//     NewFromFloat(.00000000000000001).String() // output: "0.00000000000000001"
//
// NOTE: this will panic on NaN, +/-inf
func NewFromFloat(value float64) Decimal {
	floor := math.Floor(value)

	// fast path, where float is an int
	if floor == value && value <= math.MaxInt64 && value >= math.MinInt64 {
		return NewDec(int64(value), 0)
	}

	// slow path: float is a decimal
	// HACK(vadim): do this the slow hacky way for now because the logic to
	// convert a base-2 float to base-10 properly is not trivial
	str := strconv.FormatFloat(value, 'f', -1, 64)
	dec, err := NewFromString(str)
	if err != nil {
		panic(err)
	}
	return dec
}

// NewFromFloatWithExponent converts a float64 to Decimal, with an arbitrary
// number of fractional digits.
//
// Example:
//
//     NewFromFloatWithExponent(123.456, -2).String() // output: "123.46"
//
func NewFromFloatWithExponent(value float64, exp int32) Decimal {
	mul := math.Pow(10, -float64(exp))
	floatValue := value * mul
	if math.IsNaN(floatValue) || math.IsInf(floatValue, 0) {
		panic(fmt.Sprintf("Cannot create a Decimal from %v", floatValue))
	}
	dValue := big.NewInt(round(floatValue))

	return Decimal{
		value: dValue,
		exp:   exp,
	}
}

func NewFromInt(value int64) Decimal {
	return NewDec(value, 0)
}

func NewFromUInt(value uint64) Decimal {
	return Decimal{
		value: big.NewInt(0).SetUint64(value),
		exp:   0,
	}
}

// rescale returns a rescaled version of the decimal. Returned
// decimal may be less precise if the given exponent is bigger
// than the initial exponent of the Decimal.
// NOTE: this will truncate, NOT round
//
// Example:
//
// 	d := NewDec(12345, -4)
//	d2 := d.rescale(-1)
//	d3 := d2.rescale(-4)
//	println(d1)
//	println(d2)
//	println(d3)
//
// Output:
//
//	1.2345
//	1.2
//	1.2000
//
func (d Decimal) rescale(exp int32) Decimal {
	d.ensureInitialized()
	// NOTE(vadim): must convert exps to float64 before - to prevent overflow
	diff := math.Abs(float64(exp) - float64(d.exp))
	value := new(big.Int).Set(d.value)

	expScale := new(big.Int).Exp(tenInt, big.NewInt(int64(diff)), nil)
	if exp > d.exp {
		value = value.Quo(value, expScale)
	} else if exp < d.exp {
		value = value.Mul(value, expScale)
	}

	return Decimal{
		value: value,
		exp:   exp,
	}
}

// Abs returns the absolute value of the decimal.
func (d Decimal) Abs() Decimal {
	d.ensureInitialized()
	d2Value := new(big.Int).Abs(d.value)
	return Decimal{
		value: d2Value,
		exp:   d.exp,
	}
}

// Add returns d + d2.
func (d Decimal) Add(d2 Decimal) Decimal {
	baseScale := min(d.exp, d2.exp)
	rd := d.rescale(baseScale)
	rd2 := d2.rescale(baseScale)

	d3Value := new(big.Int).Add(rd.value, rd2.value)
	return Decimal{
		value: d3Value,
		exp:   baseScale,
	}
}

// Sub returns d - d2.
func (d Decimal) Sub(d2 Decimal) Decimal {
	baseScale := min(d.exp, d2.exp)
	rd := d.rescale(baseScale)
	rd2 := d2.rescale(baseScale)

	d3Value := new(big.Int).Sub(rd.value, rd2.value)
	return Decimal{
		value: d3Value,
		exp:   baseScale,
	}
}

// Neg returns -d.
func (d Decimal) Neg() Decimal {
	val := new(big.Int).Neg(d.value)
	return Decimal{
		value: val,
		exp:   d.exp,
	}
}

// Mul returns d * d2.
func (d Decimal) Mul(d2 Decimal) Decimal {
	d.ensureInitialized()
	d2.ensureInitialized()

	expInt64 := int64(d.exp) + int64(d2.exp)
	if expInt64 > math.MaxInt32 || expInt64 < math.MinInt32 {
		// NOTE(vadim): better to panic than give incorrect results, as
		// Decimals are usually used for money
		panic(fmt.Sprintf("exponent %v overflows an int32!", expInt64))
	}

	d3Value := new(big.Int).Mul(d.value, d2.value)
	return Decimal{
		value: d3Value,
		exp:   int32(expInt64),
	}
}

// Div returns d / d2. If it doesn't divide exactly, the result will have
// DivisionPrecision digits after the decimal point.
func (d Decimal) Div(d2 Decimal) Decimal {
	return d.DivRound(d2, int32(DivisionPrecision))
}

// QuoRem does divsion with remainder
// d.QuoRem(d2,precision) returns quotient q and remainder r such that
//   d = d2 * q + r, q an integer multiple of 10^(-precision)
//   0 <= r < abs(d2) * 10 ^(-precision) if d>=0
//   0 >= r > -abs(d2) * 10 ^(-precision) if d<0
// Note that precision<0 is allowed as input.
func (d Decimal) QuoRem(d2 Decimal, precision int32) (Decimal, Decimal) {
	d.ensureInitialized()
	d2.ensureInitialized()
	if d2.value.Sign() == 0 {
		panic("decimal division by 0")
	}
	scale := -precision
	e := int64(d.exp - d2.exp - scale)
	if e > math.MaxInt32 || e < math.MinInt32 {
		panic("overflow in decimal QuoRem")
	}
	var aa, bb, expo big.Int
	var scalerest int32
	// d = a 10^ea
	// d2 = b 10^eb
	if e < 0 {
		aa = *d.value
		expo.SetInt64(-e)
		bb.Exp(tenInt, &expo, nil)
		bb.Mul(d2.value, &bb)
		scalerest = d.exp
		// now aa = a
		//     bb = b 10^(scale + eb - ea)
	} else {
		expo.SetInt64(e)
		aa.Exp(tenInt, &expo, nil)
		aa.Mul(d.value, &aa)
		bb = *d2.value
		scalerest = scale + d2.exp
		// now aa = a ^ (ea - eb - scale)
		//     bb = b
	}
	var q, r big.Int
	q.QuoRem(&aa, &bb, &r)
	dq := Decimal{value: &q, exp: scale}
	dr := Decimal{value: &r, exp: scalerest}
	return dq, dr
}

// DivRound divides and rounds to a given precision
// i.e. to an integer multiple of 10^(-precision)
//   for a positive quotient digit 5 is rounded up, away from 0
//   if the quotient is negative then digit 5 is rounded down, away from 0
// Note that precision<0 is allowed as input.
func (d Decimal) DivRound(d2 Decimal, precision int32) Decimal {
	// QuoRem already checks initialization
	q, r := d.QuoRem(d2, precision)
	// the actual rounding decision is based on comparing r*10^precision and d2/2
	// instead compare 2 r 10 ^precision and d2
	var rv2 big.Int
	rv2.Abs(r.value)
	rv2.Lsh(&rv2, 1)
	// now rv2 = abs(r.value) * 2
	r2 := Decimal{value: &rv2, exp: r.exp + precision}
	// r2 is now 2 * r * 10 ^ precision
	var c = r2.Cmp(d2.Abs())

	if c < 0 {
		return q
	}

	if d.value.Sign()*d2.value.Sign() < 0 {
		return q.Sub(NewDec(1, -precision))
	}

	return q.Add(NewDec(1, -precision))
}

// Mod returns d % d2.
func (d Decimal) Mod(d2 Decimal) Decimal {
	quo := d.Div(d2).Truncate(0)
	return d.Sub(d2.Mul(quo))
}

// Pow returns d to the power d2
func (d Decimal) Pow(d2 Decimal) Decimal {
	var temp Decimal
	if d2.IntPart() == 0 {
		return NewFromFloat(1)
	}
	temp = d.Pow(d2.Div(NewFromFloat(2)))
	if d2.IntPart()%2 == 0 {
		return temp.Mul(temp)
	}
	if d2.IntPart() > 0 {
		return temp.Mul(temp).Mul(d)
	}
	return temp.Mul(temp).Div(d)
}

// Cmp compares the numbers represented by d and d2 and returns:
//
//     -1 if d <  d2
//      0 if d == d2
//     +1 if d >  d2
//
func (d Decimal) Cmp(d2 Decimal) int {
	d.ensureInitialized()
	d2.ensureInitialized()

	if d.exp == d2.exp {
		return d.value.Cmp(d2.value)
	}

	baseExp := min(d.exp, d2.exp)
	rd := d.rescale(baseExp)
	rd2 := d2.rescale(baseExp)

	return rd.value.Cmp(rd2.value)
}

// Equal returns whether the numbers represented by d and d2 are equal.
func (d Decimal) Equal(d2 Decimal) bool {
	return d.Cmp(d2) == 0
}

// Equals is deprecated, please use Equal method instead
func (d Decimal) Equals(d2 Decimal) bool {
	return d.Equal(d2)
}

// GreaterThan (GT) returns true when d is greater than d2.
func (d Decimal) GreaterThan(d2 Decimal) bool {
	return d.Cmp(d2) == 1
}

// GreaterThanOrEqual (GTE) returns true when d is greater than or equal to d2.
func (d Decimal) GreaterThanOrEqual(d2 Decimal) bool {
	cmp := d.Cmp(d2)
	return cmp == 1 || cmp == 0
}

// LessThan (LT) returns true when d is less than d2.
func (d Decimal) LessThan(d2 Decimal) bool {
	return d.Cmp(d2) == -1
}

// LessThanOrEqual (LTE) returns true when d is less than or equal to d2.
func (d Decimal) LessThanOrEqual(d2 Decimal) bool {
	cmp := d.Cmp(d2)
	return cmp == -1 || cmp == 0
}

// Sign returns:
//
//	-1 if d <  0
//	 0 if d == 0
//	+1 if d >  0
//
func (d Decimal) Sign() int {
	if d.value == nil {
		return 0
	}
	return d.value.Sign()
}

// Exponent returns the exponent, or scale component of the decimal.
func (d Decimal) Exponent() int32 {
	return d.exp
}

// Coefficient returns the coefficient of the decimal.  It is scaled by 10^Exponent()
func (d Decimal) Coefficient() *big.Int {
	// we copy the coefficient so that mutating the result does not mutate the
	// Decimal.
	return big.NewInt(0).Set(d.value)
}

// IntPart returns the integer component of the decimal.
func (d Decimal) IntPart() int64 {
	scaledD := d.rescale(0)
	return scaledD.value.Int64()
}

// Rat returns a rational number representation of the decimal.
func (d Decimal) Rat() *big.Rat {
	d.ensureInitialized()
	if d.exp <= 0 {
		// NOTE(vadim): must negate after casting to prevent int32 overflow
		denom := new(big.Int).Exp(tenInt, big.NewInt(-int64(d.exp)), nil)
		return new(big.Rat).SetFrac(d.value, denom)
	}

	mul := new(big.Int).Exp(tenInt, big.NewInt(int64(d.exp)), nil)
	num := new(big.Int).Mul(d.value, mul)
	return new(big.Rat).SetFrac(num, oneInt)
}

// Float64 returns the nearest float64 value for d and a bool indicating
// whether f represents d exactly.
// For more details, see the documentation for big.Rat.Float64
func (d Decimal) Float64() (f float64, exact bool) {
	return d.Rat().Float64()
}

// String returns the string representation of the decimal
// with the fixed point.
//
// Example:
//
//     d := NewDec(-12345, -3)
//     println(d.String())
//
// Output:
//
//     -12.345
//
func (d Decimal) String() string {
	return d.string(true)
}

// StringFixed returns a rounded fixed-point string with places digits after
// the decimal point.
//
// Example:
//
// 	   NewFromFloat(0).StringFixed(2) // output: "0.00"
// 	   NewFromFloat(0).StringFixed(0) // output: "0"
// 	   NewFromFloat(5.45).StringFixed(0) // output: "5"
// 	   NewFromFloat(5.45).StringFixed(1) // output: "5.5"
// 	   NewFromFloat(5.45).StringFixed(2) // output: "5.45"
// 	   NewFromFloat(5.45).StringFixed(3) // output: "5.450"
// 	   NewFromFloat(545).StringFixed(-1) // output: "550"
//
func (d Decimal) StringFixed(places int32) string {
	rounded := d.Round(places)
	return rounded.string(false)
}

// Round rounds the decimal to places decimal places.
// If places < 0, it will round the integer part to the nearest 10^(-places).
//
// Example:
//
// 	   NewFromFloat(5.45).Round(1).String() // output: "5.5"
// 	   NewFromFloat(545).Round(-1).String() // output: "550"
//
func (d Decimal) Round(places int32) Decimal {
	// truncate to places + 1
	ret := d.rescale(-places - 1)

	// add sign(d) * 0.5
	if ret.value.Sign() < 0 {
		ret.value.Sub(ret.value, fiveInt)
	} else {
		ret.value.Add(ret.value, fiveInt)
	}

	// floor for positive numbers, ceil for negative numbers
	_, m := ret.value.DivMod(ret.value, tenInt, new(big.Int))
	ret.exp++
	if ret.value.Sign() < 0 && m.Cmp(zeroInt) != 0 {
		ret.value.Add(ret.value, oneInt)
	}

	return ret
}

// Floor returns the nearest integer value less than or equal to d.
func (d Decimal) Floor() Decimal {
	d.ensureInitialized()

	if d.exp >= 0 {
		return d
	}

	exp := big.NewInt(10)

	// NOTE(vadim): must negate after casting to prevent int32 overflow
	exp.Exp(exp, big.NewInt(-int64(d.exp)), nil)

	z := new(big.Int).Div(d.value, exp)
	return Decimal{value: z, exp: 0}
}

// Ceil returns the nearest integer value greater than or equal to d.
func (d Decimal) Ceil() Decimal {
	d.ensureInitialized()

	if d.exp >= 0 {
		return d
	}

	exp := big.NewInt(10)

	// NOTE(vadim): must negate after casting to prevent int32 overflow
	exp.Exp(exp, big.NewInt(-int64(d.exp)), nil)

	z, m := new(big.Int).DivMod(d.value, exp, new(big.Int))
	if m.Cmp(zeroInt) != 0 {
		z.Add(z, oneInt)
	}
	return Decimal{value: z, exp: 0}
}

// Truncate truncates off digits from the number, without rounding.
//
// NOTE: precision is the last digit that will not be truncated (must be >= 0).
//
// Example:
//
//     decimal.NewFromString("123.456").Truncate(2).String() // "123.45"
//
func (d Decimal) Truncate(precision int32) Decimal {
	d.ensureInitialized()
	if precision >= 0 && -precision > d.exp {
		return d.rescale(-precision)
	}
	return d
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Decimal) UnmarshalJSON(decimalBytes []byte) error {
	if string(decimalBytes) == "null" {
		return nil
	}

	str, err := unquoteIfQuoted(decimalBytes)
	if err != nil {
		return fmt.Errorf("Error decoding string '%s': %s", decimalBytes, err)
	}

	decimal, err := NewFromString(str)
	*d = decimal
	if err != nil {
		return fmt.Errorf("Error decoding string '%s': %s", str, err)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (d Decimal) MarshalJSON() ([]byte, error) {
	var str string
	if MarshalJSONWithoutQuotes {
		str = d.String()
	} else {
		str = "\"" + d.String() + "\""
	}
	return []byte(str), nil
}

// Scan implements the sql.Scanner interface for database deserialization.
func (d *Decimal) Scan(value interface{}) error {
	// first try to see if the data is stored in database as a Numeric datatype
	switch v := value.(type) {

	case float32:
		*d = NewFromFloat(float64(v))
		return nil

	case float64:
		// numeric in sqlite3 sends us float64
		*d = NewFromFloat(v)
		return nil

	case int64:
		// at least in sqlite3 when the value is 0 in db, the data is sent
		// to us as an int64 instead of a float64 ...
		*d = NewDec(v, 0)
		return nil

	default:
		// default is trying to interpret value stored as string
		str, err := unquoteIfQuoted(v)
		if err != nil {
			return err
		}
		*d, err = NewFromString(str)
		return err
	}
}

// Value implements the driver.Valuer interface for database serialization.
func (d Decimal) Value() (driver.Value, error) {
	return d.String(), nil
}

// StringScaled first scales the decimal then calls .String() on it.
// NOTE: buggy, unintuitive, and DEPRECATED! Use StringFixed instead.
func (d Decimal) StringScaled(exp int32) string {
	return d.rescale(exp).String()
}

func (d Decimal) string(trimTrailingZeros bool) string {
	if d.exp >= 0 {
		return d.rescale(0).value.String()
	}

	abs := new(big.Int).Abs(d.value)
	str := abs.String()

	var intPart, fractionalPart string

	// NOTE(vadim): this cast to int will cause bugs if d.exp == INT_MIN
	// and you are on a 32-bit machine. Won't fix this super-edge case.
	dExpInt := int(d.exp)
	if len(str) > -dExpInt {
		intPart = str[:len(str)+dExpInt]
		fractionalPart = str[len(str)+dExpInt:]
	} else {
		intPart = "0"

		num0s := -dExpInt - len(str)
		fractionalPart = strings.Repeat("0", num0s) + str
	}

	if trimTrailingZeros {
		i := len(fractionalPart) - 1
		for ; i >= 0; i-- {
			if fractionalPart[i] != '0' {
				break
			}
		}
		fractionalPart = fractionalPart[:i+1]
	}

	number := intPart
	if len(fractionalPart) > 0 {
		number += "." + fractionalPart
	}

	if d.value.Sign() < 0 {
		return "-" + number
	}

	return number
}

func (d *Decimal) ensureInitialized() {
	if d.value == nil {
		d.value = new(big.Int)
	}
}

// Min returns the smallest Decimal that was passed in the arguments.
//
// To call this function with an array, you must do:
//
//     Min(arr[0], arr[1:]...)
//
// This makes it harder to accidentally call Min with 0 arguments.
func Min(first Decimal, rest ...Decimal) Decimal {
	ans := first
	for _, item := range rest {
		if item.Cmp(ans) < 0 {
			ans = item
		}
	}
	return ans
}

// Max returns the largest Decimal that was passed in the arguments.
//
// To call this function with an array, you must do:
//
//     Max(arr[0], arr[1:]...)
//
// This makes it harder to accidentally call Max with 0 arguments.
func Max(first Decimal, rest ...Decimal) Decimal {
	ans := first
	for _, item := range rest {
		if item.Cmp(ans) > 0 {
			ans = item
		}
	}
	return ans
}

// Sum returns the combined total of the provided first and rest Decimals
func Sum(first Decimal, rest ...Decimal) Decimal {
	total := first
	for _, item := range rest {
		total = total.Add(item)
	}

	return total
}

// Avg returns the average value of the provided first and rest Decimals
func Avg(first Decimal, rest ...Decimal) Decimal {
	count := NewDec(int64(len(rest)+1), 0)
	sum := Sum(first, rest...)
	return sum.Div(count)
}

func min(x, y int32) int32 {
	if x >= y {
		return y
	}
	return x
}

func round(n float64) int64 {
	if n < 0 {
		return int64(n - 0.5)
	}
	return int64(n + 0.5)
}

func unquoteIfQuoted(value interface{}) (string, error) {
	var bytes []byte

	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return "", fmt.Errorf("Could not convert value '%+v' to byte array of type '%T'",
			value, value)
	}

	// If the amount is quoted, strip the quotes
	if len(bytes) > 2 && bytes[0] == '"' && bytes[len(bytes)-1] == '"' {
		bytes = bytes[1 : len(bytes)-1]
	}
	return string(bytes), nil
}
