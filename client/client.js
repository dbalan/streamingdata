var PROTO_PATH = "../streamingdata/streamingdata.proto";

var grpc = require('grpc');

var streaming_proto = grpc.load(PROTO_PATH).streamingdata

function main() {
    var client = new streaming_proto.RealTime('localhost:8000', grpc.credentials.createInsecure());

    client.getPing({}, function(err, response) {
        console.log('Got', response.resp);
    })
}

main();
