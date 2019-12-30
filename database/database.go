package database

import(
	badger "github.com/dgraph-io/badger"
)

//AddToDb Adds to the database key value pair
func AddToDb(key []byte, value []byte]) {
	opts := badger.DefaultOptions
	opts.Dir = ""
	opts.ValueDir = ""
	kv, err := badger.NewKV(&opts)
	if err != nil {
		panic(err)
	}
	defer kv.Close()
	err = kv.Set(key,value)
	if err != nil {
		panic(err)
	}
}