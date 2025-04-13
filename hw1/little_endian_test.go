package hw1_test

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

type EndianConvertible interface {
	~uint16 | ~uint32 | ~uint64
}

// Original (Big-Endian): [AA] [BB] [CC] [DD]
// Position:               3    2    1    0
// Steps:
// 1. (AA & FF) << 24: [AA] -> [AA 00 00 00] (pos 3 -> pos 0)
// 2. (BB & FF) << 8:  [BB] -> [00 BB 00 00] (pos 2 -> pos 1)
// 3. (CC & FF) >> 8:  [CC] -> [00 00 CC 00] (pos 1 -> pos 2)
// 4. (DD & FF) >> 24: [DD] -> [00 00 00 DD] (pos 0 -> pos 3)
func ToLittleEndian[T EndianConvertible](number T) T {
	switch v := any(number).(type) {
	case uint16:
		return T((v << 8) | (v >> 8))
	case uint32:
		return T(((v & 0x000000FF) << 24) | // byte 0 to 3
			((v & 0x0000FF00) << 8) | // byte 1 to 2
			((v & 0x00FF0000) >> 8) | // byte 2 to 1
			((v & 0xFF000000) >> 24)) // byte 3 to 0
	case uint64:
		return T(((v & 0x00000000000000FF) << 56) | // byte 0 to 7
			((v & 0x000000000000FF00) << 40) | // byte 1 to 6
			((v & 0x0000000000FF0000) << 24) | // byte 2 to 5
			((v & 0x00000000FF000000) << 8) | // byte 3 to 4
			((v & 0x000000FF00000000) >> 8) | // byte 4 to 3
			((v & 0x0000FF0000000000) >> 24) | // byte 5 to 2
			((v & 0x00FF000000000000) >> 40) | // byte 6 to 1
			((v & 0xFF00000000000000) >> 56)) // byte 7 to 0
	default:
		log.Panicf("unknown type: %v of %v", v, number)
		return number
	}
}

func qq[T EndianConvertible](number T) T {
	q := (number << 8) | (number >> 8)
	return q
}

func TestСonversion(t *testing.T) {
	/* uint16 */
	t.Run("uint16", func(t *testing.T) {
		tests := map[string]struct {
			number uint16
			result uint16
		}{
			"uint16 #1": {
				number: 0x0000,
				result: 0x0000,
			},
			"uint16 #2": {
				number: 0xFFFF,
				result: 0xFFFF,
			},
			"uint16 #3": {
				number: 0x1234,
				result: 0x3412,
			},
			"uint16 #4": {
				number: 0x00FF,
				result: 0xFF00,
			},
			"uint16 #5": {
				number: 0xAA55,
				result: 0x55AA,
			},
		}

		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				result := ToLittleEndian(test.number)
				assert.Equal(t, test.result, result)
			})
		}
	})
	/* uint32 */
	tests := map[string]struct {
		number uint32
		result uint32
	}{
		"test case #1": {
			number: 0x00000000,
			result: 0x00000000,
		},
		"test case #2": {
			number: 0xFFFFFFFF,
			result: 0xFFFFFFFF,
		},
		"test case #3": {
			number: 0x00FF00FF,
			result: 0xFF00FF00,
		},
		"test case #4": {
			number: 0x0000FFFF,
			result: 0xFFFF0000,
		},
		"test case #5": {
			number: 0x01020304,
			result: 0x04030201,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := ToLittleEndian(test.number)
			assert.Equal(t, test.result, result)
		})
	}
	/* uint64 */
	t.Run("uint64", func(t *testing.T) {
		tests := map[string]struct {
			number uint64
			result uint64
		}{
			"uint64 #1": {
				number: 0x0000000000000000,
				result: 0x0000000000000000,
			},
			"uint64 #2": {
				number: 0xFFFFFFFFFFFFFFFF,
				result: 0xFFFFFFFFFFFFFFFF,
			},
			"uint64 #3": {
				number: 0x00FF00FF00FF00FF,
				result: 0xFF00FF00FF00FF00,
			},
			"uint64 #4": {
				number: 0x0102030405060708,
				result: 0x0807060504030201,
			},
			"uint64 #5": {
				number: 0xAABBCCDDEEFF1122,
				result: 0x2211FFEEDDCCBBAA,
			},
		}

		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				result := ToLittleEndian(test.number)
				assert.Equal(t, test.result, result)
			})
		}
	})
}
