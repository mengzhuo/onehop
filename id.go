package onehop

import (
	"bytes"
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

func NewIDFromBytes(b []byte) (id RecordID) {

	for i := 0; i < ID_SIZE; i++ {
		id[i] = b[i]
	}
	return
}

func NewFullRecordID() (id RecordID) {

	for i := 0; i < ID_SIZE; i++ {
		id[i] = 0xff
	}
	return

}

func (r *RecordID) Cmp(cmp RecordID) int {
	return bytes.Compare(r[:], cmp[:])
}

func (r *RecordID) Equal(cmp RecordID) bool {
	return bytes.Equal(r[:], cmp[:])
}

func (id *RecordID) RShift() {

	var mask byte = 0x0

	for i, _ := range id {
		tmp_mask := id[i] << 7
		id[i] = id[i]>>1 | mask
		mask = tmp_mask
	}

}
