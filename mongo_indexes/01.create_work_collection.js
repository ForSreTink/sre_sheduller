var dbName;
var worksCollectionName = "works";
var zonesCollectionName = "zones";

create_db = (connection, dataBaseName= "workScheduler", collectionName) => {

	var checkCollectionException = function (collectionName, result) {
		if (result.ok === 0)
			throw "checkCollectionException. Create collecion " + collectionName + " failed. Code: " + result.code + "; CodeName: " + result.codeName + "; errmsg = " + result.errmsg;
	}

	db = connection.getDB(dataBaseName);
	const collections = db.getCollectionNames()
	if (!collections.includes(collectionName))
	{
		print("Create collection " + collectionName + " in db " + dataBaseName);
	
		var result = db.createCollection(collectionName);
		printjson(result);
		checkCollectionException(collectionName, result);
	}
	else
	{
		print("Collection " + collectionName + " alredy exists in db " + dataBaseName);
	}
}

create_db(conn, dbName, worksCollectionName);
create_db(conn, dbName, zonesCollectionName);