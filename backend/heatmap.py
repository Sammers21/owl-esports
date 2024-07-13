from matplotlib.colors import LinearSegmentedColormap
import matplotlib.pyplot as plt
import numpy as np
import argparse


def heroes_pick_heatmap(heroes, winrates):
    # picks = np.random.random((16, 16))
    fig = plt.figure()
    fig.patch.set_visible(False)
    # show the heatmap
    plt.imshow(picks, cmap='hot', interpolation='nearest')
    plt.axis('off')
    plt.savefig('heatmap.png', bbox_inches='tight', pad_inches=0)
    plt.close(fig)
    return 'heatmap.png'

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument("--heroes", type=str, help="heroes picked")
    parser.add_argument('--winrates', type=str, help='winrates')
    parser.add_argument('--output', type=str, help='output')
    args = parser.parse_args()
    print("Heroes: ", args.heroes)
    print("Winrates: ", args.winrates)
    print("Output: ", args.output)
    plt.ioff()
    heroes = args.heroes.split(',')
    radiant = heroes[:5] + ['Avg']
    dire = heroes[5:] + ['Avg']
    winrates = args.winrates.split(';')
    def splitAndParse(line):
        return list(map(float, line.split(',')))
    winrates_numbers = map(splitAndParse, winrates)
    np_winrates = np.array(list(winrates_numbers))
    row_averages = np.mean(np_winrates, axis=1)

    # Reshape the averages to a column vector
    row_averages = row_averages.reshape(-1, 1)

    # Horizontally stack the column vector to the original np_winrates array
    np_winrates_with_row_avg = np.hstack((np_winrates, row_averages))

    # Calculate the average of each column
    col_averages = np.mean(np_winrates_with_row_avg, axis=0)

    # Reshape the averages to a row vector
    col_averages = col_averages.reshape(1, -1)

    # Vertically stack the row vector to the np_winrates_with_row_avg array
    np_winrates = np.vstack((np_winrates_with_row_avg, col_averages))
    
    min = np_winrates.min()
    max = np_winrates.max()
    min_max_diff = max - min
    def distanceto50(x):
        return abs(x - 50)
    np_mapped = np.vectorize(distanceto50)(np_winrates)
    minTo50 = np_mapped.min()
    min50Original = np.argwhere(np_mapped == minTo50)
    min50OriginalVal = np_winrates[min50Original[0][0], min50Original[0][1]]
    print("min: ", min)
    print("max: ", max)
    print("closes_to_50: ", minTo50)
    print("min50Original: ", np_winrates[min50Original[0][0], min50Original[0][1]])
    min50Base = min50OriginalVal - min
    min50BasePerc = min50Base / min_max_diff
    print("min50Base: ", min50Base)
    print("min50BasePerc: ", min50BasePerc)
    colors = [(1, 0, 0), (1, 1, 1), (0, 1, 0)]  # Red, White, Green
    nodes = [0.0, min50BasePerc, 1.0]  # Positions for 45, 50, 55
    custom_cmap = LinearSegmentedColormap.from_list("custom_cmap", list(zip(nodes, colors)))
    fig, ax = plt.subplots(figsize=(11, 6))
    im = ax.imshow(np_winrates, cmap=custom_cmap)
    cbar = ax.figure.colorbar(im, ax=ax)
    cbar.ax.set_ylabel("Winrate %", rotation=0, va="bottom", labelpad=30)
    ax.set_yticks(np.arange(len(radiant)), labels=radiant)
    ax.xaxis.set_ticks_position('top')
    ax.set_xticks(np.arange(len(dire)), labels=dire)
    plt.setp(ax.get_xticklabels(), rotation=45, ha="left",  rotation_mode="anchor")
    for i in range(len(radiant)):
        for j in range(len(dire)):
            text = ax.text(j, i, f"{np_winrates[i, j]:.2f}%", ha="center", va="center", color="black")
    # ax.set_title("Dota 2 Winrates Heatmap")
    pick_winrate = np.mean(np_winrates[5, :5])
    if pick_winrate > 50:
        title = "Radinant/Dire win chance: {:.2f}%/{:.2f}%".format(pick_winrate, 100 - pick_winrate)
        # set title color to green
        ax.title.set_color('green')
    else:
        title = "Radinant/Dire win chance: {:.2f}%/{:.2f}%".format(pick_winrate, 100 - pick_winrate)
        ax.title.set_color('red')
    
    ax.set_title(title)

    print("Radiant Winrate: ", pick_winrate)
    fig.tight_layout()
    plt.savefig(args.output, bbox_inches='tight', pad_inches=0)
    plt.close(fig)