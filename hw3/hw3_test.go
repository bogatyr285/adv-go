package main

import (
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

type COWBuffer struct {
	data []byte
	refs *int32 // для atomic
}

func NewCOWBuffer(data []byte) COWBuffer {
	refCount := int32(1)
	return COWBuffer{
		data: data,
		refs: &refCount,
	}
}

func (b *COWBuffer) Clone() COWBuffer {
	atomic.AddInt32(b.refs, 1)
	return COWBuffer{
		data: b.data,
		refs: b.refs,
	}
}

func (b *COWBuffer) Close() {
	if atomic.AddInt32(b.refs, -1) == 0 {
		b.data = nil
		b.refs = nil
	}
}

// Update - меняет буфер. если имеется несколько ссылок, копирует буффер
func (b *COWBuffer) Update(index int, value byte) bool {
	if index < 0 || index >= len(b.data) {
		return false
	}

	if atomic.LoadInt32(b.refs) > 1 {
		newData := make([]byte, len(b.data))
		copy(newData, b.data)

		atomic.AddInt32(b.refs, -1)

		newRef := int32(1)
		b.data = newData
		b.refs = &newRef
	}

	b.data[index] = value
	return true
}

func (b *COWBuffer) String() string {
	// отдаем через unsafe для сохранения исходного адреса данных
	return unsafe.String(unsafe.SliceData(b.data), len(b.data))
}

func TestCOWBuffer(t *testing.T) {
	data := []byte{'a', 'b', 'c', 'd'}
	buffer := NewCOWBuffer(data)
	defer buffer.Close()

	copy1 := buffer.Clone()
	copy2 := buffer.Clone()

	assert.Equal(t, unsafe.SliceData(data), unsafe.SliceData(buffer.data))
	assert.Equal(t, unsafe.SliceData(buffer.data), unsafe.SliceData(copy1.data))
	assert.Equal(t, unsafe.SliceData(copy1.data), unsafe.SliceData(copy2.data))

	// log.Println(unsafe.SliceData(data), unsafe.StringData(buffer.String()))
	assert.True(t, (*byte)(unsafe.SliceData(data)) == unsafe.StringData(buffer.String()))
	assert.True(t, (*byte)(unsafe.StringData(buffer.String())) == unsafe.StringData(copy1.String()))
	assert.True(t, (*byte)(unsafe.StringData(copy1.String())) == unsafe.StringData(copy2.String()))

	assert.True(t, buffer.Update(0, 'g'))
	assert.False(t, buffer.Update(-1, 'g'))
	assert.False(t, buffer.Update(4, 'g'))

	assert.True(t, reflect.DeepEqual([]byte{'g', 'b', 'c', 'd'}, buffer.data))
	assert.True(t, reflect.DeepEqual([]byte{'a', 'b', 'c', 'd'}, copy1.data))
	assert.True(t, reflect.DeepEqual([]byte{'a', 'b', 'c', 'd'}, copy2.data))

	assert.NotEqual(t, unsafe.SliceData(buffer.data), unsafe.SliceData(copy1.data))
	assert.Equal(t, unsafe.SliceData(copy1.data), unsafe.SliceData(copy2.data))

	copy1.Close()

	previous := copy2.data
	copy2.Update(0, 'f')
	current := copy2.data

	// 1 reference - don't need to copy buffer during update
	assert.Equal(t, unsafe.SliceData(previous), unsafe.SliceData(current))

	copy2.Close()
}

/* naive buffer */

type NaiveBuffer struct {
	data []byte
	mu   sync.Mutex
}

func NewNaiveBuffer(data []byte) NaiveBuffer {
	return NaiveBuffer{
		data: data,
	}
}

func (b *NaiveBuffer) Clone() NaiveBuffer {
	b.mu.Lock()
	defer b.mu.Unlock()

	newData := make([]byte, len(b.data))
	copy(newData, b.data)
	return NaiveBuffer{
		data: newData,
	}
}

func (b *NaiveBuffer) Update(index int, value byte) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if index < 0 || index >= len(b.data) {
		return false
	}
	b.data[index] = value
	return true
}

func (b *NaiveBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.data)
}

/*
BenchmarkNaiveBuffer/New-14         	1000000000	         0.3855 ns/op	       0 B/op	       0 allocs/op
BenchmarkNaiveBuffer/Clone-14       	 8067247	       152.0 ns/op	    1024 B/op	       1 allocs/op
BenchmarkNaiveBuffer/Update-14      	231057441	         5.193 ns/op	       0 B/op	       0 allocs/op
BenchmarkNaiveBuffer/String-14      	 9426692	       136.9 ns/op	    1024 B/op	       1 allocs/op
BenchmarkNaiveBuffer/ConcurrentReads-14         	 4137111	       292.1 ns/op	    1024 B/op	       1 allocs/op
BenchmarkNaiveBuffer/ConcurrentUpdates-14       	 9338636	       130.2 ns/op	       0 B/op	       0 allocs/op
*/

func BenchmarkNaiveBuffer(b *testing.B) {
	data := make([]byte, 1024)

	b.Run("New", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewNaiveBuffer(data)
		}
	})

	b.Run("Clone", func(b *testing.B) {
		buf := NewNaiveBuffer(data)
		for i := 0; i < b.N; i++ {
			_ = buf.Clone()
		}
	})

	b.Run("Update", func(b *testing.B) {
		buf := NewNaiveBuffer(data)
		for i := 0; i < b.N; i++ {
			_ = buf.Update(i%len(data), byte(i))
		}
	})

	b.Run("String", func(b *testing.B) {
		buf := NewNaiveBuffer(data)
		for i := 0; i < b.N; i++ {
			_ = buf.String()
		}
	})

	b.Run("ConcurrentReads", func(b *testing.B) {
		buf := NewNaiveBuffer(data)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = buf.String()
			}
		})
	})

	b.Run("ConcurrentUpdates", func(b *testing.B) {
		buf := NewNaiveBuffer(data)
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				_ = buf.Update(i%len(data), byte(i))
				i++
			}
		})
	})
}

/*
BenchmarkNewCOWBuffer/New-14         	1000000000	         0.3852 ns/op	       0 B/op	       0 allocs/op
BenchmarkNewCOWBuffer/Clone-14       	445438982	         2.713 ns/op	       0 B/op	       0 allocs/op
BenchmarkNewCOWBuffer/Update-14      	1000000000	         1.001 ns/op	       0 B/op	       0 allocs/op
BenchmarkNewCOWBuffer/String-14      	1000000000	         0.3855 ns/op	       0 B/op	       0 allocs/op
BenchmarkNewCOWBuffer/ConcurrentReads-14         	1000000000	         0.07182 ns/op	       0 B/op	       0 allocs/op
BenchmarkNewCOWBuffer/ConcurrentUpdates-14       	1000000000	         0.4912 ns/op	       0 B/op	       0 allocs/op
*/
func BenchmarkNewCOWBuffer(b *testing.B) {
	data := make([]byte, 1024)

	b.Run("New", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewCOWBuffer(data)
		}
	})

	b.Run("Clone", func(b *testing.B) {
		buf := NewCOWBuffer(data)
		for i := 0; i < b.N; i++ {
			_ = buf.Clone()
		}
	})

	b.Run("Update", func(b *testing.B) {
		buf := NewCOWBuffer(data)
		for i := 0; i < b.N; i++ {
			_ = buf.Update(i%len(data), byte(i))
		}
	})

	b.Run("String", func(b *testing.B) {
		buf := NewCOWBuffer(data)
		for i := 0; i < b.N; i++ {
			_ = buf.String()
		}
	})

	b.Run("ConcurrentReads", func(b *testing.B) {
		buf := NewCOWBuffer(data)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = buf.String()
			}
		})
	})

	b.Run("ConcurrentUpdates", func(b *testing.B) {
		buf := NewCOWBuffer(data)
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				_ = buf.Update(i%len(data), byte(i))
				i++
			}
		})
	})
}
