console.log('Loading function');

var aws = require('aws-sdk');

aws.config.update({region: 'us-east-1'});

var dynamodb = new aws.DynamoDB({apiVersion: '2012-10-08'});

exports.handler = function(event, context, callback) {
    console.log('Received event:', JSON.stringify(event, null, 2))
    
    if (event.hasOwnProperty('sourceName') == false) {
        console.log("Error 'sourceName' is not present.")
        callback("I expect receive 'sourceName' field.", null)
        return
    }
    
    if (event.hasOwnProperty('chunkRows') == false) {
        console.log("Error 'chunkRows' is not present.")
        callback("I expect receive 'chunkRows' field.", null)
        return
    }

    if (event.hasOwnProperty('chunkCount') == false) {
        console.log("Error 'chunkCount' is not present.")
        callback("I expect receive 'chunkCount' field.", null)
        return
    }

    console.log("sourceName=" + event.sourceName, "chunkCount=" + event.chunkCount, "size=" + event.chunkRows.length)

    event.chunkRows.forEach(function(item) {
        dynamodb.putItem({TableName: 'movies', Item: item}, (err, data) => {
            if(err) {
                console.log(err,  err.stack);
            }
        });
    });
}
