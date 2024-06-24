package eval

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//for strftime function, copied from 'https://github.com/billhathaway/strftime'
var conversions = map[byte]string{
	'a': "Mon",         // day name abbreviated
	'A': "Monday",      // day name full
	'b': "Jan",         // month name abbreviated
	'C': "06",          // year/100 as a decimal number
	'd': "02",          // 2 digit day
	'D': "01/02/06",    // mm/dd/yy
	'e': "_2",          // day of the month as a decimal number (1-31); single digits are preceded by a blank
	'F': "2006-01-02",  // YYYY-MM-DD
	'H': "15",          // hours as decimal 01-24
	'I': "03",          // hours as decimal using 12 hour clock
	'M': "04",          // 2 digit minute
	'm': "01",          // month in decimal
	'n': "\n",          // newline
	'p': "PM",          // AM or PM
	'P': "pm",          // am or PM
	'r': "03:04:05 PM", // time in AM or PM notation
	'R': "15:04",       //HH:MM
	'S': "05",          // seconds as decimal
	't': "\t",          // tab
	'T': "15:04:05",    // time in 24 hour notation
	'Y': "2006",        // 4 digit year
	'z': "-0700",       // timezone offset from UTC
	'Z': "MST",         // timezone name or abbreviation
	'%': "%",
}

const time_name = "time"

type TimeObj struct {
	Tm    time.Time
	Valid bool
}

func NewTimeObj() Object {
	ret := &TimeObj{Tm: time.Now(), Valid: true}
	SetGlobalObj(time_name, ret)

	SetGlobalObj(time_name+".UTC", NewInteger(0))
	SetGlobalObj(time_name+".LOCAL", NewInteger(1))

	SetGlobalObj(time_name+".NANO_SECOND", NewInteger(int64(time.Nanosecond)))
	SetGlobalObj(time_name+".MICRO_SECOND", NewInteger(int64(time.Microsecond)))
	SetGlobalObj(time_name+".MILLI_SECOND", NewInteger(int64(time.Millisecond)))
	SetGlobalObj(time_name+".SECOND", NewInteger(int64(time.Second)))
	SetGlobalObj(time_name+".MINUTE", NewInteger(int64(time.Minute)))
	SetGlobalObj(time_name+".HOUR", NewInteger(int64(time.Hour)))

	//layout constant used in time.Format() method
	SetGlobalObj(time_name+".ANSIC", NewString("Mon Jan _2 15:04:05 2006"))
	SetGlobalObj(time_name+"UnixDate", NewString("Mon Jan _2 15:04:05 MST 2006"))
	SetGlobalObj(time_name+"RubyDate", NewString("Mon Jan 02 15:04:05 -0700 2006"))
	SetGlobalObj(time_name+"RFC822", NewString("02 Jan 06 15:04 MST"))
	SetGlobalObj(time_name+"RFC822Z", NewString("02 Jan 06 15:04 -0700")) // RFC822 with numeric zone
	SetGlobalObj(time_name+"RFC850", NewString("Monday, 02-Jan-06 15:04:05 MST"))
	SetGlobalObj(time_name+"RFC1123", NewString("Mon, 02 Jan 2006 15:04:05 MST"))
	SetGlobalObj(time_name+"RFC1123Z", NewString("Mon, 02 Jan 2006 15:04:05 -0700")) // RFC1123 with numeric zone
	SetGlobalObj(time_name+"RFC3339", NewString("2006-01-02T15:04:05Z07:00"))
	SetGlobalObj(time_name+"RFC3339Nano", NewString("2006-01-02T15:04:05.999999999Z07:00"))
	SetGlobalObj(time_name+"Kitchen", NewString("3:04PM"))
	// Handy time stamps.
	SetGlobalObj(time_name+"Stamp", NewString("Jan _2 15:04:05"))
	SetGlobalObj(time_name+"StampMilli", NewString("Jan _2 15:04:05.000"))
	SetGlobalObj(time_name+"StampMicro", NewString("Jan _2 15:04:05.000000"))
	SetGlobalObj(time_name+"StampNano", NewString("Jan _2 15:04:05.000000000"))

	return ret
}

const (
	TIME_OBJ                     = "TIME_OBJ"
	builtinDate_goDateTimeLayout = time.RFC1123 // "Mon, 02 Jan 2006 15:04:05 MST"
	builtinDate_goDateLayout     = "Mon, 02 Jan 2006"
	builtinDate_goTimeLayout     = "15:04:05 MST"
	builtinDate_Normal           = "2006-01-02 15:04:05"
)

func (t *TimeObj) Inspect() string {
	if t.Valid {
		v := t.ToStr("")
		return v.(*String).String
	}
	return "ERROR: Time is invalid"

}

func (t *TimeObj) Type() ObjectType { return TIME_OBJ }

func (t *TimeObj) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	case "utc":
		return t.UTC(line, args...)
	case "local":
		return t.Local(line, args...)
	case "unix": //to timestamp(UTC)
		return t.Unix(line, args...)
	case "unixNano": //to timestamp(UTC)
		return t.UnixNano(line, args...)
	case "unixLocal": //to timestamp(Local)
		return t.UnixLocal(line, args...)
	case "unixLocalNano": //to timestamp(Local)
		return t.UnixLocalNano(line, args...)
	case "fromEpoch":
		return t.FromEpoch(line, args...)
	case "toEpoch":
		return t.ToEpoch(line, args...)
	case "toStr":
		return t.ToStr(line, args...)
	case "toUTCStr":
		return t.ToUTCStr(line, args...)
	case "toISOStr":
		return t.ToISOStr(line, args...)
	case "toGMTStr":
		return t.ToGMTStr(line, args...)
	case "toDateStr":
		return t.ToDateStr(line, args...)
	case "toTimeStr":
		return t.ToTimeStr(line, args...)
	case "year":
		return t.Year(line, args...)
	case "fullYear":
		return t.FullYear(line, args...)
	case "month":
		return t.Month(line, args...)
	case "date":
		return t.Date(line, args...)
	case "day":
		return t.Day(line, args...)
	case "yearDay":
		return t.YearDay(line, args...)
	case "weekDay":
		return t.WeekDay(line, args...)
	case "hours":
		return t.Hours(line, args...)
	case "minutes":
		return t.Minutes(line, args...)
	case "seconds":
		return t.Seconds(line, args...)
	case "milliseconds":
		return t.Milliseconds(line, args...)
	case "add":
		return t.Add(line, args...)
	case "addDate":
		return t.AddDate(line, args...)
	case "after":
		return t.After(line, args...)
	case "appendFormat":
		return t.AppendFormat(line, args...)
	case "before":
		return t.Before(line, args...)
	case "clock":
		return t.Clock(line, args...)
	case "equal":
		return t.Equal(line, args...)
	case "format":
		return t.Format(line, args...)
	case "isoWeek":
		return t.ISOWeek(line, args...)
	case "isZero":
		return t.IsZero(line, args...)
	case "round":
		return t.Round(line, args...)
	case "sub":
		return t.Sub(line, args...)
	case "truncate":
		return t.Truncate(line, args...)
	case "parse":
		return t.Parse(line, args...)
	case "setValid":
		return t.SetValid(line, args...)
	case "sleep":
		return t.Sleep(line, args...)
	case "strftime":
		return t.Strftime(line, args...)
	}
	return NewError(line, NOMETHODERROR, method, t.Type())
}

func (t *TimeObj) UTC(line string, args ...Object) Object {
	if t.Valid == false {
		return NIL
	}

	t.Tm = t.Tm.UTC()
	return t
}

func (t *TimeObj) Local(line string, args ...Object) Object {
	if t.Valid == false {
		return NIL
	}

	t.Tm = t.Tm.Local()
	return t
}

//to timestamp(UTC)
func (t *TimeObj) Unix(line string, args ...Object) Object {
	if t.Valid == false {
		return NIL
	}

	ret := t.Tm.Unix()
	return NewInteger(ret)
}

//to timestamp(UTC)
func (t *TimeObj) UnixNano(line string, args ...Object) Object {
	if t.Valid == false {
		return NIL
	}

	ret := t.Tm.UnixNano()
	return NewInteger(ret)
}

//to timestamp(LOCAL)
func (t *TimeObj) UnixLocal(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	s := t.Tm.Format(builtinDate_Normal)
	loc, _ := time.LoadLocation("Local")                           //get local timezone
	theTime, _ := time.ParseInLocation(builtinDate_Normal, s, loc) //convert to time.time
	sr := theTime.Unix()

	return NewInteger(sr)
}

//to timestamp(LOCAL)
func (t *TimeObj) UnixLocalNano(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	s := t.Tm.Format(builtinDate_Normal)
	loc, _ := time.LoadLocation("Local")                           //get local timezone
	theTime, _ := time.ParseInLocation(builtinDate_Normal, s, loc) //convert to time.time
	sr := theTime.UnixNano()

	return NewInteger(sr)
}

func (t *TimeObj) FromEpoch(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	aType := args[0].Type()
	if aType != INTEGER_OBJ && aType != UINTEGER_OBJ && aType != FLOAT_OBJ {
		return NewError(line, PARAMTYPEERROR, "first", "fromEpoch", "*Integer|*UInteger|*Float", args[0].Type())
	}

	var f float64
	if aType == INTEGER_OBJ {
		f = float64(args[0].(*Integer).Int64)
	} else if aType == UINTEGER_OBJ {
		f = float64(args[0].(*UInteger).UInt64)
	} else {
		f = args[0].(*Float).Float64
	}

	//Should we return a new object or return existing object?
	//Here we use the latter(return existing object).
	aTime, err := epochToTime(f)
	if err != nil {
		t.Valid = false
		return t
	}

	t.Tm = aTime
	t.Valid = true

	return t
}

func (t *TimeObj) ToEpoch(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	f := timeToEpoch(t.Tm)
	return NewFloat(f)
}

func (t *TimeObj) ToStr(line string, args ...Object) Object {
	if len(args) != 0 && len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "0|1", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	if len(args) == 0 {
		return NewString(t.Tm.Format(builtinDate_Normal))
	}

	fmtStr, ok := args[0].(*String)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "toStr", "*String", args[0].Type())
	}

	return NewString(t.Tm.Format(fmtStr.String))

}

func (t *TimeObj) ToUTCStr(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	return NewString(t.Tm.Format(builtinDate_goDateTimeLayout))
}

func (t *TimeObj) ToISOStr(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	return NewString(t.Tm.Format("2006-01-02T15:04:05.000Z"))
}

func (t *TimeObj) ToGMTStr(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	return NewString(t.Tm.Format("Mon, 02 Jan 2006 15:04:05 GMT"))
}

func (t *TimeObj) ToDateStr(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	return NewString(t.Tm.Format(builtinDate_goDateLayout))
}

func (t *TimeObj) ToTimeStr(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	return NewString(t.Tm.Format(builtinDate_goTimeLayout))
}

func (t *TimeObj) Year(line string, args ...Object) Object {
	if t.Valid == false {
		return NIL
	}

	return NewInteger(int64(t.Tm.Year()))
}

func (t *TimeObj) FullYear(line string, args ...Object) Object {
	if t.Valid == false {
		return NIL
	}

	return NewInteger(int64(t.Tm.Year()))
}

func (t *TimeObj) Month(line string, args ...Object) Object {
	if t.Valid == false {
		return NIL
	}

	return NewInteger(int64(t.Tm.Month()))
}

func (t *TimeObj) Date(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	year, month, day := t.Tm.Date()
	arr := &Array{}
	arr.Members = append(arr.Members, NewInteger(int64(year)))
	arr.Members = append(arr.Members, NewInteger(int64(month)))
	arr.Members = append(arr.Members, NewInteger(int64(day)))

	return arr
}

func (t *TimeObj) YearDay(line string, args ...Object) Object {
	if t.Valid == false {
		return NIL
	}

	return NewInteger(int64(t.Tm.YearDay()))
}

func (t *TimeObj) WeekDay(line string, args ...Object) Object {
	if t.Valid == false {
		return NIL
	}

	return NewString(t.Tm.Weekday().String())
}

func (t *TimeObj) Day(line string, args ...Object) Object {
	if t.Valid == false {
		return NIL
	}

	return NewInteger(int64(t.Tm.Day()))
}

func (t *TimeObj) Hours(line string, args ...Object) Object {
	if t.Valid == false {
		return NIL
	}

	return NewInteger(int64(t.Tm.Hour()))
}

func (t *TimeObj) Minutes(line string, args ...Object) Object {
	if t.Valid == false {
		return NIL
	}

	return NewInteger(int64(t.Tm.Minute()))
}

func (t *TimeObj) Seconds(line string, args ...Object) Object {
	if t.Valid == false {
		return NIL
	}

	return NewInteger(int64(t.Tm.Second()))
}

func (t *TimeObj) Milliseconds(line string, args ...Object) Object {
	if t.Valid == false {
		return NIL
	}

	return NewInteger(int64(t.Tm.Nanosecond() / (100 * 100 * 100)))
}

func (t *TimeObj) SetValid(line string, args ...Object) Object {
	argLen := len(args)
	if argLen != 0 && argLen != 1 {
		return NewError(line, ARGUMENTERROR, "0|1", argLen)
	}

	if argLen == 0 {
		t.Tm, t.Valid = time.Time{}, true
		return t
	}

	tmObj, ok := args[0].(*TimeObj)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "setValid", "*TimeObj", args[0].Type())
	}

	t.Tm, t.Valid = tmObj.Tm, tmObj.Valid
	return t
}

func (t *TimeObj) Add(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	duration, ok := args[0].(*Integer)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "add", "*Integer", args[0].Type())
	}

	t.Tm = t.Tm.Add(time.Duration(duration.Int64))
	return t
}

func (t *TimeObj) AddDate(line string, args ...Object) Object {
	if len(args) != 3 {
		return NewError(line, ARGUMENTERROR, "3", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	years, ok := args[0].(*Integer)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "addDate", "*Integer", args[0].Type())
	}

	months, ok := args[1].(*Integer)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "second", "addDate", "*Integer", args[1].Type())
	}

	days, ok := args[2].(*Integer)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "third", "addDate", "*Integer", args[2].Type())
	}

	tm := t.Tm.AddDate(int(years.Int64), int(months.Int64), int(days.Int64))
	return &TimeObj{Tm: tm, Valid: true}
}

func (t *TimeObj) After(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	if t.Valid == false {
		return FALSE
	}

	tmObj, ok := args[0].(*TimeObj)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "after", "*TimeObj", args[0].Type())
	}

	b := t.Tm.After(tmObj.Tm)
	if b {
		return TRUE
	}
	return FALSE
}

func (t *TimeObj) AppendFormat(line string, args ...Object) Object {
	if len(args) != 2 {
		return NewError(line, ARGUMENTERROR, "2", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	s, ok := args[0].(*String)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "appendFormat", "*String", args[0].Type())
	}

	layout, ok := args[1].(*String)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "second", "appendFormat", "*String", args[1].Type())
	}

	ret := t.Tm.AppendFormat([]byte(s.String), layout.String)
	return NewString(string(ret))
}

func (t *TimeObj) Before(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	if t.Valid == false {
		return FALSE
	}

	tmObj, ok := args[0].(*TimeObj)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "before", "*TimeObj", args[0].Type())
	}

	b := t.Tm.Before(tmObj.Tm)
	if b {
		return TRUE
	}
	return FALSE
}

func (t *TimeObj) Clock(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	hour, min, sec := t.Tm.Clock()
	arr := &Array{}
	arr.Members = append(arr.Members, NewInteger(int64(hour)))
	arr.Members = append(arr.Members, NewInteger(int64(min)))
	arr.Members = append(arr.Members, NewInteger(int64(sec)))

	return arr
}

func (t *TimeObj) Equal(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	if t.Valid == false {
		return FALSE
	}

	tmObj, ok := args[0].(*TimeObj)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "equal", "*TimeObj", args[0].Type())
	}

	if !t.Valid || !tmObj.Valid {
		return FALSE
	}

	b := t.Tm.Equal(tmObj.Tm)
	if b {
		return TRUE
	}
	return FALSE
}

func (t *TimeObj) Format(line string, args ...Object) Object {
	if len(args) != 0 && len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "0|1", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	if len(args) == 0 {
		str := t.Tm.Format(builtinDate_Normal)
		return NewString(str)
	}

	layout, ok := args[0].(*String)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "format", "*String", args[0].Type())
	}

	str := t.Tm.Format(layout.String)
	return NewString(str)
}

func (t *TimeObj) ISOWeek(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	year, week := t.Tm.ISOWeek()
	arr := &Array{}
	arr.Members = append(arr.Members, NewInteger(int64(year)))
	arr.Members = append(arr.Members, NewInteger(int64(week)))

	return arr
}

func (t *TimeObj) IsZero(line string, args ...Object) Object {
	if len(args) != 0 {
		return NewError(line, ARGUMENTERROR, "0", len(args))
	}

	if t.Valid == false {
		return FALSE
	}

	b := t.Tm.IsZero()
	if b {
		return TRUE
	}
	return FALSE
}

func (t *TimeObj) Round(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	duration, ok := args[0].(*Integer)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "round", "*Integer", args[0].Type())
	}

	t.Tm = t.Tm.Round(time.Duration(duration.Int64))
	return t
}

func (t *TimeObj) Sub(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	tmObj, ok := args[0].(*TimeObj)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "sub", "*TimeObj", args[0].Type())
	}

	duration := t.Tm.Sub(tmObj.Tm)
	return NewInteger(int64(duration))
}

func (t *TimeObj) Truncate(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	if t.Valid == false {
		return NIL
	}

	duration, ok := args[0].(*Integer)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "truncate", "*Integer", args[0].Type())
	}

	t.Tm = t.Tm.Round(time.Duration(duration.Int64))
	return t
}

func (t *TimeObj) Parse(line string, args ...Object) Object {
	if len(args) != 2 {
		return NewError(line, ARGUMENTERROR, "2", len(args))
	}

	layout, ok := args[0].(*String)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "parse", "*String", args[0].Type())
	}

	value, ok := args[1].(*String)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "second", "parse", "*String", args[1].Type())
	}

	ret, err := time.Parse(layout.String, value.String)
	if err != nil {
		return NIL
	}
	return &TimeObj{Tm: ret, Valid: true}
}

func (t *TimeObj) Sleep(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	duration, ok := args[0].(*Integer)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "sleep", "*Integer", args[0].Type())
	}

	time.Sleep(time.Duration(duration.Int64))
	return NIL
}

func (t *TimeObj) Strftime(line string, args ...Object) Object {
	if len(args) != 1 {
		return NewError(line, ARGUMENTERROR, "1", len(args))
	}

	formatStr, ok := args[0].(*String)
	if !ok {
		return NewError(line, PARAMTYPEERROR, "first", "strftime", "*String", args[0].Type())
	}

	format := formatStr.String
	buf := bytes.Buffer{}
	var special bool
	for i := range format {
		ch := format[i]
		if special {
			val, ok := conversions[ch]
			if !ok {
				return NewNil(fmt.Sprintf("unknown conversion specifier '%%%c'", ch))
			}
			buf.WriteString(val)
			special = false
			continue
		}
		if ch == '%' {
			special = true
			continue
		}
		buf.WriteByte(ch)
	}

	return NewString(buf.String())
}

func (t *TimeObj) Scan(value interface{}) error {
	if value == nil {
		t.Tm, t.Valid = time.Time{}, false
		return nil
	}

	if tm, ok := value.(time.Time); ok {
		t.Valid = true
		t.Tm = tm
		return nil
	}

	typ := reflect.TypeOf(value)
	return fmt.Errorf("Cannot convert %s to time type", typ.Name())
}

// Value returns time value, isValid and error object
func (t TimeObj) Value() (time.Time, bool, error) {
	if !t.Valid {
		return time.Time{}, false, nil
	}
	return t.Tm, true, nil
}

func (t *TimeObj) UnmarshalJSON(buf []byte) error {
	t.Valid = true
	return json.Unmarshal(buf, &t.Tm)
}

func (t *TimeObj) MarshalJSON() ([]byte, error) {
	if !t.Valid || t.Tm.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Tm)
}

func epochToTime(value float64) (tm time.Time, err error) {
	epochWithMilli := value
	if math.IsNaN(epochWithMilli) || math.IsInf(epochWithMilli, 0) {
		err = fmt.Errorf("Invalid time %v", value)
		return
	}

	epoch := int64(epochWithMilli / 1000)
	milli := int64(epochWithMilli) % 1000

	tm = time.Unix(int64(epoch), milli*1000000).UTC()
	return
}

func timeToEpoch(time time.Time) float64 {
	return float64(time.UnixNano() / (1000 * 1000))
}

//ParseDuration with support for Y(year), M(month) and D(day)
func ParseDuration(t *TimeObj, s string) (*TimeObj, error) {
	result := &TimeObj{Tm: t.Tm, Valid: true}

	s, years, err := splitDuration(s, "Y")
	if err != nil {
		return result, err
	}

	s, months, err := splitDuration(s, "M")
	if err != nil {
		return result, err
	}

	s, days, err := splitDuration(s, "D")
	if err != nil {
		return result, err
	}

	result.Tm = result.Tm.AddDate(years, months, days)
	if s == "" {
		return result, nil
	}

	little, err := time.ParseDuration(s)
	if err != nil {
		return result, err
	}

	result.Tm = result.Tm.Add(little)
	return result, nil
}

func splitDuration(s string, delim string) (string, int, error) {
	elems := strings.Split(s, delim)
	if len(elems) == 2 {
		i64, err := strconv.ParseInt(elems[0], 10, 64)
		if err != nil {
			return "", 0, err
		}
		return elems[len(elems)-1], int(i64), nil
	}

	return s, 0, nil
}
