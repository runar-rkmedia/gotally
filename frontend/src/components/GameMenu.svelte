<script lang="ts">
	import type { PartialMessage } from '@bufbuild/protobuf'
	import { store } from '../connect-web/store'

	import { GameMode, type NewGameRequest } from '../connect-web/proto/tally/v1/board_pb'
	export let open: boolean

	import { storeHandler } from '../connect-web/store'

	const restartGame = () => {
		return storeHandler.commit(storeHandler.restartGame())
	}
	const newGame = async (options: PartialMessage<NewGameRequest>) => {
		return storeHandler.commit(storeHandler.newGame(options))
	}
</script>

<div class="container">
	<div class="score" data-score={$store.session.game.score}>
		Score: {$store.session.game.score}
	</div>
	<p class="quote">You are doing great, keep those brainfluids running!</p>
	<div class="section">
		<p>Restart the current game?</p>
		<button
			data-test-id="restart-game"
			style="--color: var(--color-orange)"
			on:click={() => {
				restartGame()
				open = false
			}}
			>Restart
		</button>
	</div>
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
</div>

<style>
	.container {
		padding: var(--size-8);
		display: flex;
		flex-direction: column;
		text-align: center;
		gap: var(--size-8);
	}
	.section {
		border: 2px solid white;
		border-radius: 20px;
		padding: 8px;
		background-color: #00380080;
	}
	.section p {
		padding-block-end: 16px;
	}
	.score {
		font-size: 2rem;
		margin-inline: auto;
	}
	.quote {
		font-size: 1.4rem;
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
