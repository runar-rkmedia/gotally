<script lang="ts">
	import type { PartialMessage } from '@bufbuild/protobuf'
	import { onMount } from 'svelte'

	import { Difficulty, GameMode, type NewGameRequest } from '../connect-web/proto/tally/v1/board_pb'
	export let open: boolean
	export let didWin: boolean

	import { newGame, storeHandler, store, type Challenge } from '../connect-web/store'
	import BoardPreview from './BoardPreview.svelte'
	import ChallengeCard from './ChallengeCard.svelte'
	import Icon from './Icon.svelte'
	import { Button } from 'flowbite-svelte'
	import GameWon from './GameWon.svelte'

	const restartGame = () => {
		if (!$store?.session.game.moves) {
			return
		}
		return storeHandler.commit(storeHandler.restartGame())
	}
	let view: 'main' | 'challenges' | 'tutorials' | 'infinite-game' = 'main'
	const close = () => {
		view = 'main'
		open = false
	}
	$: {
		if (open === false) {
			view = 'main'
		}
	}
	$: {
		console.log('why did you change', view)
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

<div class="game-menu">
	{#if didWin}
		<GameWon />
	{/if}
	{#if view === 'main'}
		<div class="buttons">
			<div class="section">
				<p>Restart the current game?</p>

				<Button
					disabled={$store.session.game.moves === 0}
					data-test-id="restart-game"
					color="purple"
					on:click={() => {
						restartGame()
						close()
					}}
				>
					<Icon icon="restart" />
					Restart
				</Button>
			</div>
			<div class="section">
				<p>Need som help? How about going through the tutorial?</p>
				<Button
					on:click={() => {
						newGame({ mode: GameMode.TUTORIAL, variant: { case: 'levelIndex', value: 0 } })
						close()
					}}
				>
					<Icon icon="tutorial" />
					New Tutorial</Button
				>
			</div>
			<div class="section">
				<p>Ready for a challenge? These are also great for learning new strategies!</p>
				<Button
					color="green"
					on:click={() => {
						console.log('why????')
						view = 'challenges'
					}}
				>
					<Icon icon="challenge" />
					New Challenge</Button
				>
			</div>
			<div class="section">
				<p>How far can you go?</p>
				<Button
					color="blue"
					on:click={() => {
						view = 'infinite-game'
					}}
				>
					<Icon icon="infinite" />
					New Infinite game</Button
				>
			</div>
		</div>
	{:else if view === 'infinite-game'}
		<p>How far can you go?</p>
		<div class="buttons">
			<div class="section">
				<p>For newcomers to the game</p>
				<Button
					color="blue"
					on:click={() => {
						newGame({
							mode: GameMode.RANDOM,
							variant: {
								value: Difficulty.EASY,
								case: 'difficulty',
							},
						})
						close()
					}}>Easy</Button
				>
			</div>
			<div class="section">
				<p>As the game is intended to be played</p>
				<Button
					color="green"
					on:click={() => {
						newGame({
							mode: GameMode.RANDOM,
							variant: {
								value: Difficulty.MEDIUM,
								case: 'difficulty',
							},
						})
						close()
					}}>Medium</Button
				>
			</div>
			<div class="section">
				<p>Tough and unfair</p>
				<Button
					color="red"
					on:click={() => {
						newGame({
							mode: GameMode.RANDOM,
							variant: {
								value: Difficulty.HARD,
								case: 'difficulty',
							},
						})
						close()
					}}>Hard</Button
				>
			</div>
		</div>
	{:else if view === 'challenges'}
		<Button on:click={() => (view = 'main')}>Back</Button>
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
		<Button
			color="alternative"
			on:click={() => {
				close()
			}}
		>
			<Icon icon="play" />
			Continue</Button
		>
	</div>
</div>

<style lang="scss">
	.game-menu {
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
		gap: var(--size-4);
		grid-template-columns: 1fr 1fr;
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
		.game-menu {
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
</style>
