<script lang="ts">
	import type { PartialMessage } from '@bufbuild/protobuf'
	import { GameMode, Rating, type NewGameRequest } from '../connect-web'
	import { storeHandler, type Challenge } from '../connect-web/store'
	import BoardPreview from './BoardPreview.svelte'
	import Icon from './Icon.svelte'

	const newGame = async (options: PartialMessage<NewGameRequest>) => {
		return storeHandler.commit(storeHandler.newGame(options))
	}
	const getRating = (r: Rating): keyof typeof Rating => {
		const s = Rating[r] as keyof typeof Rating
		if (s) {
			return s
		}
		for (const ns of Object.values(Rating)) {
			const n = Number(ns)
			if (isNaN(n)) {
				continue
			}
			if (r <= n) {
				const rating = getRating(n)
				console.log('ding', r, n, rating)
				return rating
			}
		}
		return getRating(0)
	}
	$: rating = getRating(challenge.rating)
	export let challenge: Challenge
</script>

<!-- svelte-ignore missing-declaration -->
<div
	class="rating"
	class:unplayed={rating === 'UNPLAYED'}
	class:ok={rating === 'OK'}
	class:well={rating === 'WELL'}
	class:good={rating === 'GOOD'}
	class:great={rating === 'GREAT'}
	class:superb={rating === 'SUPERB'}
	class:beyond={rating === 'BEYOND'}
	title={`Rating ${challenge.rating} ${rating}`}
>
	{#if rating === 'OK'}
		<Icon icon="star-half" />
	{:else if rating === 'WELL'}
		<Icon icon="star" />
	{:else if rating === 'GOOD'}
		<Icon icon="star" />
		<Icon icon="star" />
	{:else if rating === 'GREAT'}
		<Icon icon="star" />
		<Icon icon="star" />
		<Icon icon="star-half" />
	{:else if rating === 'SUPERB'}
		<Icon icon="star" />
		<Icon icon="star" />
		<Icon icon="star" />
	{:else if rating === 'BEYOND'}
		<img src={'/trophy-64.png'} alt="trophy" />
	{:else}
		<!-- else content here -->
	{/if}
</div>
<button
	disabled={challenge.locked}
	class={'card'}
	class:unplayed={rating === 'UNPLAYED'}
	class:rating-ok={rating === 'OK'}
	class:rating-well={rating === 'WELL'}
	class:rating-good={rating === 'GOOD'}
	class:rating-great={rating === 'GREAT'}
	class:rating-superb={rating === 'SUPERB'}
	class:rating-beyond={rating === 'BEYOND'}
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
					twopow: 0,
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
	Score: {challenge.currentUsersBestScore}
	Least moves: {challenge.currentUsersFewestMoves}
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
			background: var(--color-green-300);
			color: var(--color-black);
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
