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
    console.log("sourceName:", event.sourceName)
    
    if (event.hasOwnProperty('chunkRows') == false) {
        console.log("Error 'chunkRows' is not present.")
        callback("I expect receive 'chunkRows' field.", null)
        return
    }
    console.log("chunkRows size:", event.chunkRows.length)

    var processedRows = 0, errorsRows = 0;

    for (var i = 0, len = event.chunkRows.length; i < len; i++) {
        processedRows++;

        var item = event.chunkRows[i]
        dynamodb.putItem({TableName: 'movies', Item: item}, (err, res) => {
            if(err) {
                errorsRows++;
                console.log('item:', item);
                console.log('error:', err);
            } 
        });
    }

    console.log('processedRows:', processedRows);
    console.log('errorsRows:', errorsRows);
}