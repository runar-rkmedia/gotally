<script lang="ts">
	import type { PartialMessage } from '@bufbuild/protobuf'

	import { GameMode, type NewGameRequest } from '../connect-web/proto/tally/v1/board_pb'
	export let open: boolean

	import { storeHandler, store } from '../connect-web/store'

	const restartGame = () => {
		return storeHandler.commit(storeHandler.restartGame())
	}
	const newGame = async (options: PartialMessage<NewGameRequest>) => {
		return storeHandler.commit(storeHandler.newGame(options))
	}
</script>

<div class="container">
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
						open = false
					}}
					>Restart
				</button>
			</div>
		{/if}
		<div class="section">
			<p>Need som help? How about going through the tutorial?</p>
			<button
				style="--color: var(--color-yellow)"
				on:click={() => {
					newGame({ mode: GameMode.TUTORIAL })
					open = false
				}}>New Tutorial</button
			>
		</div>
		<div class="section">
			<p>Ready for a challenge? These are also great for learning new strategies!</p>
			<button
				style="--color: var(--color-green)"
				on:click={() => {
					newGame({ mode: GameMode.RANDOM_CHALLENGE })
					open = false
				}}>New Challenge</button
			>
		</div>
		<div class="section">
			<p>How far can you go?</p>
			<button
				style="--color: var(--color-blue)"
				on:click={() => {
					newGame({ mode: GameMode.RANDOM })
					open = false
				}}>New Infinite game</button
			>
		</div>
		<div class="section">
			<p>Back to the game</p>
			<button
				style="--color: var(--color-blue)"
				on:click={() => {
					open = false
				}}>Continue</button
			>
		</div>
	</div>
</div>

<style>
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
		grid-template-columns: 1fr 1fr;
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
	@media (max-width: 640px) {
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

	.quote {
		font-size: var(--scale-fluid-3);
		margin-inline: auto;
	}
	button {
		font-size: 15px;
		width: 100%;
		padding: 0.7em 2.7em;
		letter-spacing: 0.06em;
		position: relative;
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
	}
</style>
