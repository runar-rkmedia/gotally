<script lang="ts">
	import { onMount } from 'svelte'
	import { cellValue } from '../../components/board/cell'

	export let cells: { base: number; twopow: number }[]
	export let rows: number
	export let columns: number
	$: r = new Array(rows)
	$: c = new Array(columns)
	let values: number[][] = []
	const setup = () => {
		values = new Array(columns).fill(null).map((_, i) => {
			return new Array(rows).fill(null).map((_, j) => {
				return cellValue(cells[i * j + j])
			})
		})
	}
	onMount(() => {
		setup()
	})
</script>

<div>
	{#each c as _, i}
		{#if values[i]}
			<div>
				{#each r as _, j}
					<input type="number" bind:value={values[i][j]} />
				{/each}
			</div>
		{/if}
	{/each}
</div>
