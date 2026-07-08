package db

// ChunkPatch describes typed changes to apply to all chunks of a document in LanceDB.
// Nil/absent fields are left unchanged. ExtraJSON pointer-to-empty-string sets NULL.
type ChunkPatch struct {
	ExtraJSON *string
	MetaStr   map[int]*string  // slot (1-based) → value; nil pointer sets NULL
	MetaNum   map[int]*float64
	MetaBool  map[int]*bool
	MetaDate  map[int]*int64   // Unix timestamp seconds; nil sets NULL
	MetaArr   map[int][]string // nil slice sets NULL
}

// ToMap converts the typed patch into the map expected by tbl.Update.
func (p ChunkPatch) ToMap() map[string]interface{} {
	m := make(map[string]interface{})
	if p.ExtraJSON != nil {
		if *p.ExtraJSON == "" {
			m["extra_json"] = nil
		} else {
			m["extra_json"] = *p.ExtraJSON
		}
	}
	for slot, v := range p.MetaStr {
		m[MetadataSlotColumn("str", slot)] = v
	}
	for slot, v := range p.MetaNum {
		m[MetadataSlotColumn("num", slot)] = v
	}
	for slot, v := range p.MetaBool {
		m[MetadataSlotColumn("bool", slot)] = v
	}
	for slot, v := range p.MetaDate {
		m[MetadataSlotColumn("date", slot)] = v
	}
	for slot, v := range p.MetaArr {
		m[MetadataSlotColumn("arr", slot)] = v
	}
	return m
}

func (p ChunkPatch) IsEmpty() bool {
	return p.ExtraJSON == nil &&
		len(p.MetaStr) == 0 && len(p.MetaNum) == 0 && len(p.MetaBool) == 0 &&
		len(p.MetaDate) == 0 && len(p.MetaArr) == 0
}
