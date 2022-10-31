<script lang="ts">
	import { fade } from 'svelte/transition'

	import Cell from './board/Cell.svelte'
	import { numberFormatter } from './formatNumber'

	export let value = 0
	export let variant: 'normal' | 'error' | 'success' = 'normal'
	export let show = true
	export let label = ''
</script>

<div
	class="wrapper"
	class:show
	class:error={variant === 'error'}
	class:success={variant === 'success'}
>
	<div class="label">
		{label}:
	</div>
	<!-- {numberFormatter(value)} -->
	<div class="cellContainer" transition:fade>
		<Cell cell={{ base: value, twopow: 0 }} />
	</div>
</div>

<style>
	.wrapper {
		font-family: var(--font-mono);
		display: flex;
		flex-direction: column;
		justify-content: center;
		text-align: center;
		align-items: center;
		transition: all 150ms var(--easing-standard);
	}
	.wrapper:not(.show) {
		opacity: 0;
	}
	.cellContainer {
		display: grid;
		height: 88px;
		width: 88px;
	}
	.error {
		color: var(--color-red-300);
	}
	.success {
		color: var(--color-green-700);
	}
	.label {
		width: 100%;
	}
</style>
