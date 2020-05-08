package db

import (
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	ut "github.com/zdnscloud/cement/unittest"
	"github.com/zdnscloud/gorest/resource"
)

const ConnStr string = "user=lx password=lx host=localhost port=5432 database=lx sslmode=disable pool_max_conns=10"

type Child struct {
	resource.ResourceBase

	Name     string `db:"uk"`
	Age      uint32
	Hobbies  []string
	Scores   []int
	Birthday time.Time
	Ipaddr   net.IP
	Subnet   net.IPNet
	Talented bool
}

type Mother struct {
	resource.ResourceBase
	Age  int
	Name string
}

type MotherChild struct {
	resource.ResourceBase
	Mother string `db:"ownby"`
	Child  string `db:"referto"`
}

func initChild(store ResourceStore) {
	tx, _ := store.Begin()
	c1 := &Child{
		Name:     "ben",
		Age:      20,
		Hobbies:  []string{"movie", "music"},
		Scores:   []int{1, 3, 4},
		Birthday: time.Now(),
		Talented: true,
	}
	ip, ipnet, _ := net.ParseCIDR("1.1.1.1/24")
	c1.Ipaddr = ip
	c1.Subnet = *ipnet
	c1.SetID("c1")
	_, err := tx.Insert(c1)
	if err != nil {
		fmt.Printf("insert get err:%v\n", err.Error())
	}

	c2 := &Child{
		Name:     "nana",
		Age:      30,
		Hobbies:  []string{"movie", "music"},
		Scores:   []int{1, 2, 4},
		Birthday: time.Now(),
		Talented: true,
	}
	c2.SetID("c2")
	ip, ipnet, _ = net.ParseCIDR("1.2.1.1/16")
	c2.Ipaddr = ip
	c2.Subnet = *ipnet
	_, err = tx.Insert(c2)
	if err != nil {
		fmt.Printf("insert get err:%v\n", err.Error())
	}
	tx.Commit()
}

func initMother(store ResourceStore) {
	tx, _ := store.Begin()
	m := &Mother{
		Name: "lxq",
	}
	m.SetID("m1")
	tx.Insert(m)
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
	meta, err := NewResourceMeta([]resource.Resource{&Child{}})
	ut.Assert(t, err == nil, "")
	store, err := NewRStore(ConnStr, meta)
	ut.Assert(t, err == nil, "err str is %v", err)

	initChild(store)

	tx, _ := store.Begin()
	c, err := tx.Count("child", nil)
	ut.Equal(t, c, int64(2))
	exist, _ := tx.Exists("child", map[string]interface{}{"talented": true})
	ut.Assert(t, exist, "")
	exist, _ = tx.Exists("child", map[string]interface{}{"talented": false})
	ut.Assert(t, !exist, "")
	children := []*Child{}
	tx.Fill(map[string]interface{}{IDField: "c1"}, &children)
	ut.Equal(t, len(children), 1)
	ut.Equal(t, children[0].Scores, []int{1, 3, 4})
	tx.Rollback()

	tx, _ = store.Begin()
	c, err = tx.Update("child", map[string]interface{}{
		"hobbies": []string{"read book", "travel"},
	}, map[string]interface{}{IDField: "c1"})
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
	ut.Equal(t, children_.([]*Child)[0].GetID(), "c1")
	tx.Rollback()

	children = []*Child{}
	ci, err := GetResourceWithID(store, "c1", &children)
	ut.Assert(t, err == nil, "")
	ut.Equal(t, ci.(*Child).Age, uint32(20))

	store.Clean()
	store.Close()
}

func TestCURDEx(t *testing.T) {
	meta, err := NewResourceMeta([]resource.Resource{&Child{}})
	ut.Assert(t, err == nil, "")
	store, err := NewRStore(ConnStr, meta)
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
	count, err := tx.Exec("delete from gr_child where age >= $1 and age < $2", 25, 400)
	tx.Commit()
	ut.Equal(t, err, nil)
	ut.Equal(t, count, int64(1))

	store.Clean()
	store.Close()
}

func TestMultiToMultiRelationship(t *testing.T) {
	meta, err := NewResourceMeta([]resource.Resource{&Mother{}, &Child{}, &MotherChild{}})
	ut.Assert(t, err == nil, "")
	store, err := NewRStore(ConnStr, meta)
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
	store.Close()
}

type View struct {
	resource.ResourceBase
	Name string `db:"uk"`
}

type Zone struct {
	resource.ResourceBase
	Name string
	View string `db:"ownby"`
}

func TestOneToManyRelationship(t *testing.T) {
	meta, err := NewResourceMeta([]resource.Resource{&View{}, &Zone{}})
	ut.Assert(t, err == nil, "")
	store, err := NewRStore(ConnStr, meta)
	ut.Assert(t, err == nil, "")

	tx, _ := store.Begin()
	v1 := &View{Name: "v1"}
	v1.SetID("v1")
	v2 := &View{Name: "v2"}
	v2.SetID("v2")
	cnZone := &Zone{Name: "cn", View: "v1"}
	comZone := &Zone{Name: "com", View: "v1"}
	tx.Insert(v1)
	tx.Insert(v2)
	tx.Insert(cnZone)
	tx.Insert(comZone)
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
		IDField: "v2",
	})
	ut.Assert(t, err == nil, "")
	ut.Equal(t, c, int64(1))
	c, _ = tx.Count("zone", nil)
	ut.Equal(t, c, int64(2))
	c, err = tx.Delete("view", map[string]interface{}{
		IDField: "v1",
	})
	ut.Assert(t, err == nil, "")
	ut.Equal(t, c, int64(1))
	c, _ = tx.Count("zone", nil)
	ut.Equal(t, c, int64(0))
	tx.Commit()

	store.Clean()
	store.Close()
}

func TestGetWithLimitAndOffset(t *testing.T) {
	meta, err := NewResourceMeta([]resource.Resource{&Mother{}})
	ut.Assert(t, err == nil, "")
	store, err := NewRStore(ConnStr, meta)
	ut.Assert(t, err == nil, "")

	tx, _ := store.Begin()
	for i := 0; i < 2000; i++ {
		r, err := tx.Insert(&Mother{Age: i, Name: "m" + strconv.Itoa(i)})
		ut.Assert(t, err == nil, "")
		ut.Assert(t, r.GetID() != "", "")
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
	store.Close()
}

type Student struct {
	resource.ResourceBase
	Name      string `db:"uk"`
	Age       uint32
	Classroom string `db:"-"`
}

func TestIgnField(t *testing.T) {
	meta, err := NewResourceMeta([]resource.Resource{&Student{}})
	ut.Assert(t, err == nil, "")
	store, err := NewRStore(ConnStr, meta)
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
	store.Close()
}

type Rdata struct {
	resource.ResourceBase
	Name  string `db:"uk"`
	Type  string `db:"uk"`
	Rdata string `db:"uk"`
	Addrs []net.IP
}

func TestUniqueField(t *testing.T) {
	meta, err := NewResourceMeta([]resource.Resource{&Rdata{}})
	ut.Assert(t, err == nil, "")
	store, err := NewRStore(ConnStr, meta)
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
	ips := []net.IP{
		net.ParseIP("2.2.2.2"),
		net.ParseIP("3.3.33.3"),
	}
	_, err = tx.Insert(&Rdata{
		Name:  "n2",
		Type:  "a",
		Rdata: "2.2.2.2",
		Addrs: ips,
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
	store.Close()
}
