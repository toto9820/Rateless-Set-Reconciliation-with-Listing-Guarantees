import csv
import os
import random
from typing import List, Set, Tuple
from IBLTWithEGH import IBLTWithEGH
from IBLTWithCovArr import IBLTWithCovArr
from Method import Method
from IBLTWithRecursiveArr import IBLTWithRecursiveArr


def benchmark_set_reconciliation(symmetric_difference_size:int, 
                                 method:Method,
                                 num_trials: int, 
                                 export_to_csv: bool = False, 
                                 csv_filename: str = "results.csv",
                                 set_inside_set: bool = True):
    total_cells_transmitted = 0
    results = []

    for trial in range(1, num_trials+1):
        universe_list = list(range(1, 100*trial + 1))  
        # universe_list = list(range(1, 5*trial + 1))       

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

        if method == Method.EGH:
            sender = IBLTWithEGH(set(sender_list), universe_size)
            receiver = IBLTWithEGH(set(receiver_list), universe_size)

        elif method == Method.BINARY_COVERING_ARRAY:
            sender = IBLTWithCovArr(set(sender_list), universe_size)
            receiver = IBLTWithCovArr(set(receiver_list), universe_size)

        elif method == Method.RECURSIVE_ARRAY:
            sender = IBLTWithRecursiveArr(set(sender_list), universe_size)
            receiver = IBLTWithRecursiveArr(set(receiver_list), universe_size)

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

        total_cells_transmitted = len(receiver.iblt_sender_cells)
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

# TODO - continue and ask ori for guidance on this test maybe.
def measure_decode_success_rate(symmetric_difference_size:int,
                                 universe_size:int, 
                                 method:Method):
    
    universe_list = list(range(1, universe_size + 1))  
    receiver_list = universe_list
    

    for sender_size in range[100:1100:100]:
        sender_list = universe_list[:sender_size:]
       
    # return success_rate

if __name__ == "__main__":
    num_trials = 10 
    # num_trials = 1 

    print("IBLT + EGH:")

    # symmetric_difference_size is parameter d.
    # for symmetric_difference_size in [1, 5, 10, 20]:
    #     benchmark_set_reconciliation(symmetric_difference_size, 
    #                                  Method.EGH,
    #                                  num_trials, 
    #                                  export_to_csv=False, 
    #                                  csv_filename=f"egh_results/egh_results_receiver_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                  set_inside_set = True)

    #     benchmark_set_reconciliation(symmetric_difference_size, 
    #                                  Method.EGH,
    #                                  num_trials, 
    #                                  export_to_csv=True, 
    #                                  csv_filename=f"egh_results/egh_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                  set_inside_set = False)
        

    print("IBLT + Binary Covering Arrays:")

    #for symmetric_difference_size in [1, 5, 10, 20]:
    for symmetric_difference_size in [10, 20]:
        benchmark_set_reconciliation(symmetric_difference_size,
                                     Method.BINARY_COVERING_ARRAY, 
                                     num_trials, 
                                     export_to_csv=True, 
                                     csv_filename=f"covering_arr_results/covering_arr_results_receiver_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
                                     set_inside_set = True)

        benchmark_set_reconciliation(symmetric_difference_size, 
                                     Method.BINARY_COVERING_ARRAY,
                                     num_trials, 
                                     export_to_csv=True, 
                                     csv_filename=f"covering_arr_results/covering_arr_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
                                     set_inside_set = False)
        
    # print("IBLT + Recurisve Array:")

    # for symmetric_difference_size in [1, 5, 10, 20]:
    #     benchmark_set_reconciliation(symmetric_difference_size,
    #                                  Method.RECURSIVE_ARRAY, 
    #                                  num_trials, 
    #                                  export_to_csv=True, 
    #                                  csv_filename=f"recursive_arr_results/recursive_arr_results_receiver_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                  set_inside_set = True)

    #     benchmark_set_reconciliation(symmetric_difference_size, 
    #                                  Method.RECURSIVE_ARRAY,
    #                                  num_trials, 
    #                                  export_to_csv=True, 
    #                                  csv_filename=f"recursive_arr_results/recursive_arr_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                  set_inside_set = False)