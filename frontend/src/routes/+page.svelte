<script lang="ts">
	import 'pollen-css'
	import { createPromiseClient, createConnectTransport, ConnectError } from '@bufbuild/connect-web'
	import {
		Board,
		BoardService,
		SwipeDirection,
		type Cell,
		type GetBoardResponse
	} from '../connect-web'
	import { onMount } from 'svelte'
	import { browser } from '$app/env'
	import { animateSwipe } from '../logic'

	const transport = createConnectTransport({
		baseUrl: 'http://localhost:8080/'
		// useBinaryFormat: false
	})

	const getHint = () => {}

	const handleError = (err: ConnectError | Error) => {
		// TODO: handle error
		if (err instanceof ConnectError) {
			const { metadata, ...all } = err
			const meta: any = {}
			for (const [k, v] of metadata) {
				meta[k] = v
			}
			console.warn(`[${err.code}]: ${err.name} - ${err.message}`, all, meta)
		}
		console.warn('An error occured', err)
	}

	const go = async <P = any>(
		promise: Promise<P>
	): Promise<[P, null] | [null, ConnectError | Error]> => {
		try {
			const result = await promise
			return [result, null]
		} catch (err) {
			return [null, err as any]
		}
	}

	const client = createPromiseClient(BoardService, transport)

	let response: GetBoardResponse
	let sessionID = ''
	let boardDiv: HTMLDivElement

	onMount(async () => {
		sessionID = localStorage.getItem('sessionID') || ''
		if (!sessionID) {
			sessionID = String(Math.random())
			localStorage.setItem('sessionID', sessionID)
		}
		const [res, err] = await go(client.getBoard({}))
		if (err) {
			handleError(err)
			return
		}
		console.log(res)
		response = res

		// Set up swipes
		if (browser) {
			document.onkeydown = async (e) => {
				if (swipeLock) {
					return
				}
				if (selection.length) {
					return
				}
				switch (e.key) {
					case 'ArrowLeft':
					case 'a':
						swipe(SwipeDirection.LEFT)
						break
					case 'ArrowRight':
					case 'd':
						swipe(SwipeDirection.RIGHT)
						break
					case 'ArrowDown':
					case 's':
						swipe(SwipeDirection.DOWN)
						break
					case 'ArrowUp':
					case 'w':
						swipe(SwipeDirection.UP)
						break
					case 'h':
						getHint()
						break
					// case "c":
					//   restartButton.click()
					//   break
					case 'n':
						const [result, err] = await go(client.newGame({}))
						if (err) {
							handleError(err)
							return
						}
						response = result

						break

					default:
						console.log('key', e.key)
						break
				}
			}
		}
	})
	const cellValue = (c: Cell) => {
		if (Number(c.base) === 0) {
			return ''
		}
		return Number(c.base) * Math.pow(2, Number(c.twopow))
	}
	const animateInvalidSwipe = (direction: SwipeDirection) => {
		console.error('Not implemented', 'animateInvalidSwipe', direction)
	}
	let swipeLock = false
	const swipe = async (direction: SwipeDirection) => {
		if (!response?.board) {
			return
		}
		if (selection.length) {
			return
		}
		if (swipeLock) {
			return
		}
		const swipeOptions = {
			swipeAnimationTime: 480,
			positive: direction === SwipeDirection.DOWN || direction === SwipeDirection.RIGHT,
			vertical: direction === SwipeDirection.UP || direction === SwipeDirection.DOWN,
			boardEl: boardDiv,
			nColumns: response.board.columns,
			nRows: response.board.rows
		}
		const shouldAnimate = await animateSwipe({ ...swipeOptions, dry: true })
		console.log({ shouldAnimate })
		if (!shouldAnimate) {
			return
		}
		swipeLock = true
		const r = go(client.swipeBoard({ direction }))
		await animateSwipe(swipeOptions)
		// In case the server responds slower than the animation,
		// we set a visual indicator here.
		// On the other hand, if the server is faster, the should should see nothing of this
		boardDiv.style.opacity = '0.8'
		const [result, err] = await r
		swipeLock = false
		for (const cell of [...boardDiv.children]) {
			;(cell as HTMLElement).style.transform = ''
			;(cell as HTMLElement).style.transition = 'none'
		}
		boardDiv.style.opacity = '1'
		if (err) {
			handleError(err)
			return
		}
		if (!result.didChange) {
			animateInvalidSwipe(direction)
			return
		}
		// animateSwipe(direction)
		response.board = result.board
	}
	function createSwiper(node: HTMLElement) {
		if (!browser) {
			return
		}
		let Hammer: any
		import('hammerjs').then((h) => {
			Hammer = h.default
			const hammerTime = new Hammer(node, {
				recognizers: [[Hammer.Swipe, { direction: Hammer.DIRECTION_ALL }]]
			})
			hammerTime.on('swipe', (e: any) => {
				switch (e.direction) {
					case Hammer.DIRECTION_UP:
						swipe(SwipeDirection.UP)
						break
					case Hammer.DIRECTION_DOWN:
						swipe(SwipeDirection.DOWN)
						break
					case Hammer.DIRECTION_LEFT:
						swipe(SwipeDirection.LEFT)
						break
					case Hammer.DIRECTION_RIGHT:
						swipe(SwipeDirection.RIGHT)
						break
				}
			})
		})
	}
	let selection: number[] = []
	let selectionMap: Record<number, boolean | undefined> = {}
	let invalidSelectionMap: Record<number, boolean | undefined> = {}
	const select = async (i: number) => {
		invalidSelectionMap = {}
		if (!response.board?.cell[i]?.base) {
			invalidSelectionMap = {}
			selection = []
			selectionMap = {}
			return
		}
		const isSelected = !!selectionMap[i]
		if (isSelected) {
			console.log('should send', selection)
			const [result, err] = await go(
				client.combineCells({
					selection: {
						case: 'indexes',
						value: { index: selection }
					}
				})
			)
			if (err) {
				invalidSelectionMap = { [i]: true }
				selection = []
				selectionMap = {}
				handleError(err)
				return
			}
			selection = []
			selectionMap = {}
			console.log(result)
			response.board = result.board
			response.score = result.score
			response.moves = result.moves
			if (result.didWin) {
				setTimeout(() => {
					alert('You won!')
				}, 150)
			}
			return
		}
		selection.push(i)
		selectionMap[i] = true
	}
	$: {
		console.log({ selectionMap, selection })
	}
</script>

{#if response?.board}
	<div class="headControls">
		<div>
			<div class="score">
				Score: {response.score}
			</div>
			<div class="moves">
				Moves: {response.moves}
			</div>
		</div>
		<div>
			<small style="color: var(--color-grey-500);">
				<div>
					<div class="voteButtons">
						<button class="active" live-click="vote" live-value-vote="1">⭐</button>
						<button class="active" live-click="vote" live-value-vote="2">⭐</button>
						<button class="active" live-click="vote" live-value-vote="3">⭐</button>
						<button class="active" live-click="vote" live-value-vote="4">⭐</button>
						<button class="active" live-click="vote" live-value-vote="5">⭐</button>
					</div>
					<span style="float: right">
						{response.name}
					</span>
				</div>
			</small>
		</div>
	</div>
	<div
		bind:this={boardDiv}
		use:createSwiper
		class="board"
		style={`grid-template-columns: repeat(${response.board.columns}, 1fr); grid-template-rows: repeat(${response.board.rows}, 1fr)`}
	>
		{#each response.board.cell as c, i}
			<div
				class="cell"
				class:no-eval={invalidSelectionMap[i]}
				class:selected={selectionMap[i]}
				class:selectedLast={!!selection.length && selection[selection.length - 1]}
				class:blank={Number(c.base) === 0}
				on:click={() => select(i)}
			>
				<div>
					{cellValue(c)}
				</div>
			</div>
		{/each}
	</div>
	<div class="swipe-buttons">
		<button disabled={swipeLock} on:click={() => swipe(SwipeDirection.UP)}>Swipe up</button>
		<div>
			<button disabled={swipeLock} on:click={() => swipe(SwipeDirection.LEFT)}>Swipe Left</button>
			<button disabled={swipeLock} on:click={() => swipe(SwipeDirection.RIGHT)}>Swipe Right</button>
		</div>
		<button disabled={swipeLock} on:click={() => swipe(SwipeDirection.DOWN)}>Swipe Down</button>
	</div>
{/if}

<style>
	button:disabled {
		opacity: 0.4;
	}
	button {
		background-color: var(--color-blue);
		transition: opacity 70ms var(--easing-standard);
		min-width: 52px;
		min-height: 52px;
	}
	.swipe-buttons button {
		margin-block: var(--size-2);
	}
	.swipe-buttons > * {
		display: block;
		margin-inline: auto;
		width: max-content;
	}
	.board {
		transition: opacity 300ms var(--easing-standard);
		/* margin-inline: -4px; */
		display: grid;

		width: calc(100% + 8px);

		height: 100%;
		min-height: 60vw;
		max-height: 100vw;
		border: 2px solid var(--color-blue-700);
		border-radius: var(--radius-lg);
	}
	.cell {
		transition: transform 300ms var(--easing-standard);
		user-select: none;
		display: flex;
		justify-content: center;
		align-items: center;
		border: 2px solid var(--border-blue-700);
		margin: 2px;
		border-radius: 8px;
		background-color: var(--color-blue-700);
		font-size: 2rem;
		position: relative;
		box-shadow: var(--elevation-4);
	}
	.cell:empty,
	.cell.blank {
		opacity: 0;
	}
	.cell.hinted:not(.selected) {
		background-color: var(--color-blue-500);
		outline-color: var(--color-purple-700);
		outline-width: 5px;
		outline-style: dotted;
	}

	.cell.selected {
		background-color: var(--color-green);
		color: var(--color-black);
		transform: scale(0.9);
	}
	.cell.selected div {
		transform: scale(0.9);
		transition: transform 300ms var(--easing-standard);
	}

	.cell.selectedLast {
		background-color: var(--color-green-300);
		color: var(--color-black);
	}
	.cell:not(:empty)::after {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background-image: url(https://images.freecreatives.com/wp-content/uploads/2016/02/Sky-Blue-Textured-Background-For-Free.jpg);
		background-size: cover;
		opacity: 0.4;
		z-index: 1;
	}
	.cell.no-eval {
		animation: shake 0.82s cubic-bezier(0.36, 0.07, 0.19, 0.97) both, grow-to-normal 0.82s linear;
		transform: translate3d(0, 0, 0);
		backface-visibility: hidden;
		perspective: 1000px;
	}
	.cell.no-eval::after {
		animation: sepia 0.82s cubic-bezier(0.36, 0.07, 0.19, 0.97) both;
	}
	@keyframes grow-to-normal {
		0% {
			scale: 0.9;
		}
		80% {
			scale: 0.9;
		}
		1000% {
			scale: 0.9;
		}
	}
	@keyframes sepia {
		0% {
			filter: sepia(1);
		}
		80% {
			filter: sepia(1);
		}
		1000% {
		}
	}
	@keyframes shake {
		10%,
		90% {
			transform: translate3d(-1px, 0, 100px);
		}

		20%,
		80% {
			transform: translate3d(2px, 0, 0);
		}

		30%,
		50%,
		70% {
			transform: translate3d(-4px, 0, 0);
		}

		40%,
		60% {
			transform: translate3d(4px, 0, 0);
		}
	}
	/* Keyframes */
	@keyframes wiggle {
		0%,
		7% {
			transform: rotateZ(0);
		}
		15% {
			transform: rotateZ(-5deg);
		}
		20% {
			transform: rotateZ(3.3deg);
		}
		25% {
			transform: rotateZ(-3.3deg);
		}
		30% {
			transform: rotateZ(2deg);
		}
		35% {
			transform: rotateZ(1.3deg);
		}
		40%,
		100% {
			transform: rotateZ(0);
		}
	}
</style>
