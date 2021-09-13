package mqrpc

import (
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
	"reflect"
	"strconv"
)

// ErrNil ErrNil
var ErrNil = errors.New("mqrpc: nil returned")

// Int Int
func Int(reply interface{}, err interface{}) (int, error) {
	switch e := err.(type) {
	case string:
		if err != "" {
			return 0, fmt.Errorf(e)
		}
	case error:
		if err != nil {
			return 0, e
		}
	}
	switch reply := reply.(type) {
	case int64:
		x := int(reply)
		if int64(x) != reply {
			return 0, strconv.ErrRange
		}
		return x, nil
	case nil:
		return 0, ErrNil
	}
	return 0, fmt.Errorf("mqrpc: unexpected type for Int, got type %T", reply)
}

// Int64 is a helper that converts a command reply to 64 bit integer. If err is
// not equal to nil, then Int returns 0, err. Otherwise, Int64 converts the
// reply to an int64 as follows:
//
//  Reply type    Result
//  integer       reply, nil
//  bulk string   parsed reply, nil
//  nil           0, ErrNil
//  other         0, error
func Int64(reply interface{}, err interface{}) (int64, error) {
	switch e := err.(type) {
	case string:
		if err != "" {
			return 0, fmt.Errorf(e)
		}
	case error:
		if err != nil {
			return 0, e
		}
	}
	switch reply := reply.(type) {
	case int64:
		return reply, nil
	case nil:
		return 0, ErrNil
	}
	return 0, fmt.Errorf("mqrpc: unexpected type for Int64, got type %T", reply)
}

// Float64 is a helper that converts a command reply to 64 bit float. If err is
// not equal to nil, then Float64 returns 0, err. Otherwise, Float64 converts
// the reply to an int as follows:
//
//  Reply type    Result
//  bulk string   parsed reply, nil
//  nil           0, ErrNil
//  other         0, error
func Float64(reply interface{}, err interface{}) (float64, error) {
	switch e := err.(type) {
	case string:
		if err != "" {
			return 0, fmt.Errorf(e)
		}
	case error:
		if err != nil {
			return 0, e
		}
	}
	switch reply := reply.(type) {
	case float64:
		return reply, nil
	case nil:
		return 0, ErrNil
	}
	return 0, fmt.Errorf("mqrpc: unexpected type for Float64, got type %T", reply)
}

// String is a helper that converts a command reply to a string. If err is not
// equal to nil, then String returns "", err. Otherwise String converts the
// reply to a string as follows:
//
//  Reply type      Result
//  bulk string     string(reply), nil
//  simple string   reply, nil
//  nil             "",  ErrNil
//  other           "",  error
func String(reply interface{}, err interface{}) (string, error) {
	switch e := err.(type) {
	case string:
		if err != "" {
			return "", fmt.Errorf(e)
		}
	case error:
		if err != nil {
			return "", e
		}
	}
	switch reply := reply.(type) {
	case string:
		return reply, nil
	case nil:
		return "", ErrNil
	}
	return "", fmt.Errorf("mqrpc: unexpected type for String, got type %T", reply)
}

// Bytes is a helper that converts a command reply to a slice of bytes. If err
// is not equal to nil, then Bytes returns nil, err. Otherwise Bytes converts
// the reply to a slice of bytes as follows:
//
//  Reply type      Result
//  bulk string     reply, nil
//  simple string   []byte(reply), nil
//  nil             nil, ErrNil
//  other           nil, error
func Bytes(reply interface{}, err interface{}) ([]byte, error) {
	switch e := err.(type) {
	case string:
		if err != "" {
			return nil, fmt.Errorf(e)
		}
	case error:
		if err != nil {
			return nil, e
		}
	}
	switch reply := reply.(type) {
	case []byte:
		return reply, nil
	case nil:
		return nil, ErrNil
	}
	return nil, fmt.Errorf("mqrpc: unexpected type for Bytes, got type %T", reply)
}

// Bool is a helper that converts a command reply to a boolean. If err is not
// equal to nil, then Bool returns false, err. Otherwise Bool converts the
// reply to boolean as follows:
//
//  Reply type      Result
//  integer         value != 0, nil
//  bulk string     strconv.ParseBool(reply)
//  nil             false, ErrNil
//  other           false, error
func Bool(reply interface{}, err interface{}) (bool, error) {
	switch e := err.(type) {
	case string:
		if err != "" {
			return false, fmt.Errorf(e)
		}
	case error:
		if err != nil {
			return false, e
		}
	}
	switch reply := reply.(type) {
	case int64:
		return reply != 0, nil
	case nil:
		return false, ErrNil
	}
	return false, fmt.Errorf("mqrpc: unexpected type for Bool, got type %T", reply)
}

// StringMap is a helper that converts an array of strings (alternating key, value)
// into a map[string]string. The HGETALL and CONFIG GET commands return replies in this format.
// Requires an even number of values in result.
func StringMap(reply interface{}, err interface{}) (map[string]string, error) {
	switch e := err.(type) {
	case string:
		if err != "" {
			return nil, fmt.Errorf(e)
		}
	case error:
		if err != nil {
			return nil, e
		}
	}
	switch reply := reply.(type) {
	case map[string]string:
		return reply, nil
	case nil:
		return nil, ErrNil
	}
	return nil, fmt.Errorf("mqrpc: unexpected type for Bool, got type %T", reply)
}

// InterfaceMap InterfaceMap
func InterfaceMap(reply interface{}, err interface{}) (map[string]interface{}, error) {
	switch e := err.(type) {
	case string:
		if err != "" {
			return nil, fmt.Errorf(e)
		}
	case error:
		if err != nil {
			return nil, e
		}
	}
	switch reply := reply.(type) {
	case map[string]interface{}:
		return reply, nil
	case nil:
		return nil, ErrNil
	}
	return nil, fmt.Errorf("mqrpc: unexpected type for Bool, got type %T", reply)
}

// Marshal Marshal
func Marshal(mrsp interface{}, ff func() (reply interface{}, err interface{})) error {
	reply, err := ff()
	switch e := err.(type) {
	case string:
		if err != "" {
			return fmt.Errorf(e)
		}
	case error:
		if err != nil {
			return e
		}
	}

	rv := reflect.ValueOf(mrsp)
	if rv.Kind() != reflect.Ptr {
		//不是指针
		return fmt.Errorf("mrsp [%v] not *mqrpc.marshaler pointer type", rv.Type())
	}
	if v2, ok := mrsp.(Marshaler); ok {
		switch r := reply.(type) {
		case []byte:
			err := v2.Unmarshal(r)
			if err != nil {
				return err
			}
			return nil
		case nil:
			return ErrNil
		}
	} else {
		return fmt.Errorf("mrsp [%v] not *mqrpc.marshaler type", rv.Type())
	}
	return fmt.Errorf("mqrpc: unexpected type for %v, got type %T", reflect.ValueOf(reply), reply)
}

// Proto Proto
func Proto(mrsp interface{}, ff func() (reply interface{}, err interface{})) error {
	reply, err := ff()
	switch e := err.(type) {
	case string:
		if err != "" {
			return fmt.Errorf(e)
		}
	case error:
		if err != nil {
			return e
		}
	}

	rv := reflect.ValueOf(mrsp)
	if rv.Kind() != reflect.Ptr {
		//不是指针
		return fmt.Errorf("mrsp [%v] not *proto.Message pointer type", rv.Type())
	}
	if v2, ok := mrsp.(proto.Message); ok {
		switch r := reply.(type) {
		case []byte:
			err := proto.Unmarshal(r, v2)
			if err != nil {
				return err
			}
			return nil
		case nil:
			return ErrNil
		}
	} else {
		return fmt.Errorf("mrsp [%v] not *proto.Message type", rv.Type())
	}
	return fmt.Errorf("mqrpc: unexpected type for %v, got type %T", reflect.ValueOf(reply), reply)
}
