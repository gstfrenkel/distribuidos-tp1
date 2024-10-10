package message

type DateFilteredReleases []DateFilteredRelease
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

// package message

// import (
// 	"container/heap"
// 	"sort"
// )

// type DateFilteredReleases []DateFilteredRelease
// type DateFilteredRelease struct {
// 	GameId      int64
// 	GameName    string
// 	AvgPlaytime int64
// }

// func (r DateFilteredReleases) ToBytes() ([]byte, error) {
// 	return toBytes(r)
// }

// func DateFilteredReleasesFromBytes(b []byte) (DateFilteredReleases, error) {
// 	var r DateFilteredReleases
// 	return r, fromBytes(b, &r)
// }

// type MinHeapPlaytime DateFilteredReleases

// func (h *MinHeapPlaytime) Push(x interface{}) {
// 	*h = append(*h, x.(DateFilteredRelease))
// }

// func (h *MinHeapPlaytime) Pop() interface{} {
// 	old := *h
// 	n := len(old)
// 	item := old[n-1]
// 	*h = old[0 : n-1]
// 	return item
// }

// func (h MinHeapPlaytime) Len() int           { return len(h) }
// func (h MinHeapPlaytime) Less(i, j int) bool { return h[i].AvgPlaytime < h[j].AvgPlaytime }
// func (h MinHeapPlaytime) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

// func topNReleases(h *MinHeapPlaytime, n int) DateFilteredReleases {
// 	topReleases := make(DateFilteredReleases, 0, n)
// 	if h.Len() < n {
// 		n = h.Len()
// 	}
// 	hCopy := *h
// 	for i := 0; i < n; i++ {
// 		topReleases = append(topReleases, heap.Pop(&hCopy).(DateFilteredRelease))
// 	}
// 	sort.Slice(topReleases, func(i, j int) bool {
// 		return topReleases[i].AvgPlaytime > topReleases[j].AvgPlaytime
// 	})
// 	return topReleases
// }

// func ToTopNPlaytimeMessage(n uint8, h *MinHeapPlaytime) DateFilteredReleases {
// 	return topNReleases(h, int(n))
// }

// func (m *MinHeapPlaytime) UpdateReleases(releases DateFilteredReleases,n int) {
// 	for _, release := range releases {
// 		if m.Len() < n {
// 			heap.Push(m, release)
// 		} else if release.AvgPlaytime > (*m)[0].AvgPlaytime {
// 			heap.Pop(m)
// 			heap.Push(m, release)
// 		}
// 	}
// }

// func (m *MinHeapPlaytime) GetTopReleases() DateFilteredReleases {
// 	topReleases := make(DateFilteredReleases, m.Len())
// 	for i := range topReleases {
// 		topReleases[i] = heap.Pop(m).(DateFilteredRelease)
// 	}

// 	sort.Slice(topReleases, func(i, j int) bool {
// 		return topReleases[i].AvgPlaytime > topReleases[j].AvgPlaytime
// 	})

// 	return topReleases
// }