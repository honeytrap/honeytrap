var http = require('http');
var server = http.createServer(function(request, response) {});

var count = 0;
all_active_connections = {};

server.listen(1234, function() {
    console.log((new Date()) + ' Server is listening on port 1234');
})

var WebSocketServer = require('websocket').server;
wsServer = new WebSocketServer({
    httpServer: server
});

wsServer.on('request', function(websocket) {
    var connection = websocket.accept('echo-protocol', websocket.origin);
    
    var clients = {}; 

    // Specific id for this client & increment count
    var id = count++;

    all_active_connections[id] = connection;
    connection.id = id;

    // Store the connection method so we can loop through & contact all clients
    clients[id] = connection      

    console.log((new Date()) + ' Connection accepted [' + id + ']');     

    connection.on('open', function() {
        for(var i in clients){
            // Send a message to the client with the message
            clients[i].sendUTF('test');
        }        
    });

    // Create event listener
    connection.on('message', function(data) {
        var parsedData = JSON.parse(data.utf8Data);
        console.log(parsedData)
        if(parsedData.type == 'message') {            

            // Loop through all clients
            for(var i in clients){
                // Send a message to the client with the message
                clients[i].send(JSON.stringify(parsedData));
            }        
        }
        if(parsedData.type == 'broadcast') {
            for (i in all_active_connections)
                 all_active_connections[i].send(JSON.stringify(parsedData));
        }

        // if(parsedData.type == 'broadcast') {
        //     for (i in all_active_connections)
        //          all_active_connections[i].sendUTF(parsedData.text);
        // }
    });

    connection.on('close', function(reasonCode, description) {
        delete clients[id];
        console.log((new Date()) + ' Peer ' + connection.remoteAddress + ' disconnected.');
    });
});