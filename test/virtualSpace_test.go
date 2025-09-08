package test

import (
	memorySpace "MemoryCash_GC/VirtualSpace"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand/v2"
	"testing"
)

type testCase struct {
	value  interface{}
	data   []byte
	offset int
}

const (
	attempt   = 10
	spaceSize = 4096
)

var (
	stringsData = []string{"hello", "world", "senko"}
)

func TestVirtualSpaceHappy(t *testing.T) {
	space, err := memorySpace.NewAllocator(spaceSize)
	require.NoError(t, err)
	testCases := []testCase{}
	offset := 0
	for i := 0; i < attempt; i++ {
		if i%2 == 0 {
			dataByte := make([]byte, 8)
			number := uint64(rand.Int())
			binary.LittleEndian.PutUint64(dataByte, number)
			caseTest := testCase{
				value:  number,
				data:   dataByte,
				offset: offset,
			}
			testCases = append(testCases, caseTest)
			offset += len(dataByte)
		} else {
			word := stringsData[uint64(rand.IntN(len(stringsData)-1))]
			dataByte := []byte(word)
			caseTest := testCase{
				value:  word,
				data:   dataByte,
				offset: offset,
			}
			testCases = append(testCases, caseTest)
			offset += len(dataByte)
		}

	}

	for _, test := range testCases {
		err = space.Push(test.data)
		assert.NoError(t, err)
	}

	for _, test := range testCases {
		var data []byte
		data, err = space.GetValue(test.offset, len(test.data))
		assert.NoError(t, err)
		value := test.value
		switch value.(type) {
		case uint64:
			assert.Equal(t, binary.LittleEndian.Uint64(data), value.(string))
		case string:
			assert.Equal(t, string(data), value.(string))
		}
	}

}
