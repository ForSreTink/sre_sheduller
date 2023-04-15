function connect(user, password, authSource = "admin")
{
    var auth_db;
    var conn = db.getMongo();
    print("Authenticating on " + authSource + " with user " + user);
    auth_db = conn.getDB(authSource);
    const result = auth_db.auth(user, password);
    print(result);
    return auth_db.getMongo();
}

function get_files(directory){
    if (directory === "") {
        directory = process.cwd()
    }
    print('Search for *.js files in directory ' + directory);
    const fs = require("fs");
    const path = require("path");
    var files = [];
    const filesInDirectory = fs.readdirSync(directory);

    for (const file of filesInDirectory) {
        const absolute = path.join(directory, file);
        if (file.match("\.js$") && !file.match("script_runner\.js$")){
            files.push(absolute);
        }
    };
    return files;
}

function run_script(conn, file)
{
    print('Run script ' + file);
    load (file.toString());
}

disableTelemetry();

var scripts_folder= process.env['SCRIPTS_FOLDER'];
var user = process.env['MONGO_SCRIPTS_USER'];
var password = process.env['MONGO_SCRIPTS_PASSWORD'];
var authSource = process.env['MONGO_SCRIPTS_AUTH_SOURCE'];

var script_files = get_files(scripts_folder);
var conn = connect(user, password, authSource)
print ("Run script files: " + script_files)
script_files.forEach(function(file) {
    run_script(conn, file)
});
print ("All scripts completed")