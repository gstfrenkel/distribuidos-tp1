package message

import (
	"container/heap"
	"sort"
)


type DateFilteredReleases []DateFilteredRelease
type MinHeap DateFilteredReleases
type DateFilteredRelease struct {
	GameId      int64
	GameName    string
	AvgPlaytime int64
}

func (r DateFilteredReleases) ToBytes() ([]byte, error) {
	return toBytes(r)
}

func DateFilteredReleasesFromBytes(b []byte) (DateFilteredReleases, error) {
	var r DateFilteredReleases
	return r, fromBytes(b, &r)
}

func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(DateFilteredRelease))
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i].AvgPlaytime < h[j].AvgPlaytime } 
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func topNReleases(releases DateFilteredReleases, n int) DateFilteredReleases {
	h := &MinHeap{}
	heap.Init(h)

	for _, release := range releases {
		if h.Len() < n {
			heap.Push(h, release)
		} else if release.AvgPlaytime > (*h)[0].AvgPlaytime {
			heap.Pop(h)
			heap.Push(h, release)
		}
	}

	topReleases := make(DateFilteredReleases, h.Len())
	for i := range topReleases {
		topReleases[i] = heap.Pop(h).(DateFilteredRelease)
	}

	sort.Slice(topReleases, func(i, j int) bool {
		return topReleases[i].AvgPlaytime > topReleases[j].AvgPlaytime
	})

	return topReleases
}
func (d DateFilteredReleases) ToTopNPlaytimeMessage(n int) DateFilteredReleases{
	var result DateFilteredReleases
	result = topNReleases(d,n)
	return result
}