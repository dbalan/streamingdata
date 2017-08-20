var PROTO_PATH = "../streamingdata/streamingdata.proto";

var sleep = require('sleep');
var grpc = require('grpc');
var cmdargs = require('command-line-args');
var crypto = require("crypto");
var streaming_proto = grpc.load(PROTO_PATH).streamingdata;

function main() {
    const optionDefinitions = [
        { name: 'server', alias: 's', type: String, defaultValue: "localhost:8000" },
        { name: 'mode', type: String, defaultValue: 'stateful' },
        { name: 'count', type: Number, defaultValue: 10},
    ];
    const options = cmdargs(optionDefinitions);

    // FIXME: usage string
    if (options['mode'] != 'stateful' && options['mode'] != 'stateless') {
        console.log('choose correct mode!')
        process.exit(-1)
    }
    var client = new streaming_proto.RealTime(options['server'], grpc.credentials.createInsecure());

    if (options['mode'] == 'stateless') {
        stateLessQuery(client, options['count'], 0, 0);
        return;
    }

    var clientID = genClientID();
    var req = {'clientid': clientID, 'count': options['count']};
    var call = client.getStateFullStream(req);
    call.on('data', function(resp){
        console.log(resp);
        if (resp.hash_sum != ''){
            console.log("hash is: "+resp.hash_sum);
        }
    });
    call.on('end', function(resp) {
        console.log('stream ended');
    });
    call.on('error', function(){
        console.log('error');
    });
}

function genClientID() {
    return crypto.randomBytes(20).toString('hex');
}
function stateLessQuery(client, count, last, sum) {
    var statelessreq = {'count': count};
    if (last != 0) {
        // we're restarting
        statelessreq['lastseen'] = last;
    }
    var call = client.getStateLessStream(statelessreq);
    call.on('data', function(resp){
        // debug
        console.log('value is:', resp.current_val);
        last = resp.current_val;
        sum += resp.current_val;
        count--;
    });
    call.on('end', function(){
        console.log("sum is: "+sum);
    });
    call.on('error', function(){
        console.log('error sleeping');
        // FIXME: make this exponential, and generalize with the other API
        sleep.sleep(4);
        stateLessQuery(client, count, last, sum);
    });
}

main();
