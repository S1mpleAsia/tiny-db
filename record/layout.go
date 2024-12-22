package record

import "s1mpleasia.com/tinydb/file"

/*
	- Layout define the structure of the slot record
		+ schema: Schema of the slot
		+ offset: The offset map of all the field in slot
		+ slotSize: The size in bytes of the slot
	+--------+-----------+-----------+-----------+
	|  flag  |  field 1  |  field 2  |  .......  |
	+--------+-----------+-----------+-----------+
*/
type Layout struct {
	schema   *Schema
	offset   map[string]int32
	slotSize int32
}

func NewLayoutFromSchema(sch *Schema) *Layout {
	pos := int32(file.INT_32_BITS) // 1 bytes for store flag
	offset := make(map[string]int32)

	layout := &Layout{
		schema: sch,
	}

	for _, fieldName := range sch.fields {
		offset[fieldName] = pos
		pos += layout.lengthInBytes(fieldName)
	}

	return NewLayout(sch, offset, pos)
}

func NewLayout(sch *Schema, offset map[string]int32, slotSize int32) *Layout {
	return &Layout{
		schema:   sch,
		offset:   offset,
		slotSize: slotSize,
	}
}

func (l *Layout) Schema() *Schema {
	return l.schema
}

func (l *Layout) Offset(fieldName string) int32 {
	return l.offset[fieldName]
}

func (l *Layout) SlotSize() int32 {
	return l.slotSize
}

func (l *Layout) lengthInBytes(fieldName string) int32 {
	fieldType := l.schema.Type(fieldName)

	if fieldType == INT {
		return file.INT_32_BITS
	} else {
		return int32(file.MaxLength(int(l.schema.Length(fieldName))))
	}
}
