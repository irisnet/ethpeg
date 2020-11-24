pragma solidity ^0.6.0;

import "@openzeppelin/contracts/proxy/TransparentUpgradeableProxy.sol";

contract PeggyProxy is TransparentUpgradeableProxy {

	constructor(address logic, address admin, bytes memory data) TransparentUpgradeableProxy(logic, admin, data) public {

	}

}