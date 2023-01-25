<script lang="ts">
	import type { PartialMessage } from '@bufbuild/protobuf'

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

<div>
	<button
		on:click={() => {
			restartGame()
			open = false
		}}
		>Restart
	</button>
	<button
		on:click={() => {
			newGame({ mode: GameMode.RANDOM })
			open = false
		}}>New Random game</button
	>
	<button
		on:click={() => {
			newGame({ mode: GameMode.TUTORIAL })
			open = false
		}}>New Tutorial</button
	>
	<button
		on:click={() => {
			newGame({ mode: GameMode.RANDOM_CHALLENGE })
			open = false
		}}>New Challenge</button
	>
</div>

<style>
	div {
		padding: var(--size-8);
		display: flex;
		flex-direction: column;
		gap: var(--size-8);
	}
</style>
