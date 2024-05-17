import csv
import os
import random
from typing import List, Set, Tuple
from IBLT import IBLT


def benchmark_set_reconciliation(sender_set: Set[int], 
                                 primes: List[int], num_trials: int, 
                                 method: str = "EGH",
                                 export_to_csv: bool = False, 
                                 csv_filename: str = "results.csv",
                                 set_inside_set: bool = True):
    total_cells_transmitted = 0
    results = []

    for trial in range(1, num_trials+1):
        universe_set = set(range(1, 25*trial + 1))

        if set_inside_set:
            receiver_set = universe_set
        else:
            # receiver_set = set([7, 16, 21])
            receiver_set = set(range(1, 3 * trial + 1))
            print("receiver_set: ", receiver_set)
        
        universe_size = len(universe_set)
        sender = IBLT(sender_set, universe_size, primes, "EGH")
        receiver = IBLT(receiver_set, universe_size, primes, "EGH")

        sender.encode()
        receiver.encode()

        symmetric_difference = []

        while True:
            sender_cells = []

            sender.transmit()

            while not sender.cells_queue.empty():
                cell = sender.cells_queue.get()

                # End of IBLT cells transmitting.
                if cell == "end":
                    break

                sender_cells.append(cell)

            symmetric_difference = receiver.receive(sender_cells, set_inside_set)

            if symmetric_difference:
                sender.ack_queue.put("stop")
                break

        total_cells_transmitted = len(receiver.receiver_cells)
        results.append((trial, universe_size, total_cells_transmitted))

        print(f"Symmetric difference: {symmetric_difference}")
        print(f"Number of cells transmitted: {total_cells_transmitted:.2f}")

    if export_to_csv:
        export_results_to_csv(results, csv_filename)

def export_results_to_csv(results: List[Tuple[int, float, int]], csv_filename: str) -> None:
    with open(os.path.join("./data", csv_filename), mode='w', newline='') as file:
        writer = csv.writer(file)
        writer.writerow(["Trial", "Universe Size", "Cells Transmitted"])
        writer.writerows(results)

if __name__ == "__main__":
    # Example usage
    #sender_set = set([1, 5, 8, 11, 15, 21, 23])
    sender_set = set([1, 8, 24])
    # primes in 1 - 150.
    primes = [2, 3, 5, 7, 11, 13, 17, 19, 23, 31, 37, 41, 43, 47, 53, 59, 
    61, 67, 71, 73, 79, 83, 89, 97, 101, 103, 107, 109, 113, 127, 131, 137, 
    139, 149]
    num_trials = 8 

    print("IBLT + EGH:")
    # benchmark_set_reconciliation(sender_set, primes, num_trials, 
    #                              method="EGH", export_to_csv=True, 
    #                              csv_filename="egh_results_receiver_includes_sender.csv", 
    #                              set_inside_set = True)

    benchmark_set_reconciliation(sender_set, primes, num_trials, 
                                 method="EGH", export_to_csv=True, 
                                 csv_filename="egh_results_receiver_not_includes_sender.csv", 
                                 set_inside_set = False)