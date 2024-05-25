import csv
import os
from typing import List, Set, Tuple
from IBLTWithEGH import IBLTWithEGH


def benchmark_set_reconciliation_egh(symmetric_difference_size:int, 
                                 primes: List[int], num_trials: int, 
                                 export_to_csv: bool = False, 
                                 csv_filename: str = "results.csv",
                                 set_inside_set: bool = True):
    total_cells_transmitted = 0
    results = []

    for trial in range(1, num_trials+1):
        universe_list = list(range(1, 100*trial + 1))
        

        if set_inside_set:
            receiver_list = universe_list
            sender_list = universe_list[symmetric_difference_size:]
        else:
            receiver_list = universe_list[:symmetric_difference_size]
            print("receiver_list: ", receiver_list)
            sender_list = receiver_list[-symmetric_difference_size:]
            sender_list.extend(universe_list[len(receiver_list):len(receiver_list)+symmetric_difference_size])
            print("sender_list: ", sender_list)

        universe_size = len(universe_list)
        sender = IBLTWithEGH(set(sender_list), universe_size, primes)
        receiver = IBLTWithEGH(set(receiver_list), universe_size, primes)

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
    # primes in 1 - 400 (for EGH).
    primes = [2, 3, 5, 7, 11, 13, 17, 19, 23, 31, 37, 41, 43, 47, 53, 59, 
    61, 67, 71, 73, 79, 83, 89, 97, 101, 103, 107, 109, 113, 127, 131, 137, 
    139, 149, 151, 157, 163, 167, 173, 179, 181, 191, 193, 197, 199, 211, 
    223, 227, 229, 233, 239, 241, 251, 257, 263, 269, 271, 277, 281, 283, 293,
    307, 311, 313, 317, 331, 337, 347, 349, 353, 359, 367, 373, 379, 383, 
    389, 397]

    # num_trials = 10 
    num_trials = 10 

    print("IBLT + EGH:")

    # symmetric_difference_size is parameter d.
    for symmetric_difference_size in [1, 5, 10, 20]:
        benchmark_set_reconciliation_egh(symmetric_difference_size, primes, 
                                    num_trials, 
                                    export_to_csv=True, 
                                    csv_filename=f"egh_results/egh_results_receiver_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
                                    set_inside_set = True)

        benchmark_set_reconciliation_egh(symmetric_difference_size, primes, 
                                    num_trials, 
                                    export_to_csv=True, 
                                    csv_filename=f"egh_results/egh_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
                                    set_inside_set = False)