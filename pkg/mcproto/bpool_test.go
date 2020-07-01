package mcproto

import (
	"sync"
	"testing"
)

func BenchmarkNewBuffer(b *testing.B) {
	p := &BufferPool{
		pool: &sync.Pool{
			New: NewBuffer,
		},
	}

	for i := 0; i < b.N; i++ {
		p.Put(p.Get())
	}
}
