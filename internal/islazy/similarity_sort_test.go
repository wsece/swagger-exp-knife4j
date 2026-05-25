package islazy

import "testing"

func TestResponseBodyHammingDistance_identicalJSON(t *testing.T) {
	a := `{"code":0,"data":[1,2,3]}`
	b := "{\n  \"code\": 0,\n  \"data\": [1, 2, 3]\n}\n"
	d, err := ResponseBodyHammingDistance(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if d != 0 {
		t.Fatalf("expected 0, got %d", d)
	}
}

func TestResponseBodyHammingDistance_different(t *testing.T) {
	d, err := ResponseBodyHammingDistance(`{"a":1}`, `{"a":2}`)
	if err != nil {
		t.Fatal(err)
	}
	if d <= 0 {
		t.Fatalf("expected positive distance, got %d", d)
	}
}

func TestAssignResponseSimilarityClusters_sameGroup(t *testing.T) {
	hash := FormatResponseSimHash(ResponseSimHash(`{"ok":true}`))
	clusters := AssignResponseSimilarityClusters(
		[]uint{1, 1},
		[]int{0, 2},
		[]string{hash, hash},
		[]string{"", ""},
		[]string{"/a", "/b"},
		5,
	)
	if clusters[0] == 0 || clusters[0] != clusters[1] {
		t.Fatalf("expected same cluster, got %v", clusters)
	}
}

func TestSortIndicesByResponseSimilarity(t *testing.T) {
	keys := []SimilaritySortKey{
		{Cluster: 2, Distance: 0, TieBreak: "b"},
		{Cluster: 1, Distance: 0, TieBreak: "a"},
		{Cluster: 2, Distance: 1, TieBreak: "c"},
	}
	idx := SortIndicesByResponseSimilarity(keys)
	if len(idx) != 3 || idx[0] != 1 || idx[1] != 0 || idx[2] != 2 {
		t.Fatalf("unexpected order %v", idx)
	}
}
