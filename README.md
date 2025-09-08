# Allocator

Allocator - это система управления виртуальной памятью на языке Go, реализующая концепцию страничной организации памяти с поддержкой виртуальных и физических страниц.

## Описание

Проект представляет собой аллокатор памяти, который:

- Управляет виртуальным адресным пространством
- Использует страничную организацию памяти
- Поддерживает операции Push/Pop для стека
- Предоставляет операции MOV для записи данных по произвольному смещению
- Автоматически выделяет физическую память при необходимости

## Структура проекта

```
MemoryCash_GC/
├── VirtualSpace/
│   └── MemorySpace.go    # Основная реализация аллокатора
├── test/
│   └── virtualSpace_test.go  # Тесты
├── go.mod               # Зависимости Go
├── go.sum              # Хеши зависимостей
└── README.md           # Документация
```

##  Установка и запуск

### Требования

- Go 1.18 или выше
- Unix-подобная операционная система (Linux, macOS) для системных вызовов mmap/munmap

### Установка зависимостей

```bash
go mod tidy
```

### Запуск тестов

```bash
go test ./...
```

### Сборка проекта

```bash
go build ./...
```

## API

### Создание аллокатора

```go
allocator, err := memorySpace.NewAllocator(memorySize, memoryLimit, pageSize)
```

**Параметры:**

- `memorySize` - размер виртуальной памяти в байтах
- `memoryLimit` - максимальный лимит памяти
- `pageSize` - размер страницы в байтах

### Основные методы

#### Push - добавление данных в стек

```go
err := allocator.Push(data []byte)
```

Добавляет данные в конец стека. Автоматически выделяет новые страницы при необходимости.

#### Pop - извлечение данных из стека

```go
data := allocator.Pop(length int)
```

Извлекает указанное количество байт из конца стека. Автоматически освобождает неиспользуемые страницы.

#### MOV - запись данных по смещению

```go
err := allocator.MOV(offset int, data []byte)
```

Записывает данные по указанному смещению в виртуальной памяти.

#### GetValue - чтение данных

```go
data, err := allocator.GetValue(offset int, length int)
```

Читает данные по указанному смещению и длине.

#### ClearPage - освобождение страницы

```go
err := allocator.ClearPage(address []byte)
```

Освобождает указанную страницу памяти.

## Примеры использования

### Базовое использование

```go
package main

import (
    "fmt"
    memorySpace "MemoryCash_GC/VirtualSpace"
)

func main() {
    // Создаем аллокатор с 4KB виртуальной памяти
    allocator, err := memorySpace.NewAllocator(4096, 8192, 1024)
    if err != nil {
        panic(err)
    }

    // Добавляем данные в стек
    data := []byte("Hello, World!")
    err = allocator.Push(data)
    if err != nil {
        panic(err)
    }

    // Читаем данные
    result, err := allocator.GetValue(0, len(data))
    if err != nil {
        panic(err)
    }

    fmt.Printf("Прочитанные данные: %s\n", string(result))
}
```

### Работа с числами

```go
package main

import (
    "encoding/binary"
    "fmt"
    memorySpace "MemoryCash_GC/VirtualSpace"
)

func main() {
    allocator, err := memorySpace.NewAllocator(4096, 8192, 1024)
    if err != nil {
        panic(err)
    }

    // Записываем число
    number := uint64(12345)
    data := make([]byte, 8)
    binary.LittleEndian.PutUint64(data, number)

    err = allocator.Push(data)
    if err != nil {
        panic(err)
    }

    // Читаем число
    result, err := allocator.GetValue(0, 8)
    if err != nil {
        panic(err)
    }

    readNumber := binary.LittleEndian.Uint64(result)
    fmt.Printf("Записанное число: %d\n", readNumber)
}
```

## Архитектура

### Страничная организация

- **Физические страницы**: 20% от общего количества страниц выделяются сразу
- **Виртуальные страницы**: 80% страниц выделяются по требованию (lazy allocation)
- **Размер страницы**: настраивается при создании аллокатора

### Управление памятью

- Использует системные вызовы `mmap`/`munmap` для работы с памятью
- Автоматическое выделение страниц при записи данных
- Автоматическое освобождение страниц при Pop операциях

### Структуры данных

```go
type Allocator struct {
    realSize     int      // Размер физической памяти
    virtualSize  int      // Размер виртуальной памяти
    memoryLimit  int      // Лимит памяти
    pageSize     int      // Размер страницы
    stackPointer int      // Указатель стека
    pagesArea    []*Page  // Массив страниц
}

type Page struct {
    data []byte  // Данные страницы
}
```

##  Ограничения

1. **Платформа**: Работает только на Unix-подобных системах (Linux, macOS)
2. **Память**: Использует системные вызовы mmap, требует соответствующих прав
3. **Размер**: Ограничен лимитом памяти, заданным при создании
4. **Потокобезопасность**: Текущая реализация не является потокобезопасной

## Тестирование

Проект включает комплексные тесты:

- **TestVirtualSpaceHappy**: Тестирует основные операции Push, MOV и GetValue
- **TestVirtualSpaceError**: Тестирует обработку ошибок при превышении лимитов

Запуск тестов:

```bash
go test -v ./test/
```




