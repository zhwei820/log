package log

import (
	"testing"

	"github.com/google/uuid"
)

func Benchmark_UUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		u := uuid.New()
		_ = u.String()
	}
}

func Benchmark_UUIDWithPool(b *testing.B) {
	uuid.EnableRandPool()
	for i := 0; i < b.N; i++ {
		u := uuid.New()
		_ = u.String()
	}
}

func Benchmark_UUIDV2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		u, _ := uuid.NewUUID()
		_ = u.String()
	}
}
