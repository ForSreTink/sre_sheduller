var dbName = "workScheduler";
var worksCollectionName = "works"
var zonesCollectionName = "zones";

const create_indexes = (connection, dbName, worksCollectionName, zonesCollectionName) => {
	var checkIndexException = function (indexName, result) {
		if (result.ok === 0)
			throw "CreateIndexException. Create index " + indexName + " failed. Code: " + result.code + "; CodeName: " + result.codeName + "; errmsg = " + result.errmsg;
	}

	print("Create indexes for " + dbName);
	const db = connection.getDB(dbName);
	const worksCollection = db.getCollection(worksCollectionName);
	const zonesCollection = db.getCollection(zonesCollectionName);

    indexName = 'zoneId unique'
    print("Create " + indexName + " index for " + zonesCollectionName);
    result = zonesCollection.createIndex(
		{ 'zoneId': 1 },
		{
			'name': indexName,
            'unique': true,
			'background': true
		}
	);
    printjson(result);
    checkIndexException(indexName, result)
    
    indexName = 'workId'
    print("Create " + indexName + " index for " + worksCollectionName);
	result = worksCollection.createIndex(
		{ 'workId': 1 },
		{
			'name': 'workId',
            'unique': false,
			'background': true
		}
	);
    printjson(result);
    checkIndexException(indexName, result)

    indexName = 'startDate_zone_status_compound'
    print("Create " + indexName + " index for " + worksCollectionName);
	result = worksCollection.createIndex(
		{
			"startDate": 1,
            "zone": 1,
			"status": 1
		},
		{
			'name': indexName,
			'background': true
		}
	);
    printjson(result);
    checkIndexException(indexName, result)
}

create_indexes(conn, dbName, worksCollectionName, zonesCollectionName);
