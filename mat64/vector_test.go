package mat64

import (
	"reflect"
	"testing"

	"github.com/gonum/blas/blas64"
)

func TestNewVector(t *testing.T) {
	for i, test := range []struct {
		n      int
		data   []float64
		vector *Vector
	}{
		{
			n:    3,
			data: []float64{4, 5, 6},
			vector: &Vector{
				mat: blas64.Vector{
					Data: []float64{4, 5, 6},
					Inc:  1,
				},
				n: 3,
			},
		},
		{
			n:    3,
			data: nil,
			vector: &Vector{
				mat: blas64.Vector{
					Data: []float64{0, 0, 0},
					Inc:  1,
				},
				n: 3,
			},
		},
	} {
		v := NewVector(test.n, test.data)
		rows, cols := v.Dims()
		if rows != test.n {
			t.Errorf("unexpected number of rows for test %d: got: %d want: %d", i, rows, test.n)
		}
		if cols != 1 {
			t.Errorf("unexpected number of cols for test %d: got: %d want: 1", i, cols)
		}
		if !reflect.DeepEqual(v, test.vector) {
			t.Errorf("unexpected data slice for test %d: got: %v want: %v", i, v, test.vector)
		}
	}
}

func TestVectorAtSet(t *testing.T) {
	for i, test := range []struct {
		vector *Vector
	}{
		{
			vector: &Vector{
				mat: blas64.Vector{
					Data: []float64{0, 1, 2},
					Inc:  1,
				},
				n: 3,
			},
		},
		{
			vector: &Vector{
				mat: blas64.Vector{
					Data: []float64{0, 10, 10, 1, 10, 10, 2},
					Inc:  3,
				},
				n: 3,
			},
		},
	} {
		v := test.vector
		n := test.vector.n

		for _, row := range []int{-1, n} {
			panicked, message := panics(func() { v.At(row, 0) })
			if !panicked || message != ErrRowAccess.Error() {
				t.Errorf("expected panic for invalid row access for test %d n=%d r=%d", i, n, row)
			}
		}
		for _, col := range []int{-1, 1} {
			panicked, message := panics(func() { v.At(0, col) })
			if !panicked || message != ErrColAccess.Error() {
				t.Errorf("expected panic for invalid column access for test %d n=%d c=%d", i, n, col)
			}
		}

		for _, row := range []int{0, 1, n - 1} {
			if e := v.At(row, 0); e != float64(row) {
				t.Errorf("unexpected value for At(%d, 0) for test %d : got: %v want: %v", row, i, e, float64(row))
			}
		}

		for _, row := range []int{-1, n} {
			panicked, message := panics(func() { v.SetVec(row, 100) })
			if !panicked || message != ErrVectorAccess.Error() {
				t.Errorf("expected panic for invalid row access for test %d n=%d r=%d", i, n, row)
			}
		}

		for inc, row := range []int{0, 2} {
			v.SetVec(row, 100+float64(inc))
			if e := v.At(row, 0); e != 100+float64(inc) {
				t.Errorf("unexpected value for At(%d, 0) after SetVec(%[1]d, %v) for test %d: got: %v want: %[2]v", row, 100+float64(inc), i, e)
			}
		}
	}
}

func TestVectorMul(t *testing.T) {
	method := func(receiver, a, b Matrix) {
		type mulVecer interface {
			MulVec(a Matrix, b *Vector)
		}
		rd := receiver.(mulVecer)
		rd.MulVec(a, b.(*Vector))
	}
	denseComparison := func(receiver, a, b *Dense) {
		receiver.Mul(a, b)
	}
	legalSizeMulVec := func(ar, ac, br, bc int) bool {
		var legal bool
		if bc != 1 {
			legal = false
		} else {
			legal = ac == br
		}
		return legal
	}
	testTwoInput(t, "MulVec", &Vector{}, method, denseComparison, legalTypesNotVecVec, legalSizeMulVec, 1e-14)
}

func TestVectorScale(t *testing.T) {
	for i, test := range []struct {
		a     *Vector
		alpha float64
		want  *Vector
	}{
		{
			a:     NewVector(3, []float64{0, 1, 2}),
			alpha: 0,
			want:  NewVector(3, []float64{0, 0, 0}),
		},
		{
			a:     NewVector(3, []float64{0, 1, 2}),
			alpha: 1,
			want:  NewVector(3, []float64{0, 1, 2}),
		},
		{
			a:     NewVector(3, []float64{0, 1, 2}),
			alpha: -2,
			want:  NewVector(3, []float64{0, -2, -4}),
		},
		{
			a:     NewDense(3, 1, []float64{0, 1, 2}).ColView(0),
			alpha: 0,
			want:  NewVector(3, []float64{0, 0, 0}),
		},
		{
			a:     NewDense(3, 1, []float64{0, 1, 2}).ColView(0),
			alpha: 1,
			want:  NewVector(3, []float64{0, 1, 2}),
		},
		{
			a:     NewDense(3, 1, []float64{0, 1, 2}).ColView(0),
			alpha: -2,
			want:  NewVector(3, []float64{0, -2, -4}),
		},
		{
			a: NewDense(3, 3, []float64{
				0, 1, 2,
				3, 4, 5,
				6, 7, 8,
			}).ColView(1),
			alpha: -2,
			want:  NewVector(3, []float64{-2, -8, -14}),
		},
	} {
		var v Vector
		v.ScaleVec(test.alpha, test.a)
		if !reflect.DeepEqual(v.RawVector(), test.want.RawVector()) {
			t.Errorf("test %d: unexpected result for v = alpha * a: got: %v want: %v", i, v.RawVector(), test.want.RawVector())
		}

		v.CopyVec(test.a)
		v.ScaleVec(test.alpha, &v)
		if !reflect.DeepEqual(v.RawVector(), test.want.RawVector()) {
			t.Errorf("test %d: unexpected result for v = alpha * v: got: %v want: %v", i, v.RawVector(), test.want.RawVector())
		}
	}
}

func TestVectorAdd(t *testing.T) {
	for i, test := range []struct {
		a, b *Vector
		want *Vector
	}{
		{
			a:    NewVector(3, []float64{0, 1, 2}),
			b:    NewVector(3, []float64{0, 2, 3}),
			want: NewVector(3, []float64{0, 3, 5}),
		},
		{
			a:    NewVector(3, []float64{0, 1, 2}),
			b:    NewDense(3, 1, []float64{0, 2, 3}).ColView(0),
			want: NewVector(3, []float64{0, 3, 5}),
		},
		{
			a:    NewDense(3, 1, []float64{0, 1, 2}).ColView(0),
			b:    NewDense(3, 1, []float64{0, 2, 3}).ColView(0),
			want: NewVector(3, []float64{0, 3, 5}),
		},
	} {
		var v Vector
		v.AddVec(test.a, test.b)
		if !reflect.DeepEqual(v.RawVector(), test.want.RawVector()) {
			t.Errorf("unexpected result for test %d: got: %v want: %v", i, v.RawVector(), test.want.RawVector())
		}
	}
}

func TestVectorSub(t *testing.T) {
	for i, test := range []struct {
		a, b *Vector
		want *Vector
	}{
		{
			a:    NewVector(3, []float64{0, 1, 2}),
			b:    NewVector(3, []float64{0, 0.5, 1}),
			want: NewVector(3, []float64{0, 0.5, 1}),
		},
		{
			a:    NewVector(3, []float64{0, 1, 2}),
			b:    NewDense(3, 1, []float64{0, 0.5, 1}).ColView(0),
			want: NewVector(3, []float64{0, 0.5, 1}),
		},
		{
			a:    NewDense(3, 1, []float64{0, 1, 2}).ColView(0),
			b:    NewDense(3, 1, []float64{0, 0.5, 1}).ColView(0),
			want: NewVector(3, []float64{0, 0.5, 1}),
		},
	} {
		var v Vector
		v.SubVec(test.a, test.b)
		if !reflect.DeepEqual(v.RawVector(), test.want.RawVector()) {
			t.Errorf("unexpected result for test %d: got: %v want: %v", i, v.RawVector(), test.want.RawVector())
		}
	}
}

func TestVectorMulElem(t *testing.T) {
	for i, test := range []struct {
		a, b *Vector
		want *Vector
	}{
		{
			a:    NewVector(3, []float64{0, 1, 2}),
			b:    NewVector(3, []float64{0, 2, 3}),
			want: NewVector(3, []float64{0, 2, 6}),
		},
		{
			a:    NewVector(3, []float64{0, 1, 2}),
			b:    NewDense(3, 1, []float64{0, 2, 3}).ColView(0),
			want: NewVector(3, []float64{0, 2, 6}),
		},
		{
			a:    NewDense(3, 1, []float64{0, 1, 2}).ColView(0),
			b:    NewDense(3, 1, []float64{0, 2, 3}).ColView(0),
			want: NewVector(3, []float64{0, 2, 6}),
		},
	} {
		var v Vector
		v.MulElemVec(test.a, test.b)
		if !reflect.DeepEqual(v.RawVector(), test.want.RawVector()) {
			t.Errorf("unexpected result for test %d: got: %v want: %v", i, v.RawVector(), test.want.RawVector())
		}
	}
}

func TestVectorDivElem(t *testing.T) {
	for i, test := range []struct {
		a, b *Vector
		want *Vector
	}{
		{
			a:    NewVector(3, []float64{0.5, 1, 2}),
			b:    NewVector(3, []float64{0.5, 0.5, 1}),
			want: NewVector(3, []float64{1, 2, 2}),
		},
		{
			a:    NewVector(3, []float64{0.5, 1, 2}),
			b:    NewDense(3, 1, []float64{0.5, 0.5, 1}).ColView(0),
			want: NewVector(3, []float64{1, 2, 2}),
		},
		{
			a:    NewDense(3, 1, []float64{0.5, 1, 2}).ColView(0),
			b:    NewDense(3, 1, []float64{0.5, 0.5, 1}).ColView(0),
			want: NewVector(3, []float64{1, 2, 2}),
		},
	} {
		var v Vector
		v.DivElemVec(test.a, test.b)
		if !reflect.DeepEqual(v.RawVector(), test.want.RawVector()) {
			t.Errorf("unexpected result for test %d: got: %v want: %v", i, v.RawVector(), test.want.RawVector())
		}
	}
}
