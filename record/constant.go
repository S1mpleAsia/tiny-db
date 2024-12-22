package record

import (
	"fmt"
	"hash/fnv"
	"strings"
)

var ErrInvalidConstantType = fmt.Errorf("invalid constant type")

type Constant struct {
	ival *int32
	sval *string
}

func NewConstantWithInt(ival int32) *Constant {
	return &Constant{
		ival: &ival,
	}
}

func NewConstantWithString(sval string) *Constant {
	return &Constant{
		sval: &sval,
	}
}

func (c *Constant) AsInt() (int32, error) {
	if c.ival == nil {
		return 0, ErrInvalidConstantType
	}

	return *c.ival, nil
}

func (c *Constant) AsString() (string, error) {
	if c.sval == nil {
		return "", ErrInvalidConstantType
	}

	return *c.sval, nil
}

func (c *Constant) Equals(target *Constant) bool {
	if c.ival != nil {
		if target.ival == nil {
			return false
		}

		return c.ival == target.ival
	} else {
		if target.sval == nil {
			return false
		}

		return c.sval == target.sval
	}
}

func (c *Constant) CompareTo(target *Constant) (int, error) {
	if c.ival != nil && target.ival != nil {
		if *c.ival > *target.ival {
			return 1, nil
		} else if *c.ival < *target.ival {
			return -1, nil
		}

		return 0, nil
	}

	if c.sval != nil && target.sval != nil {
		return strings.Compare(*c.sval, *target.sval), nil
	}

	return 0, ErrInvalidConstantType
}

func (c *Constant) HashCode() int32 {
	if c.ival != nil {
		return *c.ival
	}

	if c.sval != nil {
		h := fnv.New32()
		h.Write([]byte(*c.sval))
		return int32(h.Sum32())
	}

	return 0
}

func (c *Constant) String() string {
	if c.ival != nil {
		return fmt.Sprint(*c.ival)
	}
	return fmt.Sprintf("'%s'", *c.sval)
}

func (c *Constant) AnyValue() any {
	if c.ival != nil {
		return *c.ival
	}
	if c.sval != nil {
		return *c.sval
	}
	return nil
}
