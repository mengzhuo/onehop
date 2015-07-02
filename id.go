package onehop

import (
	"crypto/rand"
)

const (
	ID_SIZE int = 16
)

type RecordID [ID_SIZE]byte

func NewRecordID() (id RecordID, err error) {
	_, err = rand.Read(id[:])
	return
}

func NewIDFromString(s string) (id RecordID) {

	for i := 0; i < ID_SIZE; i++ {
		id[i] = s[i]
	}
	return
}

func NewFullRecordID() (id RecordID) {

	for i := 0; i < ID_SIZE; i++ {
		id[i] = 0xff
	}
	return

}

func (r *RecordID) Less(cmp RecordID) bool {

	for i, v := range r {
		if v > cmp[i] {
			return false
		}
	}
	return true

}

func (r *RecordID) Equal(cmp RecordID) bool {

	for i, v := range r {

		if v != cmp[i] {
			return false
		}
	}
	return true

}

func (id *RecordID) RShift() {

	var mask byte = 0x0

	for i, _ := range id {
		tmp_mask := id[i] << 7
		id[i] = id[i]>>1 | mask
		mask = tmp_mask
	}

}
