import time
import pandas as pd
import matplotlib.pyplot as plt
from web3 import Web3
from datetime import datetime
from Utils import *

def testing():
# Connect to an Ethereum node - check later
    node1 = Web3(Web3.IPCProvider("/data/data_Tomer/node1/geth.ipc"))
    node2 = Web3(Web3.IPCProvider("/data/data_Tomer/node2/geth.ipc"))

    # Testing - not real
    # w3 = Web3(Web3.EthereumTesterProvider())

    # Example: Get the latest block number
    # Check if connected
    if node1.is_connected() and node2.is_connected():
        print("Nodes Connected to Ethereum network")

        # Get the latest block
        node1_latest_block = node1.eth.get_block('latest')
        node2_latest_block = node2.eth.get_block('latest')

        # Print block information for node 1
        print(f"Latest block number: {node1_latest_block['number']}")
        print(f"Block hash: {node1_latest_block['hash'].hex()}")
        print(f"Timestamp: {node1_latest_block['timestamp']}")
        print(f"Gas used: {node1_latest_block['gasUsed']}")
        print(f"Accounts: {len(node1.eth.accounts)}")
        # print(f"Transactions: {node1.geth.txpool.inspect()}")

        # Print block information for node 2
        print(f"Latest block number: {node2_latest_block['number']}")
        print(f"Block hash: {node2_latest_block['hash'].hex()}")
        print(f"Timestamp: {node2_latest_block['timestamp']}")
        print(f"Gas used: {node2_latest_block['gasUsed']}")
        print(f"Accounts: {len(node2.eth.accounts)}")
        # print(f"Transactions: {node2.geth.txpool.inspect()}")

        txpool1_data = node1.provider.make_request('txpool_content', [])
        first_queued_address =  list(txpool1_data['result']['queued'].keys())[0]
        first_queued_transaction = txpool1_data['result']['queued'][first_queued_address]
        print("Queued Transaction: ", [i for i in first_queued_transaction.items()])


        txpool2_data = node2.provider.make_request('txpool_content', [])
        first_queued_address =  list(txpool2_data['result']['queued'].keys())[0]
        first_queued_transaction = txpool2_data['result']['queued'][first_queued_address]
        print("Queued Transaction: ", [i for i in first_queued_transaction.items()])


        print()
        
    else:
        print("Failed to connect to Ethereum network")

def get_txpool_content(node):
    txpool_data = node.provider.make_request('txpool_content', [])
    queued = set()
    pending = set()
    
    for txs in txpool_data['result']['queued'].values():
        for tx in txs.values():
            queued.add(tx['hash'])

    for txs in txpool_data['result']['pending'].values():
        for tx in txs.values():
            pending.add(tx['hash'])
    
    return queued, pending

def integration_test():
    # Connect to Ethereum nodes
    node1 = Web3(Web3.IPCProvider("/data/data_Tomer/node1/geth.ipc"))
    node2 = Web3(Web3.IPCProvider("/data/data_Tomer/node2/geth.ipc"))

    if not (node1.is_connected() and node2.is_connected()):
        print("Failed to connect to Ethereum network")
        return

    print("Nodes Connected to Ethereum network")

    # Prepare data collection
    duration_minutes = 15  # Run for 15 minutes
    interval_seconds = 60  # Collect data every 60 seconds

    # just for check
    # duration_minutes = 1 # Run for 10 minutes
    # interval_seconds = 10  # Collect data every 60 seconds

    data_node1 = []
    data_node2 = []
    data_symmetric_difference = []

    start_time = time.time()
    end_time = start_time + (duration_minutes * 60)

    # Data collection loop
    while time.time() < end_time:
        current_time = (time.time() - start_time) / 60  # Convert to minutes
        
        queued1, pending1 = get_txpool_content(node1)
        queued2, pending2 = get_txpool_content(node2)
        
        total_txs_node1 = len(queued1) + len(pending1)
        total_txs_node2 = len(queued2) + len(pending2)
        
        all_txs_node1 = queued1.union(pending1)
        all_txs_node2 = queued2.union(pending2)
        symmetric_difference_size = len(all_txs_node1.symmetric_difference(all_txs_node2))
        
        data_node1.append([f"{current_time:.2f}", total_txs_node1, len(queued1), len(pending1)])
        print(f"Node 1 data written: Time={current_time:.2f}min, Total={total_txs_node1}, Queued={len(queued1)}, Pending={len(pending1)}")
        
        data_node2.append([f"{current_time:.2f}", total_txs_node2, len(queued2), len(pending2)])
        print(f"Node 2 data written: Time={current_time:.2f}min, Total={total_txs_node2}, Queued={len(queued2)}, Pending={len(pending2)}")
        
        data_symmetric_difference.append([f"{current_time:.2f}", symmetric_difference_size])
        print(f"Symmetric difference data written: Time={current_time:.2f}min, Symmetric Difference={symmetric_difference_size}")

        time.sleep(interval_seconds)

    # Write data to CSV using the provided function

    # Node 1 CSVs
    export_results_to_csv(['Time (minutes)', 'Total Txs'], 
                          [[row[0], row[1]] for row in data_node1], 
                          f"node1_total_txs.csv")
    
    export_results_to_csv(['Time (minutes)', 'Queued', 'Pending'], 
                          [[row[0], row[2], row[3]] for row in data_node1], 
                          f"node1_distribution.csv")
    
    # Node 2 CSVs
    export_results_to_csv(['Time (minutes)', 'Total Txs'], 
                          [[row[0], row[1]] for row in data_node2], 
                          f"node2_total_txs.csv")
    
    export_results_to_csv(['Time (minutes)', 'Queued', 'Pending'], 
                          [[row[0], row[2], row[3]] for row in data_node2], 
                          f"node2_distribution.csv")
    
    # Symmetric Difference CSV
    export_results_to_csv(['Time (minutes)', 'Symmetric Difference'], 
                          data_symmetric_difference, 
                          f"nodes_symmetric_difference.csv")

    # Create graphs
    time_data = [row[0] for row in data_node1]
    total_txs_node1_data = [row[1] for row in data_node1]
    total_txs_node2_data = [row[1] for row in data_node2]
    symmetric_difference_data = [row[1] for row in data_symmetric_difference]
    queued_data_node1 = [row[2] for row in data_node1]
    pending_data_node1 = [row[3] for row in data_node1]
    queued_data_node2 = [row[2] for row in data_node2]
    pending_data_node2 = [row[3] for row in data_node2]

    # 1. Total transactions in each node comparison
    plt.figure(figsize=(10, 6))
    plt.plot(time_data, total_txs_node1_data, label='Node 1')
    plt.plot(time_data, total_txs_node2_data, label='Node 2')
    plt.title('Total Transactions in Each Node')
    plt.xlabel('Time (minutes)')
    plt.ylabel('Number of Transactions')
    plt.legend()
    plt.savefig('total_txs_comparison.png')
    plt.close()

    # 2. Symmetric difference size
    plt.figure(figsize=(10, 6))
    plt.plot(time_data, symmetric_difference_data)
    plt.title('Symmetric Difference Size')
    plt.xlabel('Time (minutes)')
    plt.ylabel('Number of Transactions')
    plt.savefig('symmetric_difference.png')
    plt.close()

    # 3. Distribution of pending vs queued (for both nodes)
    plt.figure(figsize=(10, 6))
    plt.plot(time_data, queued_data_node1, label='Node 1 Queued')
    plt.plot(time_data, pending_data_node1, label='Node 1 Pending')
    plt.plot(time_data, queued_data_node2, label='Node 2 Queued')
    plt.plot(time_data, pending_data_node2, label='Node 2 Pending')
    plt.title('Distribution of Queued vs. Pending Transactions')
    plt.xlabel('Time (minutes)')
    plt.ylabel('Number of Transactions')
    plt.legend()
    plt.savefig('queued_vs_pending.png')
    plt.close()

    print("Graphs have been saved as PNG files.")

def files_testing():

    # Load the data from CSV files
    file1_path = '/home/tomer_local/Rateless-Set-Reconciliation-with-Listing-Guarantees/data/blockchain/node1/node1_txpool_hashes_2.csv'
    file2_path = '/home/tomer_local/Rateless-Set-Reconciliation-with-Listing-Guarantees/data/blockchain/node2/node2_txpool_hashes_2.csv'

    # Read CSV files assuming they contain hash values in a single column
    hashes1 = pd.read_csv(file1_path, header=None).squeeze().astype(str)
    hashes2 = pd.read_csv(file2_path, header=None).squeeze().astype(str)

    # Convert the hash lists to sets for fast symmetric difference computation
    set1 = set(hashes1)
    set2 = set(hashes2)

    # Compute symmetric difference
    symmetric_difference = set1.symmetric_difference(set2)

    # Output the size of the symmetric difference
    print("Symmetric Difference Size: ", len(symmetric_difference))

if __name__ == "__main__":

    # testing()

    # integration_test()

    files_testing()


    