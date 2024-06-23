import csv
import os
import random
import math
import matplotlib.pyplot as plt
from typing import List, Set, Tuple
from IBLTWithEGH import IBLTWithEGH
from IBLTWithCovArr import IBLTWithCovArr
from Method import Method
from IBLTWithRecursiveArr import IBLTWithRecursiveArr
from IBLTWithExtendedHamming import IBLTWithExtendedHamming

def benchmark_set_reconciliation(symmetric_difference_size: int, 
                                 method: Method,
                                 num_trials: int, 
                                 export_to_csv: bool = False, 
                                 csv_filename: str = "results.csv",
                                 set_inside_set: bool = True):
    total_cells_transmitted = 0
    results = []

    for trial in range(1, num_trials+1):
        universe_list = list(range(1, 100*trial + 1)) 
        universe_size = len(universe_list)       

        if set_inside_set:
            print(f"Receiver set is a super set of sender set for symmetric_difference_size {symmetric_difference_size}")
            receiver_list = universe_list
            sender_list = random.sample(universe_list, universe_size - symmetric_difference_size)
        else:
            print(f"Receiver set is not a super set of sender set symmetric_difference_size {symmetric_difference_size}")
            receiver_size = random.randint(1, universe_size - symmetric_difference_size)
            receiver_size = max(symmetric_difference_size, receiver_size)
            receiver_list = random.sample(universe_list, receiver_size)

            # print("receiver_list: ", receiver_list)

            universe_without_receiver_set = set(universe_list) - set(receiver_list)
            
            sender_list = []

            sender_list = list(universe_without_receiver_set)[:symmetric_difference_size-1]

            sender_list.extend(receiver_list[:(receiver_size-1)])
            
            # print("sender_list: ", sender_list)

        if method == Method.EGH:
            sender = IBLTWithEGH(set(sender_list), universe_size)
            receiver = IBLTWithEGH(set(receiver_list), universe_size)

        elif method == Method.BINARY_COVERING_ARRAY:
            sender = IBLTWithCovArr(set(sender_list), universe_size)
            receiver = IBLTWithCovArr(set(receiver_list), universe_size)

        elif method == Method.RECURSIVE_ARRAY:
            sender = IBLTWithRecursiveArr(set(sender_list), universe_size)
            receiver = IBLTWithRecursiveArr(set(receiver_list), universe_size)

        elif method == Method.EXTENDED_HAMMING_CODE:
            sender = IBLTWithExtendedHamming(set(sender_list), universe_size)
            receiver = IBLTWithExtendedHamming(set(receiver_list), universe_size)

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
        export_results_to_csv(["Trial", "Universe Size", "Cells Transmitted"],
                              results, csv_filename)

def export_results_to_csv(header, results, csv_filename: str) -> None:
    with open(os.path.join("./data", csv_filename), mode='w', newline='') as file:
        writer = csv.writer(file)
        writer.writerow(header)
        writer.writerows(results)

# TODO - continue and ask ori for guidance on this test maybe.
def measure_decode_success_rate(symmetric_difference_size: int, 
                                max_symmetric_diff_size: int,
                                universe_size: int, 
                                method: Method, 
                                num_trials: int = 1000,
                                set_inside_set: bool = True):
    
    universe_list = list(range(1, universe_size + 1))  

    success_probabilities = [0] * 4 * max_symmetric_diff_size

    for _ in range(num_trials):
        end_trial = False

        if set_inside_set:
            receiver_list = universe_list
            sender_list = random.sample(universe_list, universe_size - symmetric_difference_size)
        else:
            receiver_size = random.randint(1, universe_size - symmetric_difference_size)
            receiver_size = max(symmetric_difference_size, receiver_size)
            receiver_list = random.sample(universe_list, receiver_size)

            # print("receiver_list: ", receiver_list)

            universe_without_receiver_set = set(universe_list) - set(receiver_list)
            
            sender_list = []

            sender_list = list(universe_without_receiver_set)[:symmetric_difference_size-1]

            sender_list.extend(receiver_list[:(receiver_size-1)])
            
            # print("sender_list: ", sender_list)

        if method == Method.EGH:
            sender = IBLTWithEGH(set(sender_list), universe_size)
            receiver = IBLTWithEGH(set(receiver_list), universe_size)

        elif method == Method.BINARY_COVERING_ARRAY:
            sender = IBLTWithCovArr(set(sender_list), universe_size)
            receiver = IBLTWithCovArr(set(receiver_list), universe_size)

        elif method == Method.RECURSIVE_ARRAY:
            sender = IBLTWithRecursiveArr(set(sender_list), universe_size)
            receiver = IBLTWithRecursiveArr(set(receiver_list), universe_size)

        elif method == Method.EXTENDED_HAMMING_CODE:
            sender = IBLTWithExtendedHamming(set(sender_list), universe_size)
            receiver = IBLTWithExtendedHamming(set(receiver_list), universe_size)

        sender_cells = []
        reciver_cells = []

        num_iterations = 0 

        while True:
            sender.transmit()

            while not sender.cells_queue.empty():
                cell = sender.cells_queue.get()

                # End of IBLT's cells transmitting.
                if cell == "end":
                    break

                sender_cells.append(cell)

            receiver.transmit()

            while not receiver.cells_queue.empty():
                cell = receiver.cells_queue.get()

                # End of IBLT's cells transmitting.
                if cell == "end":
                    break

                reciver_cells.append(cell)

            initial_prob_idx = num_iterations

            for i in range(len(sender_cells)):
                iblt_diffrence = receiver.calc_iblt_diff(sender_cells[:initial_prob_idx+i], reciver_cells[:initial_prob_idx+i])
                symmetric_difference = receiver.listing(iblt_diffrence, with_deocde_frac = True)
                
                if symmetric_difference[0] == "Decode Failure":
                    success_probabilities[initial_prob_idx+len(sender_cells[:i])] += symmetric_difference[1]
                else:
                    success_probabilities[initial_prob_idx+len(sender_cells[:i])] += 1

                num_iterations += 1

                if num_iterations >=  len(success_probabilities):
                    end_trial = True
                    break
            
            if end_trial:
                break

    success_probabilities = [prob / num_trials for prob in success_probabilities]
    return success_probabilities

def plot_success_rate(method, universe_size, symmetric_diff_sizes, num_trials=1000,
                      export_to_csv: bool = False, 
                      csv_dir: str = "results/",
                      set_inside_set=True):

    for symmetric_diff_size in symmetric_diff_sizes:
        success_prob = measure_decode_success_rate(symmetric_diff_size, 
                                                    max(symmetric_diff_sizes),
                                                    universe_size, 
                                                    method, 
                                                    num_trials,
                                                    set_inside_set)
        
        if export_to_csv:
            if set_inside_set:
                csv_filename = f"{csv_dir}/success_probability/{csv_dir}_success_probability_receiver_includes_sender_symmetric_diff_size_{symmetric_diff_size}.csv"
            
            else:
                csv_filename = f"{csv_dir}/success_probability/{csv_dir}_success_probability_receiver_not_includes_sender_symmetric_diff_size_{symmetric_diff_size}.csv"
            export_results_to_csv(["Cells Transmitted", "Success Probability"],
                                list(enumerate(success_prob)), csv_filename)

        plt.plot(range(1, len(success_prob) + 1), success_prob, label=f'|∆|={symmetric_diff_size}')

    plt.xlabel('Number of IBLT cells')
    plt.ylabel('Success Probability')
    plt.title(f'Success Probability vs. Number of cells for {method}')
    plt.legend()
    plt.show(block=True)

if __name__ == "__main__":
    universe_size = 100
    trials = 100 

    # print("IBLT + EGH:")

    # # symmetric_difference_size is parameter d.
    # for symmetric_difference_size in [1, 3, 10, 20]:
    #     benchmark_set_reconciliation(symmetric_difference_size, 
    #                                  Method.EGH,
    #                                  num_trials=10, 
    #                                  export_to_csv=True, 
    #                                  csv_filename=f"egh_results/egh_results_receiver_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                  set_inside_set = True)

    #     benchmark_set_reconciliation(symmetric_difference_size, 
    #                                  Method.EGH,
    #                                  num_trials=10, 
    #                                  export_to_csv=True, 
    #                                  csv_filename=f"egh_results/egh_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                  set_inside_set = False)
    
    # plot_success_rate(Method.EGH, 
    #                   universe_size=universe_size, symmetric_diff_sizes=[1,3,10,20], 
    #                   num_trials=trials, export_to_csv=True, 
    #                   csv_dir=f"egh_results",
    #                   set_inside_set=True)
        

    # print("IBLT + Binary Covering Arrays:")

    # #for symmetric_difference_size in [1, 3, 10, 20]:
    # for symmetric_difference_size in [10, 20]:
    #     benchmark_set_reconciliation(symmetric_difference_size,
    #                                  Method.BINARY_COVERING_ARRAY, 
    #                                  num_trials=10, 
    #                                  export_to_csv=True, 
    #                                  csv_filename=f"covering_arr_results/covering_arr_results_receiver_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                  set_inside_set = True)

    #     benchmark_set_reconciliation(symmetric_difference_size, 
    #                                  Method.BINARY_COVERING_ARRAY,
    #                                  num_trials=10, 
    #                                  export_to_csv=True, 
    #                                  csv_filename=f"covering_arr_results/covering_arr_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                  set_inside_set = False)
        
    # print("IBLT + Recurisve Array:")

    # for symmetric_difference_size in [1, 3, 10, 20]:
    #     benchmark_set_reconciliation(symmetric_difference_size,
    #                                  Method.RECURSIVE_ARRAY, 
    #                                  num_trials=10, 
    #                                  export_to_csv=True, 
    #                                  csv_filename=f"recursive_arr_results/recursive_arr_results_receiver_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                  set_inside_set = True)

    #     benchmark_set_reconciliation(symmetric_difference_size, 
    #                                  Method.RECURSIVE_ARRAY,
    #                                  num_trials=10, 
    #                                  export_to_csv=True, 
    #                                  csv_filename=f"recursive_arr_results/recursive_arr_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                  set_inside_set = False)

    print("IBLT + Extended Hamming Code:")

    for symmetric_difference_size in [1,3]:
        benchmark_set_reconciliation(symmetric_difference_size,
                                     Method.EXTENDED_HAMMING_CODE, 
                                     num_trials=10, 
                                     export_to_csv=True, 
                                     csv_filename=f"extended_hamming_results/extended_hamming_results_receiver_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
                                     set_inside_set = True)

        benchmark_set_reconciliation(symmetric_difference_size, 
                                     Method.EXTENDED_HAMMING_CODE,
                                     num_trials=10, 
                                     export_to_csv=True, 
                                     csv_filename=f"extended_hamming_results/extended_hamming_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
                                     set_inside_set = False)

    plot_success_rate(Method.EXTENDED_HAMMING_CODE, 
                    universe_size=universe_size, symmetric_diff_sizes=[1,2,3,4,6,8], 
                    num_trials=trials, export_to_csv=True,
                    csv_dir=f"extended_hamming_results",
                    set_inside_set=True)

