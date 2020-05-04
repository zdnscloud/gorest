package db

var postgresqlTypeMap = map[Datatype]string{
	Bool:        "boolean",
	Int:         "integer",
	Uint:        "integer",
	String:      "text",
	Time:        "timestamp with time zone",
	IP:          "cidr",
	IPNet:       "inet",
	IntArray:    "integer[]",
	StringArray: "text[]",
	IPSlice:     "cidr[]",
	IPNetSlice:  "inet[]",
}

/*
func OpenPostgresql(host, user, passwd, dbname string) (*db, error) {
	port := 5432
	hostAndPort := strings.Split(host, ":")
	if len(hostAndPort) == 2 {
		host = hostAndPort[0]
		port, _ = strconv.Atoi(hostAndPort[1])
	} else {
		host = hostAndPort[0]
	}
	var conninfo = fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		host,
		port,
		user,
		dbname,
		passwd,
	)
	conn, err := pgx.Connect(context.Background(), conninfo)
	if err != nil {
		return nil, err
	} else {
		return &db{
			conn: conn,
		}, nil
	}
}
*/
