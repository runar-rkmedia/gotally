<script lang="ts">
	import { store } from '../../connect-web/store'

	const indexToCoord = (index: number, rows: number, columns: number) => {
		const row = Math.floor(index / rows)
		const column = index - row * columns
		return { row: row + 1, column: column + 1 }
	}
	const enabled = false
</script>

{#if enabled}
	<!-- content here -->
	<ul>
		{#each $store.hints as hint, i}
			<li class:done={$store.hintDoneIndex >= i}>
				{#if hint.instructionOneof.case === 'swipe'}
					{hint.instructionOneof.value}
				{:else if hint.instructionOneof.case === 'combine'}
					{hint.instructionOneof.value.index}
					{JSON.stringify(
						hint.instructionOneof.value.index.map((m) => {
							const { row, column } = indexToCoord(
								m,
								$store.session.game.board.rows,
								$store.session.game.board.columns
							)
							return `${column}x${row}`
						})
					)}
				{/if}
			</li>
		{/each}
	</ul>
{/if}

<style>
	ul {
		font-size: 0.7rem;
	}
	.done {
		opacity: 0.5;
		font-size: 0.4rem;
	}
</style>
