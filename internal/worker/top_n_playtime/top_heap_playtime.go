package top_n_playtime

import (
	"container/heap"
	"sort"
	"tp1/pkg/message"
)

type MinHeapPlaytime message.DateFilteredReleases

func (h *MinHeapPlaytime) Push(x interface{}) {
	*h = append(*h, x.(message.DateFilteredRelease))
}

func (h *MinHeapPlaytime) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

func (h MinHeapPlaytime) Len() int           { return len(h) }
func (h MinHeapPlaytime) Less(i, j int) bool { return h[i].AvgPlaytime < h[j].AvgPlaytime }
func (h MinHeapPlaytime) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func topNReleases(h *MinHeapPlaytime, n int) message.DateFilteredReleases {
	topReleases := make(message.DateFilteredReleases, 0, n)
	if h.Len() < n {
		n = h.Len()
	}
	hCopy := *h
	for i := 0; i < n; i++ {
		topReleases = append(topReleases, heap.Pop(&hCopy).(message.DateFilteredRelease))
	}
	sort.Slice(topReleases, func(i, j int) bool {
		return topReleases[i].AvgPlaytime > topReleases[j].AvgPlaytime
	})
	return topReleases
}

func ToTopNPlaytimeMessage(n uint8, h *MinHeapPlaytime) message.DateFilteredReleases {
	return topNReleases(h, int(n))
}

func (m *MinHeapPlaytime) UpdateReleases(releases message.DateFilteredReleases, n int) {
	for _, release := range releases {
		if m.Len() < n {
			heap.Push(m, release)
		} else if release.AvgPlaytime > (*m)[0].AvgPlaytime {
			heap.Pop(m)
			heap.Push(m, release)
		}
	}
}

func (m *MinHeapPlaytime) GetTopReleases() message.DateFilteredReleases {
	topReleases := make(message.DateFilteredReleases, m.Len())
	for i := range topReleases {
		topReleases[i] = heap.Pop(m).(message.DateFilteredRelease)
	}

	sort.Slice(topReleases, func(i, j int) bool {
		return topReleases[i].AvgPlaytime > topReleases[j].AvgPlaytime
	})

	return topReleases
}
