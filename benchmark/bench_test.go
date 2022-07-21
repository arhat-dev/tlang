package benchmark

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"text/template"

	"arhat.dev/tlang"
)

func pseudoFunc() int { return 0 }

func NewFuncMaps[T template.FuncMap | tlang.FuncMap]() T {
	const (
		nFuncs = 1000
	)

	fm := make(map[string]any, nFuncs)
	for i := 0; i < nFuncs; i++ {
		fm[fmt.Sprintf("fn_%d", i)] = pseudoFunc
	}

	return T(fm)
}

func BenchmarkParsing_text(b *testing.B) {
	var (
		buf strings.Builder
	)

	b.ReportAllocs()

	fm := NewFuncMaps[template.FuncMap]()
	for k := range fm {
		buf.WriteString("{{")
		buf.WriteString(k)
		buf.WriteString("}}")
	}

	input := buf.String()
	tpl := template.New("").Funcs(fm)
	for i := 0; i < 10; i++ {
		_, err := tpl.Parse(input)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.Run("parse", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tpl.Parse(input)
		}
	})

	b.Run("exec", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tpl.Execute(io.Discard, nil)
		}
	})
}

func BenchmarkParsing_tlang(b *testing.B) {
	var (
		buf strings.Builder
	)

	fm := NewFuncMaps[tlang.FuncMap]()
	for k := range fm {
		buf.WriteString(k)
		buf.WriteString("\n")
	}

	input := buf.String()
	tpl := tlang.New("").Funcs(fm)
	for i := 0; i < 10; i++ {
		_, err := tpl.Parse(input)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.Run("parse", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tpl.Parse(input)
		}
	})

	b.Run("exec", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tpl.Execute(io.Discard, nil)
		}
	})

	return
}
