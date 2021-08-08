package micro

type Model struct {
	ID        uint   `json:"id" gorm:"primary_key"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
	DeletedAt int64  `json:"deleted_at" sql:"index"`
	CreatedID int64  `json:"created_id"`
	UpdatedID int64  `json:"updated_id"`
	DeletedID int64  `json:"deleted_id"`
	CreatedBy string `json:"created_by"`
	UpdatedBy string `json:"updated_by"`
	DeletedBy string `json:"deleted_by"`
}

//type Array string
//
//func toSlice(raw string) (slice []interface{}) {
//	_ = json.Unmarshal([]byte(raw), &slice)
//	return
//}
//
//func toSet(slice []interface{}) (set interface{}) {
//
//}
//
//func toRaw(v interface{}) string {
//	raw, _ := json.Marshal(v)
//	return string(raw)
//}
//
//func (a Array) Set() Array {
//	var set = make(map[interface{}]struct{})
//	var slice []interface{}
//	for _, val := range a.Slice() {
//		if _, exists := set[val]; !exists {
//			set[val] = struct{}{}
//			slice = append(slice, val)
//		}
//	}
//	return Array()
//}
//
//func (a Array) Slice() (slice []interface{}) {
//	_ = json.Unmarshal([]byte(a), &slice)
//	return
//}
//
//func (a Array) Float64s() (slice []float64) {
//	_ = json.Unmarshal([]byte(a), &slice)
//	return
//}
//
//func (a Array) Int64s() (slice []int64) {
//	_ = json.Unmarshal([]byte(a), &slice)
//	return
//}
//
//func (a Array) Strings() (slice []string) {
//	_ = json.Unmarshal([]byte(a), &slice)
//	return
//}
