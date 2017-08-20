var PROTO_PATH = "../streamingdata/streamingdata.proto";

var sleep = require('sleep');
var grpc = require('grpc');
var cmdargs = require('command-line-args');
var crypto = require("crypto");
var sha256 = require('js-sha256');

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
        console.log('choose correct mode!');
        process.exit(-1);
    }
    var client = new streaming_proto.RealTime(options['server'], grpc.credentials.createInsecure());

    if (options['mode'] == 'stateless') {
        stateLessQuery(client, options['count'], 0, 0,1);
        return;
    }

    stateFullQuerySetup(client, options['count']);
}

function genClientID() {
    return crypto.randomBytes(20).toString('hex');
}

function stateFullQuerySetup(client, count) {
    var clientID = genClientID();
    var hash = sha256.create();
    var last_hash = '';
    stateFullQuery(client, count, clientID, last_hash, hash, 1);
}

function stateFullQuery(client, count, clientID, last_hash, hash, retry) {
    var req = {'clientid': clientID, 'count': count};
    var call = client.getStateFullStream(req);
    call.on('data', function(resp){
        console.log('got value: ', resp.current_val);
        last_hash = resp.hash_sum;;
        hash.update(resp.current_val.toString());
    });
    call.on('end', function(resp) {
        console.log('stream ended');
        console.log("our hash:" + hash.hex());
        console.log("hash from server: " + last_hash);
        if (hash.hex() != last_hash) {
            console.log("housten we might have a problem?");
        } else {
            console.log("hashes match, all good");
        }
    });
    call.on('error', function(){
        if (retry == 5) {
            console.log("5 times failed, giving up");
            process.exit(-1);
        }
        console.log('sleeping');
        sleep.sleep(retry*2);
        stateFullQuery(client, count, clientID, last_hash, hash, retry+1);
    });
}

function stateLessQuery(client, count, last, sum, retry) {
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
        if (retry == 5) {
            console.log("5 times failed, giving up");
            process.exit(-1);
        }
        console.log('error sleeping');
        // FIXME: generalize with the other API
        sleep.sleep(retry*2);
        stateLessQuery(client, count, last, sum, retry);
    });
}
main();
