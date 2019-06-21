var RLP = require('rlp')
var data = 'f839808080a0ac4042072210c81e1513a8c674af59f2911c229355a2749c1c6bb1de57a384068080c8823162844242424280808080808080808080'
var b = Buffer.from(data,'hex')

var rs = RLP.decode(b)
console.log(rs);