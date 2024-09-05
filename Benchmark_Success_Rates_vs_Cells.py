import platform
import matplotlib.pyplot as plt 
from Utils import *

def calc_decode_success_rate(symmetric_difference_size: int, 
                                max_symmetric_diff_size: int,
                                universe_size: int, 
                                method: Method, 
                                num_trials: int = 100,
                                set_inside_set: bool = True):
    
    success_rates = [0] * 20 * max_symmetric_diff_size

    for _ in range(num_trials):
        end_trial = False

        sender_iblt, receiver_iblt = generate_participants_iblts(symmetric_difference_size,
                                                                    method,
                                                                    set_inside_set)

        sender_cells = []
        reciver_cells = []

        num_iterations = 0 

        while True:
            sender_cells = sender_iblt.encode()

            # while not sender_iblt.cells_queue.empty():
            #     cell = sender_iblt.cells_queue.get()

            #     # End of IBLT's cells transmitting.
            #     if cell == "end":
            #         break

            #     sender_cells.append(cell)

            reciver_cells =  receiver_iblt.encode()

            # while not receiver_iblt.cells_queue.empty():
            #     cell = receiver_iblt.cells_queue.get()

            #     # End of IBLT's cells transmitting.
            #     if cell == "end":
            #         break

            #     reciver_cells.append(cell)

            initial_prob_idx = num_iterations

            for i in range(len(sender_cells)):
                iblt_diffrence = receiver_iblt.calc_iblt_diff(sender_cells[:initial_prob_idx+i], reciver_cells[:initial_prob_idx+i])
                symmetric_difference = receiver_iblt.listing(iblt_diffrence, with_deocde_frac = True)
                
                if symmetric_difference[0] == "Decode Failure":
                    success_rates[initial_prob_idx+len(sender_cells[:i])] += symmetric_difference[1]
                else:
                    success_rates[initial_prob_idx+len(sender_cells[:i])] += 1

                num_iterations += 1

                if num_iterations >=  len(success_rates):
                    end_trial = True
                    break
            
            if end_trial:
                break
        
        # Clean up
        del sender_iblt
        del receiver_iblt
        gc.collect()

    success_rates  = [prob / num_trials for prob in success_rates]
    return success_rates 

def benchmark_success_rates_vs_cells(method, universe_size, symmetric_diff_sizes, num_trials=1000,
                      export_to_csv: bool = False, 
                      csv_dir: str = "results/",
                      set_inside_set=True):

    for symmetric_diff_size in symmetric_diff_sizes:
        success_prob = calc_decode_success_rate(symmetric_diff_size, 
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
    # Check the system platform
    system = platform.system()

    if system == 'Linux':
        trials = 100 
        # trials = 20 

    elif system == 'Windows':
        trials = 25 

    universe_size = 1000

    print("IBLT + EGH:")

    benchmark_success_rates_vs_cells(Method.EGH, 
                      universe_size=universe_size, symmetric_diff_sizes=[1,3,10,20], 
                      num_trials=trials, export_to_csv=True, 
                      csv_dir=f"egh_results",
                      set_inside_set=False)
    
    print("IBLT + Extended Hamming Code:")

    benchmark_success_rates_vs_cells(Method.EXTENDED_HAMMING_CODE, 
                    universe_size=universe_size, symmetric_diff_sizes=[1,2,3,4,6,8], 
                    num_trials=trials, export_to_csv=True,
                    csv_dir=f"extended_hamming_results",
                    set_inside_set=False)
    
    print("IBLT + BCH:")

    benchmark_success_rates_vs_cells(Method.BCH, 
                    universe_size=universe_size, symmetric_diff_sizes=[1,3,7,10,15,20], 
                    num_trials=trials, export_to_csv=True, 
                    csv_dir=f"bch_results",
                    set_inside_set=False)
    
    