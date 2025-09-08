package test

import (
	memorySpace "MemoryCash_GC/VirtualSpace"
	"encoding/binary"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	value  interface{}
	data   []byte
	offset int
}

const (
	attempt     = 10
	spaceSize   = 4096
	memoryLimit = 4096 * 2
	pageSize    = 1024
)

var (
	stringsData = []string{"hello", "world", "senko"}
)

func TestVirtualSpaceHappy(t *testing.T) {
	space, err := memorySpace.NewAllocator(spaceSize, memoryLimit, pageSize)
	require.NoError(t, err)
	testCases := make([]testCase, 0, attempt)
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
	t.Run("Push and Get Test", func(t *testing.T) {
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
				assert.Equal(t, binary.LittleEndian.Uint64(data), value.(uint64))
			case string:
				assert.Equal(t, string(data), value.(string))
			}
		}

	})
	t.Run("Mov and Get Test", func(t *testing.T) {
		for _, test := range testCases {
			err = space.MOV(test.offset, test.data)
			assert.NoError(t, err)
		}

		for _, test := range testCases {
			var data []byte
			data, err = space.GetValue(test.offset, len(test.data))
			assert.NoError(t, err)
			value := test.value
			switch value.(type) {
			case uint64:
				assert.Equal(t, binary.LittleEndian.Uint64(data), value.(uint64))
			case string:
				assert.Equal(t, string(data), value.(string))
			}
		}

	})

}

func TestVirtualSpaceError(t *testing.T) {
	memorySize := 512
	memoryLimitSize := 1024
	pageSizeErr := 256
	space, err := memorySpace.NewAllocator(memorySize, memoryLimitSize, pageSizeErr)
	require.NoError(t, err)
	testValue := make([]byte, 1025)
	err = space.MOV(0, testValue)
	require.Error(t, err)
}
