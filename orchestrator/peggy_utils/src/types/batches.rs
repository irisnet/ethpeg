use super::*;
use crate::error::PeggyError;
use clarity::{abi::Token, Address as EthAddress};
use deep_space::{address::Address as CosmosAddress, coin::Coin};

/// This represents an individual transaction being bridged over to Ethereum
/// parallel is the OutgoingTransferTx in x/peggy/types/batch.go
#[derive(Serialize, Deserialize, Debug, Default, Clone)]
pub struct BatchTransaction {
    pub id: u64,
    pub sender: CosmosAddress,
    pub destination: EthAddress,
    pub erc20_token: ERC20Token,
    pub fee: Coin,
}

impl BatchTransaction {
    pub fn from_proto(input: peggy_proto::peggy::OutgoingTransferTx) -> Result<Self, PeggyError> {
        if input.fee.is_none() || input.erc20_token.is_none() {
            return Err(PeggyError::InvalidBridgeStateError(
                "Can not have tx with null erc20_token!".to_string(),
            ));
        }

        let fee = input.fee.unwrap();
        Ok(BatchTransaction {
            id: input.id,
            sender: input.sender.parse()?,
            destination: input.dest_address.parse()?,
            erc20_token: ERC20Token::from_proto(input.erc20_token.unwrap())?,
            fee: Coin {
                denom: fee.denom,
                amount: fee.amount.parse()?,
            },
        })
    }
}

/// the response we get when querying for a valset confirmation
#[derive(Serialize, Deserialize, Debug, Default, Clone)]
pub struct TransactionBatch {
    pub nonce: u64,
    pub transactions: Vec<BatchTransaction>,
    pub token_contracts: Vec<EthAddress>,
}

impl TransactionBatch {
    /// extracts the amounts, destinations and fees as submitted to the Ethereum contract
    /// and used for signatures
    pub fn get_checkpoint_values(&self) -> (Token, Token, Token) {
        let mut amounts = Vec::new();
        let mut destinations = Vec::new();
        for item in self.transactions.iter() {
            amounts.push(Token::Uint(item.erc20_token.amount.clone()));
            destinations.push(item.destination)
        }
        assert_eq!(amounts.len(), destinations.len());
        assert_eq!(self.token_contracts.len(), destinations.len());
        (
            Token::Dynamic(amounts),
            destinations.into(),
            self.token_contracts.clone().into(),
        )
    }

    pub fn from_proto(input: peggy_proto::peggy::OutgoingTxBatch) -> Result<Self, PeggyError> {
        let mut transactions = Vec::new();
        let mut token_contracts = Vec::new();
        for tx in input.transactions {
            let btx = BatchTransaction::from_proto(tx)?;
            token_contracts.push(btx.erc20_token.token_contract_address);
            transactions.push(btx);
        }
        Ok(TransactionBatch {
            nonce: input.batch_nonce,
            transactions,
            token_contracts,
        })
    }
}
