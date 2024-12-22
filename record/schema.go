package record

type FieldType int32

const (
	INT FieldType = iota
	VARCHAR
)

type Schema struct {
	fields []string
	info   map[string]*FieldInfo
}

func NewSchema() *Schema {
	return &Schema{
		fields: make([]string, 0),
		info:   make(map[string]*FieldInfo),
	}
}

func (schema *Schema) AddField(fieldName string, fieldType FieldType, length int32) {
	schema.fields = append(schema.fields, fieldName)
	schema.info[fieldName] = NewFieldInfo(fieldType, length)
}

func (schema *Schema) AddIntField(fieldName string) {
	schema.AddField(fieldName, INT, 0)
}

func (schema *Schema) AddStringField(fieldName string, length int32) {
	schema.AddField(fieldName, VARCHAR, length)
}

func (schema *Schema) Add(fieldName string, sch *Schema) {
	fieldType := sch.Type(fieldName)
	fieldLength := sch.Length(fieldName)

	schema.AddField(fieldName, fieldType, fieldLength)
}

func (schema *Schema) AddAll(sch *Schema) {
	for _, fieldName := range sch.fields {
		schema.Add(fieldName, sch)
	}
}

func (schema *Schema) Fields() []string {
	return schema.fields
}

func (schema *Schema) HasField(fieldName string) bool {
	_, ok := schema.info[fieldName]
	return ok
}

func (schema *Schema) Type(fieldName string) FieldType {
	field, ok := schema.info[fieldName]

	if !ok {
		panic("field not found: " + fieldName)
	}
	return field.fieldType
}

func (schema *Schema) Length(fieldName string) int32 {
	field, ok := schema.info[fieldName]

	if !ok {
		panic("field not found: " + fieldName)
	}
	return field.length
}

type FieldInfo struct {
	fieldType FieldType
	length    int32
}

func NewFieldInfo(fieldType FieldType, length int32) *FieldInfo {
	return &FieldInfo{
		fieldType: fieldType,
		length:    length,
	}
}
