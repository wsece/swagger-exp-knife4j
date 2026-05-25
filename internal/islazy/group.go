// group.go: HashGrouper, a response grouper based on SimHash Hamming distance (used for database deduplication display).
package islazy

import "gorm.io/gorm"

const DefaultResponseHammingThreshold = 5

// HashGrouper assigns the same group ID to similar hashes based on Hamming distance
type HashGrouper struct {
	Groups    []HammingGroup
	Threshold int
}

// NewHashGrouper creates a new grouper; uses default value if threshold <= 0
func NewHashGrouper(threshold int) *HashGrouper {
	if threshold <= 0 {
		threshold = DefaultResponseHammingThreshold
	}
	return &HashGrouper{Threshold: threshold}
}

// LoadFromDB restores the grouping cache from existing records
func (g *HashGrouper) LoadFromDB(conn *gorm.DB, model interface{}, hashColumn, groupColumn string) error {
	type row struct {
		Hash    string
		GroupID uint
	}
	var rows []row
	err := conn.Model(model).
		Select(hashColumn + " AS hash, " + groupColumn + " AS group_id").
		Where(hashColumn + " != '' AND " + groupColumn + " > 0").
		Group(hashColumn + ", " + groupColumn).
		Scan(&rows).Error
	if err != nil {
		return err
	}

	for _, r := range rows {
		parsed, err := ParseResponseSimHash(r.Hash)
		if err != nil || len(parsed) == 0 {
			continue
		}
		g.Groups = append(g.Groups, HammingGroup{
			GroupID: r.GroupID,
			Hash:    parsed,
		})
	}
	return nil
}

// Assign assigns a group ID to the hash and returns the Hamming distance to the group representative hash (smaller = more similar)
func (g *HashGrouper) Assign(conn *gorm.DB, model interface{}, groupColumn string, hash []byte) (groupID uint, hashStr string, distance int, err error) {
	if len(hash) == 0 {
		return 0, "", 0, nil
	}

	hashStr = FormatResponseSimHash(hash)

	for _, group := range g.Groups {
		dist, distErr := HammingDistance(hash, group.Hash)
		if distErr != nil {
			return 0, "", 0, distErr
		}
		if dist <= g.Threshold {
			return group.GroupID, hashStr, dist, nil
		}
	}

	var maxGroupID uint
	err = conn.Model(model).
		Select("COALESCE(MAX(" + groupColumn + "), 0)").
		Scan(&maxGroupID).Error
	if err != nil {
		return 0, "", 0, err
	}
	nextGroupID := maxGroupID + 1

	g.Groups = append(g.Groups, HammingGroup{
		GroupID: nextGroupID,
		Hash:    append([]byte(nil), hash...),
	})

	return nextGroupID, hashStr, 0, nil
}
