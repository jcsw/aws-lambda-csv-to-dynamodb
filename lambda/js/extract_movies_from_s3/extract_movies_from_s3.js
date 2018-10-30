console.log('Loading function');

var aws = require('aws-sdk');
var es = require('event-stream');

aws.config.update({region: 'us-east-1'});

var s3 = new aws.S3({apiVersion: '2006-03-01'});

var chunkSizeMax = 100;
var chunkSize = 10;
var headerLine = 1;

exports.handler = function(event, context, callback) {
    console.log('Received event:', JSON.stringify(event, null, 2));

    var s3BucketName = event.Records[0].s3.bucket.name;
    var s3ObjectKey = event.Records[0].s3.object.key;

    var lineCount = 0;

    var chunkLine = 0;
    var chunkCount = 0;
    var chunkRows = [];

    s3.getObject({
        Bucket: s3BucketName, Key: s3ObjectKey
    })
    .createReadStream()
    .pipe(es.split(/\r|\r?\n/))
    .pipe(es.mapSync(function (csvLine) {
        lineCount++;
        if (lineCount > headerLine) {

            chunkLine++;
            chunkRows.push(buildDynamoDBItem(csvLine));

            if(chunkLine == chunkSize) {
                chunkCount++;
                sendChunkRows(chunkRows, s3ObjectKey, chunkCount);

                chunkRows = [];
                chunkLine = 0;
                if (chunkSize < chunkSizeMax) {
                    chunkSize = chunkSize + 10;
                }
            }
        }

    })).on('end', function() {

        if(chunkLine > 0) {
            chunkCount++;
            sendChunkRows(chunkRows, s3ObjectKey, chunkCount);
        }

        console.log('lineCount=' + lineCount, 'chunkCount=' + chunkCount);

    }).on('error', function(err) {
        console.log('error=' + err,  err.stack);
    });
};

function buildDynamoDBItem(csvLine) {
    var itens = csvLine.split(",");
    return {
        imdb : {"S": itens[0]},
        year : {"N": itens[1]},
        title: {"S": itens[2]},
        code : {"S": itens[3]}
    }
}

function sendChunkRows(chunkRows, csvFileName, chunkCount) {
    console.log('sourceName=' + csvFileName, 'length=' + chunkRows.length, 'chunkCount=' + chunkCount);

    var payload = {
        sourceName: csvFileName,
        chunkRows: chunkRows,
        chunkCount: chunkCount
    }

    promiseInvoke('import_movies_in_dynamodb', payload);
}

function promiseInvoke(functionName, payload) {
    console.log('functionName=' + functionName, ' payload=' + payload);

    var lambda = new aws.Lambda();
    return lambda.invoke({
        FunctionName: functionName,
        InvocationType: 'Event',
        LogType: 'Tail',
        Payload: JSON.stringify(payload)
    }).promise();
};