import csv
import os
from typing import List, Set, Tuple
from IBLTWithEGH import IBLTWithEGH


def benchmark_set_reconciliation_egh(sender_set: Set[int], 
                                 primes: List[int], num_trials: int, 
                                 export_to_csv: bool = False, 
                                 csv_filename: str = "results.csv",
                                 set_inside_set: bool = True):
    total_cells_transmitted = 0
    results = []

    for trial in range(1, num_trials+1):
        universe_set = set(range(1, 50*trial + 1))

        if set_inside_set:
            receiver_set = universe_set
        else:
            # receiver_set = set([7, 16, 21])
            receiver_set = set(range(1, 5 * trial + 1))
            print("receiver_set: ", receiver_set)
        
        universe_size = len(universe_set)
        sender = IBLTWithEGH(sender_set, universe_size, primes)
        receiver = IBLTWithEGH(receiver_set, universe_size, primes)

        sender.generate_egh_mapping()
        receiver.generate_egh_mapping()

        symmetric_difference = []

        while True:
            sender_cells = []

            sender.transmit()

            while not sender.cells_queue.empty():
                cell = sender.cells_queue.get()

                # End of IBLT's cells transmitting.
                if cell == "end":
                    break

                sender_cells.append(cell)

            symmetric_difference = receiver.receive(sender_cells)

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
    # d = 1
    sender_set_1 = set([8])

    # d = 5
    sender_set_5 = set([1, 4, 8, 15, 23])

    # d = 20
    sender_set_20 = set(range(1, 21))

    # primes in 1 - 400 (for EGH).
    primes = [2, 3, 5, 7, 11, 13, 17, 19, 23, 31, 37, 41, 43, 47, 53, 59, 
    61, 67, 71, 73, 79, 83, 89, 97, 101, 103, 107, 109, 113, 127, 131, 137, 
    139, 149, 151, 157, 163, 167, 173, 179, 181, 191, 193, 197, 199, 211, 
    223, 227, 229, 233, 239, 241, 251, 257, 263, 269, 271, 277, 281, 283, 293,
    307, 311, 313, 317, 331, 337, 347, 349, 353, 359, 367, 373, 379, 383, 
    389, 397]

    num_trials = 10 

    print("IBLT + EGH:")

    for sender_set in [sender_set_1, sender_set_5, sender_set_20]:
        # benchmark_set_reconciliation_egh(sender_set, primes, num_trials, 
        #                             export_to_csv=True, 
        #                             csv_filename=f"egh_results/egh_results_receiver_includes_sender_d_{len(sender_set)}.csv", 
        #                             set_inside_set = True)

        benchmark_set_reconciliation_egh(sender_set, primes, num_trials, 
                                    export_to_csv=True, 
                                    csv_filename=f"egh_results/egh_results_receiver_not_includes_sender_d_{len(sender_set)}.csv", 
                                    set_inside_set = False)