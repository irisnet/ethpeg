import chai from "chai";
import { ethers } from "@nomiclabs/buidler";
import { solidity } from "ethereum-waffle";

import {deployContracts, deployContractsForUpgrade} from "../test-utils";
import {
  getSignerAddresses,
  makeCheckpoint,
  signHash,
  makeTxBatchHash,
  examplePowers
} from "../test-utils/pure";

chai.use(solidity);
const { expect } = chai;

describe("proxy upgrade tests", function() {
  it("change proxy admin", async function() {
    const signers = await ethers.getSigners();
    const peggyId = ethers.utils.formatBytes32String("foo");

    // This is the power distribution on the Cosmos hub as of 7/14/2020
    let powers = examplePowers();
    let validators = signers.slice(0, powers.length);
    const powerThreshold = 6666;

    let {
      peggy,
      testERC20,
      checkpoint: deployCheckpoint,
      peggyProxy,
      upgradedPeggyImplementation
    } = await deployContractsForUpgrade(peggyId, validators, powers, powerThreshold);

    const signersLength = signers.length;
    const newProxyAdmin = await signers[signersLength-2].getAddress();
    const proxyAdmin = await signers[signersLength-1].getAddress();

    peggyProxy = peggyProxy.connect(signers[signersLength-1])
    await expect(peggyProxy.changeAdmin(newProxyAdmin))
        .to.emit(peggyProxy,"AdminChanged")
        .withArgs(proxyAdmin, newProxyAdmin)

    peggyProxy = peggyProxy.connect(signers[signersLength-2])
    await expect(peggyProxy.changeAdmin(proxyAdmin))
        .to.emit(peggyProxy,"AdminChanged")
        .withArgs(newProxyAdmin, proxyAdmin)
  });

  it("change peggy implementation", async function() {
    const signers = await ethers.getSigners();
    const peggyId = ethers.utils.formatBytes32String("foo");

    // This is the power distribution on the Cosmos hub as of 7/14/2020
    let powers = examplePowers();
    let validators = signers.slice(0, powers.length);
    const powerThreshold = 6666;

    let {
      peggy,
      testERC20,
      checkpoint: deployCheckpoint,
      peggyProxy,
      upgradedPeggyImplementation
    } = await deployContractsForUpgrade(peggyId, validators, powers, powerThreshold);

    let upgradePeggyContract = upgradedPeggyImplementation.attach(peggy.address);
    await expect(upgradePeggyContract.testUpgrade())
        .to.be.revertedWith("function selector was not recognized and there's no fallback function");
    await expect(upgradePeggyContract.getOwnAddress())
        .to.be.revertedWith("function selector was not recognized and there's no fallback function");

    const signersLength = signers.length;
    const proxyAdmin = await signers[signersLength-1];

    peggyProxy = peggyProxy.connect(proxyAdmin);
    await expect(peggyProxy.upgradeTo(upgradedPeggyImplementation.address))
        .to.emit(peggyProxy, "Upgraded").withArgs(upgradedPeggyImplementation.address);

    upgradePeggyContract = upgradedPeggyImplementation.attach(peggy.address);

    expect(await upgradePeggyContract.testUpgrade()).to.eq("peggy upgrade is successful");
    expect(await upgradePeggyContract.getOwnAddress()).to.eq(peggyProxy.address);
  });
});
