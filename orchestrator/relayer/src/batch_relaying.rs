//! This module contains code for the batch update lifecycle. Functioning as a way for this validator to observe
//! the state of both chains and perform the required operations.

use clarity::address::Address as EthAddress;
use clarity::PrivateKey as EthPrivateKey;
use cosmos_peggy::query::get_latest_transaction_batches;
use cosmos_peggy::query::get_transaction_batch_signatures;
use cosmos_peggy::query::{get_valset,get_current_valset};
use ethereum_peggy::submit_batch::send_eth_transaction_batch;
use ethereum_peggy::utils::get_tx_batch_nonce;
use ethereum_peggy::utils::get_valset_nonce;
use peggy_proto::peggy::query_client::QueryClient as PeggyQueryClient;
use peggy_utils::types::{BatchConfirmResponse, TransactionBatch, Valset};
use std::time::Duration;
use tonic::transport::Channel;
use web30::client::Web3;

/// Check the last validator set on Ethereum, if it's lower than our latest validator
/// set then we should package and submit the update as an Ethereum transaction
pub async fn relay_batches(
    ethereum_key: EthPrivateKey,
    web3: &Web3,
    grpc_client: &mut PeggyQueryClient<Channel>,
    peggy_contract_address: EthAddress,
    timeout: Duration,
) {
    let our_ethereum_address = ethereum_key.to_public_key().unwrap();

    let latest_batches = get_latest_transaction_batches(grpc_client).await;
    if latest_batches.is_err() {
        error!("Latest batches {:?}", latest_batches);
        return;
    }
    let latest_batches = latest_batches.unwrap();
    let mut oldest_signed_batch: Option<TransactionBatch> = None;
    let mut oldest_signatures: Option<Vec<BatchConfirmResponse>> = None;
    for batch in latest_batches {
        let sigs =
            get_transaction_batch_signatures(grpc_client, batch.nonce).await;
        trace!("Got sigs {:?}", sigs);
        if let Ok(sigs) = sigs {
            // todo check that enough people have signed
            oldest_signed_batch = Some(batch);
            oldest_signatures = Some(sigs);
        } else {
            error!(
                "could not get signatures for {} with {:?}",
                batch.nonce, sigs
            );
        }
    }
    if oldest_signed_batch.is_none() {
        error!("Could not find batch with signatures! exiting");
        return;
    }
    let oldest_signed_batch = oldest_signed_batch.unwrap();

    info!("Get oldest_signed_batch{}",oldest_signed_batch.nonce); 
    let oldest_signatures = oldest_signatures.unwrap();

    let latest_ethereum_batch = get_tx_batch_nonce(
        peggy_contract_address,
        our_ethereum_address,
        web3,
    )
    .await
    .expect("Failed to get batch nonce from Ethereum");

    let latest_ethereum_valset =
        get_valset_nonce(peggy_contract_address, our_ethereum_address, web3)
            .await
            .expect("Failed to get Ethereum valset");
    let latest_cosmos_batch_nonce = oldest_signed_batch.clone().nonce;
    
    info!("Get latest_ethereum_batch{},latest_cosmos_batch_nonce {} ",latest_ethereum_batch,latest_cosmos_batch_nonce); 
    if latest_cosmos_batch_nonce > latest_ethereum_batch {
        info!(
            "We have detected latest batch {} but latest on Ethereum is {} sending an update!",
            latest_cosmos_batch_nonce, latest_ethereum_batch
        );

        if latest_ethereum_valset == 0 {
            let latest_valsets = get_current_valset(grpc_client).await;
            if latest_valsets.is_err() {
                trace!("Failed to get latest valsets!");
                // there are no latest valsets to check, possible on a bootstrapping chain maybe handle better?
                return;
            }

            let latest_cosmos_valset = latest_valsets.unwrap();
            let mut latest_valset = latest_cosmos_valset.clone();
            latest_valset.nonce = 0;

            let _res = send_eth_transaction_batch(
                latest_valset,
                oldest_signed_batch,
                &oldest_signatures,
                web3,
                timeout,
                peggy_contract_address,
                ethereum_key,
            )
            .await;
            return;
        }

        // get the current valset from the Cosmos chain
        let current_valset = get_valset(grpc_client, latest_ethereum_valset).await;
        if let Ok(Some(current_valset)) = current_valset {
            let _res = send_eth_transaction_batch(
                current_valset,
                oldest_signed_batch,
                &oldest_signatures,
                web3,
                timeout,
                peggy_contract_address,
                ethereum_key,
            )
            .await;
        } else {
            error!("Failed to get latest validator set!");
        }
    }
}
