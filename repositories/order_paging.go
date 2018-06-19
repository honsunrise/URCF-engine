package repositories

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrOrderNotSupport     = errors.New("order not support")
	ErrPagingSize          = errors.New("paging size is too large")
	ErrNoFieldCanBeOrdered = errors.New("no field can be ordered")
	ErrFieldCannotOrder    = errors.New("field can't be ordered")
	ErrFieldCannotASC      = errors.New("field can't be ordered")
	ErrFieldCannotDESC     = errors.New("field can't be ordered")
)

type Order uint8

const (
	ASC Order = 1 << iota
	DESC
)

func (o Order) String() string {
	switch o {
	case ASC:
		return "ASC"
	case DESC:
		return "DESC"
	default:
		return "ASC"
	}
}

func ParseOrder(order string) (Order, error) {
	switch strings.ToUpper(order) {
	case "ASC":
		return ASC, nil
	case "DESC":
		return DESC, nil
	default:
		return 0, ErrOrderNotSupport
	}
}

type Sort struct {
	Name  string
	Order Order
}

type OrderPaging struct {
	CanOrderFields map[string]Order
	MaxSize        uint32
}

func (o OrderPaging) BuildPagingOrder(page uint32, size uint32, sorts []Sort) (string, error) {
	if size > o.MaxSize {
		return "", ErrPagingSize
	}
	offset := page * size
	if len(sorts) == 0 {
		return fmt.Sprintf(" LIMIT %d, %d", offset, size), nil
	} else {
		if len(o.CanOrderFields) == 0 {
			return "", ErrNoFieldCanBeOrdered
		}
		var orderStr string
		for i, sort := range sorts {
			if setOrder, ok := o.CanOrderFields[sort.Name]; ok {
				if setOrder&sort.Order != 0 {
					orderStr += fmt.Sprintf("%s %s", sort.Name, sort.Order)
					if i != len(sorts)-1 {
						orderStr += ","
					}
				} else {
					if setOrder&ASC != 0 {
						return "", ErrFieldCannotDESC
					} else {
						return "", ErrFieldCannotASC
					}
				}

			} else {
				return "", ErrFieldCannotOrder
			}
		}
		return fmt.Sprintf(" ORDER BY %s LIMIT %d, %d", orderStr, offset, size), nil
	}
}
