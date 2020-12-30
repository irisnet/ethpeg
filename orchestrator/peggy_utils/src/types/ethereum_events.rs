use clarity::{Address as EthAddress, abi::derive_signature};

use deep_space::address::Address as CosmosAddress;
use num256::Uint256;
use web30::types::Log;

use crate::error::PeggyError;

/// A parsed struct representing the Ethereum event fired by the Peggy contract
/// when the validator set is updated.
#[derive(Serialize, Deserialize, Debug, Default, Clone, Eq, PartialEq, Hash)]
pub struct ValsetUpdatedEvent {
    pub nonce: Uint256,
    // we currently don't parse members, but the data is there
    //members: Vec<ValsetMember>,
}

impl ValsetUpdatedEvent {
    pub fn from_log(input: &Log) -> Result<ValsetUpdatedEvent, PeggyError> {
        // we have one indexed event so we should fine two indexes, one the event itself
        // and one the indexed nonce
        if let Some(nonce_data) = input.topics.get(1) {
            let nonce = Uint256::from_bytes_be(nonce_data);
            Ok(ValsetUpdatedEvent { nonce })
        } else {
            Err(PeggyError::InvalidEventLogError(
                "Too few topics".to_string(),
            ))
        }
    }
    pub fn from_logs(input: &[Log]) -> Result<Vec<ValsetUpdatedEvent>, PeggyError> {
        let mut res = Vec::new();
        for item in input {
            res.push(ValsetUpdatedEvent::from_log(item)?);
        }
        Ok(res)
    }
}

/// A parsed struct representing the Ethereum event fired by the Peggy contract when
/// a transaction batch is executed.
#[derive(Serialize, Deserialize, Debug, Default, Clone, Eq, PartialEq, Hash)]
pub struct TransactionBatchExecutedEvent {
    /// the nonce attached to the transaction batch that follows
    /// it throughout it's lifecycle
    pub batch_nonce: Uint256,
    /// the event nonce representing a unique ordering of events coming out
    /// of the Peggy solidity contract. Ensuring that these events can only be played
    /// back in order
    pub event_nonce: Uint256,
}

impl TransactionBatchExecutedEvent {
    pub fn from_log(input: &Log) -> Result<TransactionBatchExecutedEvent, PeggyError> { 
        let topics = (
            input.topics.get(0),
            input.topics.get(1),
        );       
        if let (Some(sig),Some(batch_nonce_data)) = topics
        {
            let event_sig = derive_signature("TransactionBatchExecutedEvent(uint256,uint256)")?; 
            if event_sig != sig.as_slice() {
                return Err(PeggyError::InvalidEventLogError(
                    "Too few topics".to_string(),
                ))
            };
            let batch_nonce = Uint256::from_bytes_be(batch_nonce_data);
            let event_nonce = Uint256::from_bytes_be(&input.data);
            Ok(TransactionBatchExecutedEvent {
                batch_nonce,
                event_nonce,
            })
        } else {
            Err(PeggyError::InvalidEventLogError(
                "Too few topics".to_string(),
            ))
        }
    }
    pub fn from_logs(input: &[Log]) -> Result<Vec<TransactionBatchExecutedEvent>, PeggyError> {
        let mut res = Vec::new();
        for item in input {
            res.push(TransactionBatchExecutedEvent::from_log(item)?);
        }
        Ok(res)
    }
    /// returns all values in the array with event nonces greater
    /// than the provided value
    pub fn filter_by_event_nonce(event_nonce: u64, input: &[Self]) -> Vec<Self> {
        let mut ret = Vec::new();
        for item in input {
            if item.event_nonce > event_nonce.into() {
                ret.push(item.clone())
            }
        }
        ret
    }
}

/// A parsed struct representing the Ethereum event fired when someone makes a deposit
/// on the Peggy contract
#[derive(Serialize, Deserialize, Debug, Default, Clone, Eq, PartialEq, Hash)]
pub struct SendToCosmosEvent {
    /// The token contract address for the deposit
    pub erc20: EthAddress,
    /// The Ethereum Sender
    pub sender: EthAddress,
    /// The Cosmos destination
    pub destination: CosmosAddress,
    /// The amount of the erc20 token that is being sent
    pub amount: Uint256,
    /// The transaction's nonce, used to make sure there can be no accidntal duplication
    pub event_nonce: Uint256,
}

impl SendToCosmosEvent {
    pub fn from_log(input: &Log) -> Result<SendToCosmosEvent, PeggyError> {
        let topics = (
            input.topics.get(0),
            input.topics.get(1),
            input.topics.get(2),
            input.topics.get(3),
        );
        if let (Some(sig),Some(erc20_data), Some(sender_data), Some(destination_data)) = topics {
            let event_sig = derive_signature("SendToCosmosEvent(address,address,bytes32,uint256,uint256)")?; 
            if event_sig != sig.as_slice() {
                return Err(PeggyError::InvalidEventLogError(
                    "Too few topics".to_string(),
                ))
            };

            let erc20 = EthAddress::from_slice(&erc20_data[12..32])?;
            let sender = EthAddress::from_slice(&sender_data[12..32])?;
            // this is required because deep_space requires a fixed length slice to
            // create an address from bytes.
            let mut c_address_bytes: [u8; 20] = [0; 20];
            c_address_bytes.copy_from_slice(&destination_data[12..32]);
            let destination = CosmosAddress::from_bytes(c_address_bytes);
            let amount = Uint256::from_bytes_be(&input.data[..32]);
            let event_nonce = Uint256::from_bytes_be(&input.data[32..]);
            Ok(SendToCosmosEvent {
                erc20,
                sender,
                destination,
                amount,
                event_nonce,
            })
        } else {
            Err(PeggyError::InvalidEventLogError(
                "Too few topics".to_string(),
            ))
        }
    }
    pub fn from_logs(input: &[Log]) -> Result<Vec<SendToCosmosEvent>, PeggyError> {
        let mut res = Vec::new();
        for item in input {
            res.push(SendToCosmosEvent::from_log(item)?);
        }
        Ok(res)
    }
    /// returns all values in the array with event nonces greater
    /// than the provided value
    pub fn filter_by_event_nonce(event_nonce: u64, input: &[Self]) -> Vec<Self> {
        let mut ret = Vec::new();
        for item in input {
            if item.event_nonce > event_nonce.into() {
                ret.push(item.clone())
            }
        }
        ret
    }
}
