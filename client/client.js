var PROTO_PATH = "../streamingdata/streamingdata.proto";

var sleep = require('sleep');
var grpc = require('grpc');
var cmdargs = require('command-line-args');

var streaming_proto = grpc.load(PROTO_PATH).streamingdata;

function main() {
    const optionDefinitions = [
        { name: 'server', alias: 's', type: String, defaultValue: "localhost:8000" },
        { name: 'mode', type: String, defaultValue: 'stateless' },
        { name: 'count', type: Number, defaultValue: 10},
    ]
    const options = cmdargs(optionDefinitions)
    // FIXME: usage string
    if (options['mode'] != 'stateful' && options['mode'] != 'stateless') {
        console.log('choose correct mode!')
        process.exit(-1)
    }
    var client = new streaming_proto.RealTime(options['server'], grpc.credentials.createInsecure());

    stateLessQuery(client, options['count'], 0, 0);
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
        console.log('value is:', resp.current_val)
        last = resp.current_val;
        sum += resp.current_val;
        count--;
    });
    call.on('end', function(){
        console.log("sum is: "+sum);
    });
    call.on('error', function(){
        console.log('error sleeping');
        sleep.sleep(4);
        stateLessQuery(client, count, last, sum);
    });
}

main();
