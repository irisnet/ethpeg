import chai from "chai";
import { ethers } from "@nomiclabs/buidler";
import { solidity } from "ethereum-waffle";
import fs from "fs";
import { Peggy } from "../typechain/Peggy";
import { TestERC20 } from "../typechain/TestERC20";
import * as bech32 from "bech32";
import { deployContracts } from "../test-utils";
import {
  getSignerAddresses,
  makeCheckpoint,
  signHash,
  makeTxBatchHash,
  examplePowers
} from "../test-utils/pure";

chai.use(solidity);
const { expect } = chai;


async function runTest(opts: {}) {


  // Prep and deploy contract
  // ========================
  const signers = await ethers.getSigners();
  const peggyId = ethers.utils.formatBytes32String("foo");
  // This is the power distribution on the Cosmos hub as of 7/14/2020
  let powers = examplePowers();
  let validators = signers.slice(0, powers.length);
  const powerThreshold = 6666;
  const {
    peggy,
    testERC20,
    checkpoint: deployCheckpoint
  } = await deployContracts(peggyId, validators, powers, powerThreshold);


  // Transfer out to Cosmos, locking coins
  // =====================================
  await testERC20.functions.approve(peggy.address, 1000);
  await expect(peggy.functions.sendToCosmos(
    testERC20.address,
    ethers.utils.formatBytes32String("myCosmosAddress"),
    1000
  )).to.emit(peggy, 'SendToCosmosEvent').withArgs(
    testERC20.address,
    await signers[0].getAddress(),
    ethers.utils.formatBytes32String("myCosmosAddress"),
    1000,
    1
  );

  expect(await testERC20.functions.balanceOf(peggy.address)).to.equal(1000);
  expect(await peggy.functions.state_lastEventNonce()).to.equal(1);



  // Do it again
  // =====================================
  await testERC20.functions.approve(peggy.address, 1000);
  await expect(peggy.functions.sendToCosmos(
    testERC20.address,
    ethers.utils.formatBytes32String("myCosmosAddress"),
    1000
  )).to.emit(peggy, 'SendToCosmosEvent').withArgs(
    testERC20.address,
    await signers[0].getAddress(),
    ethers.utils.formatBytes32String("myCosmosAddress"),
    1000,
    2
  );

  expect(await testERC20.functions.balanceOf(peggy.address)).to.equal(2000);
  expect(await peggy.functions.state_lastEventNonce()).to.equal(2);
}

describe("sendToCosmos tests", function () {
  it("works right", async function () {
    await runTest({})
  });
});

describe("transfer coin to cosmos tests", function () {
  const ethNode = "http://localhost:8545";
  const ethPrivkey = "0xc5e8f61d1ab959b397eecc0a37a6517b8e67a0e7cf1f4bce5591f3ed80199122";
  const peggyContractABI = "artifacts/Peggy.json"
  const erc20ContractABI = "artifacts/ERC20.json"
  const peggyContractAddr = "0x8858eeB3DfffA017D4BCE9801D340D36Cf895CCf"
  const erc20ContractAddr = "0x7c2C195CD6D34B8F845992d380aADB2730bB9C6F"
  const cosmosAddr = cosmosAddrToBytes32("cosmos1pzs4v88qj6u7ar3rh0g8jwtf3ngz9jjvud9jre")

  const provider = new ethers.providers.JsonRpcProvider(ethNode);
  let wallet = new ethers.Wallet(ethPrivkey, provider);

  const { abi, bytecode } = getContractArtifacts(erc20ContractABI);
  const factory = new ethers.ContractFactory(abi, bytecode, wallet);
  let erc20 = factory.attach(erc20ContractAddr) as TestERC20

  it("query erc20 balance", async function () {
    const balance = await erc20.functions.balanceOf(wallet.address)
    console.log("balance", balance.toString())
  });

  //授权peggy合约可以执行erc20代币转入功能
  it("erc20 approve peggy", async function () {
    let res = await erc20.functions.approve(peggyContractAddr, 1000);
    console.log("approve", res)
  });

  it("send erc20 token to cosmos", async function () {
    const { abi, bytecode } = getContractArtifacts(peggyContractABI);
    const factory = new ethers.ContractFactory(abi, bytecode, wallet);

    let peggy = factory.attach(peggyContractAddr) as Peggy

    let amount = "100"
    let resp = await peggy.sendToCosmos(erc20ContractAddr,
      cosmosAddr,
      amount,
      { gasLimit: 1000000 })
    console.log("sendToCosmos", resp)
  })

  it("test cosmos addr convert to byte32", function () {
    let cosmosAddr = "cosmos1pzs4v88qj6u7ar3rh0g8jwtf3ngz9jjvud9jre"
    let ethAddr = cosmosAddrToBytes32(cosmosAddr)
    console.log("addr:", ethAddr)
  })
});

function getContractArtifacts(path: string): { bytecode: string; abi: string } {
  var { bytecode, abi } = JSON.parse(fs.readFileSync(path, "utf8").toString());
  return { bytecode, abi };
}

function cosmosAddrToBytes32(str: string) {
  let ownKey = bech32.decode(str, 1023)
  let cosmosAddr = bech32.fromWords(ownKey.words)
  let err = "too long:" + cosmosAddr.length + "prefix:" + ownKey.prefix;
  if (cosmosAddr.length > 31) { throw new Error(err); }
  let padCosmosAddr = ethers.utils.padZeros(cosmosAddr, 32);
  return ethers.utils.hexlify(padCosmosAddr);
}