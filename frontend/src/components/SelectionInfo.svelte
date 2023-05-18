<script lang="ts">
	import Counter from './Counter.svelte'
	import PrimeFactors from './PrimeFactors.svelte'
	export let selectionSum: number
	export let lastSelectionValue: number
	export let selectionProduct: number
</script>

<!-- When the user has selected some cells, show some helpful information about that selection -->
<div class="selectionCounter">
	<Counter
		show={!!selectionSum}
		asCell={false}
		value={selectionSum}
		label="Sum"
		variant={lastSelectionValue * 2 < selectionSum
			? 'error'
			: lastSelectionValue * 2 === selectionSum
			? 'success'
			: 'normal'}
	/>
	<Counter
		show={selectionProduct > 1}
		asCell={true}
		value={selectionProduct}
		label="Product"
		variant={lastSelectionValue < selectionProduct / lastSelectionValue
			? 'error'
			: lastSelectionValue === selectionProduct / lastSelectionValue
			? 'success'
			: 'normal'}
	/>
	<PrimeFactors n={selectionProduct} />
</div>

<style>
	.selectionCounter {
		display: grid;
		grid-template-columns: 5fr 5fr 2fr;
		gap: 10px;
	}
</style>
