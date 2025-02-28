package main

import (
	"math"
	"math/rand"
)

// Point represents a multi-dimensional point.
type Point []float64

// distance returns the Euclidean distance between two points.
func distance(a, b Point) float64 {
	sum := 0.0
	for i := 0; i < len(a); i++ {
		d := a[i] - b[i]
		sum += d * d
	}
	return math.Sqrt(sum)
}

// scalePoint divides each element of point a by scalar.
func scalePoint(a Point, scalar float64) {
	for i := range a {
		a[i] /= scalar
	}
}

// KMeans performs k-means clustering.
// data: slice of points; k: number of clusters; maxIterations: maximum iterations.
// Returns the centroids and the cluster assignment for each point.
func KMeans(data []Point, k int, maxIterations int) ([]Point, []int) {
	centroids := make([]Point, k)
	for i := 0; i < k; i++ {
		idx := rand.Intn(len(data))
		centroids[i] = make(Point, len(data[0]))
		copy(centroids[i], data[idx])
	}

	assignments := make([]int, len(data))

	for iter := 0; iter < maxIterations; iter++ {
		changed := false

		// Assign each point to the nearest centroid.
		for i, point := range data {
			minDist := math.Inf(1)
			minIndex := 0
			for j, centroid := range centroids {
				d := distance(point, centroid)
				if d < minDist {
					minDist = d
					minIndex = j
				}
			}
			if assignments[i] != minIndex {
				assignments[i] = minIndex
				changed = true
			}
		}

		if !changed {
			break
		}

		// Update centroids by computing the mean of each cluster.
		newCentroids := make([]Point, k)
		counts := make([]int, k)
		for i := 0; i < k; i++ {
			newCentroids[i] = make(Point, len(data[0]))
			for j := range newCentroids[i] {
				newCentroids[i][j] = 0.0
			}
		}
		for i, point := range data {
			cluster := assignments[i]
			counts[cluster]++
			for j, v := range point {
				newCentroids[cluster][j] += v
			}
		}
		for i := 0; i < k; i++ {
			if counts[i] > 0 {
				scalePoint(newCentroids[i], float64(counts[i]))
			} else {
				idx := rand.Intn(len(data))
				copy(newCentroids[i], data[idx])
			}
		}
		centroids = newCentroids
	}

	return centroids, assignments
}

// GetStates returns the current state for VM persistence.
func GetStates() []byte {
	return []byte{}
}

// SetStates loads the given state.
func SetStates(states []byte) {
}
