<script lang="ts">
	export const ssr = false
	import 'pollen-css'
	import InstructionList from '../components/InstructionList.svelte'
	import {
		GameMode,
		GetHintRequest,
		NewGameRequest,
		SwipeDirection,
		type Cell
	} from '../connect-web'
	import { onMount } from 'svelte'
	import { browser } from '$app/env'
	import { animateSwipe } from '../logic'
	import type { PartialMessage } from '@bufbuild/protobuf/dist/types/message'
	import { ErrNoChange, store, storeHandler } from '../connect-web/store'
	import SwipeHint from '../components/board/SwipeHint.svelte'

	let boardDiv: HTMLDivElement
	let showCellIndex = false

	const restartGame = () => {
		return storeHandler.commit(storeHandler.restartGame())
	}
	const getHint = async (options?: PartialMessage<GetHintRequest>) => {
		return storeHandler.commit(storeHandler.getHint(options))
	}
	const newGame = async (options: PartialMessage<NewGameRequest>) => {
		return storeHandler.commit(storeHandler.newGame(options))
	}

	onMount(async () => {
		await storeHandler.commit(storeHandler.getSession())

		// Set up swipes
		if (browser) {
			document.onkeydown = async (e) => {
				if (swipeLock) {
					e.preventDefault()
					return
				}
				if (selection.length) {
					e.preventDefault()
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
					case 'c':
						restartGame()
						break
					case 'n':
						newGame({ mode: GameMode.RANDOM_CHALLENGE })
						break

					default:
						console.log('key', e.key)
						return
				}
				e.preventDefault()
			}
		}
	})
	const cellValue = (c: Cell | { base: number; twopow: number }) => {
		if (Number(c.base) === 0) {
			return ''
		}
		return Number(c.base) * Math.pow(2, Number(c.twopow))
	}
	const animateInvalidSwipe = (direction: SwipeDirection) => {
		console.error('Not implemented', 'animateInvalidSwipe', direction)
	}
	let swipeLock = false
	let _swipeQueueHandling = false
	const swipeQueue: SwipeDirection[] = []
	const swipe = async (direction: SwipeDirection) => {
		swipeQueue.push(direction)
		if (_swipeQueueHandling) {
			return
		}
		_swipeQueueHandling = true
		while (swipeQueue.length) {
			const dir = swipeQueue.pop()
			if (!dir) {
				_swipeQueueHandling = false
				return
			}
			await _swipe(dir)
		}
		_swipeQueueHandling = false
	}

	const _swipe = async (direction: SwipeDirection) => {
		if (!$store?.session?.game?.board) {
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
			nColumns: $store.session.game.board.columns,
			nRows: $store.session.game.board.rows
		}
		const shouldAnimate = await animateSwipe({ ...swipeOptions, dry: true })
		console.log({ shouldAnimate })
		if (!shouldAnimate) {
			return
		}
		// swipeLock = true
		const r = storeHandler.swipe(direction)
		await animateSwipe(swipeOptions)
		// In case the server responds slower than the animation,
		// we set a visual indicator here.
		// On the other hand, if the server is faster, the should should see nothing of this
		boardDiv.style.opacity = '0.8'
		const [_, commit, err] = await r
		swipeLock = false
		for (const cell of [...boardDiv.children] as HTMLElement[]) {
			cell.style.transform = ''
			cell.style.transition = 'none'
		}
		boardDiv.style.opacity = '1'
		if (err) {
			if (err === ErrNoChange) {
				animateInvalidSwipe(direction)
			}
			return
		}
		commit()
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
		if (!$store.session?.game?.board?.cells[i]?.base) {
			invalidSelectionMap = {}
			selection = []
			selectionMap = {}
			return
		}
		const isSelected = !!selectionMap[i]
		if (isSelected) {
			console.log('should send', selection)
			// const [result, err] = await go(
			// 	client.combineCells({
			// 		selection: {
			// 			case: 'indexes',
			// 			value: { index: selection }
			// 		}
			// 	})
			// )
			const [result, commit, err] = await storeHandler.combineCells(selection)
			if (err) {
				invalidSelectionMap = { [i]: true }
				selection = []
				selectionMap = {}
				return
			}
			commit()
			selection = []
			selectionMap = {}
			if (result.didWin) {
				setTimeout(() => {
					alert('You won!')
				}, 150)
			}
			return
		}
		selection = [...selection, i]
		selectionMap[i] = true
	}
	$: {
		console.log('store', $store)
	}
	$: nextHint = $store.hintDoneIndex >= 0 ? $store.hints[$store.hintDoneIndex + 1] : $store.hints[0]
</script>

{#if $store?.session?.game?.board}
	<div class="headControls">
		<div>
			<div class="score">
				Score: {$store.session.game.score}
			</div>
			<div class="moves">
				Moves: {$store.session.game.moves}
			</div>
		</div>
	</div>

	<div class="boardContainer">
		<SwipeHint
			instruction={nextHint?.instructionOneof.value}
			active={nextHint?.instructionOneof.case === 'swipe'}
		/>
		<div
			bind:this={boardDiv}
			use:createSwiper
			class="board"
			style={`grid-template-columns: repeat(${$store.session.game.board.columns}, 1fr); grid-template-rows: repeat(${$store.session.game.board.rows}, 1fr)`}
		>
			{#each $store.session.game.board.cells as c, i}
				<div
					class="cell"
					class:no-eval={invalidSelectionMap[i]}
					class:selected={selectionMap[i]}
					class:hinted={nextHint?.instructionOneof.case === 'combine' &&
						nextHint.instructionOneof.value.index.includes(i)}
					class:selectedLast={!!selection.length && selection[selection.length - 1] === i}
					class:blank={Number(c.base) === 0}
					on:click={() => select(i)}
				>
					<div class="cellValue">
						{cellValue(c)}
					</div>
					{#if showCellIndex}
						<div class="cellIndex">{i}</div>
					{/if}
				</div>
			{/each}
		</div>
	</div>
	<div class="bottom-controls">
		<button on:click={() => getHint()}>Hint </button>

		<div>
			<button on:click={() => restartGame()}>Restart </button>
			<button on:click={() => newGame({ mode: GameMode.RANDOM })}>New Random game</button>
			<button on:click={() => newGame({ mode: GameMode.TUTORIAL })}>New Tutorial</button>
			<button on:click={() => newGame({ mode: GameMode.RANDOM_CHALLENGE })}>New Challenge</button>
		</div>
	</div>
{/if}

<style>
	button:disabled {
		opacity: 0.4;
	}
	.bottom-controls {
		display: flex;
		justify-content: center;
		flex-direction: column;
	}
	button {
		background-color: var(--color-blue);
		transition: opacity 70ms var(--easing-standard);
		min-width: 52px;
		min-height: 52px;
		color: var(--color-black);
	}

	.boardContainer {
		position: relative;
		border: 2px solid var(--color-blue-700);
		border-radius: var(--radius-lg);
		margin-block-end: var(--size-4);
	}
	.board {
		position: relative;
		transition: opacity 300ms var(--easing-standard);
		/* margin-inline: -4px; */
		display: grid;

		height: 100%;
		min-height: 60vw;
		max-height: 100vw;
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
		position: relative;
		box-shadow: var(--elevation-4);

		opacity: 0.8;
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

	.cellValue {
		font-weight: bold;
		font-size: 2rem;
		transition-property: color, transform;
		transition-duration: 300ms;
		transition-timing-function: var(--easing-standard);
		/* transition: transform 300ms var(--easing-standard); */
	}

	.cellIndex {
		position: absolute;
		right: var(--size-1);
		bottom: var(--size-1);
		font-size: 0.8rem;
		opacity: 0.7;
	}

	.cell.selected {
		background-color: var(--color-green);
		color: var(--color-black);
		transform: scale(0.9);
	}
	.cell.selected .cellValue {
		transform: scale(1.2);
	}

	.cell.selectedLast {
		background-color: var(--color-green-300);
		color: var(--color-black);
	}
	.cell.no-eval {
		animation: shake 0.82s cubic-bezier(0.36, 0.07, 0.19, 0.97) both, grow-to-normal 0.82s linear;
		transform: translate3d(0, 0, 0);
		backface-visibility: hidden;
		perspective: 1000px;
	}
	.cell.no-eval {
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
