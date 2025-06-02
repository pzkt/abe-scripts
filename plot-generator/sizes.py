import matplotlib.pyplot as plt
import pandas as pd
import numpy as np
import random

path = "../scheme-benchmarking/Results/storage/"
files = ["charm_bsw07_ct","charm_fame_ct","charm_rw15_ct","charm_waters11_ct","charm_yahk14_ct","circl_tkn20_ct","gofe_fame_ct"]
markers = ["o","s","D","P","^","o","o","o","o","s","D"];
lib_colors = ["#003a7d"] * 5 + ["#008dff", "#d83034" , "#c701ff"] + ["#4ecb8d"] * 3

data = []
for file in files:
    try:
        csv_file = pd.read_csv(path + file + ".csv")
        selectRow = csv_file.iloc[0]
        selectRow = selectRow.drop(["index", "attributes"], errors="ignore")
        selectRow = selectRow[~selectRow.index.str.startswith("single")]
        data.append(selectRow)
    except:
        continue
plt.figure(figsize=(7, 5))

colors = ['#e41a1c','#377eb8','#4daf4a','#984ea3','#ff7f00']

for i in range(len(data)):
    if (i in []): #cull certain entries
        continue
    plt.semilogy(data[i].index, data[i].values, label=files[i] , marker=markers[i], linestyle="-", color=lib_colors[i])

x_positions = np.arange(len(data[0])) 
a = 1  # Starting value
b = 2  # Growth factor (adjust this to change the slope)

reference_line = a * (b ** x_positions)
print(reference_line)

# Plot the reference line
plt.semilogy(data[0].index, reference_line, 'r--', label=f'Exponential growth (factor {b})')

plt.legend(fontsize=10)
plt.grid(True, linestyle='--', alpha=0.7)

#plt.title(view)
plt.xlabel('Plaintext Size in Bytes')
plt.ylabel('Bytes')

plt.grid(True)

plt.tight_layout()
plt.show()