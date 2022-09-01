<script lang="ts">
	import { SwipeDirection, type Instruction } from '../connect-web'
	export let hints: Instruction[]
	export let doneIndex: number
</script>

{#if hints}
	<h3>Hints</h3>
	<ol>
		{#each hints as hint, i}
			<li class:done={doneIndex >= i}>
				{#if hint.instructionOneof.case == 'swipe'}
					{#if hint.instructionOneof.value === SwipeDirection.UP}
						Swipe up
					{:else if hint.instructionOneof.value === SwipeDirection.RIGHT}
						Swipe right
					{:else if hint.instructionOneof.value === SwipeDirection.DOWN}
						Swipe down
					{:else if hint.instructionOneof.value === SwipeDirection.LEFT}
						Swipe left
					{/if}
				{:else}
					Combine {hint.instructionOneof.value?.toJsonString()}
				{/if}
			</li>
		{/each}
	</ol>
{/if}

<style>
	.done {
		text-decoration: line-through;
		opacity: 0.7;
	}
</style>
