package db

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	ut "github.com/zdnscloud/cement/unittest"
)

var dbConf map[string]interface{} = map[string]interface{}{
	"host":     "localhost",
	"user":     "lx",
	"password": "lx",
	"dbname":   "lx",
}

type Child struct {
	Id       string
	Name     string `db:"uk"`
	Age      uint32
	Hobbies  []string
	Scores   []int
	Birthday time.Time
	Talented bool
}

type Mother struct {
	Id   string
	Age  int
	Name string
}

type MotherChild struct {
	Id     string
	Mother string `db:"ownby"`
	Child  string `db:"referto"`
}

func initChild(store ResourceStore) {
	tx, _ := store.Begin()
	c1 := &Child{
		Id:       "c1",
		Name:     "ben",
		Age:      20,
		Hobbies:  []string{"movie", "music"},
		Scores:   []int{1, 3, 4},
		Birthday: time.Now(),
		Talented: true,
	}
	_, err := tx.Insert(c1)
	if err != nil {
		fmt.Printf("insert get err:%v\n", err.Error())
	}

	c2 := &Child{
		Id:       "c2",
		Name:     "nana",
		Age:      30,
		Hobbies:  []string{"movie", "music"},
		Scores:   []int{1, 2, 4},
		Birthday: time.Now(),
		Talented: true,
	}
	_, err = tx.Insert(c2)
	if err != nil {
		fmt.Printf("insert get err:%v\n", err.Error())
	}
	tx.Commit()
}

func initMother(store ResourceStore) {
	tx, _ := store.Begin()
	tx.Insert(&Mother{
		Id:   "m1",
		Name: "lxq",
	})
	tx.Commit()
}

func initMotherChild(store ResourceStore) {
	tx, _ := store.Begin()
	tx.Insert(&MotherChild{
		Mother: "m1",
		Child:  "c1",
	})
	tx.Commit()
}

func TestCURD(t *testing.T) {
	meta, err := NewResourceMeta([]Resource{&Child{}})
	ut.Assert(t, err == nil, "")
	store, err := NewRStore(dbConf, meta)
	ut.Assert(t, err == nil, "")

	initChild(store)

	tx, _ := store.Begin()
	children := []*Child{}
	c, err := tx.Count("child", nil)
	ut.Equal(t, c, int64(2))
	tx.Fill(map[string]interface{}{"id": "c1"}, &children)
	ut.Equal(t, len(children), 1)
	ut.Equal(t, children[0].Scores, []int{1, 3, 4})
	tx.Rollback()

	tx, _ = store.Begin()
	c, err = tx.Update("child", map[string]interface{}{
		"hobbies": []string{"read book", "travel"},
	}, map[string]interface{}{"id": "c1"})
	ut.Assert(t, err == nil, "")
	ut.Equal(t, c, int64(1))
	tx.Commit()

	tx, _ = store.Begin()
	c, err = tx.Count("child", map[string]interface{}{
		"hobbies": []string{"read book", "travel"},
	})
	ut.Assert(t, err == nil, "")
	ut.Equal(t, c, int64(1))
	tx.Rollback()

	tx, _ = store.Begin()
	c, err = tx.Delete("child", map[string]interface{}{
		"name": "nana",
	})
	ut.Assert(t, err == nil, "")
	ut.Equal(t, c, int64(1))
	tx.Commit()

	tx, _ = store.Begin()
	children_, err := tx.Get("child", map[string]interface{}{
		"name": "nana",
	})
	ut.Assert(t, err == nil, "")
	ut.Equal(t, len(children_.([]*Child)), 0)
	children_, err = tx.Get("child", map[string]interface{}{
		"name": "ben",
	})
	ut.Assert(t, err == nil, "")
	ut.Equal(t, children_.([]*Child)[0].Id, "c1")
	tx.Rollback()

	store.Clean()
	store.Destroy()
}

func TestCURDEx(t *testing.T) {
	meta, err := NewResourceMeta([]Resource{&Child{}})
	ut.Assert(t, err == nil, "")
	store, err := NewRStore(dbConf, meta)
	ut.Assert(t, err == nil, "")

	initChild(store)

	tx, _ := store.Begin()
	children := []*Child{}
	tx.FillEx(&children, "select distinct age from gr_child ORDER BY age")
	ut.Equal(t, len(children), 2)

	children = []*Child{}
	tx.FillEx(&children, "select * from gr_child where age > $1 and name like $2", 20, "na%")
	ut.Equal(t, len(children), 1)

	children_, _ := tx.GetEx(ResourceType("child"), "select * from gr_child")
	ut.Equal(t, len(children_.([]*Child)), 2)

	children_, _ = tx.GetEx(ResourceType("child"), "select * from gr_child where age between $1 and $2", 1, 25)
	ut.Equal(t, len(children_.([]*Child)), 1)
	tx.Rollback()

	tx, _ = store.Begin()
	count, err := tx.DeleteEx("delete from gr_child where age >= $1 and age < $2", 25, 400)
	tx.Commit()
	ut.Equal(t, err, nil)
	ut.Equal(t, count, int64(1))

	store.Clean()
	store.Destroy()
}

func TestMultiToMultiRelationship(t *testing.T) {
	meta, err := NewResourceMeta([]Resource{&Mother{}, &Child{}, &MotherChild{}})
	ut.Assert(t, err == nil, "")
	store, err := NewRStore(dbConf, meta)
	ut.Assert(t, err == nil, "")

	initChild(store)
	initMother(store)
	initMotherChild(store)

	tx, _ := store.Begin()
	result, err := tx.GetOwned(ResourceType("mother"), "m1", ResourceType("child"))
	ut.Equal(t, err, nil)
	ut.Equal(t, len(result.([]*Child)), 1)
	tx.Rollback()

	//insert unknown mother should fail
	tx, _ = store.Begin()
	_, err = tx.Insert(&MotherChild{
		Mother: "m2",
		Child:  "c1",
	})
	ut.Assert(t, err != nil, "")
	tx.Rollback()

	//delete used child should fail
	tx, _ = store.Begin()
	_, err = tx.Delete("child", map[string]interface{}{
		"name": "ben",
	})
	ut.Assert(t, err != nil, "")
	tx.Rollback()

	store.Clean()
	store.Destroy()
}

type View struct {
	Id   string
	Name string `db:"uk"`
}

type Zone struct {
	Id   string
	Name string
	View string `db:"ownby"`
}

func TestOneToManyRelationship(t *testing.T) {
	meta, err := NewResourceMeta([]Resource{&View{}, &Zone{}})
	ut.Assert(t, err == nil, "")
	store, err := NewRStore(dbConf, meta)
	ut.Assert(t, err == nil, "")

	tx, _ := store.Begin()
	tx.Insert(&View{Id: "v1", Name: "v1"})
	tx.Insert(&View{Id: "v2", Name: "v2"})
	tx.Insert(&Zone{Name: "cn", View: "v1"})
	tx.Insert(&Zone{Name: "com", View: "v1"})
	tx.Commit()

	tx, _ = store.Begin()
	result, err := tx.Get(ResourceType("zone"), map[string]interface{}{"view": "v1"})
	ut.Equal(t, err, nil)
	ut.Equal(t, len(result.([]*Zone)), 2)
	result, err = tx.Get(ResourceType("zone"), map[string]interface{}{"view": "v2"})
	ut.Equal(t, err, nil)
	ut.Equal(t, len(result.([]*Zone)), 0)
	tx.Rollback()

	//view v3 doesn't exists
	tx, _ = store.Begin()
	_, err = tx.Insert(&Zone{Name: "cn", View: "v3"})
	ut.Assert(t, err != nil, "")
	tx.Rollback()

	//delete mother will delete owned child
	tx, _ = store.Begin()
	c, err := tx.Delete("view", map[string]interface{}{
		"id": "v2",
	})
	ut.Assert(t, err == nil, "")
	ut.Equal(t, c, int64(1))
	c, _ = tx.Count("zone", nil)
	ut.Equal(t, c, int64(2))
	c, err = tx.Delete("view", map[string]interface{}{
		"id": "v1",
	})
	ut.Assert(t, err == nil, "")
	ut.Equal(t, c, int64(1))
	c, _ = tx.Count("zone", nil)
	ut.Equal(t, c, int64(0))
	tx.Commit()

	store.Clean()
	store.Destroy()
}

func TestGetWithLimitAndOffset(t *testing.T) {
	meta, err := NewResourceMeta([]Resource{&Mother{}})
	ut.Assert(t, err == nil, "")
	store, err := NewRStore(dbConf, meta)
	ut.Assert(t, err == nil, "")

	tx, _ := store.Begin()
	for i := 0; i < 2000; i++ {
		tx.Insert(&Mother{Age: i, Name: "m" + strconv.Itoa(i)})
	}
	tx.Commit()

	tx, _ = store.Begin()
	var mothers []*Mother
	tx.Fill(map[string]interface{}{"offset": 10, "limit": 20, "orderby": "age"}, &mothers)
	ut.Equal(t, len(mothers), 20)
	for i := 0; i < 10; i++ {
		ut.Equal(t, mothers[i].Age, i+10)
	}
	tx.Commit()

	store.Clean()
	store.Destroy()
}

type Student struct {
	Id        string
	Name      string `db:"uk"`
	Age       uint32
	Classroom string `db:"-"`
}

func TestIgnField(t *testing.T) {
	meta, err := NewResourceMeta([]Resource{&Student{}})
	ut.Assert(t, err == nil, "")
	store, err := NewRStore(dbConf, meta)
	ut.Assert(t, err == nil, "")

	tx, _ := store.Begin()
	tx.Insert(&Student{
		Name:      "ben",
		Age:       40,
		Classroom: "991",
	})
	tx.Commit()

	tx, _ = store.Begin()
	var students []*Student
	tx.Fill(nil, &students)
	ut.Equal(t, len(students), 1)
	ut.Equal(t, students[0].Classroom, "")
	tx.Commit()

	store.Clean()
	store.Destroy()
}

type Rdata struct {
	Id    string
	Name  string `db:"uk"`
	Type  string `db:"uk"`
	Rdata string `db:"uk"`
}

func TestUniqueField(t *testing.T) {
	meta, err := NewResourceMeta([]Resource{&Rdata{}})
	ut.Assert(t, err == nil, "")
	store, err := NewRStore(dbConf, meta)
	ut.Assert(t, err == nil, "")

	tx, _ := store.Begin()
	_, err = tx.Insert(&Rdata{
		Name:  "n1",
		Type:  "a",
		Rdata: "1.1.1.1",
	})
	ut.Assert(t, err == nil, "")
	_, err = tx.Insert(&Rdata{
		Name:  "n1",
		Type:  "a",
		Rdata: "2.2.2.2",
	})
	ut.Assert(t, err == nil, "")
	_, err = tx.Insert(&Rdata{
		Name:  "n2",
		Type:  "a",
		Rdata: "2.2.2.2",
	})
	ut.Assert(t, err == nil, "")
	_, err = tx.Insert(&Rdata{
		Name:  "n2",
		Type:  "a",
		Rdata: "2.2.2.2",
	})
	ut.Assert(t, err != nil, "")
	tx.Rollback()

	store.Clean()
	store.Destroy()
}
