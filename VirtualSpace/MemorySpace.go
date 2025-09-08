package memorySpace

import (
	"errors"
	"fmt"
	"math"
	"syscall"
)

const (
	percentRealPages = 0.2
)

var (
	memoryLimitError = errors.New("Your size bigger than allocator memory limit")
	stackOverflow    = errors.New("stackoverflow")
)

type Allocator struct {
	realSize     int
	virtualSize  int
	memoryLimit  int
	pageSize     int
	stackPointer int
	pagesArea    []*Page
}

type Page struct {
	data []byte
}

func NewAllocator(memorySize int, memoryLimitSize int, pageSize int) (*Allocator, error) {
	if memorySize > memoryLimitSize {
		return nil, memoryLimitError
	}
	alloc := &Allocator{
		memoryLimit:  memoryLimitSize,
		stackPointer: 0,
		pageSize:     pageSize,
		pagesArea:    make([]*Page, 0, 5),
	}
	pages := math.Ceil(float64(memorySize) / float64(pageSize))
	realPages := math.Ceil(pages * percentRealPages)
	virtualPages := pages - realPages
	alloc.realSize = int(realPages) * pageSize
	alloc.virtualSize = int(virtualPages) * pageSize
	for i := 0; i < int(realPages); i++ {
		data, err := syscall.Mmap(-1,
			0,
			pageSize,
			syscall.PROT_READ|syscall.PROT_WRITE,
			syscall.MAP_ANON|syscall.MAP_PRIVATE)
		if err != nil {
			return nil, err
		}
		page := &Page{data: data}
		alloc.pagesArea = append(alloc.pagesArea, page)
	}

	for i := 0; i < int(virtualPages); i++ {
		page := &Page{data: nil}
		alloc.pagesArea = append(alloc.pagesArea, page)
	}
	return alloc, nil

}

func (a *Allocator) ClearPage(address []byte) error {
	return syscall.Munmap(address)
}

func (a *Allocator) Push(value []byte) error {
	lenghtValue := len(value)
	pageSize := a.pageSize
	stackPointer := a.stackPointer
	if stackPointer+lenghtValue > a.memoryLimit {
		return stackOverflow
	}
	pageNumber := stackPointer / a.pageSize
	pageAfterPush := (stackPointer + lenghtValue) / pageSize
	pageOffset := stackPointer % pageSize
	deltaPage := pageAfterPush - pageNumber

	for i := 0; i < deltaPage; i++ {
		page := &Page{}
		a.realSize += pageSize
		a.pagesArea = append(a.pagesArea, page)
	}

	for pageAfterPush > pageNumber {
		pageNumber++
		data, err := syscall.Mmap(-1,
			0,
			pageSize,
			syscall.PROT_READ|syscall.PROT_WRITE,
			syscall.MAP_ANON|syscall.MAP_PRIVATE)
		if err != nil {
			return err
		}
		a.pagesArea[pageNumber].data = data
	}

	for i := 0; i < lenghtValue; i++ {
		offset := (pageOffset + i) % pageSize
		additionPage := i / pageSize
		a.pagesArea[pageNumber+additionPage].data[offset] = value[i]
	}
	a.stackPointer = stackPointer + lenghtValue

	return nil
}

func (a *Allocator) Pop(valueLenght int) []byte {
	pageSize := a.pageSize
	initialStackPointer := a.stackPointer
	stackPointerAfterPop := initialStackPointer - valueLenght
	a.stackPointer = stackPointerAfterPop
	stackPointer := a.stackPointer - 1
	valueIndex := valueLenght - 1
	result := make([]byte, valueLenght)

	for i := 0; i < valueLenght; i++ {
		pageNumber := stackPointer / pageSize
		pageOffset := stackPointer % pageSize
		result[valueIndex] = a.pagesArea[pageNumber].data[pageOffset]
		stackPointer--
		valueIndex--
	}
	initialPage := initialStackPointer / pageSize
	pageAfterPop := stackPointerAfterPop / pageSize

	for initialPage > pageAfterPop {
		page := a.pagesArea[initialPage]
		err := a.ClearPage(page.data)
		if err != nil {
			break
		}
		a.pagesArea = a.pagesArea[:len(a.pagesArea)-1]
		initialPage--
	}

	return result
}

func (a *Allocator) MOV(offsetValue int, value []byte) error {
	pageSize := a.pageSize
	lenghtValue := len(value)
	offsetAfterMov := offsetValue + lenghtValue

	if offsetAfterMov > a.memoryLimit {
		return stackOverflow
	}

	pageNumber := offsetValue / pageSize
	pageOffset := offsetValue % pageSize
	pageNumberAfterMov := offsetAfterMov / pageSize
	stackPageNumber := a.stackPointer / pageSize
	deltaPage := pageNumberAfterMov - stackPageNumber

	for i := 0; i < deltaPage; i++ {
		page := &Page{}
		a.pagesArea = append(a.pagesArea, page)
	}

	for pageNumberAfterMov > pageNumber {
		pageNumber++
		data, err := syscall.Mmap(-1,
			0,
			pageSize,
			syscall.PROT_READ|syscall.PROT_WRITE,
			syscall.MAP_ANON|syscall.MAP_PRIVATE)
		if err != nil {
			return err
		}
		a.pagesArea[pageNumber].data = data
	}

	for i := 0; i < lenghtValue; i++ {
		offset := (pageOffset + i) % pageSize
		additionPage := i / pageSize
		a.pagesArea[pageNumber+additionPage].data[offset] = value[i]
	}
	if offsetAfterMov > a.stackPointer {
		a.stackPointer = offsetAfterMov
	}

	return nil
}

func (a *Allocator) GetValue(offset int, lenghtData int) ([]byte, error) {
	pageSize := a.pageSize
	if offset+lenghtData > a.stackPointer { //todo: чекнуть условия
		return nil, fmt.Errorf("offset greater than stackPointer")
	}
	result := make([]byte, lenghtData)
	for i := 0; i < lenghtData; i++ {
		pageNumber := (offset + i) / pageSize
		pageOffset := (offset + i) % pageSize
		result[i] = a.pagesArea[pageNumber].data[pageOffset]
	}
	return result, nil
}

func (a *Allocator) virtualMemorySize() int {
	return a.virtualSize
}

func (a *Allocator) physicalMemorySize() int {
	return a.realSize
}
