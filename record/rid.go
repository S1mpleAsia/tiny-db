package record

import "fmt"

type RID struct {
	blockNum int32
	slot     int32
}

func NewRID(blockNum int32, slot int32) *RID {
	return &RID{
		blockNum: blockNum,
		slot:     slot,
	}
}

func (rid *RID) BlockNumber() int32 {
	return rid.blockNum
}

func (rid *RID) Slot() int32 {
	return rid.slot
}

func (rid *RID) Equals(target *RID) bool {
	return rid.blockNum == target.blockNum && rid.slot == target.slot
}

func (rid *RID) String() string {
	return fmt.Sprintf("[blockNum: %d, slot: %d]", rid.blockNum, rid.slot)
}
