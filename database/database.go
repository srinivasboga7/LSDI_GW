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


func GetAllKeys() []string {
	Keys := make([]string,0)
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Dir = ""
		// opts.ValueDir = ""
		opts.PrefetchSize = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			Keys = append(Keys, it.Item().Key())
		}
		return nil
	})
	// fmt.Println(len(Keys))
	return Keys

	// db, err := sql.Open("mysql","Sumanth:sumanth@tcp(127.0.0.1:3306)/dag")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer db.Close()
	// stmt, err := db.Prepare("SELECT Hash_tx FROM storage")
	// rows, err := stmt.Query()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer rows.Close()
	// Hashes := make([]string,0)
	// for rows.Next() {
	// 	var u string
	// 	err := rows.Scan(&u) // check err
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	Hashes = append(Hashes, u)
	// }
	// fmt.Println(len(Hashes))
	// return Hashes
}

// GetTransaction returns transaction based on hash value.
func GetValue(key []byte) []byte {
	
	opts := badger.DefaultOptions
	opts.Dir = ""
	opts.ValueDir = ""
	kv, err := badger.NewKV(&opts)
	if err != nil {
		panic(err)
	}
	defer kv.Close()
	var item badger.KVItem
	err = kv.Get(key,&item)
	if err == ErrKeyNotFound {
		return false
	}
	return item.Value()
	//return serialize.Deserializedata(item.Value())


	// db, err := sql.Open("mysql","Sumanth:sumanth@tcp(127.0.0.1:3306)/dag")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer db.Close()
	// var Resp string
	// queryStr := "SELECT Transaction FROM storage WHERE Hash_tx = ?" // check err
	// err1 := db.QueryRow(queryStr, hash).Scan(&Resp)
	// if err1 != nil {
	// 	log.Fatal(err1)
	// }
	// return serialize.Deserializedata(DecodeToBytes(Resp))
}

// GetSignature returns signature of tranasction based on hash value.
// func GetSignature(hash string) []byte {
// 	opts := badger.DefaultOptions
// 	opts.Dir = ""
// 	opts.ValueDir = ""
// 	kv, err := badger.NewKV(&opts)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer kv.Close()
// 	var item badger.KVItem
// 	err = kv.Get([]byte(h),&item)
// 	if err == ErrKeyNotFound {
// 		return false
// 	}


// 	// db, err := sql.Open("mysql","Sumanth:sumanth@tcp(127.0.0.1:3306)/dag")
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }
// 	// defer db.Close()
// 	// Resp := make([]byte,0)
// 	// queryStr := `SELECT Signature FROM storage WHERE Hash_tx = ?` // check err
// 	// err1 := db.QueryRow(queryStr, hash).Scan(&Resp)
// 	// if err1 != nil {
// 	// 	log.Fatal(err1)
// 	// }
// 	// return Resp
// }

//CheckKey checks if a key-value pair is present in the database, returns true if present else false
func CheckKey(key []byte) bool{
	opts := badger.DefaultOptions
	opts.Dir = ""
	opts.ValueDir = ""
	kv, err := badger.NewKV(&opts)
	if err != nil {
		panic(err)
	}
	defer kv.Close()
	var item badger.KVItem
	err = kv.Get(key,&item)
	if err == ErrKeyNotFound {
		return false
	} else if err != nil {
		panic(err)
	} else {
		return true
	}

	// db, err := sql.Open("mysql","Sumanth:sumanth@tcp(127.0.0.1:3306)/dag")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer db.Close()
	// stmt, err := db.Prepare("SELECT EXISTS(SELECT 1 FROM storage WHERE Hash_tx = ?)")
	// res, err := stmt.Query(h)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer res.Close()
	// var present int
	// res.Next()
	// err1 := res.Scan(&present)
	// if err1 != nil {
	// 	log.Fatal(err1)
	// }
	// if(present==1){
	// 	return true
	// } else{
	// 	return false
	// }
}