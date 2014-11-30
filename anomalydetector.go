package anomalydetector

import (
	"log"
	"math"
	"math/rand"
	"time"
)

type AnomalyDetector struct {
	Term      int
	R         float64
	last      float64
	lastScore float64
	lastProbe float64
	Mu        float64
	Sigma     float64
	C         []float64
	Data      []float64
	DataSize  int
	Identity  []float64
	Matrix    []float64
	Multiply  []float64
}

func NewAnomalyDetector(term int, r float64) *AnomalyDetector {
	return NewAnomalyDetectorWithSource(term, r, rand.NewSource(time.Now().UnixNano()))
}

func NewAnomalyDetectorWithSource(term int, r float64, source rand.Source) *AnomalyDetector {
	v := AnomalyDetector{Term: term, R: r, Mu: 0.0, Sigma: 0, last: 0, lastScore: 0}
	v.C = make([]float64, term)
	v.Data = make([]float64, term)
	v.DataSize = 0
	v.Identity = make([]float64, term*term)
	v.Matrix = make([]float64, term*term)
	v.Multiply = make([]float64, term)

	rng := rand.New(source)
	for i := 0; i < term; i++ {
		v.C[i] = rng.Float64()
	}

	return &v
}

func (finder *AnomalyDetector) Update(x float64) float64 {
	if finder.last == x {
		return finder.lastScore
	}

	// clear
	for i := range finder.Identity {
		finder.Identity[i] = 0
	}
	for i := range finder.Matrix {
		finder.Matrix[i] = 0
	}
	for i := range finder.Multiply {
		finder.Multiply[i] = 0
	}

	length := finder.DataSize
	finder.Mu = ((1.0 - finder.R) * finder.Mu) + (finder.R * x)

	for j := 0; j < finder.Term; j++ {
		t := ((length - 1) - j)
		if t < 0 {
			t = length + t
		}
		if t < 0 {
			continue
		}
		if t <= finder.DataSize && finder.DataSize > 0 {
			finder.C[j] = (1.0-finder.R)*finder.C[j] + finder.R*(x-finder.Mu)*(finder.Data[t]-finder.Mu)
		}
	}

	// zero_matrix
	for i := range finder.Matrix {
		finder.Matrix[i] = 0
	}

	for j := 0; j < finder.Term; j++ {
		for i := j; i < finder.Term; i++ {
			v := finder.C[i-j]
			finder.Matrix[i+j*finder.Term] = v
			finder.Matrix[j+i*finder.Term] = v
		}
	}
	inverse(finder.Term, &finder.Matrix, &finder.Identity)
	multiply(finder, &finder.Identity, &finder.Multiply)

	xt := finder.Mu
	for i := 0; i < finder.DataSize; i++ {
		xt += finder.Multiply[i] * (finder.Data[i] - finder.Mu)
	}

	finder.Sigma = (1-finder.R)*finder.Sigma + finder.R*(x-xt)*(x-xt)
	shift(finder)
	finder.Data[finder.DataSize] = x
	finder.DataSize++
	finder.last = x

	probe := probe(finder, xt, x)
	result := score(probe)

	if !math.IsNaN(result) {
		finder.lastScore = result
	} else {
		result = finder.lastScore
	}

	return result
}

func inverse(size int, real, result *[]float64) {
	var v float64
	var identity []float64 = *result
	var target []float64 = *real

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if i == j {
				v = 1.0
			} else {
				v = 0
			}

			identity[j+i*size] = v
		}
	}

	for i := 0; i < size; i++ {
		buf := 1 / target[i+i*size]

		for j := 0; j < size; j++ {
			target[j+i*size] *= buf
			identity[j+i*size] *= buf
		}

		for j := 0; j < size; j++ {
			if i != j {
				buf = target[i+j*size]
				for k := 0; k < size; k++ {
					target[k+j*size] -= target[k+i*size] * buf
					identity[k+j*size] -= identity[k+i*size] * buf
				}
			}
		}
	}
}

func multiply(finder *AnomalyDetector, identity, result *[]float64) {
	n := finder.Term
	var tmp []float64 = *result
	var cc []float64 = *identity

	for k := 0; k < n; k++ {
		tmp[k] = 0
		for x := 0; x < n; x++ {
			tmp[k] += cc[x+k*n] * finder.C[x]
		}
	}
}

func shift(finder *AnomalyDetector) {
	if finder.DataSize+1 > finder.Term {
		for x := 0; x < finder.Term-1; x++ {
			finder.Data[x] = finder.Data[x+1]
		}
		finder.Data[finder.Term-1] = 0.0
		finder.DataSize--
	}
}

func probe(finder *AnomalyDetector, mu, v float64) float64 {
	var result float64 = 0.0

	if finder.Sigma == 0.0 {
		return result
	}

	result = math.Exp(-0.5*math.Pow((v-mu), 2)/finder.Sigma) / (math.Pow((2*math.Pi), 0.5) * math.Pow(finder.Sigma, 0.5))
	if math.IsNaN(result) {
		log.Printf("probe calculation failed. getting NaN")
		result = finder.lastProbe
	} else {
		finder.lastProbe = result
	}

	return result
}

func score(p float64) float64 {
	if p <= 0.0 {
		return 0.0
	} else {
		return -math.Log(p)
	}
}
