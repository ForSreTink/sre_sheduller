var dbName;
var worksCollectionName;
var zonesCollectionName;

create_db = (connection, dataBaseName= "workScheduler", collectionName = "works") => {
	print("Create collection " + collectionName + " in db " + dataBaseName);
	db = connection.getDB(dataBaseName);
	var result = db.createCollection(collectionName);
	printjson(result);
}

create_db(conn, dbName, worksCollectionName);
create_db(conn, dbName, zonesCollectionName);