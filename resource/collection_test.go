package resource

import (
	"encoding/json"
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
)

func TestCollectionToJson(t *testing.T) {
	r := &dumbResource{
		Number: 10,
	}

	rs, err := NewResourceCollection(r, nil)
	ut.Assert(t, err == nil, "")
	ut.Assert(t, rs.Resources != nil, "")
	d, _ := json.Marshal(rs)
	ut.Equal(t, string(d), `{"type":"collection","data":[]}`)

	rs2, err := NewResourceCollection(r, []*dumbResource{})
	ut.Assert(t, err == nil, "")
	ut.Assert(t, rs2.Resources != nil, "")
	ut.Equal(t, len(rs2.Resources), 0)
	d2, _ := json.Marshal(rs2)
	ut.Equal(t, string(d), string(d2))
}

func TestPagination(t *testing.T) {
	var rs []*dumbResource
	for i := 0; i < 55; i++ {
		rs = append(rs, &dumbResource{
			Number: i,
		})
	}
	r := &dumbResource{}
	r.SetType("dumbresource")
	rc, err := NewResourceCollection(r, rs)
	ut.Assert(t, err == nil, "")
	rc.ApplyPagination(&Pagination{PageSize: 10, PageNum: 5})
	ut.Assert(t, len(rc.Resources) == 10, "")
	ut.Assert(t, rc.Pagination.PageTotal == 6, "")
	ut.Assert(t, rc.Pagination.PageNum == 5, "")
	ut.Assert(t, rc.Pagination.Total == 55, "")
	ut.Assert(t, rc.Pagination.PageSize == 10, "")

	rc.ApplyPagination(&Pagination{PageSize: 20, PageNum: 5})
	ut.Assert(t, len(rc.Resources) == 10, "")
	ut.Assert(t, rc.Pagination.PageTotal == 1, "")
	ut.Assert(t, rc.Pagination.PageNum == 1, "")
	ut.Assert(t, rc.Pagination.Total == 10, "")
	ut.Assert(t, rc.Pagination.PageSize == 10, "")
}
