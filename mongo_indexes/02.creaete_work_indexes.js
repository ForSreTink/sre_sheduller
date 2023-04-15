/* run once on one of mongos */
/* conn as input argument from ansible */

var dbName;
var worksCollectionName;
var zonesCollectionName;

/* create indexes */
const create_indexes = (connection, dbName = "workScheduler", worksCollectionName = "works", zonesCollectionName = "zones") => {
	var checkIndexException = function (indexName, result) {
		if (result.ok === 0)
			throw "CreateIndexException. Create index " + indexName + " failed. Code: " + result.code + "; CodeName: " + result.codeName + "; errmsg = " + result.errmsg;
	}

	print("Create indexes for " + dbName);
	const db = connection.getDB(dbName);
	const worksCollection = db.getCollection(worksCollectionName);
	const zonesCollection = db.getCollection(zonesCollectionName);

    indexName = 'zoneId unique'
    print("Create " + indexName + " index for " + zonesCollection);
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
    
    indexName = 'workId unique'
    print("Create " + indexName + " index for " + worksCollection);
	result = worksCollection.createIndex(
		{ 'workId': 1 },
		{
			'name': 'workId',
            'unique': true,
			'background': true
		}
	);
    printjson(result);
    checkIndexException(indexName, result)

    indexName = 'startDate_zone_status_compound'
    print("Create " + indexName + " index for " + worksCollection);
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
