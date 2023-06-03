<script lang="ts">
	import { createEventDispatcher } from 'svelte'
	import Icon from './Icon.svelte'
	import { GradientButton as Button } from 'flowbite-svelte'

	const dispatch = createEventDispatcher()
	export let didWin: boolean
	export let moves: number
</script>

<div class="bottom-controls">
	<Button
		color="greenToBlue"
		size="xl"
		on:click={() => dispatch('undo')}
		disabled={didWin || moves <= 0}
	>
		<Icon icon="undo" /> Undo
	</Button>
	<Button color="red" size="xl" on:click={() => dispatch('hint')} disabled={didWin}>
		<Icon icon="help" /> Hint
	</Button>
</div>

<style>
	.bottom-controls {
		display: flex;
		justify-content: center;
		flex-direction: row;
		gap: var(--size-2);
	}
</style>
