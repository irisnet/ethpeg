use clarity::abi::Token;
use clarity::Uint256;
use clarity::{abi::encode_tokens,abi::encode_call, Address as EthAddress};
use deep_space::address::Address as CosmosAddress;
use peggy_utils::error::PeggyError;
use peggy_utils::types::*;
use sha3::{Digest, Keccak256};
use std::u64::MAX as U64MAX;
use web30::{client::Web3, jsonrpc::error::Web3Error, types::Data, types::TransactionRequest,types::UnpaddedHex};

pub fn get_correct_sig_for_address(
    address: CosmosAddress,
    confirms: &[ValsetConfirmResponse],
) -> (Uint256, Uint256, Uint256) {
    for sig in confirms {
        if sig.validator == address {
            return (
                sig.eth_signature.v.clone(),
                sig.eth_signature.r.clone(),
                sig.eth_signature.s.clone(),
            );
        }
    }
    panic!("Could not find that address!");
}

pub fn get_checkpoint_abi_encode(valset: &Valset, peggy_id: &str) -> Result<Vec<u8>, PeggyError> {
    let (eth_addresses, powers) = valset.filter_empty_addresses();
    Ok(encode_tokens(&[
        Token::FixedString(peggy_id.to_string()),
        Token::FixedString("checkpoint".to_string()),
        valset.nonce.into(),
        eth_addresses.into(),
        powers.into(),
    ]))
}

pub fn get_checkpoint_hash(valset: &Valset, peggy_id: &str) -> Result<Vec<u8>, PeggyError> {
    let locally_computed_abi_encode = get_checkpoint_abi_encode(&valset, &peggy_id);
    let locally_computed_digest = Keccak256::digest(&locally_computed_abi_encode?);
    Ok(locally_computed_digest.to_vec())
}

pub fn downcast_nonce(input: Uint256) -> Option<u64> {
    if input >= U64MAX.into() {
        None
    } else {
        let mut val = input.to_bytes_be();
        // pad to 8 bytes
        while val.len() < 8 {
            val.insert(0, 0);
        }
        let mut lower_bytes: [u8; 8] = [0; 8];
        // get the 'lowest' 8 bytes from a 256 bit integer
        lower_bytes.copy_from_slice(&val[0..val.len()]);
        Some(u64::from_be_bytes(lower_bytes))
    }
}

#[test]
fn test_downcast_nonce() {
    let mut i = 0u64;
    while i < 100_000 {
        assert_eq!(i, downcast_nonce(i.into()).unwrap());
        i += 1
    }
    let mut i: u64 = std::u32::MAX.into();
    i -= 100;
    let end = i + 100_000;
    while i < end {
        assert_eq!(i, downcast_nonce(i.into()).unwrap());
        i += 1
    }
}

/// Gets the latest validator set nonce
pub async fn get_valset_nonce(
    contract_address: EthAddress,
    caller_address: EthAddress,
    web3: &Web3,
) -> Result<u64, Web3Error> {
    let val = contract_call(
            web3,
            contract_address,
            "state_lastValsetNonce()",
            &[],
            caller_address,
        )
        .await?;
    // the go represents all nonces as u64, there's no
    // reason they should ever overflow without a user
    // submitting millions or tens of millions of dollars
    // worth of transactions. But we properly check and
    // handle that case here.
    let real_num = Uint256::from_bytes_be(&val);
    Ok(downcast_nonce(real_num).expect("Valset nonce overflow! Bridge Halt!"))
}

/// Gets the latest transaction batch nonce
pub async fn get_tx_batch_nonce(
    peggy_contract_address: EthAddress,
    erc20_contract_address: EthAddress,
    caller_address: EthAddress,
    web3: &Web3,
) -> Result<u64, Web3Error> {
    let val = contract_call(web3,
            peggy_contract_address,
            "lastBatchNonce(address)",
            &[erc20_contract_address.into()],
            caller_address,
        )
        .await?;
    // the go represents all nonces as u64, there's no
    // reason they should ever overflow without a user
    // submitting millions or tens of millions of dollars
    // worth of transactions. But we properly check and
    // handle that case here.
    let real_num = Uint256::from_bytes_be(&val);
    Ok(downcast_nonce(real_num).expect("TxBatch nonce overflow! Bridge Halt!"))
}

/// Gets the peggyID
pub async fn get_peggy_id(
    contract_address: EthAddress,
    caller_address: EthAddress,
    web3: &Web3,
) -> Result<Vec<u8>, Web3Error> {
    let val = contract_call(web3,contract_address, "state_peggyId()", &[], caller_address)
        .await?;
    Ok(val)
}

/// Gets the ERC20 symbol, should maybe be upstreamed
pub async fn get_erc20_symbol(
    contract_address: EthAddress,
    caller_address: EthAddress,
    web3: &Web3,
) -> Result<String, PeggyError> {
    let val_symbol = contract_call(web3,contract_address, "symbol()", &[], caller_address)
        .await?;
    // Pardon the unwrap, but this is temporary code, intended only for the tests, to help them
    // deal with a deprecated feature (the symbol), which will be removed soon
    Ok(String::from_utf8(val_symbol).unwrap())
}

pub async fn contract_call(
    web3: &Web3,
    contract_address: EthAddress,
    sig: &str,
    tokens: &[Token],
    own_address: EthAddress,
) -> Result<Vec<u8>, Web3Error> {
    //let our_balance = web3.eth_get_balance(own_address).await?;
    let nonce = web3.eth_get_transaction_count(own_address).await?;

    let payload = encode_call(sig, tokens)?;

    let gas_price: Uint256 = 1u8.into();
    // Geth represents gas as a u64 it will truncate leading zeros but not take
    // a value larger than u64::MAX, likewise the command will fail if we can't
    // actually pay that fee. This operation maximizes the info we can get
    let gas_limit = 2000000;
    let transaction = TransactionRequest {
        from: Some(own_address),
        to: contract_address,
        nonce: Some(UnpaddedHex(nonce)),
        gas: Some(gas_limit.into()),
        gas_price: Some(UnpaddedHex(gas_price)),
        value: Some(UnpaddedHex(0u64.into())),
        data: Some(Data(payload)),
    };

    let bytes = match web3.eth_call(transaction).await {
        Ok(val) => val,
        Err(e) => return Err(e),
    };
    Ok(bytes.0)
}
