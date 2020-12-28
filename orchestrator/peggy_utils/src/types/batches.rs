use super::*;
use crate::error::PeggyError;
use clarity::{abi::Token, Address as EthAddress};
use deep_space::address::Address as CosmosAddress;

/// This represents an individual transaction being bridged over to Ethereum
/// parallel is the OutgoingTransferTx in x/peggy/types/batch.go
#[derive(Serialize, Deserialize, Debug, Default, Clone)]
pub struct BatchTransaction {
    pub id: u64,
    pub sender: CosmosAddress,
    pub destination: EthAddress,
    pub erc20_token: ERC20Token,
    pub erc20_fee: ERC20Token,
}

impl BatchTransaction {
    pub fn from_proto(input: peggy_proto::peggy::OutgoingTransferTx) -> Result<Self, PeggyError> {
        if input.erc20_fee.is_none() || input.erc20_token.is_none() {
            return Err(PeggyError::InvalidBridgeStateError(
                "Can not have tx with null erc20_token!".to_string(),
            ));
        }
        Ok(BatchTransaction {
            id: input.id,
            sender: input.sender.parse()?,
            destination: input.dest_address.parse()?,
            erc20_token: ERC20Token::from_proto(input.erc20_token.unwrap())?,
            erc20_fee: ERC20Token::from_proto(input.erc20_fee.unwrap())?,
        })
    }
}

/// the response we get when querying for a valset confirmation
#[derive(Serialize, Deserialize, Debug, Default, Clone)]
pub struct TransactionBatch {
    pub nonce: u64,
    pub transactions: Vec<BatchTransaction>,
    pub total_fee: ERC20Token,
    pub token_contract: EthAddress,
}

impl TransactionBatch {
    /// extracts the amounts, destinations and fees as submitted to the Ethereum contract
    /// and used for signatures
    pub fn get_checkpoint_values(&self) -> (Token, Token, Token) {
        let mut amounts = Vec::new();
        let mut destinations = Vec::new();
        let mut token_contracts = Vec::new();
        for item in self.transactions.iter() {
            amounts.push(Token::Uint(item.erc20_token.amount.clone()));
            token_contracts.push(item.erc20_token.token_contract_address.clone());
            destinations.push(item.destination)
        }
        assert_eq!(amounts.len(), destinations.len());
        assert_eq!(token_contracts.len(), destinations.len());
        (
            Token::Dynamic(amounts),
            destinations.into(),
            token_contracts.into(),
        )
    }

    pub fn from_proto(input: peggy_proto::peggy::OutgoingTxBatch) -> Result<Self, PeggyError> {
        let mut transactions = Vec::new();
        let mut running_total_fee: Option<ERC20Token> = None;
        for tx in input.transactions {
            let tx = BatchTransaction::from_proto(tx)?;
            if let Some(total_fee) = running_total_fee {
                running_total_fee = Some(ERC20Token {
                    token_contract_address: total_fee.token_contract_address,
                    amount: total_fee.amount + tx.erc20_fee.amount.clone(),
                });
            } else {
                running_total_fee = Some(tx.erc20_fee.clone())
            }
            transactions.push(tx);
        }
        if let Some(total_fee) = running_total_fee {
            Ok(TransactionBatch {
                nonce: input.batch_nonce,
                transactions,
                token_contract: total_fee.token_contract_address,
                total_fee,
            })
        } else {
            Err(PeggyError::InvalidBridgeStateError(
                "Transaction batch containing no transactions!".to_string(),
            ))
        }
    }
}
