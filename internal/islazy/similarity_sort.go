// similarity_sort.go: response-body similarity (SimHash + Hamming) and stable sort keys for UI/DB.
package islazy

import (
	"sort"
)

// ResponseBodyHammingDistance returns the Hamming distance between two response bodies (0 = identical fingerprint).
// Empty bodies yield distance 0 when both empty; one empty returns 64 (max for 8-byte simhash).
func ResponseBodyHammingDistance(bodyA, bodyB string) (int, error) {
	ha := ResponseSimHash(bodyA)
	hb := ResponseSimHash(bodyB)
	if len(ha) == 0 && len(hb) == 0 {
		return 0, nil
	}
	if len(ha) == 0 || len(hb) == 0 {
		return 64, nil
	}
	return HammingDistance(ha, hb)
}

// EffectiveResponseSimHash returns stored s:<hex> hash or computes one from the body.
func EffectiveResponseSimHash(storedHash, body string) []byte {
	if storedHash != "" {
		parsed, err := ParseResponseSimHash(storedHash)
		if err == nil && len(parsed) > 0 {
			return parsed
		}
	}
	return ResponseSimHash(body)
}

// SimilaritySortKey is used to order records: similar responses share a cluster; lower distance = closer to cluster center.
type SimilaritySortKey struct {
	Cluster  int
	Distance int
	TieBreak string
}

// CompareSimilaritySortKeys returns -1 if a should appear before b.
func CompareSimilaritySortKeys(a, b SimilaritySortKey) int {
	if a.Cluster != b.Cluster {
		if a.Cluster < b.Cluster {
			return -1
		}
		return 1
	}
	if a.Distance != b.Distance {
		if a.Distance < b.Distance {
			return -1
		}
		return 1
	}
	if a.TieBreak < b.TieBreak {
		return -1
	}
	if a.TieBreak > b.TieBreak {
		return 1
	}
	return 0
}

type similarityItem struct {
	index    int
	hash     []byte
	groupID  uint
	distance int
	tieBreak string
	cluster  int
}

// AssignResponseSimilarityClusters groups records by DB response_group_id when present, otherwise by Hamming threshold.
// Returns cluster ids 1..N (0 = empty / no fingerprint). Larger clusters get smaller ids when possible.
func AssignResponseSimilarityClusters(
	responseGroupID []uint,
	similarityDistance []int,
	storedSimHash []string,
	responseBody []string,
	tieBreak []string,
	threshold int,
) []int {
	n := len(responseGroupID)
	if n == 0 {
		return nil
	}
	if threshold <= 0 {
		threshold = DefaultResponseHammingThreshold
	}
	if len(similarityDistance) < n {
		similarityDistance = padIntSlice(similarityDistance, n)
	}
	if len(storedSimHash) < n {
		storedSimHash = padStringSlice(storedSimHash, n)
	}
	if len(responseBody) < n {
		responseBody = padStringSlice(responseBody, n)
	}
	if len(tieBreak) < n {
		tieBreak = padStringSlice(tieBreak, n)
	}

	items := make([]similarityItem, n)
	for i := 0; i < n; i++ {
		items[i] = similarityItem{
			index:    i,
			hash:     EffectiveResponseSimHash(storedSimHash[i], responseBody[i]),
			groupID:  responseGroupID[i],
			distance: similarityDistance[i],
			tieBreak: tieBreak[i],
		}
	}

	nextCluster := 1
	groupToCluster := make(map[uint]int)
	clusterRep := make(map[int][]byte)

	for i := range items {
		if items[i].groupID == 0 {
			continue
		}
		cid, ok := groupToCluster[items[i].groupID]
		if !ok {
			cid = nextCluster
			nextCluster++
			groupToCluster[items[i].groupID] = cid
			if len(items[i].hash) > 0 {
				clusterRep[cid] = append([]byte(nil), items[i].hash...)
			}
		}
		items[i].cluster = cid
	}

	for i := range items {
		if items[i].cluster != 0 {
			continue
		}
		if len(items[i].hash) == 0 {
			items[i].cluster = 0
			continue
		}
		assigned := false
		for cid, rep := range clusterRep {
			dist, err := HammingDistance(items[i].hash, rep)
			if err != nil {
				continue
			}
			if dist <= threshold {
				items[i].cluster = cid
				assigned = true
				break
			}
		}
		if !assigned {
			cid := nextCluster
			nextCluster++
			items[i].cluster = cid
			clusterRep[cid] = append([]byte(nil), items[i].hash...)
		}
	}

	out := make([]int, n)
	for _, it := range items {
		out[it.index] = it.cluster
	}
	return out
}

// BuildSimilaritySortKeys builds per-record sort keys from parallel slices (same length).
func BuildSimilaritySortKeys(
	responseGroupID []uint,
	similarityDistance []int,
	storedSimHash []string,
	responseBody []string,
	tieBreak []string,
	threshold int,
) []SimilaritySortKey {
	n := len(responseGroupID)
	clusters := AssignResponseSimilarityClusters(
		responseGroupID, similarityDistance, storedSimHash, responseBody, tieBreak, threshold,
	)
	keys := make([]SimilaritySortKey, n)
	for i := 0; i < n; i++ {
		dist := 0
		if i < len(similarityDistance) {
			dist = similarityDistance[i]
		}
		tb := ""
		if i < len(tieBreak) {
			tb = tieBreak[i]
		}
		cluster := 1_000_000
		if i < len(clusters) && clusters[i] > 0 {
			cluster = clusters[i]
		}
		keys[i] = SimilaritySortKey{Cluster: cluster, Distance: dist, TieBreak: tb}
	}
	return keys
}

// SortIndicesByResponseSimilarity returns record indices ordered by similarity cluster (ascending cluster, distance, tie-break).
func SortIndicesByResponseSimilarity(keys []SimilaritySortKey) []int {
	indices := make([]int, len(keys))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(i, j int) bool {
		return CompareSimilaritySortKeys(keys[indices[i]], keys[indices[j]]) < 0
	})
	return indices
}

func padIntSlice(s []int, n int) []int {
	out := make([]int, n)
	copy(out, s)
	return out
}

func padStringSlice(s []string, n int) []string {
	out := make([]string, n)
	copy(out, s)
	return out
}
