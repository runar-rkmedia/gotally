<script lang="ts">
	import { store } from '../connect-web/store'
	import Icon from './Icon.svelte'
	export let game: {
		score: number
		moves: number
		board: { id: string; name: string }
	} = $store.session.game
	export let username = $store.session.username
	export let sessionId = $store.session.sessionId
	export let showGameMenu: boolean
</script>

<div class="headControls">
	<div class="info">
		<div>
			<div class="score" data-score={game.score} data-testid="score">
				Score: {game.score}
			</div>
			<small class="boardName" title={game.board.id}>{game.board.name}</small>
			<div class="moves" data-moves={game.moves}>
				Moves: {game.moves}
			</div>
		</div>
	</div>
	<p>Hi, {username} ({sessionId})</p>
	<button
		class="icon-only"
		data-testid="menu"
		on:click={() => (showGameMenu = true)}
		aria-roledescription="Show menu"
	>
		<Icon icon="settings" color="white" />
	</button>
</div>

<style lang="scss">
	.boardName {
		opacity: 0.7;
		float: right;
	}
	button {
		&.icon-only {
			min-height: 48px;
			background: unset;
			border: unset;
			display: flex;
			justify-content: center;
			align-items: center;
		}
	}
	.headControls {
		display: flex;
		justify-content: space-between;
	}
</style>
