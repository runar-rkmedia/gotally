<script lang="ts">
	import { Button } from 'flowbite-svelte'
	import { store } from '../connect-web/store'
	import Icon from './Icon.svelte'
	import { fade, fly, blur, scale } from 'svelte/transition'
	export let showGameMenu: boolean
</script>

<div class="headControls">
	<div class="info">
		<div class="">
			<div class="score" data-score={$store.session.game.score} data-testid="score">
				<span class="label">Score: </span>
				<span class="value">
					{#key $store.session.game.score}
						<span in:scale out:scale>
							{$store.session.game.score}
						</span>
					{/key}
				</span>
			</div>
			<div class="moves" data-moves={$store.session.game.moves}>
				<span class="label"> Moves: </span>
				<span class="value">
					{#key $store.session.game.moves}
						<span in:scale out:scale>
							{$store.session.game.moves}
						</span>
					{/key}
				</span>
			</div>
		</div>
	</div>
	<Button
		data-testid="menu"
		on:click={() => (showGameMenu = true)}
		aria-roledescription="Show menu"
		pill={true}
		class="!p-2"
		color="blue"
		outline={true}
		size="xl"
	>
		<Icon icon="settings" color="white" class="" />
	</Button>
</div>

<style lang="scss">
	.label {
		margin-inline-end: var(--size-3);
		min-width: 5ch;
		display: inline-block;
	}
	.value {
		position: relative;
		display: inline-block;
		vertical-align: top;
	}
	.value span {
		position: absolute;
		left: 0;
		top: 0;
	}
	.headControls {
		display: flex;
		justify-content: space-between;
		font-size: 1.8rem;
	}
</style>
