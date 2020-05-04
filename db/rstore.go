package db

import (
	"fmt"
	"reflect"
	"text/template"
	"time"

	"github.com/zdnscloud/cement/reflector"
	"github.com/zdnscloud/gorest/resource"
)

type RStore struct {
	db   *db
	meta *ResourceMeta
}

type RStoreTx struct {
	*Tx
	meta *ResourceMeta
}

const (
	joinSqlTemplateContent string = "select {{.OwnedTable}}.* from {{.OwnedTable}} inner join {{.RelTable}} on ({{.OwnedTable}}.id={{.RelTable}}.{{.Owned}} and {{.RelTable}}.{{.Owner}}=$1)"
)

var joinSqlTemplate *template.Template

func init() {
	joinSqlTemplate, _ = template.New("").Parse(joinSqlTemplateContent)
}

func NewRStore(connInfo map[string]interface{}, meta *ResourceMeta) (ResourceStore, error) {
	db, err := OpenDB(connInfo)
	if err != nil {
		return nil, err
	}

	for _, descriptor := range meta.GetDescriptors() {
		db.Exec(createTableSql(descriptor))
	}

	return &RStore{db, meta}, nil
}

func (store *RStore) Destroy() {
	store.db.CloseDB()
}

func (store *RStore) Clean() {
	rs := store.meta.Resources()
	for i := len(rs); i > 0; i-- {
		store.db.DropTable(resourceTableName(rs[i-1]))
	}
}

func (store *RStore) Begin() (Transaction, error) {
	tx, err := store.db.Begin()
	if err != nil {
		return nil, err
	} else {
		return &RStoreTx{tx, store.meta}, nil
	}
}

func (tx *RStoreTx) Insert(r resource.Resource) (resource.Resource, error) {
	r.SetCreationTimestamp(time.Now())
	sql, args, err := insertSqlArgsAndID(tx.meta, r)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(sql, args...)
	if err != nil {
		return nil, err
	} else {
		return r, err
	}
}

func (tx *RStoreTx) GetOwned(owner ResourceType, ownerID string, owned ResourceType) (interface{}, error) {
	goTyp, err := tx.meta.GetGoType(owned)
	if err != nil {
		return nil, err
	}
	sp := reflector.NewSlicePointer(reflect.PtrTo(goTyp))
	sql, args, err := joinSelectSqlAndArgs(tx.meta, owner, owned, ownerID)
	if err != nil {
		return nil, err
	}

	err = tx.getWithSql(sql, args, sp)
	if err != nil {
		return nil, err
	} else {
		return reflect.ValueOf(sp).Elem().Interface(), nil
	}
}

func (tx *RStoreTx) FillOwned(owner ResourceType, ownerID string, out interface{}) error {
	r, err := reflector.GetStructPointerInSlice(out)
	if err != nil {
		return err
	}

	sql, args, err := joinSelectSqlAndArgs(tx.meta, owner, ResourceDBType(r.(resource.Resource)), ownerID)
	if err != nil {
		return err
	}

	return tx.getWithSql(sql, args, out)
}

func (tx *RStoreTx) Get(typ ResourceType, conds map[string]interface{}) (interface{}, error) {
	goTyp, err := tx.meta.GetGoType(typ)
	if err != nil {
		return nil, err
	}
	sp := reflector.NewSlicePointer(reflect.PtrTo(goTyp))
	err = tx.Fill(conds, sp)
	if err != nil {
		return nil, err
	} else {
		return reflect.ValueOf(sp).Elem().Interface(), nil
	}
}

func (tx *RStoreTx) Fill(conds map[string]interface{}, out interface{}) error {
	r, err := reflector.GetStructPointerInSlice(out)
	if err != nil {
		return err
	}

	sql, args, err := selectSqlAndArgs(tx.meta, ResourceDBType(r.(resource.Resource)), conds)
	if err != nil {
		return err
	}
	return tx.getWithSql(sql, args, out)
}

func (tx *RStoreTx) getWithSql(sql string, args []interface{}, out interface{}) error {
	rows, err := tx.Query(sql, args...)
	if err != nil {
		return err
	}

	return rowsToResources(rows, out)
}

func (tx *RStoreTx) Exists(typ ResourceType, conds map[string]interface{}) (bool, error) {
	sql, params, err := existsSqlAndArgs(tx.meta, typ, conds)
	if err != nil {
		return false, err
	}

	return tx.existsWithSql(sql, params...)
}

func (tx *RStoreTx) existsWithSql(sql string, params ...interface{}) (bool, error) {
	rows, err := tx.Query(sql, params...)
	if err != nil {
		return false, err
	}

	var exist bool
	//there should only one row
	for rows.Next() {
		if err := rows.Scan(&exist); err != nil {
			return false, err
		}
	}
	return exist, nil
}

func (tx *RStoreTx) Count(typ ResourceType, conds map[string]interface{}) (int64, error) {
	sql, params, err := countSqlAndArgs(tx.meta, typ, conds)
	if err != nil {
		return 0, err
	}

	return tx.countWithSql(sql, params...)
}

func (tx *RStoreTx) CountEx(typ ResourceType, sql string, params ...interface{}) (int64, error) {
	if tx.meta.Has(typ) == false {
		return 0, fmt.Errorf("unknown resource type %v", typ)
	}
	return tx.countWithSql(sql, params...)
}

func (tx *RStoreTx) countWithSql(sql string, params ...interface{}) (int64, error) {
	rows, err := tx.Query(sql, params...)
	if err != nil {
		return 0, err
	}

	var count int64
	//there should only one row
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, err
		}
	}

	return count, nil
}

func (tx *RStoreTx) Update(typ ResourceType, nv map[string]interface{}, conds map[string]interface{}) (int64, error) {
	sql, args, err := updateSqlAndArgs(tx.meta, typ, nv, conds)
	if err != nil {
		return 0, err
	}

	return tx.Exec(sql, args...)
}

func (tx *RStoreTx) Delete(typ ResourceType, conds map[string]interface{}) (int64, error) {
	sql, args, err := deleteSqlAndArgs(tx.meta, typ, conds)
	if err != nil {
		return 0, err
	}

	return tx.Exec(sql, args...)
}

func (tx *RStoreTx) DeleteEx(sql string, params ...interface{}) (int64, error) {
	return tx.Exec(sql, params...)
}

func (tx *RStoreTx) GetEx(typ ResourceType, sql string, params ...interface{}) (interface{}, error) {
	rt, err := tx.meta.GetGoType(typ)
	if err != nil {
		return nil, err
	}
	sp := reflector.NewSlicePointer(reflect.PtrTo(rt))
	err = tx.FillEx(sp, sql, params...)
	if err != nil {
		return nil, err
	} else {
		return reflect.ValueOf(sp).Elem().Interface(), nil
	}
}

func (tx *RStoreTx) FillEx(out interface{}, sql string, params ...interface{}) error {
	return tx.getWithSql(sql, params, out)
}

func (tx *RStoreTx) Exec(sql string, params ...interface{}) (int64, error) {
	result, err := tx.Tx.Exec(sql, params...)
	if err != nil {
		return 0, err
	} else {
		return result.RowsAffected(), nil
	}
}
