import matplotlib.pyplot as plt
import pandas as pd
import random

path = "../scheme-benchmarking/Results/performance/"
files = ["charm_bsw07","charm_fame","charm_rw15","charm_waters11","charm_yahk14","circl_tkn20","gofe_fame","openabe_waters11","rabe_bsw07","rabe_fame","rabe_ghw11"]
markers = ["o","s","D","P","^","o","o","o","o","s","D"];
lib_colors = ["#003a7d"] * 5 + ["#008dff", "#d83034" , "#c701ff"] + ["#4ecb8d"] * 3

data = []
for file in files:
    data.append(pd.read_csv(path + file + ".csv"))

plt.figure(figsize=(7, 5))

view = "and decrypt"

colors = ['#e41a1c','#377eb8','#4daf4a','#984ea3','#ff7f00']
for i in range(len(files)):
    if (i in [0,2,3,4,8]): #cull certain entries
        continue
    if (view in data[i]):
        plt.plot(data[i]['attributes'], data[i][view], label=files[i] , marker=markers[i], linestyle="-", color=lib_colors[i])

plt.legend(fontsize=10)
plt.grid(True, linestyle='--', alpha=0.7)

#plt.title(view)
plt.xlabel('Number of Attributes')
plt.ylabel('Time [s]')

plt.grid(True)

plt.savefig(f"{view}.png", bbox_inches='tight', pad_inches=0.1)

plt.tight_layout()
plt.show()