<script lang="ts">
	import { createEventDispatcher } from 'svelte'
	import Icon from './Icon.svelte'

	const dispatch = createEventDispatcher()
	export let didWin: boolean
	export let moves: number
</script>

<div class="bottom-controls">
	<button data-testid="undo" on:click={() => dispatch('undo')} disabled={didWin || moves <= 0}>
		<Icon icon="undo" color="white" /> Undo
	</button>
	<button data-testid="hint" on:click={() => dispatch('hint')} disabled={didWin}>
		<Icon icon="help" color="white" /> Hint
	</button>
</div>

<style>
	.bottom-controls {
		display: flex;
		justify-content: center;
		flex-direction: column;
	}
	button:disabled {
		opacity: 0.4;
	}
	button {
		cursor: pointer;
		transition: opacity 70ms var(--easing-standard);
		min-width: 52px;
		min-height: 52px;
		color: var(--color-white);
		background-color: var(--color-primary);
	}
</style>
