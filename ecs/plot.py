import seaborn
import pandas as pd
import matplotlib.pyplot as plt
#from pathlib import Path

seaborn.set_style("dark")

#files = Path("./results").rglob('*.csv')  # .rglob to get subdirectories
#files = ["results/go/release/native/5000_0.txt", "results/rust/release/native/5000_0.txt"];
# files = [
#     "results/rust/release/native/1000_0.txt",
#     "results/go/release/native/1000_0.txt",
#     "results/rust/release/nativeSplit/1000_0.txt",
#     "results/go/release/nativeSplit/1000_0.txt",
#     "results/rust/release/ecs/1000_0.txt",
#     "results/go/release/ecs/1000_0.txt",
#     "results/rust/release/ecs-slow/1000_0.txt",
#     "results/go/release/ecs-slow/1000_0.txt",
#     ];

languages = ["go", "rust"]
programs = ["native", "nativeSplit", "ecs"]
sizes = ["1000", "5000", "10000"]
#sizes = ["1000", "2000", "3000", "4000", "5000", "6000", "7000", "8000", "9000", "10000"]
collisions = ["0"]

files = []
for l in languages:
    for p in programs:
        for s in sizes:
            for c in collisions:
                newFile = {
                    "path": "results/"+l+"/release/"+p+"/"+s+"_"+c+".txt",
                    "lang": l,
                    "prog": p,
                    "size": s,
                    "col": c,
                }
                files.append(newFile)

dfs = list()
for f in files:
    print(f)
    data = pd.read_csv(f['path'], delim_whitespace=True)
    # .stem is method for pathlib objects to get the filename w/o the extension
    #    data['file'] = f
    data['lang'] = f['lang']
    data['prog'] = f['prog']
    data['size'] = f['size']
    data['col'] = f['col']
    dfs.append(data)

df = pd.concat(dfs, ignore_index=True)

print(df)

seaborn.relplot(
    data=df,
    x="Iter", y="Time", hue="lang", col="prog",
    row="size",
#    hue="smoker", style="smoker", size="size",
#    kind="line", errorbar="sd",
)
plt.show()

#res = seaborn.violinplot(x=dfs['Time'])
#plt.show();

# csv = pandas.read_csv("results/go/release/native/1000_0.txt", delim_whitespace=True)
# csv2 = pandas.read_csv("results/rust/release/native/1000_0.txt", delim_whitespace=True)

# print (csv['Time'].describe())
# print (csv2['Time'].describe())
# #res = seaborn.pointplot(x=csv['Iter'], y=csv['Time'])
# #res = seaborn.violinplot(x=csv['Time'])
# #res = seaborn.violinplot(x=csv2['Time'])
# #plt.show()

