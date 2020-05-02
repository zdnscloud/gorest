package db

import (
	"strconv"
	"strings"
	"testing"
	"time"

	ut "github.com/zdnscloud/cement/unittest"
)

type child struct {
	Id       string
	Name     string `sql:"uk"`
	Age      uint32
	Hobbies  []string
	Scores   []int
	Birthday time.Time
	Talented bool
}

func (t *child) Validate() error {
	return nil
}

type tuser struct {
	Id   string
	Name string `sql:"uk"`
	Age  int
	CId  int
}

func (t *tuser) Validate() error {
	return nil
}

type tuserTview struct {
	Id    string
	Tuser string `sql:"ownby"`
	Tview string `sql:"referto"`
}

func (t *tuserTview) Validate() error {
	return nil
}

type tview struct {
	Id   string
	Name string `sql:"uk"`
}

func (t *tview) Validate() error {
	return nil
}

type trr struct {
	Id    string
	Name  string `sql:"uk"`
	Tview string `sql:"referto,uk"`
	Ttl   int
}

func (t *trr) Validate() error {
	return nil
}

type tnest struct {
	Id    string
	Name  string         `sql:"uk"`
	Inner map[string]int `sql:"-"`
}

func (t *tnest) Validate() error {
	return nil
}

func TestCURDRecord(t *testing.T) {
	var store ResourceStore
	var err error

	mr, err := NewResourceMeta([]Resource{&child{}})
	ut.Assert(t, err == nil, "err should be nil but %v", err)
	store, err = NewRStore("./foo.db", "zdns", "zdns", "zdns", mr)
	ut.Equal(t, err, nil)

	tx, _ := store.Begin()
	birthDay := time.Now()
	c := &child{
		Name:     "ben",
		Age:      20,
		Hobbies:  []string{"movie", "music"},
		Scores:   []int{1, 3, 4},
		Birthday: birthDay,
		Talented: true,
	}
	r, err := tx.Insert(c)
	newStudent, ok := r.(*child)
	ut.Equal(t, ok, true)
	ut.NotEqual(t, newStudent.Id, "")

	children := []child{}
	tx.Fill(map[string]interface{}{"id": newStudent.Id}, &children)
	ut.Equal(t, len(children), 1)
	ut.Equal(t, children[0].Birthday.Unix(), birthDay.Unix())
	tx.Commit()

	tx, _ = store.Begin()
	c = &child{
		Id:       "xxxxxxx",
		Name:     "benxxxx",
		Age:      20,
		Hobbies:  []string{},
		Scores:   []int{},
		Birthday: time.Now(),
		Talented: true,
	}

	_, err = tx.Insert(c)
	ut.Equal(t, err, nil)

	children = []child{}
	tx.Fill(map[string]interface{}{"id": "xxxxxxx"}, &children)
	ut.Equal(t, len(children), 1)
	ut.Equal(t, len(children[0].Hobbies), 0)
	ut.Equal(t, len(children[0].Scores), 0)
	tx.Delete("child", map[string]interface{}{"id": "xxxxxxx"})
	tx.Commit()

	//tx automatic rollback
	tx, _ = store.Begin()
	c = &child{
		Name:     "nana",
		Age:      20,
		Hobbies:  []string{"movie", "music"},
		Scores:   []int{1, 3, 4},
		Birthday: time.Now(),
		Talented: false,
	}
	tx.Insert(c)
	c = &child{
		Name:     "ben",
		Age:      20,
		Hobbies:  []string{"movie", "music"},
		Scores:   []int{1, 3, 4},
		Birthday: time.Now(),
		Talented: false,
	}
	_, err = tx.Insert(c)
	ut.NotEqual(t, err, nil)
	err = tx.Commit()
	ut.NotEqual(t, err, nil)
	err = tx.RollBack()
	ut.NotEqual(t, err, nil)

	tx, _ = store.Begin()
	children = []child{}
	tx.Fill(map[string]interface{}{"Age": 20}, &children)
	ut.Equal(t, len(children), 1)

	c = &child{
		Name:     "nana",
		Age:      20,
		Hobbies:  []string{"movie", "music"},
		Scores:   []int{1, 3, 4},
		Birthday: time.Now(),
		Talented: true,
	}
	tx.Insert(c)
	studentsInterface, err := tx.Get("child", map[string]interface{}{"Age": 20})
	ut.Equal(t, err, nil)

	children, ok = studentsInterface.([]child)
	ut.Equal(t, ok, true)

	ut.Equal(t, len(children), 2)
	for _, s := range children {
		ut.Assert(t, s.Talented, "child talented isn't stored correctly")
	}

	rows, err := tx.Update("child", map[string]interface{}{"Age": uint32(0xffff)}, map[string]interface{}{"Name": "ben"})
	ut.Equal(t, err, nil)
	ut.Equal(t, rows, int64(1))

	children = []child{}
	err = tx.Fill(map[string]interface{}{"Age": uint32(0xffff)}, &children)
	ut.Equal(t, err, nil)
	ut.Equal(t, len(children), 1)
	ut.NotEqual(t, children[0].Id, "")

	rows, err = tx.Delete("child", map[string]interface{}{"Age": 20})
	ut.Equal(t, err, nil)
	ut.Equal(t, rows, int64(1))

	rows, err = tx.Delete("child", map[string]interface{}{"Age": 20})
	ut.Equal(t, rows, int64(0))

	children = []child{}
	err = tx.Fill(map[string]interface{}{"Age": 20}, &children)
	ut.Equal(t, err, nil)
	ut.Equal(t, len(children), 0)

	rows, err = tx.Delete("child", map[string]interface{}{})
	ut.Equal(t, err, nil)
	ut.Equal(t, rows, int64(1))

	children = []child{}
	err = tx.Fill(map[string]interface{}{}, &children)
	ut.Equal(t, err, nil)
	ut.Equal(t, len(children), 0)
	tx.RollBack()

	store.Clean()
	store.Destroy()
}

func TestGetEx(t *testing.T) {
	var store ResourceStore
	var err error

	mr, err := NewResourceMeta([]Resource{&child{}})
	ut.Assert(t, err == nil, "err should be nil but %v", err)
	store, err = NewRStore("./foo.db", "zdns", "zdns", "zdns", mr)
	ut.Equal(t, err, nil)

	tx, _ := store.Begin()
	for i := 0; i < 200; i++ {
		c := &child{
			Name: "c" + strconv.Itoa(i),
			Age:  uint32(i),
		}
		tx.Insert(c)
	}

	for i := 0; i < 200; i++ {
		c := &child{
			Name: "a" + strconv.Itoa(i),
			Age:  uint32(i),
		}
		tx.Insert(c)
	}
	tx.Commit()

	tx, _ = store.Begin()
	children := []child{}
	tx.FillEx(&children, "select distinct age from zc_child ORDER BY age")
	ut.Equal(t, len(children), 200)

	children = []child{}
	tx.FillEx(&children, "select * from zc_child where age > $1 and name like $2", 100, "c%")
	ut.Equal(t, len(children), 99)

	childrenI, _ := tx.GetEx(ResourceType("child"), "select * from zc_child")
	children, _ = childrenI.([]child)
	ut.Equal(t, len(children), 400)

	childrenI, _ = tx.GetEx(ResourceType("child"), "select * from zc_child where age between $1 and $2", 1, 10)
	children, _ = childrenI.([]child)
	ut.Equal(t, len(children), 20)

	tx.Commit()

	tx, _ = store.Begin()
	count, err := tx.DeleteEx("delete from zc_child where age >= $1 and age < $2", 50, 100)
	tx.Commit()
	ut.Equal(t, err, nil)
	ut.Equal(t, count, int64(100))

	store.Clean()
	store.Destroy()
}

func TestJoinSelect(t *testing.T) {
	var store ResourceStore
	var err error

	mr, _ := NewResourceMeta([]Resource{&tuser{}, &tview{}, &tuserTview{}, &trr{}})
	ut.Assert(t, err == nil, "err should be nil but %v", err)
	store, _ = NewRStore("./foo.db", "zdns", "zdns", "zdns", mr)

	tx, err := store.Begin()
	hasAnyuser, _ := tx.Exists(ResourceType("tuser"), map[string]interface{}{})
	ut.Equal(t, hasAnyuser, false)

	u := &tuser{Id: "1", Name: "ben"}
	tx.Insert(u)
	v := &tview{Id: "1", Name: "v1"}
	tx.Insert(v)
	v = &tview{Id: "2", Name: "v2"}
	tx.Insert(v)
	v = &tview{Id: "12", Name: "v12"}
	tx.Insert(v)

	hasAnyuser, _ = tx.Exists(ResourceType("tuser"), map[string]interface{}{})
	ut.Equal(t, hasAnyuser, true)

	uv := &tuserTview{Id: "1", Tuser: "1", Tview: "1"}
	tx.Insert(uv)

	uv = &tuserTview{Id: "2", Tuser: "1", Tview: "2"}
	tx.Insert(uv)

	result, err := tx.GetOwned(ResourceType("tuser"), "1", ResourceType("tview"))
	ut.Equal(t, err, nil)

	views, ok := result.([]tview)
	ut.Assert(t, ok, "get user owned view failed")
	ut.Equal(t, len(views), 2)
	tx.Commit()

	tx, _ = store.Begin()
	rr := &trr{Id: "1", Name: "a.cn", Tview: "1", Ttl: 10}
	tx.Insert(rr)
	rr = &trr{Id: "2", Name: "a.cn", Tview: "2", Ttl: 20}
	tx.Insert(rr)
	rr = &trr{Id: "3", Name: "a.a.cn", Tview: "12", Ttl: 30}
	tx.Insert(rr)

	count, _ := tx.Count(ResourceType("trr"), map[string]interface{}{})
	ut.Equal(t, count, int64(3))
	count, _ = tx.Count(ResourceType("trr"), map[string]interface{}{"id": 2})
	ut.Equal(t, count, int64(1))
	count, _ = tx.Count(ResourceType("trr"), map[string]interface{}{"name": "a.cn", "search": "name"})
	ut.Equal(t, count, int64(3))
	count, _ = tx.Count(ResourceType("trr"), map[string]interface{}{"name": "a.cn", "tview": "2", "search": "name"})
	ut.Equal(t, count, int64(1))
	count, _ = tx.Count(ResourceType("trr"), map[string]interface{}{"name": "a.cn", "tview": "2", "search": "name,tview"})
	ut.Equal(t, count, int64(2))
	count, err = tx.Count(ResourceType("trr"), map[string]interface{}{"tview": "2,12", "match_list": "tview"})
	ut.Equal(t, count, int64(2))
	count, err = tx.Count(ResourceType("trr"), map[string]interface{}{"ttl": "10,20,30", "match_list": "ttl"})
	ut.Equal(t, count, int64(3))
	count, err = tx.Count(ResourceType("trr"), map[string]interface{}{"ttl": "11,20,30", "match_list": "ttl"})
	ut.Equal(t, count, int64(2))

	count, _ = tx.Count(ResourceType("trr"), map[string]interface{}{"name": "a.cn"})
	ut.Equal(t, count, int64(2))
	count, _ = tx.CountEx(ResourceType("trr"), "select count(*) from zc_trr where id=$1", 2)
	ut.Equal(t, count, int64(1))

	hasTrr, _ := tx.Exists(ResourceType("trr"), map[string]interface{}{})
	ut.Equal(t, hasTrr, true)
	hasTrr, _ = tx.Exists(ResourceType("trr"), map[string]interface{}{"id": 10000})
	ut.Equal(t, hasTrr, false)
	hasTrr, _ = tx.Exists(ResourceType("trr"), map[string]interface{}{"id": 2})
	ut.Equal(t, hasTrr, true)

	rrs := []trr{}
	tx.Fill(map[string]interface{}{"Name": "a.cn"}, &rrs)
	ut.Equal(t, len(rrs), 2)
	tx.RollBack()

	tx, _ = store.Begin()
	result, err = tx.GetOwned(ResourceType("tuser"), "2", ResourceType("tview"))
	views, _ = result.([]tview)
	if len(views) != 0 {
		t.Fatal("should return null views")
	}
	tx.RollBack()

	store.Clean()
	store.Destroy()
}

func TestGetWithLimitAndOffset(t *testing.T) {
	var store ResourceStore
	mr, _ := NewResourceMeta([]Resource{&tuser{}})
	store, _ = NewRStore("./foo.db", "zdns", "zdns", "zdns", mr)

	tx, _ := store.Begin()
	for i := 0; i < 2000; i++ {
		if i < 1000 {
			tx.Insert(&tuser{Id: strconv.Itoa(i), Name: "ben" + strconv.Itoa(i), Age: 40, CId: i})
		} else {
			tx.Insert(&tuser{Id: strconv.Itoa(i), Name: "nana" + strconv.Itoa(i), Age: 50, CId: i})
		}
	}
	tx.Commit()

	tx, _ = store.Begin()
	var users []tuser
	tx.Fill(map[string]interface{}{"age": 40, "offset": 10, "limit": 10, "orderby": "CId"}, &users)
	ut.Equal(t, len(users), 10)
	for i := 0; i < 10; i++ {
		ut.Assert(t, strings.HasPrefix(users[i].Name, "ben"), "")
		ut.Equal(t, users[i].CId, i+10)
	}
	tx.Commit()

	store.Clean()
	store.Destroy()
}

func TestSearch(t *testing.T) {
	var store ResourceStore
	mr, _ := NewResourceMeta([]Resource{&tuser{}})
	store, _ = NewRStore("./foo.db", "zdns", "zdns", "zdns", mr)

	tx, _ := store.Begin()
	tx.Insert(&tuser{Name: "ben", Age: 330, CId: 0})
	tx.Insert(&tuser{Name: "bean", Age: 30, CId: 0})
	tx.Insert(&tuser{Name: "baan", Age: 40, CId: 0})
	tx.Commit()

	tx, _ = store.Begin()
	var users []tuser
	tx.Fill(map[string]interface{}{"name": "be", "search": "name"}, &users)
	ut.Equal(t, len(users), 2)

	users = []tuser{}
	tx.Fill(map[string]interface{}{"name": "be", "age": 30, "search": "name"}, &users)
	ut.Equal(t, len(users), 1)

	users_, err := tx.Get("tuser", map[string]interface{}{"name": "b", "search": "name"})
	ut.Equal(t, err, nil)
	users, ok := users_.([]tuser)
	ut.Assert(t, ok, "get should return tusers")
	ut.Equal(t, len(users), 3)

	users_, err = tx.Get("tuser", map[string]interface{}{"age": "330,30", "match_list": "age"})
	users, ok = users_.([]tuser)
	ut.Assert(t, ok, "get should return tusers")
	ut.Equal(t, len(users), 2)
	tx.Commit()

	store.Clean()
	store.Destroy()
}

func TestErrorMessage(t *testing.T) {
	var store ResourceStore

	mr, _ := NewResourceMeta([]Resource{&tuser{}, &tview{}, &tuserTview{}, &trr{}})
	store, _ = NewRStore("./foo.db", "zdns", "zdns", "zdns", mr)

	tx, _ := store.Begin()
	u := &tuser{Id: "1", Name: "ben"}
	tx.Insert(u)
	u = &tuser{Id: "2", Name: "ben2"}
	tx.Insert(u)
	tx.Commit()

	tx, _ = store.Begin()
	u = &tuser{Id: "1", Name: "ben"}
	_, err := tx.Insert(u)
	ut.Assert(t, strings.HasPrefix(err.Error(), DuplicateErrorMsg), "duplicate resource error")
	tx.RollBack()

	tx, _ = store.Begin()
	_, err = tx.Update("tuser", map[string]interface{}{"id": "1"}, map[string]interface{}{"id": "2"})
	ut.Assert(t, strings.HasPrefix(err.Error(), DuplicateErrorMsg), "duplicate resource error")
	tx.RollBack()

	tx, _ = store.Begin()
	uv := &tuserTview{Id: "1", Tuser: "1", Tview: "1"}
	_, err = tx.Insert(uv)
	ut.Assert(t, strings.HasPrefix(err.Error(), RelatedNoneExistsErrorMsg), "refer to resource doesn't exists")
	tx.RollBack()

	store.Clean()
	store.Destroy()
}

func TestIgnField(t *testing.T) {
	var store ResourceStore
	mr, _ := NewResourceMeta([]Resource{&tnest{}})
	store, _ = NewRStore("./foo.db", "zdns", "zdns", "zdns", mr)

	tx, _ := store.Begin()
	u := &tnest{Id: "1", Name: "ben", Inner: map[string]int{"good": 1}}
	tx.Insert(u)
	tx.Commit()

	tx, _ = store.Begin()
	var nests []tnest
	tx.Fill(map[string]interface{}{"id": 1}, &nests)
	ut.Equal(t, len(nests), 1)
	ut.Equal(t, nests[0].Inner, map[string]int(nil))
	tx.Commit()

	store.Clean()
	store.Destroy()
}

type trrr struct {
	Id   string
	Name string `sql:"suk"`
	Age  int    `sql:"suk"`
}

func (t *trrr) Validate() error {
	return nil
}

func TestUniqueField(t *testing.T) {
	var store ResourceStore
	mr, _ := NewResourceMeta([]Resource{&trrr{}})
	store, _ = NewRStore("./foo.db", "zdns", "zdns", "zdns", mr)

	tx, _ := store.Begin()
	rr := &trrr{Id: "1", Name: "a.cn", Age: 1}
	_, err := tx.Insert(rr)
	ut.Assert(t, err == nil, "err should be nil but %v", err)
	tx.Commit()

	tx, _ = store.Begin()
	rr = &trrr{Id: "2", Name: "a.cn", Age: 2}
	_, err = tx.Insert(rr)
	ut.Assert(t, err != nil, "duplicate name should raise")
	tx.RollBack()

	tx, _ = store.Begin()
	rr = &trrr{Id: "2", Name: "a.cn.", Age: 1}
	_, err = tx.Insert(rr)
	ut.Assert(t, err != nil, "duplicate age should raise")
	tx.RollBack()

	tx, _ = store.Begin()
	rr = &trrr{Id: "2", Name: "a.cn.", Age: 2}
	_, err = tx.Insert(rr)
	ut.Assert(t, err == nil, "no error should get but:%v", err)
	tx.Commit()

	store.Clean()
	store.Destroy()
}

func BenchmarkInsert(b *testing.B) {
	var store ResourceStore
	mr, _ := NewResourceMeta([]Resource{&trrr{}})
	store, _ = NewRStore("./foo.db", "zdns", "zdns", "zdns", mr)
	for i := 0; i < b.N; i++ {
		tx, _ := store.Begin()
		rr := &trrr{Name: "a.cn.", Age: 1}
		tx.Insert(rr)
		tx.Commit()
	}

	store.Clean()
	store.Destroy()
}

func BenchmarkBatchInsert(b *testing.B) {
	var store ResourceStore
	mr, _ := NewResourceMeta([]Resource{&trrr{}})
	store, _ = NewRStore("./foo.db", "zdns", "zdns", "zdns", mr)
	tx, _ := store.Begin()
	for i := 0; i < b.N; i++ {
		rr := &trrr{Name: "a.cn.", Age: 1}
		tx.Insert(rr)
		if i%30 == 0 {
			tx.Commit()
			tx, _ = store.Begin()
		}
	}
	tx.Commit()
	store.Clean()
	store.Destroy()
}
