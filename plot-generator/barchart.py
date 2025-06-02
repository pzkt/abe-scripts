import matplotlib.pyplot as plt
import pandas as pd

# Define your custom color mapping
color_map = {
    'charm bsw07': '#003a7d',
    'charm fame': '#003a7d', 
    'charm rw15': '#003a7d',
    'charm waters11': '#003a7d',  
    'charm yahk14': '#003a7d', 
    'circl tkn20': '#008dff',
    'gofe fame': '#d83034',  
    'openABE waters11': '#c701ff', 
    'rabe bsw07': '#4ecb8d',
    'rabe fame': '#4ecb8d',
    'rabe ghw11': '#4ecb8d'
}

# Read and sort data
df = pd.read_csv('data.csv', header=None, names=['Name', 'Value'])
df_sorted = df.sort_values('Value', ascending=False)

# Create list of colors in the sorted order
colors = [color_map.get(name, '#888888') for name in df_sorted['Name']]  # Default gray for unmapped categories

# Create bar chart with custom colors
plt.figure(figsize=(10, 6))
bars = plt.bar(df_sorted['Name'], df_sorted['Value'], color=colors)

# Customize chart
plt.ylabel('Time [ms]')
plt.xticks(rotation=45, ha='right')

# Add value labels
for bar in bars:
    height = bar.get_height()
    plt.text(bar.get_x() + bar.get_width()/2., height,
             f'{height:.2f}',
             ha='center', va='bottom')

plt.subplots_adjust(
    bottom=0.2,  # More bottom space for more bars
    top=0.9,
    left=0.1,
    right=0.95
)

plt.show()