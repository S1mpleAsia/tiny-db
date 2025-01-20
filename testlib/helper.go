package testlib

// T=Table, B=Blocks, R=Records, V=DistinctValues
// | T        | B(T)   | R(T)     | V(T, F)                |
// |----------|--------|----------|------------------------|
// | STUDENT  | 4,500  | 45,000   | 45,000 for F=SId       |
// |          |        |          | 44,960 for F=SName     |
// |          |        |          | 50 for F=GradYear      |
// |          |        |          | 40 for F=MajorId       |
// | DEPT     | 2      | 40       | 40 for F=DId, DName    |
// | COURSE   | 25     | 500      | 500 for F=CId, Title   |
// |          |        |          | 40 for F=DeptId        |
// | SECTION  | 2,500  | 25,000   | 25,000 for F=SectId    |
// |          |        |          | 500 for F=CourseId     |
// |          |        |          | 250 for F=Prof         |
// |          |        |          | 50 for F=YearOffered   |
// | ENROLL   | 50,000 | 1,500,000| 1,500,000 for F=EId    |
// |          |        |          | 25,000 for F=SectionId |
// |          |        |          | 45,000 for F=StudentId |
// |          |        |          | 14 for F=Grade         |

type Student struct {
	SId      int
	SName    string
	GradYear int
	MajorId  int
}

type Dept struct {
	DId   int
	DName string
}

type Course struct {
	CId    int
	Title  string
	DeptId int
}

type Section struct {
	SectId      int
	CourseId    int
	Prof        string
	YearOffered int
}

type Enroll struct {
	EId       int
	SectionId int
	StudentId int
	Grade     string
}

var grades = []string{
	"A+", "A", "A-",
	"B+", "B", "B-",
	"C+", "C", "C-",
	"D+", "D", "D-",
	"F+", "F",
}

var studentExamples = []Student{
	{1, "joe", 2021, 10},
	{2, "amy", 2020, 20},
	{3, "max", 2022, 10},
	{4, "sue", 2022, 20},
	{5, "bob", 2020, 30},
	{6, "kim", 2019, 20},
	{7, "art", 2021, 30},
	{8, "pat", 2022, 10},
	{9, "lee", 2021, 10},
	{10, "dan", 2020, 20},
}

var deptExamples = []Dept{
	{10, "compsci"},
	{20, "math"},
	{30, "drama"},
}

// func InsertSmallTestData(t *testing.T, db *server.TinyDB) error {
// 	t.Helper()

// 	t.Log("--- Start insert small test data ---")

// 	tx, err := db.NewTx()
// 	if err != nil {
// 		return err
// 	}

// 	// plann
// }
