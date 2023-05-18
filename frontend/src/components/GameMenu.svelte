<script lang="ts">
	import type { PartialMessage } from '@bufbuild/protobuf'
	import { onMount } from 'svelte'

	import { Difficulty, GameMode, type NewGameRequest } from '../connect-web/proto/tally/v1/board_pb'
	export let open: boolean

	import { storeHandler, store, type Challenge } from '../connect-web/store'
	import BoardPreview from './BoardPreview.svelte'
	import ChallengeCard from './ChallengeCard.svelte'
	import Icon from './Icon.svelte'

	const restartGame = () => {
		if (!$store?.session.game.moves) {
			return
		}
		return storeHandler.commit(storeHandler.restartGame())
	}
	const newGame = async (options: PartialMessage<NewGameRequest>) => {
		return storeHandler.commit(storeHandler.newGame(options))
	}
	let view: 'main' | 'challenges' | 'tutorials' | 'infinite-game' = 'challenges'
	const close = () => {
		view = 'main'
		open = false
	}
	$: {
		if (open === false) {
			view = 'main'
		}
	}
	let challenges: Challenge[] = []
	onMount(async () => {
		const r = await storeHandler.commit(storeHandler.getChallenges({}))
		if (r.error) {
			console.error(r)
			return
		}
		if (!r.result) {
			return
		}
		challenges = r.result
	})
</script>

<div class="container">
	{#if view === 'main'}
		<!-- content here -->
		<div class="buttons">
			{#if $store.session.game.moves}
				<div class="section">
					<p>Restart the current game?</p>

					<button
						disabled={$store.session.game.moves === 0}
						data-test-id="restart-game"
						style="--color: var(--color-orange)"
						on:click={() => {
							restartGame()
							close()
						}}
					>
						<Icon icon="restart" />
						Restart
					</button>
				</div>
			{/if}
			<div class="section">
				<p>Need som help? How about going through the tutorial?</p>
				<button
					style="--color: var(--color-yellow)"
					on:click={() => {
						newGame({ mode: GameMode.TUTORIAL })
						close()
					}}
				>
					<Icon icon="tutorial" />
					New Tutorial</button
				>
			</div>
			<div class="section">
				<p>Ready for a challenge? These are also great for learning new strategies!</p>
				<button
					style="--color: var(--color-green)"
					on:click={() => {
						view = 'challenges'
					}}
				>
					<Icon icon="challenge" />
					New Challenge</button
				>
			</div>
			<div class="section">
				<p>How far can you go?</p>
				<button
					style="--color: var(--color-blue)"
					on:click={() => {
						view = 'infinite-game'
					}}
				>
					<Icon icon="infinite" />
					New Infinite game</button
				>
			</div>
		</div>
	{:else if view === 'infinite-game'}
		<p>How far can you go?</p>
		<div class="buttons">
			<div class="section">
				<p>For newcomers to the game</p>
				<button
					style="--color: var(--color-blue)"
					on:click={() => {
						newGame({
							mode: GameMode.RANDOM,
							variant: {
								value: Difficulty.EASY,
								case: 'difficulty',
							},
						})
						close()
					}}>Easy</button
				>
			</div>
			<div class="section">
				<p>As the game is intended to be played</p>
				<button
					style="--color: var(--color-green)"
					on:click={() => {
						newGame({
							mode: GameMode.RANDOM,
							variant: {
								value: Difficulty.MEDIUM,
								case: 'difficulty',
							},
						})
						close()
					}}>Medium</button
				>
			</div>
			<div class="section">
				<p>Tough and unfair</p>
				<button
					style="--color: var(--color-orange)"
					on:click={() => {
						newGame({
							mode: GameMode.RANDOM,
							variant: {
								value: Difficulty.HARD,
								case: 'difficulty',
							},
						})
						close()
					}}>Hard</button
				>
			</div>
		</div>
	{:else if view === 'challenges'}
		<div class="cards">
			{#each challenges as c}
				<div class="card">
					<ChallengeCard
						challenge={c}
						on:click={() => {
							newGame({
								mode: GameMode.RANDOM_CHALLENGE,
								variant: {
									value: c.id,
									case: 'id',
								},
							})
							close()
						}}
					/>
				</div>
			{/each}
		</div>
	{/if}
	<div class="section">
		<p>Back to the game</p>
		<button
			style="--color: var(--color-gray-700)"
			on:click={() => {
				close()
			}}
		>
			<Icon icon="play" />
			Continue</button
		>
	</div>
</div>

<style lang="scss">
	.container {
		margin-top: env(titlebar-area-height);
		padding: var(--size-8);
		display: flex;
		flex-direction: column;
		flex-wrap: wrap;
		justify-content: space-between;
		align-items: space-between;
		text-align: center;
	}
	.buttons {
		margin-block-start: var(--size-8);
		display: grid;
		grid-template-columns: 1fr;
		gap: var(--size-4);
	}
	.section {
		border: 2px solid white;
		border-radius: 20px;
		padding: var(--size-2);
		background-color: #00380080;
		display: flex;
		flex-direction: column;
		justify-content: space-between;
	}
	.section p {
		padding-block-end: 16px;
	}

	.quote {
		font-size: var(--scale-fluid-3);
		margin-inline: auto;
	}
	.buttons button {
		font-size: 15px;
		padding: 0.7em 2.7em;
		letter-spacing: 0.06em;
		font-family: inherit;
		border-radius: 0.6em;
		overflow: hidden;
		transition: all 0.3s;
		line-height: 1.4em;
		border: 2px solid var(--color);
		background: linear-gradient(
			to right,
			rgba(27, 253, 156, 0.1) 1%,
			transparent 40%,
			transparent 60%,
			rgba(27, 253, 156, 0.1) 100%
		);
		color: var(--color);
		box-shadow: inset 0 0 10px rgba(27, 253, 156, 0.4), 0 0 9px 3px rgba(27, 253, 156, 0.1);
		&:disabled {
			color: var(--color-gray-400);
		}
	}
	.cards {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--size-4);
		.card {
			position: relative;
			width: 100%;
			height: 100%;
			display: grid;
		}
	}

	@media (max-width: 640px) {
		.cards,
		.buttons {
			grid-template-columns: 1fr;
		}
		.container {
			padding: var(--size-2);
			/* gap: var(--size-4); */
		}
		.section {
			border: unset;
			background: unset;
			padding: var(--size-1);
		}
		.section p {
			padding-block-end: var(--size-0);
		}
	}
	.preview {
		overflow: hidden;
		border: 1px solid var(--color-black);
	}
</style>
