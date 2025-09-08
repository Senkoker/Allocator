package memorySpace

import (
	"fmt"
	"math"
	"syscall"
)

const (
	pageSize         = 1024 //1 kb
	percentRealPages = 0.2
)

type Allocator struct {
	RealSize     int
	VirtualSize  int
	memoryLimit  int
	stackPointer int
	pagesArea    []*Page
}
type Page struct {
	data []byte
}

func NewAllocator(memorySize int) (*Allocator, error) {
	pages := math.Ceil(float64(memorySize) / float64(pageSize))
	realPages := math.Ceil(pages * percentRealPages)
	virtualPages := pages - realPages
	alloc := &Allocator{
		stackPointer: 0,
		pagesArea:    make([]*Page, 0, 5),
	}
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

func (a *Allocator) Push(value []byte) error {
	lenghtValue := len(value)
	stackPointer := a.stackPointer
	pageNumber := stackPointer / pageSize
	pageAfterPush := (stackPointer + lenghtValue) / pageSize
	pageOffset := stackPointer % pageSize
	deltaPage := pageAfterPush - pageNumber

	for i := 0; i < deltaPage; i++ {
		page := &Page{}
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

func (a *Allocator) MOV(offsetValue int, value []byte) error {
	lenghtValue := len(value)
	offsetAfterMov := offsetValue + lenghtValue
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
	if offset+lenghtData > a.stackPointer { //todo: чекнуть условия
		return nil, fmt.Errorf("offset grather than stackPointer")
	}
	result := make([]byte, lenghtData)
	for i := 0; i < lenghtData; i++ {
		pageNumber := (offset + i) % pageSize
		pageOffset := (offset + i) / pageSize
		result[i] = a.pagesArea[pageNumber].data[pageOffset]
	}
	return result, nil
}
