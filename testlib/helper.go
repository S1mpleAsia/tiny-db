package testlib

import (
	"fmt"
	"slices"
	"strconv"
	"testing"

	"s1mpleasia.com/tinydb/plan"
	"s1mpleasia.com/tinydb/server"
	"s1mpleasia.com/tinydb/transaction"
)

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

// setup 10 `student`, 3 `dept` records
func InsertSmallTestData(t *testing.T, db *server.TinyDB) error {
	t.Helper()

	t.Log("--- Start insert small test data ---")

	tx, err := db.NewTx()
	if err != nil {
		return err
	}

	planner := db.Planner()

	createStudentTable(t, planner, tx)
	createDeptTable(t, planner, tx)
	insertDepts(t, planner, tx, deptExamples)
	insertStudents(t, planner, tx, studentExamples)

	mdm := db.MetadataMgmt()
	err = mdm.ForceRefreshStatistics(tx)
	if err != nil {
		return err
	}

	tx.Commit()

	t.Log("End InsertSmallTestData")
	return nil
}

// setup 10 `student`, 100 `enroll` records
func InsertMiddleTestData(t *testing.T, db *server.TinyDB) error {
	t.Helper()

	t.Log("--- Start insert middle test data ---")

	tx, err := db.NewTx()
	if err != nil {
		return err
	}

	planner := db.Planner()
	createStudentTable(t, planner, tx)
	createEnrollTable(t, planner, tx)
	insertStudents(t, planner, tx, studentExamples)

	enrolls := []Enroll{}
	eid := 0

	for _, s := range studentExamples {
		for j := 0; j < 10; j++ {
			enrolls = append(enrolls, Enroll{eid, eid % 25, s.SId, grades[eid%len(grades)]})
			eid++
		}
	}

	slices.Reverse(enrolls)
	insertEnrolls(t, planner, tx, enrolls)

	mdm := db.MetadataMgmt()
	err = mdm.ForceRefreshStatistics(tx)
	if err != nil {
		return err
	}
	tx.Commit()

	t.Log("End InsertMiddleTestData")
	return nil
}

// Setup full `student`, `dept`, `course`, `section`, `enroll` records
func InsertLargeTestData(t *testing.T, db *server.TinyDB) error {
	t.Helper()

	t.Log("Start InsertLargeTestData")

	tx, err := db.NewTx()
	if err != nil {
		return err
	}

	planner := db.Planner()

	createStudentTable(t, planner, tx)
	createDeptTable(t, planner, tx)
	createCourseTable(t, planner, tx)
	createSectionTable(t, planner, tx)
	createEnrollTable(t, planner, tx)

	numStudents := 450  // 1%
	numDepts := 40      // 100%
	numCourses := 50    // 10%
	numSections := 250  // 10%
	numEnrolls := 1_500 // 0.1%

	students := make([]Student, 0, numStudents)
	for sid := range numStudents {
		students = append(students, Student{
			SId:      sid,
			SName:    "student_" + strconv.Itoa(sid),
			MajorId:  sid%40 + 1,
			GradYear: 1974 + sid%50,
		})
	}
	insertStudents(t, planner, tx, students)

	depts := make([]Dept, 0, numDepts)
	for did := range numDepts {
		depts = append(depts, Dept{
			DId:   did,
			DName: "dept_" + strconv.Itoa(did),
		})
	}
	insertDepts(t, planner, tx, depts)

	courses := make([]Course, 0, numCourses)
	for cid := range numCourses {
		courses = append(courses, Course{
			CId:    cid,
			Title:  "course_" + strconv.Itoa(cid),
			DeptId: cid%numDepts + 1,
		})
	}
	insertCourses(t, planner, tx, courses)

	sections := make([]Section, 0, numSections)
	for sectid := range numSections {
		sections = append(sections, Section{
			SectId:      sectid,
			CourseId:    sectid%numCourses + 1,
			Prof:        "prof_" + strconv.Itoa(sectid),
			YearOffered: 1974 + sectid%50,
		})
	}
	insertSections(t, planner, tx, sections)

	enrolls := make([]Enroll, 0, numEnrolls)
	for eid := range numEnrolls {
		enrolls = append(enrolls, Enroll{
			EId:       eid,
			SectionId: eid%numSections + 1,
			StudentId: eid%numStudents + 1,
			Grade:     grades[eid%len(grades)],
		})
	}
	slices.Reverse(enrolls)
	insertEnrolls(t, planner, tx, enrolls)

	mdm := db.MetadataMgmt()
	err = mdm.ForceRefreshStatistics(tx)
	if err != nil {
		return err
	}

	t.Log("End InsertLargeTestData")
	return nil
}

func createStudentTable(t *testing.T, planner *plan.Planner, tx *transaction.Transaction) {
	_, err := planner.ExecuteUpdate("create table student(sid int, sname varchar(10), gradyear int, majorid int) ", tx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = planner.ExecuteUpdate("create index majorid_idx on student(majorid)", tx)
	if err != nil {
		t.Fatal(err)
	}
}

func insertStudents(t *testing.T, planner *plan.Planner, tx *transaction.Transaction, students []Student) {
	for _, s := range students {
		query := fmt.Sprintf("insert into student(sid, sname, gradyear, majorid) values(%d, '%s', %d, %d)", s.SId, s.SName, s.GradYear, s.MajorId)
		_, err := planner.ExecuteUpdate(query, tx)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func createDeptTable(t *testing.T, planner *plan.Planner, tx *transaction.Transaction) {
	_, err := planner.ExecuteUpdate("create table dept(did int, dname varchar(8))", tx)
	if err != nil {
		t.Fatal(err)
	}
}

func insertDepts(t *testing.T, planner *plan.Planner, tx *transaction.Transaction, depts []Dept) {
	for _, d := range depts {
		query := fmt.Sprintf("insert into dept(did, dname) values(%d, '%s')", d.DId, d.DName)
		_, err := planner.ExecuteUpdate(query, tx)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func createCourseTable(t *testing.T, planner *plan.Planner, tx *transaction.Transaction) {
	_, err := planner.ExecuteUpdate("create table course(cid int, title varchar(12), deptid int)", tx)
	if err != nil {
		t.Fatal(err)
	}
}

func insertCourses(t *testing.T, planner *plan.Planner, tx *transaction.Transaction, courses []Course) {
	for _, c := range courses {
		query := fmt.Sprintf("insert into course(cid, title, deptid) values(%d, '%s', %d)", c.CId, c.Title, c.DeptId)
		_, err := planner.ExecuteUpdate(query, tx)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func createSectionTable(t *testing.T, planner *plan.Planner, tx *transaction.Transaction) {
	_, err := planner.ExecuteUpdate("create table section(sectid int, courseid int, prof varchar(12), yearoffered int)", tx)
	if err != nil {
		t.Fatal(err)
	}
}

func insertSections(t *testing.T, planner *plan.Planner, tx *transaction.Transaction, sections []Section) {
	for _, s := range sections {
		query := fmt.Sprintf("insert into section(sectid, courseid, prof, yearoffered) values(%d, %d, '%s', %d)", s.SectId, s.CourseId, s.Prof, s.YearOffered)
		_, err := planner.ExecuteUpdate(query, tx)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func createEnrollTable(t *testing.T, planner *plan.Planner, tx *transaction.Transaction) {
	_, err := planner.ExecuteUpdate("create table enroll(eid int, studentid int, sectionid int, grade varchar(2))", tx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = planner.ExecuteUpdate("create index studentid_idx on enroll(studentid)", tx)
	if err != nil {
		t.Fatal(err)
	}
}

func insertEnrolls(t *testing.T, planner *plan.Planner, tx *transaction.Transaction, enrolls []Enroll) {
	for _, e := range enrolls {
		query := fmt.Sprintf("insert into enroll(eid, studentid, sectionid, grade) values(%d, %d, %d, '%s')", e.EId, e.StudentId, e.SectionId, e.Grade)
		_, err := planner.ExecuteUpdate(query, tx)
		if err != nil {
			t.Fatal(err)
		}
	}
}
