// ExpressJS Setup
const express = require('express');
const app = express();
var bodyParser = require('body-parser');

// ExpressJS Setup
const { FileSystemWallet, Gateway } = require('fabric-network');
const fs = require('fs');
const path = require('path');
const ccpPath = path.resolve(__dirname,'connection.json');
const ccpJSON = fs.readFileSync(ccpPath, 'utf8');
const ccp = JSON.parse(ccpJSON);

// Constants
const PORT = 3000;
const HOST = '0.0.0.0';

let today = new Date();   

let year = today.getFullYear(); // 년도
let month = today.getMonth() + 1;  // 월
let date = today.getDate();  // 날짜

const present = (year + '-' + month + '-' + date)

// use static file
app.use(express.static(path.join(__dirname, 'views')));

// configure app to use body-parser
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({ extended: false }));

// main page routing
app.get('/', (req, res)=>{
    res.sendFile(__dirname + '/index-shareshares.html');
});

//거래 정보 입력 라우팅

app.post('/putTrading', async(req, res)=>{
    
    const ndate = present;
    const nsharecode = req.body.nsharecode;
    const nprice = req.body.nprice;
    const namount = req.body.namount;

    try {
        console.log(`Trading post routing - ${nsharecode}`);

        cc_call('putTrading', [ndate,nsharecode,namount,nprice], res)

    }
    catch (error) {
        console.error(`Failed to submit transaction: ${error}`);
    }

});

//거래 정보 출력 라우팅

app.get('/getTrading', async(req, res)=>{
    const trading = "trading"
    try {
        console.log(`Trading routing`);

        cc_call('getHoldShare', [trading], res)

    }
    catch (error) {
        console.error(`Failed to evaluate transaction: ${error}`);
    }
});

//포트폴리오 출력 라우팅

app.get('/getPortfolio', async(req, res)=>{
    const portfolio = "portfolio"
    try {
        console.log(`portflio routing`);

        cc_call('getHoldShare', [portfolio], res)

    }
    catch (error) {
        console.error(`Failed to evaluate transaction: ${error}`);
    }
});

//종목 출력 라우팅

app.get('/getHoldShare', async(req, res)=>{
    const sharecode = req.query.psharecode;
    try {

        console.log(`holdshare routing - ${sharecode}`);

        cc_call('getHoldShare', [sharecode], res)

    }
    catch (error) {
        console.error(`Failed to evaluate transaction: ${error}`);
    }
});

//이력 출력 라우팅

app.get('/getHistoryForShare', async(req, res)=>{
    const sharecode = req.query.hsharecode;
    try {

        console.log(`History routing - ${sharecode}`);

        cc_call('getHistoryForShare', [sharecode], res)

    }
    catch (error) {
        console.error(`Failed to evaluate transaction: ${error}`);
    }
});

async function cc_call(fn_name, args, res){


    const walletPath = path.join(process.cwd(), 'wallet');
    const wallet = new FileSystemWallet(walletPath);
    console.log(`Wallet path: ${walletPath}`);

    const userExists = await wallet.exists('user1');
    if (!userExists) {
        console.log(`cc_call`);
        console.log('An identity for the user "user1" does not exist in the wallet');
        console.log('Run the registerUser.js application before retrying');
        return;
    }
    const gateway = new Gateway();
    await gateway.connect(ccpPath, { wallet, identity: 'user1', discovery: { enabled: true, asLocalhost: true } });
    const network = await gateway.getNetwork('mychannel');
    const contract = network.getContract('shareshares');

    var result;

    if(fn_name == 'putTrading'){
        result = await contract.submitTransaction('putTrading', args[0], args[1], args[2], args[3]);
        const myobj = {result: "success"}
        res.status(200).json(myobj)
    }else if(fn_name == 'getHoldShare'){
        result = await contract.evaluateTransaction('getHoldShare', args[0]);
        const myobj = JSON.parse(result)
        res.status(200).json(myobj)
    }else if(fn_name == 'getHistoryForShare'){
        result = await contract.evaluateTransaction('getHistoryForShare', args[0]);
        const myobj = JSON.parse(result)
        res.status(200).json(myobj)
    }else{
        result = 'not supported function'
    }

    gateway.disconnect();

    return ;
}

// server start
app.listen(PORT, HOST);
console.log(`Running on http://${HOST}:${PORT}`);