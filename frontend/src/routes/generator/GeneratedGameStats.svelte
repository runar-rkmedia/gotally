<script lang="ts">
	import type { GameStats } from '../../connect-web'

	// import type { GameStats } from '../../connect-web'

	export let stats: GameStats
</script>

{#if stats}
	<div>
		<table>
			<thead>
				<tr>
					<th>Cell Count</th>
					<th>Duplicate Factors</th>
					<th>Duplicate Values</th>
					<th>Hints</th>
					<th>Ideal Moves</th>
					<th>Ideal Moves Solution Index</th>
					<th>Max Score</th>
					<th>Max Score Solution Index</th>
					<th>Score On Ideal</th>
					<th>Unique Factors</th>
					<th>Unique Hints</th>
					<th>Unique Values</th>
					<th>With Value Count</th>
					<th>Solution Stats</th>
				</tr></thead
			>
			<tbody>
				<tr>
					<td>{stats.cellCount}</td>
					<td>{stats.duplicateFactors}</td>
					<td>{stats.duplicateValues}</td>
					<td>{stats.hints?.length}</td>
					<td>{stats.idealMoves}</td>
					<td>{stats.idealMovesSolutionIndex}</td>
					<td>{stats.maxScore}</td>
					<td>{stats.maxScoreSolutionIndex}</td>
					<td>{stats.scoreOnIdeal}</td>
					<td>{stats.uniqueFactors.join(', ')}</td>
					<td>{stats.uniqueHints}</td>
					<td>{stats.uniqueValues.join(', ')}</td>
					<td>{stats.withValueCount}</td>
					{#each stats.solutionStats as sol}
						<tr>
							<td title={JSON.stringify(sol)}>{sol.instructionTag.length}</td>
							<td>
								{#each sol.instructionTag as tag}
									{#if tag.isSwipe}
										👆
									{/if}
									{#if tag.isAddition}
										+
									{/if}
									{#if tag.isMultiplication}
										X
									{/if}
									{#if tag.twoPow}
										{tag.twoPow}
									{/if}
								{/each}
							</td>
						</tr>
					{/each}
				</tr>
			</tbody>
		</table>
	</div>
{/if}

<style>
	div {
		overflow: scroll;
		color: var(--color-blue-700);
		padding-inline: var(--size-5);
	}
	thead {
		height: 180px;
	}
	td {
		text-align: center;
	}
	thead tr :nth-child(even),
	tbody :nth-child(even) {
		color: var(--color-green-700);
	}
</style>
