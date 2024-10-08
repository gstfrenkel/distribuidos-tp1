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

func topNReleases(h *MinHeap, n int) DateFilteredReleases {
	topReleases := make(DateFilteredReleases, 0, n)
	if h.Len() < n {
		n = h.Len()
	}
	hCopy := *h
	for i := 0; i < n; i++ {
		topReleases = append(topReleases, heap.Pop(&hCopy).(DateFilteredRelease))
	}
	sort.Slice(topReleases, func(i, j int) bool {
		return topReleases[i].AvgPlaytime > topReleases[j].AvgPlaytime
	})
	return topReleases
}

func ToTopNPlaytimeMessage(n uint8, h *MinHeap) DateFilteredReleases {
	return topNReleases(h, int(n))
}

func (m *MinHeap) UpdateReleases(releases DateFilteredReleases) {
	for _, release := range releases {
		if m.Len() < 10 {
			heap.Push(m, release)
		} else if release.AvgPlaytime > (*m)[0].AvgPlaytime {
			heap.Pop(m)   
			heap.Push(m, release)
		}
	}
}

func (m *MinHeap) GetTopReleases() DateFilteredReleases {
	topReleases := make(DateFilteredReleases, m.Len())
	for i := range topReleases {
		topReleases[i] = heap.Pop(m).(DateFilteredRelease)
	}
	
	sort.Slice(topReleases, func(i, j int) bool {
		return topReleases[i].AvgPlaytime > topReleases[j].AvgPlaytime
	})

	return topReleases
}