<script lang="ts">
	import { onMount } from 'svelte'

	import { GameMode } from '../connect-web'

	import { store, storeHandler } from '../connect-web/store'
	import FancyText from './FancyText.svelte'

	import Giphy from './Giphy.svelte'
	import Pyro from './Pyro.svelte'
	export let open = true
	let seed = 0
	let userName = ''
	function newSeed() {
		seed = Math.floor(Math.random() * 1e12)
	}
	function getAppState() {
		try {
			return JSON.parse(localStorage.getItem('appState') || '{}')
		} catch (err) {
			return {
				userName: '',
			}
		}
	}

	onMount(async () => {
		newSeed()
		const appState = getAppState()
		if (appState) {
			userName = String(appState.userName || '')
		}
	})
	const tags = [
		{ headline: 'Fantastic' },
		{ headline: 'You won' },
		{ headline: 'Awesome' },
		{ headline: 'Great' },
		{ headline: 'Superb' },
		{ headline: 'Wonderful', gif: 'hug win' },
	]
	$: tag = tags[seed % tags.length]
	function submitToBoard() {
		console.log('submitt', userName)
		if (userName.length < 3) {
			console.log('bailearly', userName)
			return
		}
		if (userName.length > 16) {
			console.log('bailearly logn', userName)
			return
		}
		console.log('setting', userName)
		const appState = getAppState()
		appState.userName = userName
		localStorage.setItem('appState', JSON.stringify(appState))
	}
	const votes = [
		[1, '🤢', 'Terrible'],
		[2, '👎', 'No fun'],
		[3, '🫤', 'OK'],
		[4, '😊', 'Good'],
		[5, '🤗', 'Great'],
	] as const
</script>

<div class="wrapper">
	<FancyText>
		<span class="fancy">{tag.headline}</span>
	</FancyText>
	<FancyText>
		<span class="fancy-score">{$store?.session?.game?.score || 0} points</span>
	</FancyText>

	<p>You are a fantastic person and deserve a big hug!</p>

	{#if open}
		<!-- Pyro component is a bit expensive to mount, so we don't preload it -->
		<Pyro />
	{/if}
	<hr />
	{#if $store.session.game.board?.id}
		<p>Would you like to have your nickname on the scoreboard?</p>

		<form on:submit|preventDefault={submitToBoard}>
			<label>
				Nickname:
				<input
					minlength="3"
					maxlength="16"
					name="userName"
					bind:value={userName}
					placeholder="User name"
				/>
			</label>
			<button type="submit">Send</button>
		</form>

		{#if userName.length > 3}
			<hr />
			<p>How did you like this board?</p>
			<!-- content here -->
			<div class="voteButtons">
				{#each votes as [vote, emoji, desc]}
					<button
						class:active={$store.usersVotes[$store.session.game.board?.id]?.funVote === vote}
						on:click={() =>
							storeHandler.commit(
								storeHandler.vote({
									funVote: vote,
									userName,
								})
							)}
					>
						{emoji}
						<div>{desc}</div>
					</button>
				{/each}
			</div>
		{/if}
		<hr />
	{/if}

	<button on:click={() => storeHandler.commit(storeHandler.newGame({ mode: GameMode.RANDOM }))}
		>New Random game</button
	>
	<button on:click={() => storeHandler.commit(storeHandler.newGame({ mode: GameMode.TUTORIAL }))}
		>New Tutorial</button
	>
	<button
		on:click={() => storeHandler.commit(storeHandler.newGame({ mode: GameMode.RANDOM_CHALLENGE }))}
		>New Challenge</button
	>
	<Giphy tag={tag.gif || tag.headline} />
</div>

<style>
	.fancy {
		font-size: 2rem;
	}
	.fancy-score {
		font-size: 3rem;
	}
	.wrapper {
		text-align: center;
	}
	.voteButtons {
		display: flex;
		justify-content: center;
		/* flex-direction: column-reverse; */
	}

	.voteButtons button {
		all: unset;
		cursor: pointer;
		min-width: 52px;
		min-height: 48px;
		padding-inline-end: 5px;
		text-align: center;
		transition-property: transform, filter;
		transition-duration: 150ms;
		transition-timing-function: var(--easing-standard);
		opacity: 0.8;
		position: relative;
	}
	.voteButtons div {
		position: relative;
	}
	.voteButtons div::before {
		z-index: -1;
		position: absolute;
		content: '';
		top: 0;
		right: 0;
		bottom: 0;
		left: 0;
		background-color: var(--color-grey-800);
		border-radius: var(--radius-md);
		opacity: 1;
	}

	.voteButtons button.active {
		pointer-events: none;
		opacity: 1;
		transform: translateY(-10px) scale(2);
		z-index: 1;
	}
</style>
