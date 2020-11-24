import { Peggy } from "../typechain/Peggy";
import { UpgradedPeggy } from "../typechain/UpgradedPeggy";
import { PeggyProxy } from "../typechain/PeggyProxy";
import { TestERC20 } from "../typechain/TestERC20";
import { ethers } from "@nomiclabs/buidler";
import { makeCheckpoint, signHash, getSignerAddresses } from "./pure";
import { BigNumberish } from "ethers/utils";
import { Signer } from "ethers";

type DeployContractsOptions = {
  corruptSig?: boolean;
};

export async function deployContracts(
  peggyId: string = "foo",
  validators: Signer[],
  powers: number[],
  powerThreshold: number,
  opts?: DeployContractsOptions
) {
  const TestERC20 = await ethers.getContractFactory("TestERC20");
  const testERC20 = (await TestERC20.deploy()) as TestERC20;

  const PeggyContract = await ethers.getContractFactory("Peggy");
  const PeggyProxyContract = await ethers.getContractFactory("PeggyProxy");

  const valAddresses = await getSignerAddresses(validators);

  const checkpoint = makeCheckpoint(valAddresses, powers, 0, peggyId);

  const peggyImplementation = (await PeggyContract.deploy()) as Peggy;
  await peggyImplementation.deployed();

  const signers = await ethers.getSigners();
  const signersLength = signers.length;
  const proxyAdmin = await signers[signersLength-1].getAddress();
  const peggyProxy = (await PeggyProxyContract.deploy(peggyImplementation.address, proxyAdmin, [/*empty function call data*/])) as PeggyProxy;
  await peggyProxy.deployed();

  var peggy = peggyImplementation.attach(peggyProxy.address);
  await peggy.initialize(peggyId, powerThreshold, valAddresses, powers);

  return { peggy, testERC20, checkpoint };
}

export async function deployContractsForUpgrade(
    peggyId: string = "foo",
    validators: Signer[],
    powers: number[],
    powerThreshold: number,
    opts?: DeployContractsOptions
) {
  const TestERC20 = await ethers.getContractFactory("TestERC20");
  const testERC20 = (await TestERC20.deploy()) as TestERC20;

  const PeggyContract = await ethers.getContractFactory("Peggy");
  const UpgradedPeggyContract = await ethers.getContractFactory("UpgradedPeggy");
  const PeggyProxyContract = await ethers.getContractFactory("PeggyProxy");

  const valAddresses = await getSignerAddresses(validators);

  const checkpoint = makeCheckpoint(valAddresses, powers, 0, peggyId);

  const peggyImplementation = (await PeggyContract.deploy()) as Peggy;
  await peggyImplementation.deployed();

  const upgradedPeggyImplementation = (await UpgradedPeggyContract.deploy()) as UpgradedPeggy;
  await upgradedPeggyImplementation.deployed();

  const signers = await ethers.getSigners();
  const signersLength = signers.length;
  const proxyAdmin = await signers[signersLength-1].getAddress();
  const peggyProxy = (await PeggyProxyContract.deploy(peggyImplementation.address, proxyAdmin, [/*empty function call data*/])) as PeggyProxy;
  await peggyProxy.deployed();

  var peggy = peggyImplementation.attach(peggyProxy.address);
  await peggy.initialize(peggyId, powerThreshold, valAddresses, powers);

  return { peggy, testERC20, checkpoint, peggyProxy, upgradedPeggyImplementation };
}
