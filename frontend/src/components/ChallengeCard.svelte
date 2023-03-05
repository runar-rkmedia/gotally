<script lang="ts">
	import type { PartialMessage } from '@bufbuild/protobuf'
	import { GameMode, type NewGameRequest } from '../connect-web'
	import { storeHandler, type Challenge } from '../connect-web/store'
	import BoardPreview from './BoardPreview.svelte'
	import Icon from './Icon.svelte'
	import Pyro from './Pyro.svelte'

	const newGame = async (options: PartialMessage<NewGameRequest>) => {
		return storeHandler.commit(storeHandler.newGame(options))
	}
	const getRating = (n: number) => {
		switch (true) {
			case !n:
				return 'unplayed' as const
			case n < 20:
				return 'ok' as const
			case n < 40:
				return 'well' as const
			case n < 60:
				return 'good' as const
			case n < 80:
				return 'great' as const
			case n <= 100:
				return 'superb' as const
			case n > 100:
				return 'beyond' as const
		}
	}
	$: rating = getRating(challenge.rating)
	export let challenge: Challenge
</script>

<!-- svelte-ignore missing-declaration -->
<div
	class="rating"
	class:unplayed={rating === 'unplayed'}
	class:ok={rating === 'ok'}
	class:well={rating === 'well'}
	class:good={rating === 'good'}
	class:great={rating === 'great'}
	class:superb={rating === 'superb'}
	class:beyond={rating === 'beyond'}
	title={`Rating ${challenge.rating} ${rating}`}
>
	{#if rating === 'ok'}
		<Icon icon="star-half" />
	{:else if rating === 'well'}
		<Icon icon="star" />
	{:else if rating === 'good'}
		<Icon icon="star" />
		<Icon icon="star-half" />
	{:else if rating === 'great'}
		<Icon icon="star" />
		<Icon icon="star" />
	{:else if rating === 'superb'}
		<Icon icon="star" />
		<Icon icon="star" />
		<Icon icon="star-half" />
	{:else if rating === 'beyond'}
		<img src={'/trophy-64.png'} alt="trophy" />
	{:else}
		<!-- else content here -->
	{/if}
</div>
<button
	disabled={challenge.locked}
	class={'card'}
	class:unplayed={rating === 'unplayed'}
	class:rating-ok={rating === 'ok'}
	class:rating-well={rating === 'well'}
	class:rating-good={rating === 'good'}
	class:rating-great={rating === 'great'}
	class:rating-superb={rating === 'superb'}
	class:rating-beyond={rating === 'beyond'}
	on:click
>
	<div class="title">
		{#if challenge.locked}
			<Icon icon="lock" />
			<!-- content here -->
		{/if}
		{challenge.name}
	</div>
	<div class="preview">
		<BoardPreview
			columns={challenge.rows}
			rows={challenge.columns}
			cells={challenge.cells ||
				new Array(challenge.rows * challenge.columns || 30).fill(null).map((_, i) => ({
					base: Math.random() > 0.6 ? 0 : Math.ceil(Math.random() * 12),
					twopow: 0
				}))}
		/>
	</div>
</button>
<div class="ideal">
	{#if challenge.locked}
		Solve previous challenges to unlock
	{:else if challenge.moves}
		{#if challenge.moves === challenge.ideal}
			Par
		{:else if challenge.moves < challenge.ideal}
			{challenge.moves - challenge.ideal} moves better than par
		{:else}
			{challenge.moves - challenge.ideal} moves off from par
		{/if}
	{:else}
		Solve the game at {challenge.ideal} moves or less to get a superb rating
	{/if}
</div>
<div class="score">
	Score: {challenge.score}
</div>

<style lang="scss">
	button {
		position: relative;
		&:disabled {
			color: var(--color-grey-400);
			.preview {
				filter: grayscale(100%);
			}
		}
		border: 1px solid var(--color-grey-400);
		border-radius: 0.6em;
		overflow: hidden;
		.title {
			background: var(--color-grey-600);
			z-index: 1;
			position: inherit;
		}
		preview {
			overflow: hidden;
		}

		&.rating-beyond .title {
			background: var(--color-blue-700);
		}
		&.rating-superb .title {
			background: var(--color-green-700);
		}
		&.rating-great .title {
			background: var(--color-purple-300);
		}
		&.rating-good .title {
			background: var(--color-orange-500);
			color: var(--color-black);
		}
		&.rating-well .title {
			background: var(--color-yellow-300);
			color: var(--color-black);
		}
		&.rating-ok .title {
			background: var(--color-yellow-300);
			color: var(--color-black);
		}
	}
	.rating {
		position: absolute;
		z-index: 2;
		right: var(--size-1);
		color: hotpink;
		&.superb {
			color: goldenrod;
		}
		&.well {
			color: black;
		}
		&.ok {
			color: black;
		}
		&.good {
			color: black;
		}
		> img {
			position: absolute;
			top: -10px;
			right: -20px;
			margin-inline: -10px;
		}
	}
</style>
