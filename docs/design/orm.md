# Features
1. Primitive types to postgresql data type
- int, string, bool
- []int, []string
1. Resource integrity
1. Resource relationship

# Convensions
1. Table name is the snake format of the resource go struct name
1. Id is unique and is a string, when not specified will be generated use uuid.
1. use `pk` tag to speicify primiary key
1. use `uk` tag to include several composite attribute as unique constraint
1. `Get` and `Fill` interface returns slice of resource pointer

# Relationship between resource
1. Use
```go
type View struct {
    Id string
    Name string 
    Acl string `db:"referto"`
}

type Acl struct {
    Id string
    Ips []string
}
```
- `Acl` in `View` is valid acl id
- Used acl couldn't be deleted
1. One to many ownership  
```go
type View struct {
    Id string
    Name string 
    Acl string `db:"referto"`
}

type Zone struct {
    Id string
    View string `db:"ownby"`
}
```
- `View` in `Zone` is valid view id
- When view delete, all the zone owned by that view will be deleted too

1. Many to many ownership
```go
type View struct {
    Id string
    Name string 
    Acl string `db:"referto"`
}

type User struct {
    Id string
    Email string `db:"uk"`
}

type UserView struct {
    Id string
    User string `db:"ownby"`
    View string `db:"referto"`
}

views, err := tx.GetOwned("user", "user_id", "view") 
```
1. The relationship is a seperate table which means a seperate go struct
1. Delete owner will delete the relationship but not the resource being owned
1. The owned resource cann't be deleted once it is owned by any owner.
