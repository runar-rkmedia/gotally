<script lang="ts">
	import { fade } from 'svelte/transition'
	import { primeFactors } from './board/cell'
	import Cell from './board/Cell.svelte'

	export let n: number
	$: primes = Object.entries(
		primeFactors(n).reduce((r, f) => {
			if (r[f]) {
				r[f]++
			} else {
				r[f] = 1
			}
			return r
		}, {} as Record<number, number>)
	)
</script>

{#if n > 1}
	<div class="wrapper">
		{#each primes as [prime, primecount]}
			<!-- content here -->
			<div class="factor" data-factor={prime} transition:fade>
				<span class="prime" title={prime}>
					{prime}
				</span>
				<sup>{primecount}</sup>
			</div>
		{/each}
	</div>
{/if}

<style>
	.wrapper {
		display: flex;
		flex-wrap: wrap;
		gap: var(--size-2);
		flex-direction: column;
		justify-content: end;
		height: 100%;
		max-height: 100px;
	}

	[data-factor='1'] {
		--c: var(--color-grey-400);
	}
	[data-factor='2'] {
		--c: var(--color-orange-700);
	}
	[data-factor='3'] {
		--c: var(--color-green-700);
	}
	[data-factor='5'] {
		--c: var(--color-blue-700);
	}
	[data-factor='7'] {
		--c: var(--color-purple-700);
	}
	[data-factor='11'] {
		--c: var(--color-red-700);
	}
	.factor {
		background-color: var(--c);
		padding-inline: var(--size-3);
		border-radius: var(--radius-lg);
		height: min-content;
		text-align: center;
	}
</style>
