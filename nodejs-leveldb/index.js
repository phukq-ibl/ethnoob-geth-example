var levelup = require('levelup')
var leveldown = require('leveldown')
var RLP = require('rlp')

var down = leveldown('../state/data',{
    keyEncoding: 'binary',
    valueEncoding: 'binary',
})
var db = levelup(down)

db.createReadStream({
    start: '',
    end:'',
    limit:100
})
    .on('data', function (data) {
        // console.log(data.key.toString('hex'), '=>',  RLP.decode(data.value))
        parseData =  data.value.toString('hex')
        if(parseData.length > 45) {
            parseData =  RLP.decode(data.value)
        }
        console.log(data.key.toString('hex'), '=>',  parseData)
    })
    .on('error', function (err) {
        console.log('Oh my!', err)
    })
    .on('close', function () {
        console.log('Stream closed')
    })
    .on('end', function () {
        console.log('Stream ended')
    })

    